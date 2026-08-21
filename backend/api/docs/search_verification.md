# Search verification

| Search | Result | Verification |
| --- | --- | --- |
| Full text | PASS | Worker creates `search_text`; global search targets it. |
| Time range | PASS | Range filter uses `@timestamp`. |
| Level | PASS | Searches root/data level aliases. |
| Service | PASS | Searches `data.service` and root alias. |
| Environment / Product | PASS | Product is mandatory; Environment is scoped filter. |
| Project | PASS | Matches canonical and compatibility paths. |
| Feature / Sub Feature | PASS | Category/path-ID filters match hierarchy paths. |
| Trace ID / Request ID | PASS | Root and data aliases are searchable. |
| Custom field | PASS (fixed) | `custom_fields.*` maps to `data.custom_fields.*`. |
| Log Explorer | PASS | Uses `POST /v1/logs/search` response; frontend build passed. |

`go test ./internal/main_logs/usecase/...` passed. Live API verification remains dependent on the API process running the current build.