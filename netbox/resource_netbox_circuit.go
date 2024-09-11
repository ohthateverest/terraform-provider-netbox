package netbox

import (
	"strconv"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/circuits"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxCircuitStatusOptions = []string{"planned", "provisioning", "active", "offline", "deprovisioning", "decommissioning"}

func resourceNetboxCircuit() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCircuitCreate,
		Read:   resourceNetboxCircuitRead,
		Update: resourceNetboxCircuitUpdate,
		Delete: resourceNetboxCircuitDelete,

		Description: `:meta:subcategory:Circuits:From the [official documentation](https://docs.netbox.dev/en/stable/features/circuits/#circuits_1):

> A communications circuit represents a single physical link connecting exactly two endpoints, commonly referred to as its A and Z terminations. A circuit in NetBox may have zero, one, or two terminations defined. It is common to have only one termination defined when you don't necessarily care about the details of the provider side of the circuit, e.g. for Internet access circuits. Both terminations would likely be modeled for circuits which connect one customer site to another.
>
> Each circuit is associated with a provider and a user-defined type. For example, you might have Internet access circuits delivered to each site by one provider, and private MPLS circuits delivered by another. Each circuit must be assigned a circuit ID, each of which must be unique per provider.`,

		Schema: map[string]*schema.Schema{
			"provider_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"cid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"commit_rate": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"install_date": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"termination_date": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxCircuitStatusOptions, false),
				Description:  buildValidValueDescription(resourceNetboxCircuitStatusOptions),
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNetboxCircuitCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	data := models.WritableCircuit{}

	cid := d.Get("cid").(string)
	data.Cid = &cid

	data.Status = d.Get("status").(string)
	descriptionValue, ok := d.GetOk("description")
	if ok {
		data.Description = descriptionValue.(string)
	} else {
		data.Description = ""
	}

	installDateValue, ok := d.GetOk("install_date")
	if ok {
		installDateStr := installDateValue.(string) // Get the string value
		var parsedInstallDate strfmt.Date
		err := parsedInstallDate.UnmarshalText([]byte(installDateStr)) // Parse it into a strfmt.Date
		if err == nil {
			data.InstallDate = &parsedInstallDate // Assign the parsed date if successful
		} else {
			return err // Return the error from UnmarshalText if parsing fails
		}
	} else {
		data.InstallDate = nil // Set to nil if not provided
	}

	terminationDateValue, ok := d.GetOk("termination_date")
	if ok {
		terminationDateStr := terminationDateValue.(string) // Get the string value
		var parsedTerminationDate strfmt.Date
		err := parsedTerminationDate.UnmarshalText([]byte(terminationDateStr)) // Parse it into a strfmt.Date
		if err == nil {
			data.TerminationDate = &parsedTerminationDate // Assign the parsed date if successful
		} else {
			return err // Return the error from UnmarshalText if parsing fails
		}
	} else {
		data.TerminationDate = nil // Set to nil if not provided
	}

	providerIDValue, ok := d.GetOk("provider_id")
	if ok {
		data.Provider = int64ToPtr(int64(providerIDValue.(int)))
	}

	commitRateValue, ok := d.GetOk("commit_rate")
	if ok {
		data.CommitRate = int64ToPtr(int64(commitRateValue.(int)))
	}

	typeIDValue, ok := d.GetOk("type_id")
	if ok {
		data.Type = int64ToPtr(int64(typeIDValue.(int)))
	}

	tenantIDValue, ok := d.GetOk("tenant_id")
	if ok {
		data.Tenant = int64ToPtr(int64(tenantIDValue.(int)))
	}
	ct, ok := d.GetOk(customFieldsKey)
	if ok {
		data.CustomFields = ct
	}

	data.Comments = d.Get("comments").(string)

	//data.Tags = []*models.NestedTag{}
	data.Tags, _ = getNestedTagListFromResourceDataSet(api, d.Get(tagsKey))

	params := circuits.NewCircuitsCircuitsCreateParams().WithData(&data)

	res, err := api.Circuits.CircuitsCircuitsCreate(params, nil)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(res.GetPayload().ID, 10))

	return resourceNetboxCircuitRead(d, m)
}

func resourceNetboxCircuitRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := circuits.NewCircuitsCircuitsReadParams().WithID(id)

	res, err := api.Circuits.CircuitsCircuitsRead(params, nil)

	if err != nil {
		if errresp, ok := err.(*circuits.CircuitsCircuitsReadDefault); ok {
			errorcode := errresp.Code()
			if errorcode == 404 {
				// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-.html
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("cid", res.GetPayload().Cid)
	d.Set("status", res.GetPayload().Status.Value)

	if res.GetPayload().Provider != nil {
		d.Set("provider_id", res.GetPayload().Provider.ID)
	} else {
		d.Set("provider_id", nil)
	}

	if res.GetPayload().CommitRate != nil {
		d.Set("commit_rate", res.GetPayload().CommitRate)
	} else {
		d.Set("commit_rate", nil)
	}

	if res.GetPayload().Type != nil {
		d.Set("type_id", res.GetPayload().Type.ID)
	} else {
		d.Set("type_id", nil)
	}

	if res.GetPayload().Tenant != nil {
		d.Set("tenant_id", res.GetPayload().Tenant.ID)
	} else {
		d.Set("tenant_id", nil)
	}

	if res.GetPayload().Description != "" {
		d.Set("description", res.GetPayload().Description)
	} else {
		d.Set("description", "")
	}

	if res.GetPayload().InstallDate != nil {
		d.Set("install_date", res.GetPayload().InstallDate)
	} else {
		d.Set("install_date", nil)
	}

	if res.GetPayload().TerminationDate != nil {
		d.Set("termination_date", res.GetPayload().TerminationDate)
	} else {
		d.Set("termination_date", nil)
	}
	cf := getCustomFields(res.GetPayload().CustomFields)
	if cf != nil {
		d.Set(customFieldsKey, cf)
	}
	d.Set("comments", res.GetPayload().Comments)

	d.Set(tagsKey, getTagListFromNestedTagList(res.GetPayload().Tags))
	return nil
}

func resourceNetboxCircuitUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	data := models.WritableCircuit{}

	cid := d.Get("cid").(string)
	data.Cid = &cid

	data.Status = d.Get("status").(string)
	descriptionValue, ok := d.GetOk("description")
	if ok {
		data.Description = descriptionValue.(string)
	} else {
		data.Description = ""
	}

	installDateValue, ok := d.GetOk("install_date")
	if ok {
		installDateStr := installDateValue.(string) // Get the string value
		var parsedInstallDate strfmt.Date
		err := parsedInstallDate.UnmarshalText([]byte(installDateStr)) // Parse it into a strfmt.Date
		if err == nil {
			data.InstallDate = &parsedInstallDate // Assign the parsed date if successful
		} else {
			return err // Return the error from UnmarshalText if parsing fails
		}
	} else {
		data.InstallDate = nil // Set to nil if not provided
	}

	terminationDateValue, ok := d.GetOk("termination_date")
	if ok {
		terminationDateStr := terminationDateValue.(string) // Get the string value
		var parsedTerminationDate strfmt.Date
		err := parsedTerminationDate.UnmarshalText([]byte(terminationDateStr)) // Parse it into a strfmt.Date
		if err == nil {
			data.TerminationDate = &parsedTerminationDate // Assign the parsed date if successful
		} else {
			return err // Return the error from UnmarshalText if parsing fails
		}
	} else {
		data.TerminationDate = nil // Set to nil if not provided
	}

	providerIDValue, ok := d.GetOk("provider_id")
	if ok {
		data.Provider = int64ToPtr(int64(providerIDValue.(int)))
	}
	commitRateValue, ok := d.GetOk("commit_rate")
	if ok {
		data.CommitRate = int64ToPtr(int64(commitRateValue.(int)))
	}

	typeIDValue, ok := d.GetOk("type_id")
	if ok {
		data.Type = int64ToPtr(int64(typeIDValue.(int)))
	}

	tenantIDValue, ok := d.GetOk("tenant_id")
	if ok {
		data.Tenant = int64ToPtr(int64(tenantIDValue.(int)))
	}
	ct, ok := d.GetOk(customFieldsKey)
	if ok {
		data.CustomFields = ct
	}

	data.Comments = d.Get("comments").(string)

	//data.Tags = []*models.NestedTag{}
	data.Tags, _ = getNestedTagListFromResourceDataSet(api, d.Get(tagsKey))

	params := circuits.NewCircuitsCircuitsPartialUpdateParams().WithID(id).WithData(&data)

	_, err := api.Circuits.CircuitsCircuitsPartialUpdate(params, nil)
	if err != nil {
		return err
	}

	return resourceNetboxCircuitRead(d, m)
}

func resourceNetboxCircuitDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := circuits.NewCircuitsCircuitsDeleteParams().WithID(id)

	_, err := api.Circuits.CircuitsCircuitsDelete(params, nil)
	if err != nil {
		if errresp, ok := err.(*circuits.CircuitsCircuitsDeleteDefault); ok {
			if errresp.Code() == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}
	return nil
}
