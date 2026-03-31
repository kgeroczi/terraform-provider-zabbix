---
page_title: "zabbix_template_group Resource - terraform-provider-zabbix"
subcategory: ""
description: |-
  Manages a Zabbix template group. Template groups are used to organize templates (since Zabbix 6.2).
---

# zabbix_template_group (Resource)

Manages a Zabbix template group.

Template groups were introduced in Zabbix 6.2 as a separation from host groups. Templates must be assigned to template groups (not host groups).

## Example Usage

```terraform
resource "zabbix_template_group" "example" {
  name = "My Templates"
}
```

## Argument Reference

- `name` - (Required) Name of the template group.

## Attribute Reference

- `id` - The ID of the template group.

## Import

Template groups can be imported using the ID:

```
terraform import zabbix_template_group.example 12345
```
