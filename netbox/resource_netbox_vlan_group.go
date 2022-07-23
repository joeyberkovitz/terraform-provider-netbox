package netbox

import (
	"strconv"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/ipam"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxVlanGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVlanGroupCreate,
		Read:   resourceNetboxVlanGroupRead,
		Update: resourceNetboxVlanGroupUpdate,
		Delete: resourceNetboxVlanGroupDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"tags": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				Set:      schema.HashString,
			},
			"min_vid": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"max_vid": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  4094,
			},
			"scope_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"scope_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceNetboxVlanGroupCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	data := models.VLANGroup{}

	name := d.Get("name").(string)
	slug := d.Get("slug").(string)
	description := d.Get("description").(string)
	min_vid := d.Get("min_vid").(int)
	max_vid := d.Get("max_vid").(int)

	data.Name = &name
	data.Slug = &slug
	data.Description = description
	data.MinVid = int64(min_vid)
	data.MaxVid = int64(max_vid)

	if scopeType, ok := d.GetOk("scope_type"); ok {
		data.ScopeType = scopeType.(string)
	}

	if scopeID, ok := d.GetOk("scope_id"); ok {
		data.ScopeID = int64ToPtr(int64(scopeID.(int)))
	}

	data.Tags, _ = getNestedTagListFromResourceDataSet(api, d.Get("tags"))

	params := ipam.NewIpamVlanGroupsCreateParams().WithData(&data)
	res, err := api.Ipam.IpamVlanGroupsCreate(params, nil)
	if err != nil {
		return err
	}
	d.SetId(strconv.FormatInt(res.GetPayload().ID, 10))

	return resourceNetboxVlanGroupRead(d, m)
}

func resourceNetboxVlanGroupRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := ipam.NewIpamVlanGroupsReadParams().WithID(id)

	res, err := api.Ipam.IpamVlanGroupsRead(params, nil)
	if err != nil {
		errorcode := err.(*ipam.IpamVlanGroupsReadDefault).Code()
		if errorcode == 404 {
			// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
			d.SetId("")
			return nil
		}
		return err
	}

	group := res.GetPayload()

	d.Set("name", group.Name)
	d.Set("slug", group.Slug)
	d.Set("description", group.Description)
	d.Set("tags", getTagListFromNestedTagList(group.Tags))
	d.Set("scope_type", group.ScopeType)
	d.Set("min_vid", group.MinVid)
	d.Set("max_vid", group.MaxVid)

	if group.ScopeID != nil {
		d.Set("scope_id", *group.ScopeID)
	}

	return nil
}

func resourceNetboxVlanGroupUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	data := models.VLANGroup{}

	name := d.Get("name").(string)
	slug := d.Get("slug").(string)
	description := d.Get("description").(string)
	min_vid := d.Get("min_vid").(int)
	max_vid := d.Get("max_vid").(int)

	data.Name = &name
	data.Slug = &slug
	data.Description = description
	data.MinVid = int64(min_vid)
	data.MaxVid = int64(max_vid)

	if scopeType, ok := d.GetOk("scope_type"); ok {
		data.ScopeType = scopeType.(string)
	}

	if scopeID, ok := d.GetOk("scope_id"); ok {
		data.ScopeID = int64ToPtr(int64(scopeID.(int)))
	}

	data.Tags, _ = getNestedTagListFromResourceDataSet(api, d.Get("tags"))

	params := ipam.NewIpamVlanGroupsUpdateParams().WithID(id).WithData(&data)
	_, err := api.Ipam.IpamVlanGroupsUpdate(params, nil)
	if err != nil {
		return err
	}
	return resourceNetboxVlanGroupRead(d, m)
}

func resourceNetboxVlanGroupDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := ipam.NewIpamVlanGroupsDeleteParams().WithID(id)
	_, err := api.Ipam.IpamVlanGroupsDelete(params, nil)
	if err != nil {
		return err
	}

	return nil
}
