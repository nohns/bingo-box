package main

import (
	"context"
	"fmt"
	"os"

	bingo "github.com/nohns/bingo-box/server"
	"github.com/nohns/bingo-box/server/bcrypt"
	"github.com/nohns/bingo-box/server/config"
	"github.com/nohns/bingo-box/server/http"
	"github.com/nohns/bingo-box/server/logger"
	"github.com/nohns/bingo-box/server/postgres"
)

type App struct {
	Log        logger.Logger
	HTTPServer *http.Server
	Conf       config.Conf
}

// Boostrap the server application
func (a *App) Bootstrap(ctx context.Context) error {

	// Run dependency prechecks...
	// Validate configuration
	err := a.Precheck()
	if err != nil {
		a.Log.Err("failed application prechecks")
		return err
	}

	// Setup bcrypt hasher
	hasher := bcrypt.NewHasher()

	// Setup postgres DB dependency
	db, err := postgres.NewDB(ctx, a.Conf.DSN(), a.Conf.DB.MigrationsPath)
	if err != nil {
		return err
	}

	// Migrate database up if enabled
	if a.Conf.DB.Migrate {
		if err = db.MigrateDown(); err != nil {
			return err
		}
		if err = db.MigrateUp(); err != nil {
			return err
		}
	}

	// Setup postgres repo dependencies
	userRepo := postgres.NewUserRepository(db)

	// Setup domain services
	userSvc := bingo.NewUserService(userRepo, hasher)

	// Setup HTTP rest server
	a.HTTPServer = http.NewServer(a.Conf.HTTP.APIKey)
	a.HTTPServer.UserService = userSvc
	a.HTTPServer.Addr = a.Conf.HTTPListenAddr()
	a.HTTPServer.Log = a.Log

	return nil
}

// Run checks need before application can start
func (a *App) Precheck() error {
	// Validate configuration
	err := a.Conf.Validate()
	if err != nil {
		return err
	}

	return nil
}

// Serve the applicaion. Atm. it is only the http server
func (a *App) Run(errChan chan<- error) {
	a.Log.Infof("Now serving http request on address %s...\n", a.HTTPServer.Addr)
	errChan <- a.HTTPServer.Serve()
}

// Runs when server application closes
func (a *App) Close() error {
	fmt.Fprint(os.Stdout, "\n")
	a.Log.Info("App terminated. Goodbye.")

	return nil
}

func NewApp() *App {
	return &App{
		Log: logger.New(),
	}
}
