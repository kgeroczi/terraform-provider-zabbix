package provider

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/kgeroczi/go-zabbix-api"
)

// resourceSLA terraform resource handler
func resourceSLA() *schema.Resource {
	return &schema.Resource{
		Create: resourceSLACreate,
		Read:   resourceSLARead,
		Update: resourceSLAUpdate,
		Delete: resourceSLADelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Name of the SLA.",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Required:     true,
			},
			"period": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "SLA reporting period. 0=daily, 1=weekly, 2=monthly, 3=quarterly, 4=annually.",
				ValidateFunc: validation.IntBetween(0, 4),
				Required:     true,
			},
			"slo": &schema.Schema{
				Type:         schema.TypeFloat,
				Description:  "Service level objective (percent).",
				ValidateFunc: validation.FloatBetween(0.0, 100.0),
				Required:     true,
			},
			"effective_date": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Effective date as Unix timestamp.",
				Optional:    true,
			},
			"timezone": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Timezone for SLA calculations.",
				Optional:    true,
			},
			"status": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "SLA status. 0=enabled, 1=disabled.",
				ValidateFunc: validation.IntBetween(0, 1),
				Optional:     true,
				Default:      0,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "SLA description.",
				Optional:    true,
			},
			"service_tag": &schema.Schema{
				Type:        schema.TypeList,
				Description: "Service tag filters for this SLA.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tag": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringIsNotWhiteSpace,
							Required:     true,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"operator": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(0, 1),
							Optional:     true,
							Default:      0,
						},
					},
				},
			},
			"schedule": &schema.Schema{
				Type:        schema.TypeList,
				Description: "SLA schedule windows as seconds from week start.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"period_from": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(0, 604800),
							Required:     true,
						},
						"period_to": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(0, 604800),
							Required:     true,
						},
					},
				},
			},
			"excluded_downtime": &schema.Schema{
				Type:        schema.TypeList,
				Description: "Excluded downtime windows as Unix timestamps.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringIsNotWhiteSpace,
							Required:     true,
						},
						"period_from": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"period_to": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func buildSLAServiceTags(d *schema.ResourceData) []zabbix.SLAServiceTag {
	rawTags := d.Get("service_tag").([]interface{})
	tags := make([]zabbix.SLAServiceTag, len(rawTags))
	for i, raw := range rawTags {
		tag := raw.(map[string]interface{})
		tags[i] = zabbix.SLAServiceTag{
			Tag:      tag["tag"].(string),
			Value:    tag["value"].(string),
			Operator: zabbix.SLATagOperatorType(tag["operator"].(int)),
		}
	}
	return tags
}

func flattenSLAServiceTags(tags []zabbix.SLAServiceTag) []interface{} {
	val := make([]interface{}, len(tags))
	for i, tag := range tags {
		val[i] = map[string]interface{}{
			"tag":      tag.Tag,
			"value":    tag.Value,
			"operator": int(tag.Operator),
		}
	}
	return val
}

func buildSLASchedule(d *schema.ResourceData) []zabbix.SLASchedule {
	rawSchedule := d.Get("schedule").([]interface{})
	schedule := make([]zabbix.SLASchedule, len(rawSchedule))
	for i, raw := range rawSchedule {
		window := raw.(map[string]interface{})
		schedule[i] = zabbix.SLASchedule{
			PeriodFrom: window["period_from"].(int),
			PeriodTo:   window["period_to"].(int),
		}
	}
	return schedule
}

func flattenSLASchedule(schedule []zabbix.SLASchedule) []interface{} {
	val := make([]interface{}, len(schedule))
	for i, window := range schedule {
		val[i] = map[string]interface{}{
			"period_from": window.PeriodFrom,
			"period_to":   window.PeriodTo,
		}
	}
	return val
}

