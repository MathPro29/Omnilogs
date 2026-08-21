# Payload mapping

| Incoming payload | Stored Elasticsearch document | Search API | Log Explorer |
| --- | --- | --- | --- |
| Nested JSON / arrays | `data` and compatible `payload` retain the original structure | Returned in `raw.data` / `raw.payload` | Raw inspector and nested field lookup |
| `custom_fields.*` | `data.custom_fields.*`, plus root `custom_fields` alias | Dynamic custom-field query supports `custom_fields.*` | Favorite/custom columns |
| `metadata` | `data.metadata`, root `metadata` alias | Returned raw; searchable by dynamic path | Metadata panel |
| `headers`, `request`, `query` | Retained under `data`; method/path copied when present | Returned raw; dynamic search supported | Request/header/query panels |
| Request/response body | `data.request` / `data.response` (or payload aliases) | Returned raw | Request/response body panels |
| `trace_id`, `request_id` | Root aliases and `data.*`; `source_request_id` maps to request ID | Searchable by `trace_id` / `request_id` | Trace panel |
| `timestamp` | `@timestamp` and index timestamp | Time-range search | Timestamp column |
| `level` / `log_level` | Root `level`, preserved in `data` | Level search | Level column |
| Product/environment | API-key scoped root `product_id` / `environment_id` | Mandatory Product filter; optional Environment filter | Scope selector/result fields |
| Project/feature/sub feature | Canonical `project_id`, `category_id`, `feature_path_ids`, `feature_full_path` | Project/category scoped filters | Hierarchy column |

All incoming fields are retained in the raw document unless a configured masking rule replaces a sensitive value.