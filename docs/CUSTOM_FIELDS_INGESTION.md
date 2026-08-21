# Omnilogs custom-fields ingestion

Omnilogs no longer assigns a log to a project or feature from HTTP paths, route keys, or routing rules. The source Backend must send the hierarchy that owns each log.

```json
{
  "custom_fields": {
    "product_code": "ASSETWISE",
    "environment_code": "UAT",
    "project_code": "CONDO",
    "category_code": "FINANCE",
    "feature_code": "RECEIPTS",
    "sub_feature_code": "CREATE"
  }
}
```

IDs are also accepted when the source already has them: `product_id`, `environment_id`, `project_id`, and `category_id`. Codes are preferred because they are stable across environments. Product and environment scope still come from the authenticated ingest key/request and are validated against the custom fields.

Omnilogs normalizes valid values into the searchable document fields (`product_id`, `environment_id`, `project_id`, `category_id`, and feature path fields) while retaining the original `custom_fields`. The Log Explorer can therefore filter Product, Environment, Project, Category, and Sub-feature without a mapping configuration.

The old top-level hierarchy fields and `routing.explicitHierarchy` remain readable during migration, but route/rule matching is not used by the worker.