func buildSLAExcludedDowntimes(d *schema.ResourceData) []zabbix.SLAExcludedDowntime {
	rawDowntimes := d.Get("excluded_downtime").([]interface{})
	downtimes := make([]zabbix.SLAExcludedDowntime, len(rawDowntimes))
	for i, raw := range rawDowntimes {
		entry := raw.(map[string]interface{})
		downtimes[i] = zabbix.SLAExcludedDowntime{
			Name:       entry["name"].(string),
			PeriodFrom: int64(entry["period_from"].(int)),
			PeriodTo:   int64(entry["period_to"].(int)),
		}
	}
	return downtimes
}

func flattenSLAExcludedDowntimes(downtimes []zabbix.SLAExcludedDowntime) []interface{} {
	val := make([]interface{}, len(downtimes))
	for i, entry := range downtimes {
		val[i] = map[string]interface{}{
			"name":        entry.Name,
			"period_from": int(entry.PeriodFrom),
			"period_to":   int(entry.PeriodTo),
		}
	}
	return val
}

func buildSLAObject(d *schema.ResourceData) zabbix.SLA {
	return zabbix.SLA{
		SLAID:             d.Id(),
		Name:              d.Get("name").(string),
		Period:            zabbix.SLAPeriodType(d.Get("period").(int)),
		SLO:               d.Get("slo").(float64),
		EffectiveDate:     int64(d.Get("effective_date").(int)),
		Timezone:          d.Get("timezone").(string),
		Status:            zabbix.SLAStatusType(d.Get("status").(int)),
		Description:       d.Get("description").(string),
		ServiceTags:       buildSLAServiceTags(d),
		Schedule:          buildSLASchedule(d),
		ExcludedDowntimes: buildSLAExcludedDowntimes(d),
	}
}

// terraform SLA create function
func resourceSLACreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)

	item := buildSLAObject(d)
	item.SLAID = ""
	items := []zabbix.SLA{item}

	err := api.SLAsCreate(items)
	if err != nil {
		return err
	}

	log.Trace("created SLA: %s (id: %s)", items[0].Name, items[0].SLAID)

	d.SetId(items[0].SLAID)

	return resourceSLARead(d, m)
}

// slaRead generic SLA read function
func slaRead(d *schema.ResourceData, m interface{}, params zabbix.Params) error {
	api := m.(*zabbix.API)

	slas, err := api.SLAsGet(params)
	if err != nil {
		return err
	}

	if len(slas) < 1 {
		d.SetId("")
		return nil
	}
	if len(slas) > 1 {
		return errors.New("multiple SLAs found")
	}
	t := slas[0]

	log.Debug("Got SLA: %s (id: %s)", t.Name, t.SLAID)

	d.SetId(t.SLAID)
	d.Set("name", t.Name)
	d.Set("period", int(t.Period))
	d.Set("slo", t.SLO)
	d.Set("effective_date", int(t.EffectiveDate))
	d.Set("timezone", t.Timezone)
	d.Set("status", int(t.Status))
	d.Set("description", t.Description)
	d.Set("service_tag", flattenSLAServiceTags(t.ServiceTags))
	d.Set("schedule", flattenSLASchedule(t.Schedule))
	d.Set("excluded_downtime", flattenSLAExcludedDowntimes(t.ExcludedDowntimes))

	return nil
}

// resourceSLARead terraform resource read handler
func resourceSLARead(d *schema.ResourceData, m interface{}) error {
	log.Debug("Lookup of SLA with id %s", d.Id())

	return slaRead(d, m, zabbix.Params{
		"slaids":                  d.Id(),
		"selectServiceTags":       "extend",
		"selectSchedule":          "extend",
		"selectExcludedDowntimes": "extend",
	})
}

// resourceSLAUpdate terraform resource update handler
func resourceSLAUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)

	item := buildSLAObject(d)
	items := []zabbix.SLA{item}

	err := api.SLAsUpdate(items)
	if err != nil {
		return err
	}

	return resourceSLARead(d, m)
}

// resourceSLADelete terraform resource delete handler
func resourceSLADelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)
	return api.SLAsDeleteByIds([]string{d.Id()})
}
