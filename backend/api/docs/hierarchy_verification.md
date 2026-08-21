# Hierarchy routing verification

Flow: API Key Product/Environment -> queue scope check -> routing rule/source/API-key defaults -> hierarchy resolver -> Elasticsearch document -> scoped Search API -> Log Explorer.

| Check | Result |
| --- | --- |
| Product cannot be overridden by payload | PASS — handler binds Product from API key. |
| Environment cannot be overridden by payload | PASS — handler binds Environment from API key and validates payload scope. |
| Project belongs to Product | PASS (fixed) — unknown/cross-product `project_id` or code is rejected. |
| Feature/Sub Feature belongs to Product/Project | PASS (fixed) — unknown/cross-scope category or feature code is rejected. |
| Search does not cross Product | PASS — Product term filter and backend access check are mandatory. |
| Search does not cross Environment | PASS — Environment filter is applied when scoped. |

Unclassified logs remain supported only when no hierarchy identifier is supplied.