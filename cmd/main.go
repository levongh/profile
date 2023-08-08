//go:generate swag init -parseDependency -d ../internal/api -g ../../cmd/main.go
package main

import (
	golog "log"
	"runtime"
	"time"

	_ "github.com/lib/pq" // postgres driver

	"github.com/levongh/profile/internal/api"
	"github.com/levongh/profile/internal/config"
	"github.com/levongh/profile/internal/log"
)

// @title Profile API
// @version 1.0
// @description Entrypoint for profile related requests.
// @termsOfService http://swagger.io/terms/

// @contact.name CONTACT NAME
// @contact.url http://www.contact.url
// @contact.email contact@email.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8030
// @BasePath /api/v1

func main() {
	cfg, err := config.Read()
	if err != nil {
		golog.Fatal(err)
	}
	// cfg.HostWithoutProtocol()
	// change swagger host per deployment, see HOST env var
	// swaggerSettings.SwaggerInfo.Host = cfg.HostWithoutProtocol()

	// Register database option data
	// columns.RegisterOptionData()

	logger, err := log.NewLogger(cfg.Mode, cfg.LogLevel)

	if err != nil {
		golog.Fatal(err)
	}

	s, err := api.NewServer(cfg, logger)
	if err != nil {
		golog.Fatalf("can't start server: %s", err)
	}

	defer func() {
		// Close jeafer tracer after server stopped.
		err := s.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	// TODO: greaceful stutdown
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			s.Logger.Debugf("Number of goroutines: %d", runtime.NumGoroutine())
		}
	}()

	golog.Fatal(s.Start(cfg.Port))
}
