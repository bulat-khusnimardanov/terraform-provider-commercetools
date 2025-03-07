---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "commercetools_customer_group Resource - terraform-provider-commercetools"
subcategory: ""
description: |-
  A Customer can be a member of a customer group (for example reseller, gold member). Special prices can be assigned to specific products based on a customer group.
  See also the Customer Group API Documentation https://docs.commercetools.com/api/projects/customerGroups
---

# commercetools_customer_group (Resource)

A Customer can be a member of a customer group (for example reseller, gold member). Special prices can be assigned to specific products based on a customer group.

See also the [Customer Group API Documentation](https://docs.commercetools.com/api/projects/customerGroups)

## Example Usage

```terraform
resource "commercetools_customer_group" "standard" {
  key  = "my-customer-group-key"
  name = "Standard Customer Group"
}

resource "commercetools_customer_group" "golden" {
  key  = "my-customer-group-key"
  name = "Golden Customer Group"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Unique within the project

### Optional

- `custom` (Block List, Max: 1) (see [below for nested schema](#nestedblock--custom))
- `key` (String) User-specific unique identifier for the customer group

### Read-Only

- `id` (String) The ID of this resource.
- `version` (Number)

<a id="nestedblock--custom"></a>
### Nested Schema for `custom`

Required:

- `type_id` (String)

Optional:

- `fields` (Map of String)
