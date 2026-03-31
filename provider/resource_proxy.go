package provider

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/kgeroczi/go-zabbix-api"
)

// resourceProxy terraform resource handler
func resourceProxy() *schema.Resource {
	return &schema.Resource{
		Create: resourceProxyCreate,
		Read:   resourceProxyRead,
		Update: resourceProxyUpdate,
		Delete: resourceProxyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Name of the proxy.",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Required:     true,
			},
			"operating_mode": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Type of proxy. Possible values: 0 - active proxy; 1 - passive proxy.",
				ValidateFunc: validation.IntBetween(0, 1),
				Required:     true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Description of the proxy.",
				//ValidateFunc: validation.StringIsNotWhiteSpace,
				Optional: true,
			},
			"tls_connect": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Connections to host.	Possible values: 1 - (default) No encryption; 2 - PSK; 4 - certificate.",
				ValidateFunc: validation.IntBetween(1, 4),
				Optional:     true,
				Default:      1,
			},
			"tls_accept": &schema.Schema{
				Type:         schema.TypeInt,
				Description:  "Connections from host. This is a bitmask field, any combination of possible bitmap values is acceptable. Possible bitmap values: 1 - (default) No encryption; 2 - PSK; 4 - certificate.",
				ValidateFunc: validation.IntBetween(1, 7),
				Optional:     true,
				Default:      1,
			},
			"tls_issuer": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Certificate issuer.",
				//ValidateFunc: validation.StringIsNotWhiteSpace,
				Optional: true,
			},
			"tls_subject": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Certificate subject.",
				//ValidateFunc: validation.StringIsNotWhiteSpace,
				Optional: true,
			},
			"tls_psk_identity": &schema.Schema{
				Type:        schema.TypeString,
				Description: "PSK identity. Do not put sensitive information in the PSK identity, it is transmitted unencrypted over the network to inform a receiver which PSK to use. Required if tls_connect is set to \"PSK\", or tls_accept contains the \"PSK\" bit.",
				//ValidateFunc: validation.StringIsNotWhiteSpace,
				Optional: true,
			},
			"tls_psk": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The preshared key, at least 32 hex digits. Required if tls_connect is set to \"PSK\", or tls_accept contains the \"PSK\" bit.",
				//ValidateFunc: validation.StringIsNotWhiteSpace,
				Optional:  true,
				Sensitive: true,
			},
			"allowed_addresses": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Comma-delimited IP addresses or DNS names of active Zabbix proxy.",
				Optional:    true,
			},
			"proxy_address": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Deprecated: Use allowed_addresses instead.",
				Deprecated:  "Use allowed_addresses instead",
				Optional:    true,
			},
			"address": &schema.Schema{
				Type:        schema.TypeString,
				Description: "IP address or DNS name to connect to for passive proxies.",
				Optional:    true,
			},
			"port": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Port number to connect to for passive proxies. Default: 10051.",
				Optional:    true,
			},
		},
	}
}

// dataProxy terraform data handler
func dataProxy() *schema.Resource {
	return &schema.Resource{
		Read: dataProxyRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "Name of the proxy.",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Required:     true,
			},
		},
	}
}

// dataProxyRead read handler for data resource
func dataProxyRead(d *schema.ResourceData, m interface{}) error {
	params := zabbix.Params{
		//"selectInterface": "extend",
		"filter": map[string]interface{}{},
	}

	lookups := []string{"name"}
	for _, k := range lookups {
		if v, ok := d.GetOk(k); ok {
			params["filter"].(map[string]interface{})[k] = v
		}
	}

	if len(params["filter"].(map[string]interface{})) < 1 {
		return errors.New("no proxy lookup attribute")
	}
	log.Debug("performing proxy data lookup")

	return proxyRead(d, m, params)
}

