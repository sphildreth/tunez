# Tunez — Test Strategy

**Last updated:** 2026-01-03

## Quick Reference

```bash
# Run all tests
cd src && go test ./...

# Run with verbose output
go test ./... -v

# Run specific test
go test ./internal/app/... -run TestScreensGolden

# Update golden files after UI changes
go test ./internal/app/... -run TestScreensGolden -update

# Skip slow interactive tests
go test ./... -short
```

## Test Types

### Unit tests
- Queue operations (`internal/queue`)
- Config validation (`internal/config`)
- Player IPC (`internal/player`)

### Provider contract tests
Single suite run against each Provider:
- paging behavior
- browse flows
- search sanity
- GetStream returns usable URL

### TUI Integration tests (teatest)
Located in `internal/app/integration_test.go`:

| Test | Purpose |
|------|---------|
| `TestScreensGolden` | Golden file comparison for all screens |
| `TestViewOutput` | Validates expected text appears on each screen |
| `TestKeyboardShortcuts` | Verifies j/k/? keybindings work |
| `TestInteractiveNavigation` | Full interactive session simulation |
| `TestNavigation` | Unit test for navigation state changes |

### Golden Files
Located in `internal/app/testdata/TestScreensGolden/`:
- `now_playing_empty.golden` - Now Playing with no track
- `library_artists.golden` - Library showing artist list
- `library_albums.golden` - Library showing album list  
- `library_tracks.golden` - Library showing track list
- `queue_empty.golden` - Empty queue screen
- `search_empty.golden` - Search with no query
- `config_screen.golden` - Configuration screen
- `help_overlay.golden` - Help overlay visible

**Updating golden files:** After intentional UI changes, regenerate with:
```bash
go test ./internal/app/... -run TestScreensGolden -update
```

## For Coding Agents

### Verifying UI Changes
After modifying any render function in `app.go`:

1. **Run golden tests** to see if output changed:
   ```bash
   go test ./internal/app/... -run TestScreensGolden -v
   ```

2. **If test fails**, review the diff shown in output

3. **If change is intentional**, update golden files:
   ```bash
   go test ./internal/app/... -run TestScreensGolden -update
   ```

4. **Verify specific screens** contain expected elements:
   ```bash
   go test ./internal/app/... -run TestViewOutput -v
   ```

### Testing Keybindings
After modifying keybinding behavior:
```bash
go test ./internal/app/... -run TestKeyboardShortcuts -v
```

### Testing Navigation Flow
After modifying screen transitions:
```bash
go test ./internal/app/... -run TestNavigation -v
go test ./internal/app/... -run TestInteractiveNavigation -v
```

### Adding New Screen Tests
To add golden file test for a new screen:

1. Add test case in `TestScreensGolden`:
   ```go
   {
       name: "my_new_screen",
       setup: func(m Model) Model {
           m = initializeModel(m, prov)
           m.screen = screenMyNew
           return m
       },
   },
   ```

2. Run with `-update` to generate golden file:
   ```bash
   go test ./internal/app/... -run TestScreensGolden/my_new_screen -update
   ```

## Melodee API Development Environment
For development and testing of the Melodee API Provider, use a mock server or staging environment rather than production APIs. Options include:
- Recorded HTTP fixtures (see integration tests above)
- Local mock server implementing the Melodee API contract
- Dedicated staging/dev instance of Melodee server

## Fake mpv
A fake IPC server for deterministic player tests:
- Accept `command`
- Emit `property-change`
- Emit `end-file`
