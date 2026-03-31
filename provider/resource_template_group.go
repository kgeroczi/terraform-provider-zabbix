package provider

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/kgeroczi/go-zabbix-api"
)

// resourceTemplateGroup terraform resource handler
func resourceTemplateGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceTemplateGroupCreate,
		Read:   resourceTemplateGroupRead,
		Update: resourceTemplateGroupUpdate,
		Delete: resourceTemplateGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Description:  "Template Group Name",
				Required:     true,
			},
		},
	}
}

// dataTemplateGroup terraform data handler
func dataTemplateGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataTemplateGroupRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Description:  "Template Group Name",
				Required:     true,
			},
		},
	}
}

// terraform templategroup create function
func resourceTemplateGroupCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)

	item := zabbix.TemplateGroup{
		Name: d.Get("name").(string),
	}

	items := []zabbix.TemplateGroup{item}

	err := api.TemplateGroupsCreate(items)

	if err != nil {
		return err
	}

	log.Trace("created TemplateGroup: %s (id: %s)", items[0].Name, items[0].GroupID)

	d.SetId(items[0].GroupID)

	return resourceTemplateGroupRead(d, m)
}

// templateGroupRead terraform template group read function
func templateGroupRead(d *schema.ResourceData, m interface{}, params zabbix.Params) error {
	api := m.(*zabbix.API)

	templateGroups, err := api.TemplateGroupsGet(params)

	if err != nil {
		return err
	}

	if len(templateGroups) < 1 {
		d.SetId("")
		return nil
	}
	if len(templateGroups) > 1 {
		return errors.New("multiple template groups found")
	}
	t := templateGroups[0]

	log.Debug("Got TemplateGroup: %s (id: %s)", t.Name, t.GroupID)

	d.SetId(t.GroupID)
	d.Set("name", t.Name)

	return nil
}

// dataTemplateGroupRead terraform data resource read handler
func dataTemplateGroupRead(d *schema.ResourceData, m interface{}) error {
	return templateGroupRead(d, m, zabbix.Params{
		"filter": map[string]interface{}{
			"name": d.Get("name"),
		},
	})
}

// resourceTemplateGroupRead terraform resource read handler
func resourceTemplateGroupRead(d *schema.ResourceData, m interface{}) error {
	log.Debug("Lookup of TemplateGroup with id %s", d.Id())

	return templateGroupRead(d, m, zabbix.Params{
		"groupids": d.Id(),
	})
}

// resourceTemplateGroupUpdate terraform resource update handler
func resourceTemplateGroupUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)

	item := zabbix.TemplateGroup{
		GroupID: d.Id(),
		Name:    d.Get("name").(string),
	}

	items := []zabbix.TemplateGroup{item}

	err := api.TemplateGroupsUpdate(items)

	if err != nil {
		return err
	}

	return resourceTemplateGroupRead(d, m)
}

// resourceTemplateGroupDelete terraform resource delete handler
func resourceTemplateGroupDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)
	return api.TemplateGroupsDeleteByIds([]string{d.Id()})
}
