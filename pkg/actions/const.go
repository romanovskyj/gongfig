package actions

import "net/http"

// Timeout - how long http client should wait before terminating connection
const Timeout = 10

// DefaultURL keeps url when kong api is accessed with port forwarding (as mentioned in readme)
const DefaultURL = "http://localhost:8001"

// ServicesPath has Kong admin services path
const ServicesPath = "services"

// RoutesPath has Kong admin routes path
const RoutesPath = "routes"

// CertificatesPath has Kong admin certificates path
const CertificatesPath = "certificates"

// ConsumersPath has Kong admin consumers path
const ConsumersPath = "consumers"

// PluginsPath has Kong admin plugins path
const PluginsPath = "plugins"

// Resource is a representation of corresponding type object and path in Kong
type Resource struct {
	Path   string
	Struct interface{}
}

// Apis - list of apis for import/export with corresponding structure types for parsing values
// Be aware it should be in the same order as it is going to be deleted, e.g. firstly we delete
// routes and then services as route has service foreign key
var Apis = []string{RoutesPath, ServicesPath, CertificatesPath, ConsumersPath, PluginsPath}

// ExportResourceBundles is a slice of elements with resource path and corresponding struct type
// in order to store elements in config while exporting using a loop, without duplicating a code.
// Services and routes are not here as they handled separately in export procedure.
var ExportResourceBundles  = []Resource{
	{CertificatesPath, &Certificate{}},
	{ConsumersPath, &Consumer{}},
	{PluginsPath, &Plugin{}},
}

// ImportResourceBundles is the same as ExportResourceBundles but for the import
var ImportResourceBundles = []Resource{
	{CertificatesPath, &Certificate{}},
	{ConsumersPath, &Consumer{}},
}

//Service struct - is used for managing services
type Service struct {
	Id string		   `json:"id,omitempty" mapstructure:"id"`
	Name string        `json:"name" mapstructure:"name"`
	Host string        `json:"host" mapstructure:"host"`
	Path string        `json:"path,omitempty" mapstructure:"path"`
	Port int           `json:"port" mapstructure:"port"`
	Protocol string    `json:"protocol" mapstructure:"protocol"`
	ConnectTimeout int `json:"connect_timeout" mapstructure:"connect_timeout"`
	ReadTimeout int    `json:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout int   `json:"write_timeout" mapstructure:"write_timeout"`
	Routes []Route     `json:"routes,omitempty"`
}

//Route struct - is used for managing routes
type Route struct {
	Id string 		   `json:"id,omitempty" mapstructure:"id"`
	Paths []string     `json:"paths" mapstructure:"paths"`
	Service *Service   `json:"service,omitempty" mapstructure:"service"`
	StripPath bool     `json:"strip_path" mapstructure:"strip_path"`
	PreserveHost bool  `json:"preserve_host" mapstructure:"preserve_host"`
	RegexPriority int  `json:"regex_priority" mapstructure:"regex_priority"`
	Hosts []string     `json:"hosts" mapstructure:"hosts"`
	Protocols []string `json:"protocols" mapstructure:"protocols"`
	Methods []string   `json:"methods" mapstructure:"methods"`
}

// Certificate - for obtaining certificates from the server
type Certificate struct {
	Id string     `json:"id,omitempty" mapstructure:"id"`
	Cert string   `json:"cert" mapstructure:"cert"`
	Key string    `json:"key" mapstructure:"key"`
	Snis []string `json:"snis" mapstructure:"snis"`
}

// Consumer - for obtaining consumers from the server
type Consumer struct {
	Id string         `json:"id,omitempty" mapstructure:"id"`
	CustomId string   `json:"custom_id,omitempty" mapstructure:"custom_id"`
	Username string   `json:"username,omitempty" mapstructure:"username"`
}

//Plugin struct - is used for managing plugins
type Plugin struct {
	Id string   			      `json:"id,omitempty" mapstructure:"id"`
	Name string 				  `json:"name" mapstructure:"name"`
	Config map[string]interface{} `json:"config" mapstructure:"config"`
	Enabled bool                  `json:"enabled" mapstructure:"enabled"`
	ServiceId string              `json:"service_id,omitempty" mapstructure:"service_id"`
	RouteId string                `json:"route_id,omitempty" mapstructure:"route_id"`
	ConsumerId string             `json:"consumer_id,omitempty" mapstructure:"consumer_id"`
}

// ResourceInstance can be both service or route
type ResourceInstance struct {
	Id string `mapstructure:"id"`
}

// LocalResource is needed for obtaining local id of resource during import
// It needs later to match newly created resources and local that we keep in config
// in order to keep relations between them (e.g. plugin has service_id etc)
type LocalResource struct {
	Id string `mapstructure:"id"`
}

type ConnectionBundle struct {
	Client *http.Client
	Url string
	ReqLimitChan chan bool
}

type Message struct {
	Message string
}