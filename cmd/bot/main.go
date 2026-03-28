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
	"github.com/theluckymadman/memotogo/internal/client/salute"
	"github.com/theluckymadman/memotogo/internal/config"
	"github.com/theluckymadman/memotogo/internal/deps"
	"github.com/theluckymadman/memotogo/internal/handler"
	"github.com/theluckymadman/memotogo/internal/llm"
	"github.com/theluckymadman/memotogo/internal/llm/tool"
	"github.com/theluckymadman/memotogo/internal/middleware"
	"github.com/theluckymadman/memotogo/internal/model"
	"github.com/theluckymadman/memotogo/internal/queue"
	"github.com/theluckymadman/memotogo/internal/repository"
	"github.com/theluckymadman/memotogo/internal/workers"
)

var _ workers.Transcriptor = (*salute.SaluteAdapter)(nil)
var _ workers.LLM = (*gigachat.GigaChatAdapter)(nil)

func run() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("init logger error: %w", err)
	}
	defer logger.Sync()

	cfg := config.NewConfig()
	deps := initGlobalDeps(logger, cfg)

	pg, err := repository.NewPGStorage(cfg.DSN, cfg.DBInitMode, cfg.DBMigrationDir, deps)
	if err != nil {
		return fmt.Errorf("new storage: %w", err)
	}
	userRepo := repository.NewStorage[model.User](pg.DB, "users")
	meetRepo := repository.NewStorage[model.Meeting](pg.DB, "meetings")
	transcriptRepo := repository.NewTranscriptionRepo(pg.DB.DB)
	summaryRepo := repository.NewSummaryTextRepo(pg.DB.DB)

	saluteAdapter := initSalute(cfg, deps)
	gigaAdapter := initGiga(cfg, deps)

	toolRegistry := tool.NewRegistry()
	listMeeting := tool.NewListMeetings(meetRepo, deps)
	getMeeting := tool.NewGetMeeting(meetRepo, deps)
	toolRegistry.Add(listMeeting)
	toolRegistry.Add(getMeeting)
	userState := llm.NewState(20, 5, deps)
	llmSvc := llm.NewLLMService(gigaAdapter, cfg.LLMSVSSettings.ChatAssitSystemPropmpt, deps, userState, toolRegistry, 5)

	jobQueue := queue.NewQueue(5)

	bot, err := client.NewBot(cfg.TeleToken)
	if err != nil {
		return err
	}
	h := handler.NewHandler(deps, jobQueue, bot, llmSvc, userRepo, meetRepo)
	m := middleware.NewMiddleware(userRepo, deps)
	bot.Handle("/start", h.Start)
	bot.Handle(telebot.OnText, h.OnText, m.Auth)
	bot.Handle(telebot.OnVoice, h.OnVoice, m.Auth)
	bot.Handle(telebot.OnAudio, h.OnAudio, m.Auth)
	bot.Handle("/find", h.OnFind, m.Auth)
	bot.Handle("/list", h.OnList, m.Auth)
	bot.Handle("/get", h.OnGet, m.Auth)
	bot.Handle("/chat", h.OnText, m.Auth)

	var wg = sync.WaitGroup{}
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go bot.Start()

	taskCreator := workers.NewTaskCreator(jobQueue, saluteAdapter, meetRepo, transcriptRepo, deps)

	var creatorWg sync.WaitGroup
	for id := range cfg.ParallelTaskCnt {
		wg.Add(1)
		creatorWg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer creatorWg.Done()
			taskCreator.Start(signalCtx, id)
		}(id)
	}

	notifier := workers.NewNotifier(bot, deps, userRepo)
	wg.Add(1)
	go func() {
		defer wg.Done()
		notifier.Start(signalCtx)
	}()

	taskExecutor := workers.NewTaskExecutor(cfg.ParallelTaskCnt, cfg.RunWorkersInterval, saluteAdapter, meetRepo, transcriptRepo, deps)
	wg.Add(1)
	go func() {
		defer wg.Done()
		taskExecutor.Start(signalCtx)
	}()

	//Only 1 parallel request is allowed in the free tier; it is temporarily hard-coded.
	summarizer := workers.NewSummarizer(1, cfg.RunWorkersInterval, bot, gigaAdapter, cfg.LLMSVSSettings.VoiceAssistSystemPrompt, meetRepo, summaryRepo, deps)
	wg.Add(1)
	go func() {
		defer wg.Done()
		summarizer.Start(signalCtx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		creatorWg.Wait()
		logger.Info("shutdown signal received", zap.Error(signalCtx.Err()))
		bot.Stop()
		logger.Info("bot stopped gracefully")
	}()

	wg.Wait()
	close(jobQueue.Jobs)
	logger.Info("job queue stopped")

	return nil
}

func initGlobalDeps(logger *zap.Logger, cfg *config.Config) *deps.Deps {
	httpClient := client.NewHTTPClient(cfg.HTTPTimeout)
	errMsg := "Hmm, it looks like something went wrong, but we are already aware of it and working on a fix"
	deps := deps.NewDeps(logger, httpClient, errMsg)
	return deps
}

func initSalute(cfg *config.Config, deps *deps.Deps) *salute.SaluteAdapter {
	saluteAuth := oauth.NewOAuth(
		cfg.SaluteOAuthURL,
		cfg.SaluteAuthKey,
		cfg.SaluteScope,
		cfg.SaluteOAuthRefreshTokenBeforeExp,
		deps,
	)
	saluteClient := salute.NewSalute(
		cfg.SaluteURL,
		saluteAuth,
		deps,
	)
	saluteAdapter := salute.NewSaluteAdapter(saluteClient, deps)

	return saluteAdapter
}

func initGiga(cfg *config.Config, deps *deps.Deps) *gigachat.GigaChatAdapter {
	gigaAuth := oauth.NewOAuth(
		cfg.GigaChatOAuthURL,
		cfg.GigaChatAuthKey,
		cfg.GigaChatScope,
		cfg.GigaOAuthRefreshTokenBeforeExp,
		deps,
	)
	gigaClient := gigachat.NewGigaChat(
		cfg.GigaChatURL,
		cfg.GigaChatModel,
		gigaAuth,
		deps,
	)
	gigaAdapter := gigachat.NewGigaChatAdapter(gigaClient, deps)

	return gigaAdapter
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
