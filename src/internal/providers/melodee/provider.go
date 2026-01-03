package melodee

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/tunez/tunez/internal/provider"
)

type Config struct {
	BaseURL    string
	Username   string
	Password   string
	PageSize   int
	CacheDB    string
	HTTPClient *http.Client
}

type Provider struct {
	cfg    Config
	client *http.Client
	token  string
	caps   provider.Capabilities
}

func New() *Provider {
	return &Provider{
		caps: provider.Capabilities{
			provider.CapPlaylists: true,
			provider.CapLyrics:    true,
			provider.CapArtwork:   true,
		},
	}
}

func (p *Provider) ID() string   { return "melodee" }
func (p *Provider) Name() string { return "Melodee" }

func (p *Provider) Capabilities() provider.Capabilities { return p.caps }

func (p *Provider) Initialize(ctx context.Context, profileCfg any) error {
	raw, ok := profileCfg.(map[string]any)
	if !ok {
		return provider.ErrInvalidConfig
	}
	cfg, err := parseConfig(raw)
	if err != nil {
		return err
	}
	p.cfg = cfg
	if p.cfg.HTTPClient != nil {
		p.client = p.cfg.HTTPClient
	} else {
		p.client = &http.Client{Timeout: 8 * time.Second}
	}
	if err := p.authenticate(ctx); err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}
	return nil
}

func parseConfig(raw map[string]any) (Config, error) {
	cfg := Config{PageSize: 100}
	if v, ok := raw["base_url"].(string); ok {
		cfg.BaseURL = v
	}
	if v, ok := raw["username"].(string); ok {
		cfg.Username = v
	}
	if v, ok := raw["password"].(string); ok {
		cfg.Password = v
	}
	if v, ok := raw["password_env"].(string); ok && cfg.Password == "" {
		cfg.Password = os.Getenv(v)
	}
	if v, ok := raw["page_size"].(int64); ok && v > 0 {
		cfg.PageSize = int(v)
	}
	if cfg.BaseURL == "" {
		return Config{}, provider.ErrInvalidConfig
	}
	return cfg, nil
}

func (p *Provider) Health(ctx context.Context) (bool, string) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.BaseURL+"/health", nil)
	resp, err := p.client.Do(req)
	if err != nil {
		return false, err.Error()
	}
	resp.Body.Close()
	return resp.StatusCode < 500, resp.Status
}

func (p *Provider) authenticate(ctx context.Context) error {
	body := map[string]string{"username": p.cfg.Username, "password": p.cfg.Password}
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.BaseURL+"/api/v1/auth/authenticate", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return provider.ErrUnauthorized
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("auth status %d", resp.StatusCode)
	}
	var r struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if r.AccessToken == "" {
		return errors.New("empty token")
	}
	p.token = r.AccessToken
	return nil
}

func (p *Provider) authHeader(req *http.Request) {
	if p.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.token)
	}
}

func (p *Provider) doRequest(req *http.Request) (*http.Response, error) {
	p.authHeader(req)
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		// Re-authenticate
		if err := p.authenticate(req.Context()); err != nil {
			return nil, err // Return auth error
		}
		// Retry once
		p.authHeader(req)
		return p.client.Do(req)
	}
	return resp, nil
}

func (p *Provider) ListArtists(ctx context.Context, req provider.ListReq) (provider.Page[provider.Artist], error) {
	page, err := getPaged[provider.Artist](ctx, p, "/api/v1/artists", req)
	if err != nil {
		return provider.Page[provider.Artist]{}, err
	}
	return page, nil
}

func (p *Provider) GetArtist(ctx context.Context, id string) (provider.Artist, error) {
	return getOne[provider.Artist](ctx, p, "/api/v1/artists/"+id)
}

func (p *Provider) ListAlbums(ctx context.Context, artistId string, req provider.ListReq) (provider.Page[provider.Album], error) {
	path := "/api/v1/albums"
	if artistId != "" {
		path = "/api/v1/artists/" + url.PathEscape(artistId) + "/albums"
	}
	return getPaged[provider.Album](ctx, p, path, req)
}

func (p *Provider) GetAlbum(ctx context.Context, id string) (provider.Album, error) {
	return getOne[provider.Album](ctx, p, "/api/v1/albums/"+id)
}

func (p *Provider) ListTracks(ctx context.Context, albumId string, artistId string, playlistId string, req provider.ListReq) (provider.Page[provider.Track], error) {
	switch {
	case playlistId != "":
		return getPaged[provider.Track](ctx, p, "/api/v1/playlists/"+url.PathEscape(playlistId)+"/songs", req)
	case albumId != "":
		return getPaged[provider.Track](ctx, p, "/api/v1/albums/"+url.PathEscape(albumId)+"/songs", req)
	case artistId != "":
		// fallback: search songs by artist
		res, err := p.Search(ctx, "artist:"+artistId, req)
		return res.Tracks, err
	default:
		return getPaged[provider.Track](ctx, p, "/api/v1/search/songs", req)
	}
}

