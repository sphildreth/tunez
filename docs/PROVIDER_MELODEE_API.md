# Tunez — Melodee API Provider (Built-in)

**Last updated:** 2026-01-02  
**Provider ID:** `melodee`

## Overview
Connects to a Melodee server over HTTPS for browsing/searching and streaming tracks.

## Endpoints (Phase 1)
- `POST /api/v1/auth/authenticate`
- `POST /api/v1/auth/refresh-token`
- `GET /api/v1/artists?page=&pageSize=`
- `GET /api/v1/artists/{id}/albums`
- `GET /api/v1/albums?page=&pageSize=`
- `GET /api/v1/albums/{id}/songs`
- `GET /api/v1/user/playlists?page=&limit=`
- `GET /api/v1/playlists/{apiKey}/songs?page=&pageSize=`
- `GET /api/v1/search/songs?q=&page=&pageSize=`
- `GET /api/v1/songs/{id}`

## Auth
- Authenticate on Initialize
- Refresh token before expiry
- Map 401/403 → `ErrUnauthorized`

## Streaming
- Prefer `Song.streamUrl`
- If headers required, return them in `StreamInfo.Headers`

## MVP acceptance
- Browse/search/play streamed tracks without blocking UI
