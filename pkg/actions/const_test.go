package actions

var TestEmailService = Service{
	Name: "email-service",
	Host: "email.tld",
	Path: "/api/v1",
	Routes: []Route{
		{
			Paths: []string{
				"/rest/emails",
			},
		},
	},
}

var TestCertificate = Certificate{
	"--certificate--",
	"--key--",
	[]string{"domain.tld"},
}
