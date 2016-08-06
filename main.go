//go:generate goagen bootstrap -d github.com/codeclysm/mara-api/design

package main

import (
	"flag"

	"github.com/codeclysm/mara-api/app"
	"github.com/codeclysm/mara-api/auth"
	"github.com/codeclysm/mara-api/calendar"
	"github.com/codeclysm/rdbutils"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	"github.com/goadesign/goa/middleware/security/jwt"
)

var (
	port       = flag.String("port", ":9000", "The port where the api would listen")
	signingKey = flag.String("signingKey", "secret", "The key used to create jwt tokens")

	// Auth is the client to handle authentication
	Auth auth.Client
	// Calendar is the client that handles the appointments
	Calendar calendar.Client
)

func main() {
	// Create service
	service := goa.New("mara")

	// Users Client
	dbu := rdbutils.Database{Name: "mara", Table: "users"}
	if err := dbu.Connect(); err != nil {
		panic("Missing database")
	}
	Auth = auth.Client{DB: &dbu, SigningKey: *signingKey}

	// Calendar Client
	dbc := rdbutils.Database{Name: "mara", Table: "appointments"}
	if err := dbc.Connect(); err != nil {
		panic("Missing database")
	}
	Calendar = calendar.Client{DB: &dbc}

	// Mount middleware
	service.Use(middleware.RequestID())
	service.Use(middleware.LogRequest(true))
	service.Use(middleware.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	// Mount Jwt security
	app.UseJWTMiddleware(service, jwt.New(*signingKey, nil, app.NewJWTSecurity()))

	// Mount "auth" controller
	c := NewAuthController(service)
	app.MountAuthController(service, c)
	// Mount "calendar" controller
	c2 := NewCalendarController(service)
	app.MountCalendarController(service, c2)

	// Start service
	if err := service.ListenAndServe(":9000"); err != nil {
		service.LogError("startup", "err", err)
	}
}