// terraform proxy create function
func resourceProxyCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)

	proxy := zabbix.Proxy{
		ProxyID:          d.Id(),
		Name:             d.Get("name").(string),
		OperatingMode:    d.Get("operating_mode").(int),
		Description:      d.Get("description").(string),
		TLSConnect:       d.Get("tls_connect").(int),
		TLSAccept:        d.Get("tls_accept").(int),
		TLSIssuer:        d.Get("tls_issuer").(string),
		TLSSubject:       d.Get("tls_subject").(string),
		TLSPSKIdentity:   d.Get("tls_psk_identity").(string),
		TLSPSK:           d.Get("tls_psk").(string),
		AllowedAddresses: proxyGetAllowedAddresses(d),
		Address:          d.Get("address").(string),
		Port:             d.Get("port").(string),
	}

	proxies := []zabbix.Proxy{proxy}

	err := api.ProxiesCreate(proxies)

	if err != nil {
		return err
	}

	log.Trace("created Proxy: %s (id: %s)", proxies[0].Name, proxies[0].ProxyID)

	d.SetId(proxies[0].ProxyID)

	return resourceProxyRead(d, m)
}

// proxyRead common proxy read function
func proxyRead(d *schema.ResourceData, m interface{}, params zabbix.Params) error {
	api := m.(*zabbix.API)

	log.Debug("Lookup of proxy")

	proxys, err := api.ProxiesGet(params)

	if err != nil {
		return err
	}

	if len(proxys) < 1 {
		d.SetId("")
		return nil
	}
	if len(proxys) > 1 {
		return errors.New("multiple proxys found")
	}
	proxy := proxys[0]

	log.Debug("Got proxy: %s (id: %s)", proxy.Name, proxy.ProxyID)

	d.SetId(proxy.ProxyID)
	d.Set("name", proxy.Name)
	d.Set("operating_mode", proxy.OperatingMode)
	d.Set("description", proxy.Description)
	d.Set("tls_connect", proxy.TLSConnect)
	d.Set("tls_accept", proxy.TLSAccept)
	d.Set("tls_issuer", proxy.TLSIssuer)
	d.Set("tls_subject", proxy.TLSSubject)
	d.Set("tls_psk_identity", proxy.TLSPSKIdentity)
	d.Set("tls_psk", proxy.TLSPSK)
	d.Set("allowed_addresses", proxy.AllowedAddresses)
	d.Set("address", proxy.Address)
	d.Set("port", proxy.Port)

	return nil
}

// proxyGetAllowedAddresses returns allowed_addresses, falling back to deprecated proxy_address
func proxyGetAllowedAddresses(d *schema.ResourceData) string {
	if v, ok := d.GetOk("allowed_addresses"); ok {
		return v.(string)
	}
	return d.Get("proxy_address").(string)
}

// resourceProxyRead terraform resource read handler
func resourceProxyRead(d *schema.ResourceData, m interface{}) error {
	log.Debug("Lookup of Proxy with id %s", d.Id())

	return proxyRead(d, m, zabbix.Params{
		"proxyids": d.Id(),
	})
}

// resourceProxyUpdate terraform resource update handler
func resourceProxyUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)

	proxy := zabbix.Proxy{
		ProxyID:          d.Id(),
		Name:             d.Get("name").(string),
		OperatingMode:    d.Get("operating_mode").(int),
		Description:      d.Get("description").(string),
		TLSConnect:       d.Get("tls_connect").(int),
		TLSAccept:        d.Get("tls_accept").(int),
		TLSIssuer:        d.Get("tls_issuer").(string),
		TLSSubject:       d.Get("tls_subject").(string),
		TLSPSKIdentity:   d.Get("tls_psk_identity").(string),
		TLSPSK:           d.Get("tls_psk").(string),
		AllowedAddresses: proxyGetAllowedAddresses(d),
		Address:          d.Get("address").(string),
		Port:             d.Get("port").(string),
	}

	proxies := []zabbix.Proxy{proxy}

	err := api.ProxiesUpdate(proxies)

	if err != nil {
		return err
	}

	return resourceProxyRead(d, m)
}

// resourceProxyDelete terraform resource delete handler
func resourceProxyDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)
	return api.ProxiesDeleteByIds([]string{d.Id()})
}
