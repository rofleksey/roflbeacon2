package telegram

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/samber/do"
	"log/slog"
	"roflbeacon2/pkg/config"
	"roflbeacon2/pkg/database"
	"roflbeacon2/pkg/util"
	"sync"
)

type Service struct {
	tgBot   *bot.Bot
	cfg     *config.Config
	queries *database.Queries

	m     sync.Mutex
	state BotState
}

func New(di *do.Injector) (*Service, error) {
	cfg := do.MustInvoke[*config.Config](di)

	service := &Service{
		cfg:     cfg,
		queries: do.MustInvoke[*database.Queries](di),
		state: BotState{
			Stage: "idle",
		},
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(service.handleUpdates),
	}

	b, err := bot.New(cfg.Telegram.Token, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	service.tgBot = b

	return service, nil
}

func (s *Service) SendMessage(ctx context.Context, chatID int64, text string) {
	if _, err := s.tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: util.ToPtr(true),
		},
	}); err != nil {
		slog.ErrorContext(ctx, "Failed to send message",
			slog.Int64("chat_id", chatID),
			slog.String("text", text),
			slog.Any("error", err),
		)
	}
}

func (s *Service) initCommands(ctx context.Context) {
	cmds := []models.BotCommand{
		{
			Command:     "/list",
			Description: "Показать всех",
		},
		{
			Command:     "/history",
			Description: "История",
		},
		{
			Command:     "/addfence",
			Description: "Добавить ограду",
		},
		{
			Command:     "/deletefence",
			Description: "Удалить ограду",
		},
		{
			Command:     "/cancel",
			Description: "Отменить текущее действие",
		},
	}

	if _, err := s.tgBot.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: cmds,
	}); err != nil {
		slog.ErrorContext(ctx, "Failed to set commands",
			slog.Any("error", err),
		)
	}
}

func (s *Service) Run(ctx context.Context) {
	s.initCommands(ctx)
	s.tgBot.Start(ctx)
}
