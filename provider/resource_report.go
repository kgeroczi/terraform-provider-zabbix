package provider

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/kgeroczi/go-zabbix-api"
)

// resourceReport terraform resource handler
func resourceReport() *schema.Resource {
	return &schema.Resource{
		Create: resourceReportCreate,
		Read:   resourceReportRead,
		Update: resourceReportUpdate,
		Delete: resourceReportDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"userid": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Owner user ID of the report.",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Required:     true,
			},
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Name of the report.",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Required:     true,
			},
			"dashboardid": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Dashboard ID used for report content.",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Required:     true,
			},
			"period": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Reporting period. 0=previous day, 1=previous week, 2=previous month, 3=previous year.",
				ValidateFunc: validation.IntBetween(0, 3),
				Required:     true,
			},
			"cycle": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Scheduling cycle. 0=daily, 1=weekly, 2=monthly, 3=yearly.",
				ValidateFunc: validation.IntBetween(0, 3),
				Required:     true,
			},
			"start_time": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Report start time value accepted by Zabbix API.",
				Optional:    true,
				Default:     0,
			},
			"weekdays": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Bitmask for weekdays (Mon=1, Tue=2, Wed=4, Thu=8, Fri=16, Sat=32, Sun=64).",
				ValidateFunc: validation.IntBetween(0, 127),
				Optional:     true,
				Default:      0,
			},
			"active_since": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Active since timestamp string accepted by Zabbix.",
				Optional:    true,
			},
			"active_till": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Active till timestamp string accepted by Zabbix.",
				Optional:    true,
			},
			"subject": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Email subject.",
				Optional:    true,
			},
			"message": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Email message body.",
				Optional:    true,
			},
			"status": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Report status. 0=disabled, 1=enabled.",
				ValidateFunc: validation.IntBetween(0, 1),
				Optional:     true,
				Default:      1,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Description of the report.",
				Optional:    true,
			},
			"user": &schema.Schema{
				Type:        schema.TypeList,
				Description: "User recipients.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"userid": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringIsNotWhiteSpace,
							Required:     true,
						},
						"access_userid": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"exclude": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(0, 1),
							Optional:     true,
							Default:      0,
						},
					},
				},
			},
			"user_group": &schema.Schema{
				Type:        schema.TypeList,
				Description: "User group recipients.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"usrgrpid": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringIsNotWhiteSpace,
							Required:     true,
						},
						"access_userid": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"state": {
				Type:        schema.TypeInt,
				Description: "Current delivery state (read-only).",
				Computed:    true,
			},
			"lastsent": {
				Type:        schema.TypeInt,
				Description: "Unix timestamp of last successful send (read-only).",
				Computed:    true,
			},
			"info": {
				Type:        schema.TypeString,
				Description: "Information about report delivery state (read-only).",
				Computed:    true,
			},
		},
	}
}

func buildReportUsers(d *schema.ResourceData) zabbix.ReportUsers {
	rawUsers := d.Get("user").([]interface{})
	users := make(zabbix.ReportUsers, len(rawUsers))
	for i, raw := range rawUsers {
		user := raw.(map[string]interface{})
		users[i] = zabbix.ReportUser{
			UserID:       user["userid"].(string),
			AccessUserID: user["access_userid"].(string),
			Exclude:      user["exclude"].(int),
		}
	}
	return users
}

func flattenReportUsers(users zabbix.ReportUsers) []interface{} {
	val := make([]interface{}, len(users))
	for i, user := range users {
		val[i] = map[string]interface{}{
			"userid":        user.UserID,
			"access_userid": user.AccessUserID,
			"exclude":       user.Exclude,
		}
	}
	return val
}

func buildReportUserGroups(d *schema.ResourceData) zabbix.ReportUserGroups {
	rawGroups := d.Get("user_group").([]interface{})
	groups := make(zabbix.ReportUserGroups, len(rawGroups))
	for i, raw := range rawGroups {
		group := raw.(map[string]interface{})
		groups[i] = zabbix.ReportUserGroup{
			UserGroupID:  group["usrgrpid"].(string),
			AccessUserID: group["access_userid"].(string),
		}
	}
	return groups
}

