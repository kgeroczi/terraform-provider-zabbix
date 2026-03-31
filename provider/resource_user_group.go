package provider

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/kgeroczi/go-zabbix-api"
)

// resourceUserGroup terraform resource handler
func resourceUserGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserGroupCreate,
		Read:   resourceUserGroupRead,
		Update: resourceUserGroupUpdate,
		Delete: resourceUserGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Description:  "Name of the user group.",
				Required:     true,
			},
			"debug_mode": &schema.Schema{
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntBetween(0, 1),
				Description:  "Whether debug mode is enabled or disabled.",
				Optional:     true,
				Default:      0,
			},
			"gui_access": &schema.Schema{
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntBetween(0, 3),
				Description:  "Frontend authentication method of the users in the group.",
				Optional:     true,
				Default:      0,
			},
			"status": &schema.Schema{
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntBetween(0, 1),
				Description:  "Whether the user group is enabled or disabled. For deprovisioned users, the user group cannot be enabled.",
				Optional:     true,
				Default:      0,
			},
			"host_permission": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"permission": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(0, 3),
							Required:     true,
						},
					},
				},
			},
			"template_permission": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "Template group permissions",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"permission": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(0, 3),
							Required:     true,
						},
					},
				},
			},
		},
	}
}

func resourceHostGroupPermissionsV1(d *schema.ResourceData) []zabbix.UserGroupPermission {
	var permissionsRequests []zabbix.UserGroupPermission

	permissions := d.Get("host_permission").([]interface{})
	for i := range permissions {
		permission := permissions[i].(map[string]interface{})
		permissionsRequest := zabbix.UserGroupPermission{
			ID:         permission["id"].(string),
			Permission: permission["permission"].(int),
		}

		permissionsRequests = append(permissionsRequests, permissionsRequest)
	}
	return permissionsRequests
}

func resourceTemplateGroupPermissionsV1(d *schema.ResourceData) []zabbix.UserGroupPermission {
	var permissionsRequests []zabbix.UserGroupPermission

	permissions := d.Get("template_permission").([]interface{})
	for i := range permissions {
		permission := permissions[i].(map[string]interface{})
		permissionsRequest := zabbix.UserGroupPermission{
			ID:         permission["id"].(string),
			Permission: permission["permission"].(int),
		}

		permissionsRequests = append(permissionsRequests, permissionsRequest)
	}
	return permissionsRequests
}

// dataUserGroup terraform data handler
func dataUserGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataUserGroupRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Description:  "Name of the user group.",
				Required:     true,
			},
		},
	}
}

// terraform usergroup create function
func resourceUserGroupCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)

	item := zabbix.UserGroup{
		Name:                     d.Get("name").(string),
		DebugMode:                d.Get("debug_mode").(int),
		GUIAccess:                d.Get("gui_access").(int),
		Status:                   d.Get("status").(int),
		Permissions:              resourceHostGroupPermissionsV1(d),
		TemplateGroupPermissions: resourceTemplateGroupPermissionsV1(d),
	}

	items := []zabbix.UserGroup{item}

	err := api.UserGroupsCreate(items)

	if err != nil {
		return err
	}

	log.Trace("created UserGroup: %s (id: %s)", items[0].Name, items[0].UserGroupID)

	d.SetId(items[0].UserGroupID)

	return resourceUserGroupRead(d, m)
}

// userGroupRead terraform user group read function
func userGroupRead(d *schema.ResourceData, m interface{}, params zabbix.Params) error {
	api := m.(*zabbix.API)

	UserGroups, err := api.UserGroupsGet(params)

	if err != nil {
		return err
	}

	if len(UserGroups) < 1 {
		d.SetId("")
		return nil
	}
	if len(UserGroups) > 1 {
		return errors.New("multiple UserGroups found")
	}
	t := UserGroups[0]

	log.Debug("Got UserGroup: %s (id: %s)", t.Name, t.UserGroupID)

	d.SetId(t.UserGroupID)
	d.Set("name", t.Name)
	d.Set("debug_mode", t.DebugMode)
	d.Set("gui_access", t.GUIAccess)
	d.Set("status", t.Status)

	return nil
}

// dataUserGroupRead terraform data resource read handler
func dataUserGroupRead(d *schema.ResourceData, m interface{}) error {
	return userGroupRead(d, m, zabbix.Params{
		"filter": map[string]interface{}{
			"name": d.Get("name"),
		},
	})
}

// resourceUserGroupRead terraform resource read handler
func resourceUserGroupRead(d *schema.ResourceData, m interface{}) error {
	log.Debug("Lookup of UserGroup with id %s", d.Id())

	return userGroupRead(d, m, zabbix.Params{
		"usrgrpids": d.Id(),
	})
}

// resourceUserGroupUpdate terraform resource update handler
func resourceUserGroupUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)

	item := zabbix.UserGroup{
		UserGroupID:              d.Id(),
		Name:                     d.Get("name").(string),
		DebugMode:                d.Get("debug_mode").(int),
		GUIAccess:                d.Get("gui_access").(int),
		Status:                   d.Get("status").(int),
		Permissions:              resourceHostGroupPermissionsV1(d),
		TemplateGroupPermissions: resourceTemplateGroupPermissionsV1(d),
	}

	items := []zabbix.UserGroup{item}

	err := api.UserGroupsUpdate(items)

	if err != nil {
		return err
	}

	return resourceUserGroupRead(d, m)
}

// resourceUserGroupDelete terraform resource delete handler
func resourceUserGroupDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)
	return api.UserGroupsDeleteByIds([]string{d.Id()})
}
