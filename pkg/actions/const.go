package actions

const Services = "services"
const Routes = "routes"

type Resource struct {
	path string
	fields []string
}

var Apis = [...]Resource{
	{
	Services,
	[]string{"name", "id", "host", "protocol", "read_timeout", "port", "path", "retries", "write_timeout"},
	},
	{
	Routes,
	[]string{"id", "paths", "hosts", "methods", "protocols"},
	},
}