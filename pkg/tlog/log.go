package tlog

import (
	slogmulti "github.com/samber/slog-multi"
	slogtelegram "github.com/samber/slog-telegram/v2"
	"log/slog"
	"os"
	"roflbeacon2/pkg/build"
	"roflbeacon2/pkg/config"
)

func Init(cfg *config.Config) error {
	logHandlers := []slog.Handler{slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	})}

	//if cfg.Log.LokiUrl != "" {
	//	lokiCfg, err := loki.NewDefaultConfig(cfg.Log.LokiUrl)
	//	if err != nil {
	//		return fmt.Errorf("failed to create loki config: %w", err)
	//	}
	//
	//	client, err := loki.New(lokiCfg)
	//	if err != nil {
	//		return fmt.Errorf("failed to create loki client: %w", err)
	//	}
	//
	//	logHandlers = append(logHandlers, slogloki.Option{
	//		Level:     slog.LevelDebug,
	//		Client:    client,
	//		AddSource: true,
	//	}.NewLokiHandler())
	//}

	if cfg.Log.TelegramToken != "" && cfg.Log.TelegramChatID != "" {
		logHandlers = append(logHandlers, slogtelegram.Option{
			Level:     slog.LevelError,
			Token:     cfg.Log.TelegramToken,
			Username:  cfg.Log.TelegramChatID,
			AddSource: true,
		}.NewTelegramHandler())
	}

	multiHandler := slogmulti.Fanout(logHandlers...)
	ctxHandler := &contextHandler{multiHandler}

	logger := slog.New(ctxHandler).With(
		slog.String("app", "api"),
		slog.String("app_tag", build.Tag),
	)
	slog.SetDefault(logger)

	return nil
}
