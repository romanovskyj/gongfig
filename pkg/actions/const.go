package actions

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

// Resource is a representation of corresponding type object and path in Kong
type Resource struct {
	Path   string
	Struct interface{}
}

// Apis - list of apis for import/export with corresponding structure types for parsing values
// Be aware it should be in the same order as it is going to be deleted, e.g. firstly we delete
// routes and then services as route has service foreign key
var Apis = []string{RoutesPath, ServicesPath, CertificatesPath, ConsumersPath}

// ExportResourceBundles is a slice of elements with resource path and corresponding struct type
// in order to store elements in config while exporting using a loop, without duplicating a code.
// Services and routes are not here as they handled separately in export procedure.
var ExportResourceBundles  = []Resource{
	{CertificatesPath, &Certificate{}},
	{ConsumersPath, &Consumer{}},
}

// ImportResourceBundles is the same as ExportResourceBundles but for the import
var ImportResourceBundles = []Resource{
	{CertificatesPath, &Certificate{}},
	{ConsumersPath, &Consumer{}},
}

//Service struct - is used for managing services
type Service struct {
	InternalId string  `json:"internal_id,omitempty" mapstructure:"id"`
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
	InternalId string `json:"internal_id,omitempty" mapstructure:"id"`
	Paths []string `json:"paths" mapstructure:"paths"`
	Service *Service `json:"service,omitempty" mapstructure:"service"`
	StripPath bool	`json:"strip_path" mapstructure:"strip_path"`
	PreserveHost bool `json:"preserve_host" mapstructure:"preserve_host"`
	RegexPriority int `json:"regex_priority" mapstructure:"regex_priority"`
	Hosts []string `json:"hosts" mapstructure:"hosts"`
	Protocols []string `json:"protocols" mapstructure:"protocols"`
	Methods []string `json:"methods" mapstructure:"methods"`
}

// Certificate - for obtaining certificates from the server
type Certificate struct {
	Cert string `json:"cert" mapstructure:"cert"`
	Key string `json:"key" mapstructure:"key"`
	Snis []string `json:"snis" mapstructure:"snis"`
}

// Consumer - for obtaining consumers from the server
type Consumer struct {
	InternalId string `json:"internal_id" mapstructure:"id"`
	CustomId string `json:"custom_id" mapstructure:"custom_id"`
	Username string `json:"username" mapstructure:"username"`
}

// ResourceInstance can be both service or route
type ResourceInstance struct {
	Id string `mapstructure:"id"`
}
