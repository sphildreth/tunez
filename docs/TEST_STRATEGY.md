# Tunez — Test Strategy

**Last updated:** 2026-01-02

## Unit tests
- Queue operations
- Keybinding parsing/dispatch
- mpv IPC encode/decode

## Provider contract tests
Single suite run against each Provider:
- paging behavior
- browse flows
- search sanity
- GetStream returns usable URL

## Integration tests (build-tagged)
- Filesystem provider on fixture library
- Melodee provider using mocked HTTP (recorded fixtures)

## Fake mpv
A fake IPC server for deterministic player tests:
- Accept `command`
- Emit `property-change`
- Emit `end-file`
