# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog and this project follows Semantic Versioning.

## [v0.3.2] - 2026-04-20

### Changed
- Upgraded `github.com/kgeroczi/go-zabbix-api` to `v0.3.2`.

### Fixed
- Improved compatibility with Zabbix 7.x JSON payload/response formats by consuming go-zabbix-api `v0.3.2` model fixes:
  - String-backed numeric field decoding for macros, services, and SLAs.
  - Correct host group payload field mapping for host create/update requests.
  - Required SLA payload field handling for `service_tags` and `effective_date`.
  - Item payload cleanup for optional fields not accepted by all item types.
- Provider-side defaults and schemas adjusted for more predictable create payloads:
  - `monitored_by` default set to server mode.
  - Removed invalid default `interfaceid` value for item resources.
  - Marked SNMP passphrase/secret-like fields as sensitive in host/SNMP schemas.

### Validation
- go test -mod=mod ./... passes.
- terraform -chdir=terraform_local validate passes with local development override.

## [v0.3.1] - 2026-03-31

### Added
- New resource: zabbix_sla
- New resource: zabbix_service
- New resource: zabbix_report
- New resource docs: proxy, user, user_group
- New data source docs: user
- Acceptance tests for new resources:
  - resource_sla_test.go
  - resource_service_test.go
  - resource_report_test.go

### Changed
- Upgraded github.com/kgeroczi/go-zabbix-api to v0.3.1.
- Migrated provider implementation from terraform-plugin-sdk v1 to terraform-plugin-sdk/v2.
- Updated provider entrypoint wiring and test scaffolding for SDK v2.
- Replaced legacy hash helper usage with schema.HashString.
- Refreshed provider documentation from current schema definitions.
- Updated README resource index and testing/docs generation guidance.

### Fixed
- Acceptance test precheck now correctly supports:
  - URL aliases: ZABBIX_URL or ZABBIX_SERVER_URL
  - Token auth: ZABBIX_TOKEN or ZABBIX_API_TOKEN
  - Username/password aliases: ZABBIX_USER or ZABBIX_USERNAME, and ZABBIX_PASS or ZABBIX_PASSWORD

### Validation
- go test -mod=mod ./... passes.
- go vet ./... passes.
