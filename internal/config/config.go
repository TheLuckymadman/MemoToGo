package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/theluckymadman/memotogo/internal/repository"
	"gopkg.in/yaml.v3"
)

type Config struct {
	TeleToken                        string                `env:"TELEGRAM_TOKEN"`
	DSN                              string                `env:"DSN"`
	DBInitMode                       repository.DBInitMode `env:"DB_INIT_MODE"`
	DBMigrationDir                   string                `env:"DB_MIGRATION_DIR"`
	HTTPTimeout                      time.Duration         `env:"HTTP_TIMEOUT"`
	GigaChatURL                      string                `env:"GIGACHAT_URL"`
	GigaChatModel                    string                `env:"GIGACHAT_MODEL"`
	GigaChatOAuthURL                 string                `env:"GIGACHAT_OAUTH_URL"`
	GigaChatAuthKey                  string                `env:"GIGACHAT_AUTH_KEY"`
	GigaOAuthRefreshTokenBeforeExp   time.Duration         `env:"GIGA_REFRESH_TOKEN_BEFORE"`
	GigaChatScope                    string                `env:"GIGACHAT_SCOPE"`
	SaluteURL                        string                `env:"SALUTE_URL"`
	SaluteOAuthURL                   string                `env:"SALUTE_OAUTH_URL"`
	SaluteAuthKey                    string                `env:"SALUTE_AUTH_KEY"`
	SaluteScope                      string                `env:"SALUTE_SCOPE"`
	SaluteOAuthRefreshTokenBeforeExp time.Duration         `env:"SALUTE_REFRESH_TOKEN_BEFORE"`
	LLMSVCSettingsFile               string                `yaml:"LLM_SVC_SETTINGS_FILE"`
	LLMSVSSettings                   LLMSVSSettings
}

func NewConfig() *Config {
	cfg := Config{
		DBInitMode:                       repository.External,
		DBMigrationDir:                   "migrations",
		HTTPTimeout:                      30 * time.Second,
		GigaChatURL:                      "https://gigachat.devices.sberbank.ru/api/v1",
		GigaChatModel:                    "GigaChat-2-Pro",
		GigaChatOAuthURL:                 "https://ngw.devices.sberbank.ru:9443/api/v2/oauth",
		GigaChatScope:                    "GIGACHAT_API_PERS",
		GigaOAuthRefreshTokenBeforeExp:   1 * time.Minute,
		SaluteURL:                        "https://smartspeech.sber.ru/rest/v1",
		SaluteOAuthURL:                   "https://ngw.devices.sberbank.ru:9443/api/v2/oauth",
		SaluteOAuthRefreshTokenBeforeExp: 1 * time.Minute,
		SaluteScope:                      "SALUTE_SPEECH_PERS",
		LLMSVCSettingsFile:               "llmsvcsettings.yaml",
	}
	err := godotenv.Load()
	if err != nil {
		log.Print(err)
	}

	err = env.Parse(&cfg)
	if err != nil {
		log.Printf("parse envs: %v", err)
	}

	dbInitMode, err := cfg.DBInitMode.String()
	if err != nil {
		log.Fatalf("marshal db init mode: %v", err)
	}

	flag.StringVar(&cfg.TeleToken, "tele-token", cfg.TeleToken, "telegram token")
	flag.StringVar(&cfg.DSN, "d", cfg.DSN, "DB DSN|URI")
	flag.StringVar(&dbInitMode, "m", dbInitMode, "DB init mode, use external, internal, forceinternal")
	flag.StringVar(&cfg.DBMigrationDir, "migr-dir", cfg.DBMigrationDir, "DB migration scripts dir path")

	flag.StringVar(&cfg.GigaChatURL, "g-url", cfg.GigaChatURL, "gigachat url")
	flag.StringVar(&cfg.GigaChatOAuthURL, "g-oauth-url", cfg.GigaChatOAuthURL, "gigachat oauth url")
	//flag.StringVar(&cfg.GigaChatAuthKey, "g-oauth-key", "", "gigachat oauth authorization key")
	flag.StringVar(&cfg.GigaChatScope, "g-scope", cfg.GigaChatScope, "gigachat scope")
	flag.StringVar(&cfg.GigaChatModel, "g-model", cfg.GigaChatModel, "gigachat model")
	var gigachatRefreshTokenBeforeExp int64 = int64(cfg.GigaOAuthRefreshTokenBeforeExp.Seconds())
	flag.Int64Var(&gigachatRefreshTokenBeforeExp, "g-token-upd-before", gigachatRefreshTokenBeforeExp, "refresh token before it exipres time in sec")

	flag.StringVar(&cfg.SaluteURL, "s-url", cfg.SaluteURL, "salute url")
	flag.StringVar(&cfg.SaluteOAuthURL, "s-oauth-url", cfg.SaluteOAuthURL, "salute oauth url")
	//flag.StringVar(&cfg.SaluteAuthKey, "s-oauth-key", "", "salute oauth authorization key")
	flag.StringVar(&cfg.SaluteScope, "s-scope", cfg.SaluteScope, "salute scope")
	var saluteRefreshTokenBeforeExp int64 = int64(cfg.SaluteOAuthRefreshTokenBeforeExp.Seconds())
	flag.Int64Var(&saluteRefreshTokenBeforeExp, "s-token-upd-before", saluteRefreshTokenBeforeExp, "refresh token before it exipres time in sec")

	flag.StringVar(&cfg.LLMSVCSettingsFile, "llm-cfg", cfg.LLMSVCSettingsFile, "llm service settings yaml file")

	var httpTimeout int64 = int64(cfg.HTTPTimeout.Seconds())
	flag.Int64Var(&httpTimeout, "t", httpTimeout, "http request timeout in sec")

	flag.Parse()

	cfg.HTTPTimeout = time.Duration(httpTimeout) * time.Second
	cfg.SaluteOAuthRefreshTokenBeforeExp = time.Duration(saluteRefreshTokenBeforeExp) * time.Second
	cfg.GigaOAuthRefreshTokenBeforeExp = time.Duration(gigachatRefreshTokenBeforeExp) * time.Second

	if err = cfg.DBInitMode.UnmarshalText([]byte(dbInitMode)); err != nil {
		log.Fatalf("failed to parse DB init mode: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}

	llm := LLMSVSSettings{}
	if err = llm.LoadYaml(cfg.LLMSVCSettingsFile); err != nil {
		log.Fatalf("LLM service load config: %v", err)
	}
	cfg.LLMSVSSettings = llm

	return &cfg
}

func (c *Config) Validate() error {
	if c.TeleToken == "" {
		return fmt.Errorf("TELEGRAM_TOKEN is required")
	}

	if c.DSN == "" {
		return fmt.Errorf("DB DSN is required")
	}

	if c.SaluteAuthKey == "" {
		return fmt.Errorf("SALUTE_AUTH_KEY is required")
	}

	if c.SaluteScope == "" {
		return fmt.Errorf("SALUTE_SCOPE is required")
	}

	return nil
}

type LLMSVSSettings struct {
	VoiceAssistSystemPrompt string `yaml:"VoiceAssistSystemPrompt"`
	ChatAssitSystemPropmpt  string `yaml:"ChatAssitSystemPropmpt"`
}

func (l *LLMSVSSettings) LoadYaml(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file %w", err)
	}

	err = yaml.Unmarshal(file, &l)
	if err != nil {
		return fmt.Errorf("unmarshal file %w", err)
	}
	return nil
}
