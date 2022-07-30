package netbox

import (
	"strconv"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/extras"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxWebhookCreate,
		Read:   resourceNetboxWebhookRead,
		Update: resourceNetboxWebhookUpdate,
		Delete: resourceNetboxWebhookDelete,

		Schema: map[string]*schema.Schema{
			"additional_headers": {
				Type:        schema.TypeString,
				Description: "User-supplied HTTP headers to be sent with the request in addition to the HTTP content type. Headers should be defined in the format <code>Name: Value</code>. Jinja2 template processing is supported with the same context as the request body (below).",
				Optional:    true,
			},
			"body_template": {
				Type:        schema.TypeString,
				Description: "Jinja2 template for a custom request body. If blank, a JSON object representing the change will be included. Available context data includes: <code>event</code>, <code>model</code>, <code>timestamp</code>, <code>username</code>, <code>request_id</code>, and <code>data</code>.",
				Optional:    true,
			},
			"ca_file_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The specific CA certificate file to use for SSL verification. Leave blank to use the system defaults.",
				Default:     "",
			},
			"conditions": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A set of conditions which determine whether the webhook will be generated.",
				Default:     "",
			},
			"content_types": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				Set:      schema.HashString,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"http_method": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "HTTP method for webhook: [GET POST PUT PATCH DELETE]",
			},
			"http_content_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The complete list of official content types is available <a href=\"https://www.iana.org/assignments/media-types/media-types.xhtml\">here</a>.",
			},
			"payload_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "This URL will be called using the HTTP method defined when the webhook is called. Jinja2 template processing is supported with the same context as the request body.",
			},
			"secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "When provided, the request will include a 'X-Hook-Signature' header containing a HMAC hex digest of the payload body using the secret as the key. The secret is not transmitted in the request.",
			},
			"ssl_verification": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"type_create": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Call this webhook when a matching object is created.",
			},
			"type_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Call this webhook when a matching object is deleted.",
			},
			"type_update": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Call this webhook when a matching object is updated.",
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceNetboxWebhookCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	data := models.Webhook{
		AdditionalHeaders: d.Get("additional_headers").(string),
		BodyTemplate:      d.Get("body_template").(string),
		CaFilePath:        strToPtr(d.Get("ca_file_path").(string)),
		Conditions:        strToPtr(d.Get("conditions").(string)),
		ContentTypes:      toNList[string](d.Get("content_types")),
		Enabled:           d.Get("enabled").(bool),
		HTTPContentType:   d.Get("http_content_type").(string),
		HTTPMethod:        d.Get("http_method").(string),
		Name:              strToPtr(d.Get("name").(string)),
		PayloadURL:        strToPtr(d.Get("payload_url").(string)),
		Secret:            d.Get("secret").(string),
		SslVerification:   d.Get("ssl_verification").(bool),
		TypeCreate:        nToPtr(d.Get("type_create").(bool)),
		TypeDelete:        nToPtr(d.Get("type_delete").(bool)),
		TypeUpdate:        nToPtr(d.Get("type_update").(bool)),
	}

	params := extras.NewExtrasWebhooksCreateParams().WithData(&data)
	res, err := api.Extras.ExtrasWebhooksCreate(params, nil)
	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(res.GetPayload().ID, 10))

	return resourceNetboxWebhookRead(d, m)
}

func resourceNetboxWebhookRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := extras.NewExtrasWebhooksReadParams().WithID(id)

	res, err := api.Extras.ExtrasWebhooksRead(params, nil)
	if err != nil {
		errorcode := err.(*extras.ExtrasWebhooksReadDefault).Code()
		if errorcode == 404 {
			// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
			d.SetId("")
			return nil
		}
		return err
	}

	webhook := res.GetPayload()

	d.Set("additional_headers", webhook.AdditionalHeaders)
	d.Set("body_template", webhook.BodyTemplate)
	d.Set("enabled", webhook.Enabled)
	d.Set("http_content_type", webhook.HTTPContentType)
	d.Set("http_method", webhook.HTTPMethod)
	d.Set("secret", webhook.Secret)
	d.Set("ssl_verification", webhook.SslVerification)
	d.Set("content_types", webhook.ContentTypes)

	if webhook.CaFilePath != nil {
		d.Set("ca_file_path", *webhook.CaFilePath)
	} else {
		d.Set("ca_file_path", "")
	}

	if webhook.Conditions != nil {
		d.Set("conditions", *webhook.Conditions)
	} else {
		d.Set("conditions", "")
	}

	if webhook.Name != nil {
		d.Set("name", *webhook.Name)
	} else {
		d.Set("name", "")
	}

	if webhook.PayloadURL != nil {
		d.Set("payload_url", *webhook.PayloadURL)
	} else {
		d.Set("payload_url", "")
	}

	if webhook.TypeCreate != nil {
		d.Set("type_create", *webhook.TypeCreate)
	} else {
		d.Set("type_create", false)
	}

	if webhook.TypeUpdate != nil {
		d.Set("type_update", *webhook.TypeUpdate)
	} else {
		d.Set("type_update", false)
	}

	if webhook.TypeDelete != nil {
		d.Set("type_delete", *webhook.TypeDelete)
	} else {
		d.Set("type_delete", false)
	}

	return nil
}

func resourceNetboxWebhookUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	data := models.Webhook{
		AdditionalHeaders: d.Get("additional_headers").(string),
		BodyTemplate:      d.Get("body_template").(string),
		CaFilePath:        strToPtr(d.Get("ca_file_path").(string)),
		Conditions:        strToPtr(d.Get("conditions").(string)),
		ContentTypes:      toNList[string](d.Get("content_types")),
		Enabled:           d.Get("enabled").(bool),
		HTTPContentType:   d.Get("http_content_type").(string),
		HTTPMethod:        d.Get("http_method").(string),
		Name:              strToPtr(d.Get("name").(string)),
		PayloadURL:        strToPtr(d.Get("payload_url").(string)),
		Secret:            d.Get("secret").(string),
		SslVerification:   d.Get("ssl_verification").(bool),
		TypeCreate:        nToPtr(d.Get("type_create").(bool)),
		TypeDelete:        nToPtr(d.Get("type_delete").(bool)),
		TypeUpdate:        nToPtr(d.Get("type_update").(bool)),
	}

	params := extras.NewExtrasWebhooksUpdateParams().WithID(id).WithData(&data)
	_, err := api.Extras.ExtrasWebhooksUpdate(params, nil)
	if err != nil {
		return err
	}
	return resourceNetboxWebhookRead(d, m)
}

func resourceNetboxWebhookDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := extras.NewExtrasWebhooksDeleteParams().WithID(id)
	_, err := api.Extras.ExtrasWebhooksDelete(params, nil)
	if err != nil {
		return err
	}

	return nil
}
