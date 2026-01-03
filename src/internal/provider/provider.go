package provider

import "context"

type Capability string

const (
	CapPlaylists Capability = "playlists"
	CapLyrics    Capability = "lyrics"
	CapArtwork   Capability = "artwork"
)

type Capabilities map[Capability]bool

type ListReq struct {
	Cursor   string
	PageSize int
	Sort     string
}

type Page[T any] struct {
	Items      []T
	NextCursor string
	TotalHint  int
}

type StreamInfo struct {
	URL     string
	Headers map[string]string
}

type Provider interface {
	ID() string
	Name() string
	Capabilities() Capabilities

	Initialize(ctx context.Context, profileCfg any) error
	Health(ctx context.Context) (bool, string)

	ListArtists(ctx context.Context, req ListReq) (Page[Artist], error)
	GetArtist(ctx context.Context, id string) (Artist, error)

	ListAlbums(ctx context.Context, artistId string, req ListReq) (Page[Album], error)
	GetAlbum(ctx context.Context, id string) (Album, error)

	ListTracks(ctx context.Context, albumId string, artistId string, playlistId string, req ListReq) (Page[Track], error)
	GetTrack(ctx context.Context, id string) (Track, error)

	Search(ctx context.Context, q string, req ListReq) (SearchResults, error)

	ListPlaylists(ctx context.Context, req ListReq) (Page[Playlist], error)
	GetPlaylist(ctx context.Context, id string) (Playlist, error)

	GetStream(ctx context.Context, trackId string) (StreamInfo, error)

	GetLyrics(ctx context.Context, trackId string) (Lyrics, error)
	GetArtwork(ctx context.Context, ref string, sizePx int) (Artwork, error)
}

type SearchResults struct {
	Tracks    Page[Track]
	Albums    Page[Album]
	Artists   Page[Artist]
	Playlists Page[Playlist]
}

type Artist struct {
	ID         string
	Name       string
	SortName   string
	AlbumCount int
	TrackCount int
}

type Album struct {
	ID         string
	Title      string
	ArtistID   string
	ArtistName string
	Year       int
	TrackCount int
	ArtworkRef string
}

type Track struct {
	ID          string
	Title       string
	ArtistID    string
	ArtistName  string
	AlbumID     string
	AlbumTitle  string
	Year        int
	DurationMs  int
	TrackNo     int
	DiscNo      int
	Codec       string
	BitrateKbps int
	ArtworkRef  string
	StreamURL   string
}

type Playlist struct {
	ID         string
	Name       string
	TrackCount int
}

type Lyrics struct {
	Text string
}

type Artwork struct {
	Data     []byte
	MimeType string
}
