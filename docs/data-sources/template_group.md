---
page_title: "zabbix_template_group Data Source - terraform-provider-zabbix"
subcategory: ""
description: |-
  Use this data source to look up a Zabbix template group by name.
---

# zabbix_template_group (Data Source)

Use this data source to look up a Zabbix template group by name.

## Example Usage

```terraform
data "zabbix_template_group" "example" {
  name = "Templates"
}
```

## Argument Reference

- `name` - (Required) Name of the template group.

## Attribute Reference

- `id` - The ID of the template group.
