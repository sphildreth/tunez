---
description: "Security baseline for Tunez (mpv + filesystem + HTTP)"
applyTo: "**/*.{go,sh,yml}"
---

## Security mindset
- Treat all external input as untrusted:
  - file paths from user input
  - URLs from APIs
  - metadata fields
- Avoid command injection:
  - never build shell strings
  - use argument arrays (`exec.CommandContext`)
- Validate and normalize paths before reading/writing.
- Set timeouts on HTTP clients; limit response sizes where practical.
- Prefer safe defaults: deny-by-default for risky operations.
