package actions

var TestEmailService = ServicePrepared{
	Name: "email-service",
	Host: "email.tld",
	Path: "/api/v1",
	Routes: []RoutePrepared{
		{
			Paths: []string{
				"/rest/emails",
			},
		},
	},
}

var TestCertificate = CertificatePrepared{
	"--certificate--",
	"--key--",
	[]string{"domain.tld"},
}
