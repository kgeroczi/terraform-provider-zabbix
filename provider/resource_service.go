package provider

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/kgeroczi/go-zabbix-api"
)

// resourceService terraform resource handler
func resourceService() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceCreate,
		Read:   resourceServiceRead,
		Update: resourceServiceUpdate,
		Delete: resourceServiceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Name of the service.",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Required:     true,
			},
			"algorithm": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Service status calculation algorithm. 0=none, 1=worst child, 2=best child.",
				ValidateFunc: validation.IntBetween(0, 2),
				Optional:     true,
				Default:      0,
			},
			"sortorder": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Service sort order among siblings.",
				Optional:    true,
				Default:     0,
			},
			"weight": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Service weight used by status rule calculations.",
				ValidateFunc: validation.IntAtLeast(0),
				Optional:     true,
				Default:      0,
			},
			"propagation_rule": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Propagation rule to parent service. 0=as-is, 1=increase, 2=decrease, 3=ignore, 4=fixed.",
				ValidateFunc: validation.IntBetween(0, 4),
				Optional:     true,
				Default:      0,
			},
			"propagation_value": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Propagation value used with propagation_rule.",
				ValidateFunc: validation.IntAtLeast(0),
				Optional:     true,
				Default:      0,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Service description.",
				Optional:    true,
			},
			"tag": &schema.Schema{
				Type:        schema.TypeList,
				Description: "Service tags.",
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
					},
				},
			},
			"problem_tag": &schema.Schema{
				Type:        schema.TypeList,
				Description: "Problem tag filters for the service.",
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
			"status_rule": &schema.Schema{
				Type:        schema.TypeList,
				Description: "Service status rules.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(0, 5),
							Required:     true,
						},
						"limit_n": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntAtLeast(0),
							Required:     true,
						},
						"limit_s": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntAtLeast(0),
							Required:     true,
						},
						"new_status": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntAtLeast(0),
							Required:     true,
						},
					},
				},
			},
			"parents": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "Parent service IDs.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"children": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "Child service IDs.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"status": {
				Type:        schema.TypeInt,
				Description: "Current service status as calculated by Zabbix.",
				Computed:    true,
			},
			"uuid": {
				Type:        schema.TypeString,
				Description: "Service UUID (read-only).",
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeInt,
				Description: "Creation time as Unix timestamp (read-only).",
				Computed:    true,
			},
		},
	}
}

func buildServiceTags(d *schema.ResourceData) []zabbix.ServiceTag {
	rawTags := d.Get("tag").([]interface{})
	tags := make([]zabbix.ServiceTag, len(rawTags))
	for i, raw := range rawTags {
		tag := raw.(map[string]interface{})
		tags[i] = zabbix.ServiceTag{
			Tag:   tag["tag"].(string),
			Value: tag["value"].(string),
		}
	}
	return tags
}

func flattenServiceTags(tags []zabbix.ServiceTag) []interface{} {
	val := make([]interface{}, len(tags))
	for i, tag := range tags {
		val[i] = map[string]interface{}{
			"tag":   tag.Tag,
			"value": tag.Value,
		}
	}
	return val
}

func buildServiceProblemTags(d *schema.ResourceData) []zabbix.ServiceProblemTag {
	rawTags := d.Get("problem_tag").([]interface{})
	tags := make([]zabbix.ServiceProblemTag, len(rawTags))
	for i, raw := range rawTags {
		tag := raw.(map[string]interface{})
		tags[i] = zabbix.ServiceProblemTag{
			Tag:      tag["tag"].(string),
			Value:    tag["value"].(string),
			Operator: zabbix.ServiceTagOperatorType(tag["operator"].(int)),
		}
	}
	return tags
}

