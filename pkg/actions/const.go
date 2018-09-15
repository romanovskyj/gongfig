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

// Resource is a representation of corresponding type object and path in Kong
type Resource struct {
	Path   string
	Struct interface{}
}

// Apis - list of apis for import/export with corresponding structure types for parsing values
// Be aware it should be in the same order as it is going to be deleted, e.g. firstly we delete
// routes and then services as route has service foreign key
var Apis = []string{RoutesPath, ServicesPath, CertificatesPath}

// ResourceBundles is a slice of elements with resource path and corresponding struct type
// in order to store elements in config while exporting using a loop, without duplicating a code.
// Services and routes are not here as they handled separately in export procedure.
var ResourceBundles  = []Resource{
	{CertificatesPath, &CertificatePrepared{}},
}


// Service - for obtaining services from the server
// for importing configuration every time so name is enough for identifying it
type Service struct {
	Id string `mapstructure:"id"`
	Name string `mapstructure:"name"`
	Host string `mapstructure:"host"`
	Path string `mapstructure:"path"`
	Port int `mapstructure:"port"`
	Protocol string `mapstructure:"protocol"`
	ConnectTimeout int `mapstructure:"connect_timeout"`
	ReadTimeout int `mapstructure:"read_timeout"`
	WriteTimeout int `mapstructure:"write_timeout"`
	Routes []RoutePrepared
}

// ServicePrepared - service object without Id field as id will be different
type ServicePrepared struct {
	Name string `json:"name" mapstructure:"name"`
	Host string `json:"host" mapstructure:"host"`
	Path string `json:"path,omitempty" mapstructure:"path"`
	Port int `json:"port" mapstructure:"port"`
	Protocol string `json:"protocol" mapstructure:"protocol"`
	ConnectTimeout int `json:"connect_timeout" mapstructure:"connect_timeout"`
	ReadTimeout int `json:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout int `json:"write_timeout" mapstructure:"write_timeout"`
	Routes []RoutePrepared `json:"routes,omitempty"`
}

// Route - for obtaining routes from the server
type Route struct {
	Paths []string `mapstructure:"paths"`
	Service Service `mapstructure:"service"`
	StripPath bool `mapstructure:"strip_path"`
	PreserveHost bool `mapstructure:"preserve_host"`
	RegexPriority int `mapstructure:"regex_priority"`
	Hosts []string `mapstructure:"hosts"`
	Protocols []string `mapstructure:"protocols"`
	Methods []string `mapstructure:"methods"`
}

// RoutePrepared - route object without Service field as route is already nested inside the server
type RoutePrepared struct {
	Paths []string `json:"paths" mapstructure:"paths"`
	StripPath bool	`json:"strip_path" mapstructure:"strip_path"`
	PreserveHost bool `json:"preserve_host" mapstructure:"preserve_host"`
	RegexPriority int `json:"regex_priority" mapstructure:"regex_priority"`
	Hosts []string `json:"hosts" mapstructure:"hosts"`
	Protocols []string `json:"protocols" mapstructure:"protocols"`
	Methods []string `json:"methods" mapstructure:"methods"`
}

// CertificatePrepared - for obtaining certificates from the server
type CertificatePrepared struct {
	Cert string `json:"cert" mapstructure:"cert"`
	Key string `json:"key" mapstructure:"key"`
	Snis []string `json:"snis" mapstructure:"snis"`
}

// ResourceInstance can be both service or route
type ResourceInstance struct {
	Id string `mapstructure:"id"`
}
