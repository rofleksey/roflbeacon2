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
	"strings"
)

type Service struct {
	tgBot   *bot.Bot
	cfg     *config.Config
	queries *database.Queries
}

func New(di *do.Injector) (*Service, error) {
	cfg := do.MustInvoke[*config.Config](di)

	service := &Service{
		cfg:     cfg,
		queries: do.MustInvoke[*database.Queries](di),
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

func (s *Service) handleUpdates(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil {
		s.handleMessage(ctx, b, update.Message)
	}
}

func (s *Service) handleMessage(ctx context.Context, b *bot.Bot, msg *models.Message) {
	acc, err := s.queries.GetAccountByChatID(ctx, &msg.Chat.ID)
	if err != nil {
		return
	}

	if acc.ChatID == nil {
		return
	}

	switch strings.TrimSpace(msg.Text) {
	case "/list":
		s.handleList(ctx, &acc)
	}
}

func (s *Service) handleList(ctx context.Context, selfAcc *database.Account) {
	accounts, err := s.queries.GetAllAccounts(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get all accounts",
			slog.Any("error", err),
		)
		return
	}

	var builder strings.Builder

	for _, acc := range accounts {
		if acc.ID == selfAcc.ID {
			continue
		}

		// TODO: improve this
		lastUpdates, err := s.queries.GetLastUpdateByAccountID(ctx, acc.ID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get last update",
				slog.Any("error", err),
			)
			return
		}

		if len(lastUpdates) == 0 {
			continue
		}

		lastUpdate := lastUpdates[0]
		loc := lastUpdate.Data.Location

		builder.WriteString("*")
		builder.WriteString(acc.Name)
		builder.WriteString("* (")
		builder.WriteString(util.TimeAgo(lastUpdate.Created))
		builder.WriteString(")\n")

		if loc == nil {
			builder.WriteString("Местоположение не определено")
		} else {
			mapLink := util.GenerateYandexLink(loc.Latitude, loc.Longitude)
			builder.WriteString(fmt.Sprintf("[На карте](%s)\n", mapLink))

			if loc.Address != nil {
				builder.WriteString(fmt.Sprintf("📍 %s\n", *loc.Address))
			} else {
				builder.WriteString("Адрес не определен\n")
			}
		}

		if lastUpdate.Data.Battery != nil {
			builder.WriteString(fmt.Sprintf("🔋 %d%% ", lastUpdate.Data.Battery.Level))

			if lastUpdate.Data.Battery.Charging {
				builder.WriteString("⚡")
			}

			builder.WriteString("\n")
		}

		builder.WriteString("\n\n")
	}

	s.SendMessage(ctx, *selfAcc.ChatID, builder.String())
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
