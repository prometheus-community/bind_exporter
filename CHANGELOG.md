# v0.2.0 / unreleased

- [CHANGE] Rename label in `bind_incoming_requests_total` from `name` to `opcode`
- [CHANGE] Rename flag `-bind.statsuri` to `-bind.stats-url`
- [CHANGE] Duplicated queries are not an error and get now exported as `bind_query_duplicates_total`
- [FEATURE] Add support for BIND statistics v3
- [FEATURE] Automatically detect BIND statistics version and use correct client
- [FEATURE] Provide option to control exported statistics with `-bind.stats-groups`
- [FEATURE] Export number of queries causing recursion as `bind_query_recursions_total`
- [FEATURE] Export `bind_boot_time_seconds` (v2+v3) and `bind_config_time_seconds` (v3 only)