func flattenServiceProblemTags(tags []zabbix.ServiceProblemTag) []interface{} {
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

func buildServiceStatusRules(d *schema.ResourceData) []zabbix.ServiceStatusRule {
	rawRules := d.Get("status_rule").([]interface{})
	rules := make([]zabbix.ServiceStatusRule, len(rawRules))
	for i, raw := range rawRules {
		rule := raw.(map[string]interface{})
		rules[i] = zabbix.ServiceStatusRule{
			Type:      zabbix.ServiceStatusRuleType(rule["type"].(int)),
			Limit_n:   rule["limit_n"].(int),
			LimitS:    rule["limit_s"].(int),
			NewStatus: rule["new_status"].(int),
		}
	}
	return rules
}

func flattenServiceStatusRules(rules []zabbix.ServiceStatusRule) []interface{} {
	val := make([]interface{}, len(rules))
	for i, rule := range rules {
		val[i] = map[string]interface{}{
			"type":       int(rule.Type),
			"limit_n":    rule.Limit_n,
			"limit_s":    rule.LimitS,
			"new_status": rule.NewStatus,
		}
	}
	return val
}

func buildServiceIDs(rawSet *schema.Set) zabbix.ServiceIDs {
	rawIDs := rawSet.List()
	ids := make(zabbix.ServiceIDs, len(rawIDs))
	for i, raw := range rawIDs {
		ids[i] = zabbix.ServiceID{ServiceID: raw.(string)}
	}
	return ids
}

func flattenServiceIDs(ids zabbix.ServiceIDs) []string {
	val := make([]string, len(ids))
	for i, id := range ids {
		val[i] = id.ServiceID
	}
	return val
}

func buildServiceObject(d *schema.ResourceData) zabbix.Service {
	return zabbix.Service{
		ServiceID:        d.Id(),
		Name:             d.Get("name").(string),
		Algorithm:        zabbix.ServiceAlgorithmType(d.Get("algorithm").(int)),
		Sortorder:        d.Get("sortorder").(int),
		Weight:           d.Get("weight").(int),
		PropagationRule:  zabbix.ServicePropagationRuleType(d.Get("propagation_rule").(int)),
		PropagationValue: d.Get("propagation_value").(int),
		Description:      d.Get("description").(string),
		Tags:             buildServiceTags(d),
		ProblemTags:      buildServiceProblemTags(d),
		StatusRules:      buildServiceStatusRules(d),
		Parents:          buildServiceIDs(d.Get("parents").(*schema.Set)),
		Children:         buildServiceIDs(d.Get("children").(*schema.Set)),
	}
}

// terraform service create function
func resourceServiceCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)

	item := buildServiceObject(d)
	item.ServiceID = ""
	items := []zabbix.Service{item}

	err := api.ServicesCreate(items)
	if err != nil {
		return err
	}

	log.Trace("created Service: %s (id: %s)", items[0].Name, items[0].ServiceID)

	d.SetId(items[0].ServiceID)

	return resourceServiceRead(d, m)
}

// serviceRead generic service read function
func serviceRead(d *schema.ResourceData, m interface{}, params zabbix.Params) error {
	api := m.(*zabbix.API)

	services, err := api.ServicesGet(params)
	if err != nil {
		return err
	}

	if len(services) < 1 {
		d.SetId("")
		return nil
	}
	if len(services) > 1 {
		return errors.New("multiple services found")
	}
	t := services[0]

	log.Debug("Got Service: %s (id: %s)", t.Name, t.ServiceID)

	d.SetId(t.ServiceID)
	d.Set("name", t.Name)
	d.Set("algorithm", int(t.Algorithm))
	d.Set("sortorder", t.Sortorder)
	d.Set("weight", t.Weight)
	d.Set("propagation_rule", int(t.PropagationRule))
	d.Set("propagation_value", t.PropagationValue)
	d.Set("description", t.Description)
	d.Set("tag", flattenServiceTags(t.Tags))
	d.Set("problem_tag", flattenServiceProblemTags(t.ProblemTags))
	d.Set("status_rule", flattenServiceStatusRules(t.StatusRules))
	d.Set("parents", flattenServiceIDs(t.Parents))
	d.Set("children", flattenServiceIDs(t.Children))
	d.Set("status", t.Status)
	d.Set("uuid", t.UUID)
	d.Set("created_at", int(t.CreatedAt))

	return nil
}

// resourceServiceRead terraform resource read handler
func resourceServiceRead(d *schema.ResourceData, m interface{}) error {
	log.Debug("Lookup of Service with id %s", d.Id())

	return serviceRead(d, m, zabbix.Params{
		"serviceids":        d.Id(),
		"selectTags":        "extend",
		"selectProblemTags": "extend",
		"selectStatusRules": "extend",
		"selectParents":     "extend",
		"selectChildren":    "extend",
	})
}

// resourceServiceUpdate terraform resource update handler
func resourceServiceUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)

	item := buildServiceObject(d)
	items := []zabbix.Service{item}

	err := api.ServicesUpdate(items)
	if err != nil {
		return err
	}

	return resourceServiceRead(d, m)
}

// resourceServiceDelete terraform resource delete handler
func resourceServiceDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)
	return api.ServicesDeleteByIds([]string{d.Id()})
}
