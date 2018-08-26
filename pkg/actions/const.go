package actions

// Timeout - how long http client should wait before terminating connection
const Timeout = 10

// DefaultURL keeps url when kong api is accessed with port forwarding (as mentioned in readme)
const DefaultURL = "http://localhost:8001"

// ServicesKey has Kong admin services path
const ServicesKey = "services"

// RoutesKey has Kong admin routes path
const RoutesKey = "routes"

// Apis - list of apis for import/export
var Apis = []string{ServicesKey, RoutesKey}

// Service - for obtaining data from the server
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

// Route - for obtaining data from the server
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

// ResourceInstance can be both service or route
type ResourceInstance struct {
	Id string `mapstructure:"id"`
}
