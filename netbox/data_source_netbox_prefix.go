package netbox

import (
	"errors"
	"strconv"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/ipam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxPrefix() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNetboxPrefixRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cidr": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"cidr", "vlan_vid"},
				ValidateFunc: validation.IsCIDR,
			},
			"vlan_vid": {
				Type:         schema.TypeInt,
				Optional:     true,
				AtLeastOneOf: []string{"cidr", "vlan_vid"},
				ValidateFunc: validation.IntBetween(1, 4094),
			},
			"vrf_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceNetboxPrefixRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	params := ipam.NewIpamPrefixesListParams()

	if cidr, ok := d.Get("cidr").(string); ok && cidr != "" {
		params.Prefix = strToPtr(cidr)
	}

	if vid, ok := d.Get("vlan_vid").(int); ok && vid != 0 {
		params.VlanVid = float64ToPtr(float64(vid))
	}

	limit := int64(2) // Limit of 2 is enough
	params.Limit = &limit

	res, err := api.Ipam.IpamPrefixesList(params, nil)
	if err != nil {
		return err
	}

	if *res.GetPayload().Count > int64(1) {
		return errors.New("More than one result. Specify a more narrow filter")
	}
	if *res.GetPayload().Count == int64(0) {
		return errors.New("No result")
	}
	result := res.GetPayload().Results[0]
	d.Set("id", result.ID)
	if result.Vrf != nil {
		d.Set("vrf_id", result.Vrf.ID)
	}
	d.SetId(strconv.FormatInt(result.ID, 10))
	return nil
}
