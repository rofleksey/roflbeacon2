package telegram

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"roflbeacon2/pkg/database"
	"strings"
)

func (s *Service) handleHistoryCallback(ctx context.Context, acc *database.Account, dto HistoryCallbackDTO, query *models.CallbackQuery) {
	if _, err := s.tgBot.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    query.Message.Message.Chat.ID,
		MessageID: query.Message.Message.ID,
	}); err != nil {
		slog.ErrorContext(ctx, "Failed to delete message",
			slog.Any("error", err),
		)
		return
	}

	targetAcc, err := s.queries.GetAccount(ctx, dto.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get account",
			slog.Any("error", err),
		)
		return
	}

	updates, err := s.queries.GetLatestUpdatesByAccountID(ctx, dto.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get latest updates",
			slog.Any("error", err),
		)
		return
	}

	var result []string

	for _, update := range updates {
		result = append(result, s.formatUpdate(&targetAcc, update, nil))
	}

	s.SendMessage(ctx, *acc.ChatID, strings.Join(result, "\n\n"))
}
