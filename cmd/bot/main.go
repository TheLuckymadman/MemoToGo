package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
	"gopkg.in/telebot.v3"

	"github.com/theluckymadman/memotogo/internal/client"
	"github.com/theluckymadman/memotogo/internal/client/gigachat"
	"github.com/theluckymadman/memotogo/internal/client/oauth"
	"github.com/theluckymadman/memotogo/internal/config"
	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/handler"
	"github.com/theluckymadman/memotogo/internal/middleware"
	"github.com/theluckymadman/memotogo/internal/model"
	"github.com/theluckymadman/memotogo/internal/queue"
	"github.com/theluckymadman/memotogo/internal/repository"
	"github.com/theluckymadman/memotogo/internal/workers"
)

var _ workers.Transcriptor = (*client.Salute)(nil)

func run() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("init logger error: %w", err)
	}
	defer logger.Sync()

	cfg := config.NewConfig()
	httpClient := client.NewHTTPClient(cfg.HTTPTimeout)
	errMsg := "Hmm, it looks like something went wrong, but we are already aware of it and working on a fix"
	deps := deps.NewDeps(logger, httpClient, errMsg)

	pg, err := repository.NewPGStorage(cfg.DSN, cfg.DBInitMode, cfg.DBMigrationDir, deps)
	if err != nil {
		return fmt.Errorf("new storage: %w", err)
	}

	userRepo := repository.NewStorage[model.User](pg.DB, "users")
	meetRepo := repository.NewStorage[model.Meeting](pg.DB, "meetings")

	bot, err := client.NewBot(cfg.TeleToken)
	if err != nil {
		return err
	}

	saluteAuth := oauth.NewOAuth(
		cfg.SaluteOAuthURL,
		cfg.SaluteAuthKey,
		cfg.SaluteScope,
		cfg.SaluteOAuthRefreshTokenBeforeExp,
		deps,
	)
	salute := client.NewSalute(
		cfg.SaluteURL,
		saluteAuth,
		deps,
	)

	gigaAuth := oauth.NewOAuth(
		cfg.GigaChatOAuthURL,
		cfg.GigaChatAuthKey,
		cfg.GigaChatScope,
		cfg.GigaOAuthRefreshTokenBeforeExp,
		deps,
	)
	giga := gigachat.NewGigaChat(
		cfg.GigaChatURL,
		cfg.GigaChatModel,
		gigaAuth,
		deps,
	)

	q := queue.NewQueue(5)
	h := handler.NewHandler(deps, q, bot, userRepo, meetRepo)
	m := middleware.NewMiddleware(userRepo, deps)
	bot.Handle("/start", h.Start)
	bot.Handle(telebot.OnText, h.OnText, m.Auth)
	bot.Handle(telebot.OnVoice, h.OnVoice, m.Auth)
	bot.Handle(telebot.OnAudio, h.OnAudio, m.Auth)
	bot.Handle("/find", h.OnFind, m.Auth)
	bot.Handle("/list", h.OnList, m.Auth)
	bot.Handle("/get", h.OnGet, m.Auth)
	go bot.Start()

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	w := workers.NewWorker(q, bot, salute, giga, meetRepo, deps)
	go w.Start()

	n := workers.NewNotifier(bot, deps, userRepo, signalCtx)
	n.Start()

	var wg = sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-signalCtx.Done()
		logger.Info("graceful shutdown initiated")
		bot.Stop()
		logger.Info("bot stopped gracefully")
	}()

	wg.Wait()

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
