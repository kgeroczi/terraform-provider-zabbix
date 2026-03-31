package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/kgeroczi/go-zabbix-api"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

// Provider definition
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:         schema.TypeString,
				Required:     false,
				Optional:     true,
				Description:  "Zabbix API username",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"ZABBIX_USER", "ZABBIX_USERNAME"}, nil),
			},
			"password": &schema.Schema{
				Type:         schema.TypeString,
				Required:     false,
				Optional:     true,
				Sensitive:    true,
				Description:  "Zabbix API password",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"ZABBIX_PASS", "ZABBIX_PASSWORD"}, nil),
			},
			"token": &schema.Schema{
				Type:         schema.TypeString,
				Required:     false,
				Optional:     true,
				Sensitive:    true,
				Description:  "Zabbix API token",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"ZABBIX_TOKEN", "ZABBIX_API_TOKEN"}, nil),
			},
			"url": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Zabbix API url",
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"ZABBIX_URL", "ZABBIX_SERVER_URL"}, nil),
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
			"tls_insecure": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Disable TLS certificate checking (for testing use only)",
				Optional:    true,
				Default:     false,
			},
			"serialize": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Serialize API requests, if required due to API race conditions",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"zabbix_host":           dataHost(),
			"zabbix_proxy":          dataProxy(),
			"zabbix_hostgroup":      dataHostgroup(),
			"zabbix_template_group": dataTemplateGroup(),
			"zabbix_template":       dataTemplate(),
			"zabbix_user":           dataUser(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"zabbix_trigger":       resourceTrigger(),
			"zabbix_proto_trigger": resourceProtoTrigger(),
			"zabbix_template":      resourceTemplate(),
			"zabbix_hostgroup":     resourceHostgroup(),
			"zabbix_host":           resourceHost(),
			"zabbix_template_group": resourceTemplateGroup(),

			"zabbix_graph":       resourceGraph(),
			"zabbix_proto_graph": resourceProtoGraph(),

			"zabbix_item_trapper":       resourceItemTrapper(),
			"zabbix_proto_item_trapper": resourceProtoItemTrapper(),
			"zabbix_lld_trapper":        resourceLLDTrapper(),

			"zabbix_item_http":       resourceItemHttp(),
			"zabbix_proto_item_http": resourceProtoItemHttp(),
			"zabbix_lld_http":        resourceLLDHttp(),

			"zabbix_item_simple":       resourceItemSimple(),
			"zabbix_proto_item_simple": resourceProtoItemSimple(),
			"zabbix_lld_simple":        resourceLLDSimple(),

			"zabbix_item_external":       resourceItemExternal(),
			"zabbix_proto_item_external": resourceProtoItemExternal(),
			"zabbix_lld_external":        resourceLLDExternal(),

			"zabbix_item_internal":       resourceItemInternal(),
			"zabbix_proto_item_internal": resourceProtoItemInternal(),
			"zabbix_lld_internal":        resourceLLDInternal(),

			"zabbix_item_snmp":       resourceItemSnmp(),
			"zabbix_proto_item_snmp": resourceProtoItemSnmp(),
			"zabbix_lld_snmp":        resourceLLDSnmp(),

			"zabbix_item_snmptrap":       resourceItemSnmpTrap(),
			"zabbix_proto_item_snmptrap": resourceProtoItemSnmpTrap(),

			"zabbix_item_agent":       resourceItemAgent(),
			"zabbix_proto_item_agent": resourceProtoItemAgent(),
			"zabbix_lld_agent":        resourceLLDAgent(),

			"zabbix_item_calculated":       resourceItemCalculated(),
			"zabbix_proto_item_calculated": resourceProtoItemCalculated(),

			"zabbix_item_dependent":       resourceItemDependent(),
			"zabbix_proto_item_dependent": resourceProtoItemDependent(),
			"zabbix_lld_dependent":        resourceLLDDependent(),

			"zabbix_user":       resourceUser(),
			"zabbix_user_group": resourceUserGroup(),

			"zabbix_proxy": resourceProxy(),
		},
		ConfigureFunc: providerConfigure,
	}
}

// providerConfigure configure this provider
func providerConfigure(d *schema.ResourceData) (meta interface{}, err error) {
	log.Trace("Started zabbix provider init")

	api, apierr := zabbix.NewAPI(zabbix.Config{
		Url:         d.Get("url").(string),
		TlsNoVerify: d.Get("tls_insecure").(bool),
		Serialize:   d.Get("serialize").(bool),
	})
	if apierr != nil {
		return nil, apierr
	}

	if d.Get("token").(string) != "" {
		_, err = api.Token(d.Get("token").(string))
	} else {
		_, err = api.Login(d.Get("username").(string), d.Get("password").(string))
	}
	meta = api
	log.Trace("Started zabbix provider got error: %+v", err)

	return
}

// tagGenerate build tag structs from terraform inputs
func tagGenerate(d *schema.ResourceData) (tags zabbix.Tags) {
	set := d.Get("tag").(*schema.Set).List()
	tags = make(zabbix.Tags, len(set))

	for i := 0; i < len(set); i++ {
		current := set[i].(map[string]interface{})
		tags[i] = zabbix.Tag{
			Tag:   current["key"].(string),
			Value: current["value"].(string),
		}
	}

	return
}

// flattenTags convert response to terraform input
func flattenTags(list zabbix.Tags) *schema.Set {
	set := schema.NewSet(func(i interface{}) int {
		m := i.(map[string]interface{})
		return hashcode.String(m["key"].(string) + "V" + m["value"].(string))
	}, []interface{}{})
	for i := 0; i < len(list); i++ {
		set.Add(map[string]interface{}{
			"key":   list[i].Tag,
			"value": list[i].Value,
		})
	}
	return set
}
