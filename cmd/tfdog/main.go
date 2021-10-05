package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.sr.ht/~mcldresner/tfdog/bot"
	"git.sr.ht/~mcldresner/tfdog/config"
	"git.sr.ht/~mcldresner/tfdog/logger"
	"git.sr.ht/~mcldresner/tfdog/scheduler"
	"git.sr.ht/~mcldresner/tfdog/version"
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

	sc := getScheduler(cfg, log)
	b := getBot(cfg, log, sc)

	handleStop(sc, b, log)

	log.Info("starting...")
	sc.Start()
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

func getBot(cfg ini.File, log *zap.Logger, sc *scheduler.Scheduler) *tb.Bot {
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

	startText, ok := botCfg["help_text"]
	if !ok {
		cfgLog.Panic("config must contain help text field")
	}

	b, err := bot.NewBot(
		pollerTimeout,
		token,
		sc,
		helpText,
		startText,
	)
	if err != nil {
		log.With(zap.Error(err)).Panic("failed to create bot")
	}

	return b
}

func getScheduler(cfg ini.File, log *zap.Logger) *scheduler.Scheduler {
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

	sc := scheduler.NewScheduler(interval)

	return sc
}

func handleStop(sc *scheduler.Scheduler, b *tb.Bot, log *zap.Logger) {
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		s := <-termChan
		log.With(zap.Stringer("signal", s)).Info("received signal. terminating...")
		b.Stop()
		sc.Stop()
	}()
}