func (p *Provider) GetTrack(ctx context.Context, id string) (provider.Track, error) {
	return getOne[provider.Track](ctx, p, "/api/v1/songs/"+url.PathEscape(id))
}

func (p *Provider) Search(ctx context.Context, q string, req provider.ListReq) (provider.SearchResults, error) {
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = p.cfg.PageSize
	}
	offset := parseCursor(req.Cursor)
	u, _ := url.Parse(p.cfg.BaseURL + "/api/v1/search/songs")
	qp := u.Query()
	qp.Set("q", q)
	qp.Set("page", strconv.Itoa(offset/pageSize+1))
	qp.Set("pageSize", strconv.Itoa(pageSize))
	u.RawQuery = qp.Encode()
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := p.doRequest(httpReq)
	if err != nil {
		return provider.SearchResults{}, mapHTTPError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return provider.SearchResults{}, provider.ErrUnauthorized
	}
	if resp.StatusCode == http.StatusNotFound {
		return provider.SearchResults{}, provider.ErrNotFound
	}
	if resp.StatusCode >= 500 {
		return provider.SearchResults{}, provider.ErrTemporary
	}
	var data pagedResponse[provider.Track]
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return provider.SearchResults{}, err
	}
	next := ""
	if data.HasMore {
		next = fmt.Sprintf("%d", offset+pageSize)
	}
	return provider.SearchResults{
		Tracks: provider.Page[provider.Track]{Items: data.Items, NextCursor: next, TotalHint: data.Total},
	}, nil
}

func (p *Provider) ListPlaylists(ctx context.Context, req provider.ListReq) (provider.Page[provider.Playlist], error) {
	return getPaged[provider.Playlist](ctx, p, "/api/v1/user/playlists", req)
}

func (p *Provider) GetPlaylist(ctx context.Context, id string) (provider.Playlist, error) {
	return getOne[provider.Playlist](ctx, p, "/api/v1/playlists/"+url.PathEscape(id))
}

func (p *Provider) GetStream(ctx context.Context, trackId string) (provider.StreamInfo, error) {
	track, err := p.GetTrack(ctx, trackId)
	if err != nil {
		return provider.StreamInfo{}, err
	}
	if track.StreamURL == "" {
		return provider.StreamInfo{}, provider.ErrNotFound
	}
	return provider.StreamInfo{URL: track.StreamURL, Headers: map[string]string{"Authorization": "Bearer " + p.token}}, nil
}

func (p *Provider) GetLyrics(ctx context.Context, trackId string) (provider.Lyrics, error) {
	track, err := p.GetTrack(ctx, trackId)
	if err != nil {
		return provider.Lyrics{}, err
	}
	return provider.Lyrics{Text: track.AlbumTitle}, nil
}

func (p *Provider) GetArtwork(ctx context.Context, ref string, sizePx int) (provider.Artwork, error) {
	return provider.Artwork{}, provider.ErrNotSupported
}

type pagedResponse[T any] struct {
	Items   []T  `json:"items"`
	HasMore bool `json:"hasMore"`
	Total   int  `json:"total"`
}

func getPaged[T any](ctx context.Context, p *Provider, path string, req provider.ListReq) (provider.Page[T], error) {
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = p.cfg.PageSize
	}
	offset := parseCursor(req.Cursor)
	u, _ := url.Parse(p.cfg.BaseURL + path)
	q := u.Query()
	q.Set("page", strconv.Itoa(offset/pageSize+1))
	q.Set("pageSize", strconv.Itoa(pageSize))
	u.RawQuery = q.Encode()
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := p.doRequest(httpReq)
	if err != nil {
		return provider.Page[T]{}, mapHTTPError(err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return provider.Page[T]{}, provider.ErrUnauthorized
	case http.StatusNotFound:
		return provider.Page[T]{}, provider.ErrNotFound
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return provider.Page[T]{}, provider.ErrRateLimited
	}
	if resp.StatusCode >= 500 {
		return provider.Page[T]{}, provider.ErrTemporary
	}
	var data pagedResponse[T]
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return provider.Page[T]{}, err
	}
	next := ""
	if data.HasMore {
		next = fmt.Sprintf("%d", offset+pageSize)
	}
	return provider.Page[T]{Items: data.Items, NextCursor: next, TotalHint: data.Total}, nil
}

func getOne[T any](ctx context.Context, p *Provider, path string) (T, error) {
	var zero T
	u := p.cfg.BaseURL + path
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := p.doRequest(req)
	if err != nil {
		return zero, mapHTTPError(err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return zero, provider.ErrUnauthorized
	case http.StatusNotFound:
		return zero, provider.ErrNotFound
	}
	if resp.StatusCode >= 500 {
		return zero, provider.ErrTemporary
	}
	if resp.StatusCode >= 400 {
		return zero, fmt.Errorf("http status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&zero); err != nil {
		return zero, err
	}
	return zero, nil
}

func parseCursor(cur string) int {
	if cur == "" {
		return 0
	}
	var off int
	fmt.Sscanf(cur, "%d", &off)
	return off
}

func mapHTTPError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return provider.ErrTemporary
	}
	return err
}
