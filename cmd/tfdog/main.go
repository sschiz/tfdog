package main

import (
	"database/sql"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"git.sr.ht/~mcldresner/tfdog/recovery"

	"git.sr.ht/~mcldresner/tfdog/config"
	"git.sr.ht/~mcldresner/tfdog/logger"
	"git.sr.ht/~mcldresner/tfdog/repository"
	"git.sr.ht/~mcldresner/tfdog/service"
	"git.sr.ht/~mcldresner/tfdog/transport/bot"
	"git.sr.ht/~mcldresner/tfdog/version"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vaughan0/go-ini"
	"go.uber.org/zap"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	cfg := getConfig()
	log := getLogger(cfg)
	defer func(log *zap.Logger) {
		_ = log.Sync()
	}(log)

	repo := getRepository(cfg, log)
	defer func(repo repository.Repository) {
		err := repo.Close()
		if err != nil {
			log.With(zap.Error(err)).Error("failed to close repository")
		}
	}(repo)

	srv := getService(cfg, log, repo)
	defer func(srv service.Service) {
		err := srv.Close()
		if err != nil {
			log.With(zap.Error(err)).Error("failed to close service")
		}
	}(srv)

	b := getBot(cfg, log, srv)

	recoveryFromRepository(srv, repo, b, log)
	handleStop(b, log)

	log.Info("starting...")
	b.Start()
}

func getConfig() ini.File {
	const argsLen = 2
	if len(os.Args) < argsLen {
		panic("config path must be passed")
	}

	cfgPath := os.Args[1]
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		panic(err)
	}

	return cfg
}

func getLogger(cfg ini.File) *zap.Logger {
	value, ok := cfg.Get("logger", "level")
	if !ok {
		panic("config must contain logger level field")
	}

	var isProduction bool
	switch value {
	case "production":
		isProduction = true
	case "development":
	default:
		panic("logger level must be equal to one of values: production or development")
	}

	log, err := logger.NewLogger(isProduction)
	if err != nil {
		panic(err)
	}

	const name = "tfdog"
	log = log.
		With(
			zap.String("version", version.Version),
			zap.String("app", name),
			zap.String("env", value),
		)

	zap.ReplaceGlobals(log)

	return log
}

func getBot(cfg ini.File, log *zap.Logger, srv service.Service) *tb.Bot {
	cfgLog := log.Named("config").With(zap.String("section", "bot"))
	botCfg := cfg.Section("bot")

	token, ok := botCfg["token"]
	if !ok {
		cfgLog.Panic("config must contain token field")
	}

	pollerTimeoutStr, ok := botCfg["poller_timeout"]
	if !ok {
		pollerTimeoutStr = "10s"
	}
	pollerTimeout, err := time.ParseDuration(pollerTimeoutStr)
	if err != nil {
		cfgLog.With(zap.Error(err)).Panic("failed to parse poller timeout")
	}

	helpText, ok := botCfg["help_text"]
	if !ok {
		cfgLog.Panic("config must contain help text field")
	}
	helpText = strings.ReplaceAll(helpText, "\\n", "\n")

	startText, ok := botCfg["start_text"]
	if !ok {
		cfgLog.Panic("config must contain help text field")
	}
	startText = strings.ReplaceAll(helpText, "\\n", "\n")

	b, err := bot.NewBot(
		pollerTimeout,
		token,
		srv,
		helpText,
		startText,
	)
	if err != nil {
		log.With(zap.Error(err)).Panic("failed to create bot")
	}

	return b
}

func handleStop(b *tb.Bot, log *zap.Logger) {
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		s := <-termChan
		log.With(zap.Stringer("signal", s)).Info("received signal. terminating...")
		b.Stop()
	}()
}

func getRepository(cfg ini.File, log *zap.Logger) repository.Repository {
	dsn, ok := cfg.Get("database", "data_source_name")
	if !ok {
		log.
			Named("config").
			With(zap.String("section", "database")).
			Panic("config must contain data_source_name field")
	}

	err := migrate(dsn)
	if err != nil {
		log.
			Named("migration").
			With(zap.Error(err)).
			Panic("failed to migrate database")
	}

	repo, err := repository.NewSqliteRepository(dsn)
	if err != nil {
		log.
			Named("repository").
			With(zap.Error(err)).
			Panic("failed to create new sqlite repo")
	}

	return repo
}

const migrateQuery = `
CREATE TABLE IF NOT EXISTS subscriptions
(
    user_id  int,
    app_name text,
    link     text
);
`

func migrate(dsn string) error {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	_, err = db.Exec(migrateQuery)
	if err != nil {
		return err
	}

	return nil
}

func getService(cfg ini.File, log *zap.Logger, repo repository.Repository) service.Service {
	value, ok := cfg.Get("scheduler", "interval")
	if !ok {
		log.
			Named("config").
			With(
				zap.String("section", "scheduler"),
			).
			Panic("config must contain interval field")
	}

	interval, err := time.ParseDuration(value)
	if err != nil {
		log.With(zap.Error(err)).Panic("failed to parse interval")
	}

	srv := service.NewService(repo, interval)
	return srv
}

func recoveryFromRepository(srv service.Service, repo repository.Repository, b *tb.Bot, log *zap.Logger) {
	err := recovery.ServiceFromRepository(srv, repo, b)
	if err != nil {
		log.With(zap.Error(err)).Panic("failed to recovery service from repository")
	}
}
