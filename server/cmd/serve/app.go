package main

import (
	"context"
	"fmt"
	"os"
	"time"

	bingo "github.com/nohns/bingo-box/server"
	"github.com/nohns/bingo-box/server/bcrypt"
	"github.com/nohns/bingo-box/server/config"
	"github.com/nohns/bingo-box/server/http"
	"github.com/nohns/bingo-box/server/logger"
	"github.com/nohns/bingo-box/server/mail"
	"github.com/nohns/bingo-box/server/mongo"
)

type App struct {
	Log        logger.Logger
	HTTPServer *http.Server
	Conf       config.Conf
	Mailer     *mail.Mailer
}

// Boostrap the server application
func (a *App) Bootstrap(ctx context.Context) error {

	// Setup bcrypt hasher
	hasher := bcrypt.NewHasher()

	// Setup mongodb dependency
	mongoCtx, mongoCancel := context.WithTimeout(ctx, 5*time.Second)
	db, err := mongo.New(mongoCtx, a.Conf.ConnURI())
	mongoCancel()
	if err != nil {
		return err
	}

	a.Mailer = mail.NewMailer(a.Conf.Mail.MGDomain, a.Conf.Mail.MGAPIKey)
	a.Mailer.BaseDownloadLink = a.Conf.Mail.DLLinkBase

	// Setup repos dependencies
	userRepo := mongo.NewUserRepository(db)
	invRepo := mongo.NewInvitationRepository(db)
	playerRepo := mongo.NewPlayerRepository(db)
	gameRepo := mongo.NewGameRepository(db)
	cardRepo := mongo.NewCardRepository(db)

	// Setup domain services
	userSvc := bingo.NewUserService(userRepo, hasher)
	gameSvc := bingo.NewGameService(gameRepo, cardRepo)
	invSvc := bingo.NewInvitationService(invRepo, playerRepo)
	playerSvc := bingo.NewPlayerService(playerRepo)

	// Setup HTTP rest server
	a.HTTPServer = http.NewServer(a.Conf.HTTP.APIKey)
	a.HTTPServer.UserService = userSvc
	a.HTTPServer.GameService = gameSvc
	a.HTTPServer.InvitationService = invSvc
	a.HTTPServer.PlayerService = playerSvc

	a.HTTPServer.Addr = a.Conf.HTTPListenAddr()
	a.HTTPServer.Log = a.Log

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
