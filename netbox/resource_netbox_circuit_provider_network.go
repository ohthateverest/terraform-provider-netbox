package netbox

import (
	"strconv"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/circuits"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCircuitProviderNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCircuitProviderNetworkCreate,
		Read:   resourceNetboxCircuitProviderNetworkRead,
		Update: resourceNetboxCircuitProviderNetworkUpdate,
		Delete: resourceNetboxCircuitProviderNetworkDelete,

		Description: `:meta:subcategory:Circuits:From the [official documentation](https://docs.netbox.dev/en/stable/features/circuits/#providers):

> A circuit provider is any entity which provides some form of connectivity of among sites or organizations within a site. While this obviously includes carriers which offer Internet and private transit service, it might also include Internet exchange (IX) points and even organizations with whom you peer directly. Each circuit within NetBox must be assigned a provider and a circuit ID which is unique to that provider.
>
> Each provider may be assigned an autonomous system number (ASN), an account number, and contact information.`,

		Schema: map[string]*schema.Schema{
			"provider_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"service_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 200),
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNetboxCircuitProviderNetworkCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	data := models.WritableProviderNetwork{}

	providerIDValue, ok := d.GetOk("provider_id")
	if ok {
		data.Provider = int64ToPtr(int64(providerIDValue.(int)))
	}
	name := d.Get("name").(string)
	data.Name = &name
	ct, ok := d.GetOk(customFieldsKey)
	if ok {
		data.CustomFields = ct
	}

	data.Tags, _ = getNestedTagListFromResourceDataSet(api, d.Get(tagsKey))
	data.Comments = getOptionalStr(d, "comments", false)
	data.ServiceID = getOptionalStr(d, "service_id", false)
	data.Description = getOptionalStr(d, "description", false)

	params := circuits.NewCircuitsProviderNetworksCreateParams().WithData(&data)

	res, err := api.Circuits.CircuitsProviderNetworksCreate(params, nil)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(res.GetPayload().ID, 10))

	return resourceNetboxCircuitProviderNetworkRead(d, m)
}

func resourceNetboxCircuitProviderNetworkRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := circuits.NewCircuitsProviderNetworksReadParams().WithID(id)

	res, err := api.Circuits.CircuitsProviderNetworksRead(params, nil)

	if err != nil {
		if errresp, ok := err.(*circuits.CircuitsProvidersReadDefault); ok {
			errorcode := errresp.Code()
			if errorcode == 404 {
				// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("name", res.GetPayload().Name)
	d.Set("provider_id", res.GetPayload().Provider)
	d.Set("service_id", res.GetPayload().ServiceID)
	d.Set("description", res.GetPayload().Description)

	cf := getCustomFields(res.GetPayload().CustomFields)
	if cf != nil {
		d.Set(customFieldsKey, cf)
	}
	d.Set("comments", res.GetPayload().Comments)

	d.Set(tagsKey, getTagListFromNestedTagList(res.GetPayload().Tags))

	return nil
}

func resourceNetboxCircuitProviderNetworkUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	data := models.WritableProviderNetwork{}

	providerIDValue, ok := d.GetOk("provider_id")
	if ok {
		data.Provider = int64ToPtr(int64(providerIDValue.(int)))
	}
	name := d.Get("name").(string)
	data.Name = &name
	ct, ok := d.GetOk(customFieldsKey)
	if ok {
		data.CustomFields = ct
	}

	data.Tags, _ = getNestedTagListFromResourceDataSet(api, d.Get(tagsKey))
	data.Comments = getOptionalStr(d, "comments", false)
	data.ServiceID = getOptionalStr(d, "service_id", false)
	data.Description = getOptionalStr(d, "description", false)

	params := circuits.NewCircuitsProviderNetworksPartialUpdateParams().WithID(id).WithData(&data)

	_, err := api.Circuits.CircuitsProviderNetworksPartialUpdate(params, nil)
	if err != nil {
		return err
	}

	return resourceNetboxCircuitProviderNetworkRead(d, m)
}

func resourceNetboxCircuitProviderNetworkDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := circuits.NewCircuitsProviderNetworksDeleteParams().WithID(id)

	_, err := api.Circuits.CircuitsProviderNetworksDelete(params, nil)
	if err != nil {
		return err
	}
	return nil
}
