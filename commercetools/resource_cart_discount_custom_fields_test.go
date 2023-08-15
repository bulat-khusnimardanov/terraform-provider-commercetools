package commercetools

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccCartDiscount_CustomField(t *testing.T) {
	resourceName := "commercetools_cart_discount.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCartDiscountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNewCartDiscountConfigWithCustomField(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", "test"),
					func(s *terraform.State) error {
						result, err := testGetCartDiscount(s, resourceName)
						if err != nil {
							return err
						}

						assert.NotNil(t, result)
						assert.NotNil(t, result.Custom)
						assert.NotNil(t, result.Custom.Fields)
						assert.EqualValues(t, "foobar", result.Custom.Fields["my-string-field"])
						assert.EqualValues(t, []any{"ENUM-1", "ENUM-3"}, result.Custom.Fields["my-enum-set-field"])
						assert.EqualValues(t, map[string]interface{}{"centAmount": float64(150000), "currencyCode": "EUR", "fractionDigits": float64(2), "type": "centPrecision"}, result.Custom.Fields["my-money-field"])
						return nil
					},
				),
			},
		},
	})
}

func testAccNewCartDiscountConfigWithCustomField() string {
	return hclTemplate(`
		resource "commercetools_type" "test" {
			key = "test-for-cart-discount"
			name = {
				en = "for cart-discount"
			}
			description = {
				en = "Custom Field for cart-discount resource"
			}

			resource_type_ids = ["cart-discount"]

			field {
				name = "my-string-field"
				label = {
					en = "My Custom string field"
				}
				type {
					name = "String"
				}
			}

			field {
				name = "my-enum-set-field"
				label = {
					en = "My Custom Set of enums field"
				}
				type {
					name = "Set"
					element_type {
						name = "Enum"
						value {
							key   = "ENUM-1"
							label = "ENUM 1"
						}
						value {
							key   = "ENUM-2"
							label = "ENUM 2"
						}
						value {
							key   = "ENUM-3"
							label = "ENUM 3"
						}
					}
				}
			}

			field {
				name = "my-money-field"
				label = {
					en = "My Custom money field"
				}
				required = true
				type {
					name = "Money"
				}
				input_hint = "SingleLine"
			}
		}

		resource "commercetools_cart_discount" "test" {
			key = "test"
			name = {
				en = "test-cart-discount"
			}

			description = {
				en = "Cart discount fir custom fields"
			}

			value {
				type = "absolute"
				money {
				  currency_code = "EUR"
				  cent_amount   = 1000
				}
			}
			predicate = "lineItemExists(price.discount.id is not defined) = true and totalPrice >= \"50 GBP\""
			target {
				type      = "lineItems"
				predicate = "price.discount.id is not defined"
			}
			sort_order             = "0.1111"
			is_active              = false
			requires_discount_code = true
			stacking_mode          = "Stacking"

			custom {
				type_id = commercetools_type.test.id
				fields = {
					"my-string-field" = "foobar"
					"my-enum-set-field" = jsonencode(["ENUM-1", "ENUM-3"])
					"my-money-field" = jsonencode({
						"type" : "centPrecision",
						"currencyCode" : "EUR",
						"centAmount" : 150000,
						"fractionDigits" : 2
					})
				}
			}
		}
	`, map[string]any{})
}
