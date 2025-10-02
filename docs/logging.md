# Logging Cheatsheet

The service uses structured logs (JSON by default). Tune verbosity with environment variables?no code changes needed.

## Quick Reference

| Variable | Default | Description |
| --- | --- | --- |
| `LOG_LEVEL` | `INFO` | One of `DEBUG`, `INFO`, `WARN`, `ERROR`. |
| `LOG_FORMAT` | `json` | Set to `text` for human-readable dev output. |
| `ENV` | `local` | Implicitly adjusts some logging presets (see below). |

## Recommended Settings

| Environment | Suggested Values |
| --- | --- |
| Local dev | `LOG_LEVEL=DEBUG`, `LOG_FORMAT=text` |
| QA/Stage | `LOG_LEVEL=INFO`, `LOG_FORMAT=json` |
| Production | `LOG_LEVEL=WARN`, `LOG_FORMAT=json` |

## Fields Emitted

Each log record includes:
- `ts`: timestamp (RFC3339)
- `level`: log level
- `msg`: message text
- `request.id`: correlation ID when available
- `tenant.id`: resolved tenant (if any)
- `actor.id`, `actor.role`: populated after auth

## Tips
- Wrap curl requests with `X-Request-ID` to correlate logs.
- Avoid enabling `DEBUG` in production; it emits sensitive token claims.
- Structured logs play nicely with Stackdriver, Datadog, and ELK.
