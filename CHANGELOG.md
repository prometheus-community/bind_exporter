## 0.6.1 / 2023-03-22

* [BUGFIX] Fix unmarshall error for negative values #166

## 0.6.0 / 2022-11-09

* [FEATURE] Add REFUSED label for metric bind_resolver_response_errors_total #125
* [ENHANCEMENT] Decode resp.Body directly, without ioutil.ReadAll #84
* [ENHANCEMENT] Update exporter-toolkit to support new listen options #151

## 0.5.0 / 2021-11-23

* [FEATURE] Add support for RCODE metrics. #113
* [BUGFIX] handle non integer values for zone serial. #97

## 0.4.0 / 2021-01-14

* [CHANGE] Replace legacy common/log with promlog #85
* [FEATURE] Add current recursive clients metric #74
* [FEATURE] Add zone serial numbers as metrics #91
* [FEATURE] Add TLS and basic authentication #94
* [BUGFIX] Use uint64 for counters in v3 xml #70
* [BUGFIX] Fix Gauge type for large gauges #90

## 0.3.0 / 2020-01-08

* [FEATURE] Support zone stats, enable some initial zone transfer metrics #49
* [ENHANCEMENT] Better flag defaults #50
* [BUGFIX] Fix parsing on 32bit systems. #58

## 0.2.0 / 2017-08-28

* [CHANGE] Rename label in `bind_incoming_requests_total` from `name` to `opcode`
* [CHANGE] Rename flag `-bind.statsuri` to `-bind.stats-url`
* [CHANGE] Duplicated queries are not an error and get now exported as `bind_query_duplicates_total`
* [FEATURE] Add support for BIND statistics v3
* [FEATURE] Automatically detect BIND statistics version and use correct client
* [FEATURE] Provide option to control exported statistics with `-bind.stats-groups`
* [FEATURE] Export number of queries causing recursion as `bind_query_recursions_total`
* [FEATURE] Export `bind_boot_time_seconds` (v2+v3) and `bind_config_time_seconds` (v3 only)
