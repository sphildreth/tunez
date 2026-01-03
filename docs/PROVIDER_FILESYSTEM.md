# Tunez — Filesystem Provider (Built-in)

**Last updated:** 2026-01-02  
**Provider ID:** `filesystem`

## Overview
Loads a local music library rooted at directories, builds an index for browse/search, and returns `file://` URLs for mpv playback.

## Indexing
- SQLite index recommended
- Incremental scan using mtime/size
- Tag-first, folder/filename fallback

## Browse/Search
- Artists/Albums/Tracks with paging
- Search using FTS or prefix queries

## Playback
- `GetStream(trackId)` returns `file:///absolute/path`

## MVP acceptance
- Scan, browse, search, play a local track reliably
