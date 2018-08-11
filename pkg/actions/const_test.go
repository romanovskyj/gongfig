package actions

var TestService = ServicePrepared{
	Name: "email-service",
	Host: "email.tld",
	Path: "/api/v1",
	Routes: []RoutePrepared{
		{
			Paths: []string{
				"/rest/reports",
			},
		},
	},
}
