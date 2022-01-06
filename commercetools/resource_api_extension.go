package commercetools

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceAPIExtension() *schema.Resource {
	return &schema.Resource{
		Description: "Create a new API extension to extend the bevahiour of an API with business logic. " +
			"Note that API extensions affect the performance of the API it is extending. If it fails, the whole API " +
			"call fails \n\n" +
			"Also see the [API Extension API Documentation](https://docs.commercetools.com/api/projects/api-extensions)",
		Create: resourceAPIExtensionCreate,
		Read:   resourceAPIExtensionRead,
		Update: resourceAPIExtensionUpdate,
		Delete: resourceAPIExtensionDelete,
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "User-specific unique identifier for the extension",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"destination": {
				Description: "[Destination](https://docs.commercetools.com/api/projects/api-extensions#destination) " +
					"Details where the extension can be reached",
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateDestinationType,
						},
						// HTTP specific fields
						"url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"azure_authentication": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"authorization_header": {
							Type:     schema.TypeString,
							Optional: true,
						},

						// AWSLambda specific fields
						"arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"access_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"access_secret": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"trigger": {
				Description: "Array of [Trigger](https://docs.commercetools.com/api/projects/api-extensions#trigger) " +
					"Describes what triggers the extension",
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type_id": {
							Description: "Currently, cart, order, payment, and customer are supported",
							Type:        schema.TypeString,
							Required:    true,
						},
						"actions": {
							Description: "Currently, Create and Update are supported",
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"timeout_in_ms": {
				Description: "Extension timeout in milliseconds",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func validateDestinationType(val interface{}, key string) (warns []string, errs []error) {
	var v = strings.ToLower(val.(string))

	switch v {
	case
		"http",
		"awslambda":
		return
	default:
		errs = append(errs, fmt.Errorf("%q not a valid value for %q", val, key))
	}
	return
}

func resourceAPIExtensionCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var extension *platform.Extension

	triggers := resourceAPIExtensionGetTriggers(d)
	destination, err := resourceAPIExtensionGetDestination(d)
	if err != nil {
		return err
	}

	draft := platform.ExtensionDraft{
		Key:         stringRef(d.Get("key")),
		Destination: destination,
		Triggers:    triggers,
		TimeoutInMs: intRef(d.Get("timeout_in_ms")),
	}

	err = resource.Retry(20*time.Second, func() *resource.RetryError {
		var err error

		extension, err = client.Extensions().Post(draft).Execute(context.Background())
		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if extension == nil {
		return fmt.Errorf("Error creating extension")
	}

	d.SetId(extension.ID)
	d.Set("version", extension.Version)

	return resourceAPIExtensionRead(d, m)
}

func resourceAPIExtensionRead(d *schema.ResourceData, m interface{}) error {
	log.Print("[DEBUG] Reading extensions from commercetools")
	client := getClient(m)

	extension, err := client.Extensions().WithId(d.Id()).Get().Execute(context.Background())

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if extension == nil {
		log.Print("[DEBUG] No extensions found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following extensions:")
		log.Print(stringFormatObject(extension))

		d.Set("version", extension.Version)
		d.Set("key", extension.Key)
		d.Set("destination", extension.Destination)
		d.Set("trigger", extension.Triggers)
		d.Set("timeout_in_ms", extension.TimeoutInMs)
	}
	return nil
}

func resourceAPIExtensionUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	input := platform.ExtensionUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.ExtensionUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.ExtensionSetKeyAction{Key: &newKey})
	}

	if d.HasChange("trigger") {
		triggers := resourceAPIExtensionGetTriggers(d)
		input.Actions = append(
			input.Actions,
			&platform.ExtensionChangeTriggersAction{Triggers: triggers})
	}

	if d.HasChange("destination") {
		destination, err := resourceAPIExtensionGetDestination(d)
		if err != nil {
			return err
		}
		input.Actions = append(
			input.Actions,
			&platform.ExtensionChangeDestinationAction{Destination: destination})
	}

	if d.HasChange("timeout_in_ms") {
		newTimeout := d.Get("timeout_in_ms").(int)
		input.Actions = append(
			input.Actions,
			&platform.ExtensionSetTimeoutInMsAction{TimeoutInMs: &newTimeout})
	}

	_, err := client.Extensions().WithId(d.Id()).Post(input).Execute(context.Background())
	if err != nil {
		return err
	}

	return resourceAPIExtensionRead(d, m)
}

func resourceAPIExtensionDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.Extensions().WithId(d.Id()).Delete().WithQueryParams(platform.ByProjectKeyExtensionsByIDRequestMethodDeleteInput{
		Version: version,
	}).Execute(context.Background())
	if err != nil {
		return err
	}
	return nil
}

//
// Helper methods
//

func resourceAPIExtensionGetDestination(d *schema.ResourceData) (platform.Destination, error) {
	input := d.Get("destination").(map[string]interface{})
	switch strings.ToLower(input["type"].(string)) {
	case "http":
		auth, err := resourceAPIExtensionGetAuthentication(input)
		if err != nil {
			return nil, err
		}

		return platform.ExtensionHttpDestination{
			Url:            input["url"].(string),
			Authentication: &auth,
		}, nil
	case "awslambda":
		return platform.ExtensionAWSLambdaDestination{
			Arn:          input["arn"].(string),
			AccessKey:    input["access_key"].(string),
			AccessSecret: input["access_secret"].(string),
		}, nil
	default:
		return nil, fmt.Errorf("Extension type %s not implemented", input["type"])
	}
}

func resourceAPIExtensionGetAuthentication(destInput map[string]interface{}) (platform.ExtensionHttpDestinationAuthentication, error) {
	authKeys := [2]string{"authorization_header", "azure_authentication"}
	count := 0
	for _, key := range authKeys {
		if _, ok := destInput[key]; ok {
			count++
		}
	}
	if count > 1 {
		return nil, fmt.Errorf(
			"In the destination only one of the auth values should be definied: %q", authKeys)
	}

	if authVal, ok := destInput["authorization_header"]; ok {
		return &platform.ExtensionAuthorizationHeaderAuthentication{
			HeaderValue: authVal.(string),
		}, nil
	}
	if authVal, ok := destInput["azure_authentication"]; ok {
		return &platform.ExtensionAzureFunctionsAuthentication{
			Key: authVal.(string),
		}, nil
	}

	return nil, nil
}

func resourceAPIExtensionGetTriggers(d *schema.ResourceData) []platform.ExtensionTrigger {
	input := d.Get("trigger").([]interface{})
	var result []platform.ExtensionTrigger

	for _, raw := range input {
		i := raw.(map[string]interface{})
		typeID := i["resource_type_id"].(string)

		actions := []platform.ExtensionAction{}
		for _, item := range expandStringArray(i["actions"].([]interface{})) {
			actions = append(actions, platform.ExtensionAction(item))
		}

		result = append(result, platform.ExtensionTrigger{
			ResourceTypeId: platform.ExtensionResourceTypeId(typeID),
			Actions:        actions,
		})
	}

	return result
}
