package actions

var TestEmailService = Service{
	Id: "service1",
	Name: "email-service",
	Host: "email.tld",
	Path: "/api/v1",
	Routes: []Route{
		{
			Id: "route1",
			Paths: []string{
				"/rest/emails",
			},
		},
	},
}

var TestCertificate = Certificate{
	"certificate1",
	"--certificate--",
	"--key--",
	[]string{"domain.tld"},
}

var TestPlugin = Plugin{
	"plugin1",
	"test-plugin",
	map[string]interface{}{"key": "value"},
	true,
	TestEmailService.Id,
	"",
	"",
}