func flattenReportUserGroups(groups zabbix.ReportUserGroups) []interface{} {
	val := make([]interface{}, len(groups))
	for i, group := range groups {
		val[i] = map[string]interface{}{
			"usrgrpid":      group.UserGroupID,
			"access_userid": group.AccessUserID,
		}
	}
	return val
}

func buildReportObject(d *schema.ResourceData) zabbix.Report {
	return zabbix.Report{
		ReportID:    d.Id(),
		UserID:      d.Get("userid").(string),
		Name:        d.Get("name").(string),
		DashboardID: d.Get("dashboardid").(string),
		Period:      zabbix.ReportPeriodType(d.Get("period").(int)),
		Cycle:       zabbix.ReportCycleType(d.Get("cycle").(int)),
		StartTime:   d.Get("start_time").(int),
		Weekdays:    d.Get("weekdays").(int),
		ActiveSince: d.Get("active_since").(string),
		ActiveTill:  d.Get("active_till").(string),
		Subject:     d.Get("subject").(string),
		Message:     d.Get("message").(string),
		Status:      zabbix.ReportStatusType(d.Get("status").(int)),
		Description: d.Get("description").(string),
		Users:       buildReportUsers(d),
		UserGroups:  buildReportUserGroups(d),
	}
}

// terraform report create function
func resourceReportCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)

	item := buildReportObject(d)
	item.ReportID = ""
	items := []zabbix.Report{item}

	err := api.ReportsCreate(items)
	if err != nil {
		return err
	}

	log.Trace("created Report: %s (id: %s)", items[0].Name, items[0].ReportID)

	d.SetId(items[0].ReportID)

	return resourceReportRead(d, m)
}

// reportRead generic report read function
func reportRead(d *schema.ResourceData, m interface{}, params zabbix.Params) error {
	api := m.(*zabbix.API)

	reports, err := api.ReportsGet(params)
	if err != nil {
		return err
	}

	if len(reports) < 1 {
		d.SetId("")
		return nil
	}
	if len(reports) > 1 {
		return errors.New("multiple reports found")
	}
	t := reports[0]

	log.Debug("Got Report: %s (id: %s)", t.Name, t.ReportID)

	d.SetId(t.ReportID)
	d.Set("userid", t.UserID)
	d.Set("name", t.Name)
	d.Set("dashboardid", t.DashboardID)
	d.Set("period", int(t.Period))
	d.Set("cycle", int(t.Cycle))
	d.Set("start_time", t.StartTime)
	d.Set("weekdays", t.Weekdays)
	d.Set("active_since", t.ActiveSince)
	d.Set("active_till", t.ActiveTill)
	d.Set("subject", t.Subject)
	d.Set("message", t.Message)
	d.Set("status", int(t.Status))
	d.Set("description", t.Description)
	d.Set("user", flattenReportUsers(t.Users))
	d.Set("user_group", flattenReportUserGroups(t.UserGroups))
	d.Set("state", int(t.State))
	d.Set("lastsent", int(t.LastSent))
	d.Set("info", t.Info)

	return nil
}

// resourceReportRead terraform resource read handler
func resourceReportRead(d *schema.ResourceData, m interface{}) error {
	log.Debug("Lookup of Report with id %s", d.Id())

	return reportRead(d, m, zabbix.Params{
		"reportids":        d.Id(),
		"selectUsers":      "extend",
		"selectUserGroups": "extend",
	})
}

// resourceReportUpdate terraform resource update handler
func resourceReportUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)

	item := buildReportObject(d)
	items := []zabbix.Report{item}

	err := api.ReportsUpdate(items)
	if err != nil {
		return err
	}

	return resourceReportRead(d, m)
}

// resourceReportDelete terraform resource delete handler
func resourceReportDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)
	return api.ReportsDeleteByIds([]string{d.Id()})
}
