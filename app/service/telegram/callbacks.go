package telegram

import (
	"context"
	"log/slog"
	"roflbeacon2/pkg/database"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (s *Service) handleCancelCallback(ctx context.Context, acc *database.Account, query *models.CallbackQuery) {
	if _, err := s.tgBot.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    acc.ChatID,
		MessageID: query.Message.Message.ID,
	}); err != nil {
		slog.ErrorContext(ctx, "Failed to delete message",
			slog.Any("error", err),
		)
		return
	}
}

func (s *Service) handleHistoryCallback(ctx context.Context, acc *database.Account, dto HistoryCallbackDTO, query *models.CallbackQuery) {
	if _, err := s.tgBot.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    acc.ChatID,
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

func (s *Service) handleDeleteFenceCallback(ctx context.Context, acc *database.Account, dto DeleteFenceCallbackDTO, query *models.CallbackQuery) {
	if _, err := s.tgBot.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    acc.ChatID,
		MessageID: query.Message.Message.ID,
	}); err != nil {
		slog.ErrorContext(ctx, "Failed to delete message",
			slog.Any("error", err),
		)
		return
	}

	if err := s.queries.DeleteFence(ctx, dto.ID); err != nil {
		slog.ErrorContext(ctx, "Failed to delete fence",
			slog.Any("error", err),
		)
		return
	}

	s.SendMessage(ctx, *acc.ChatID, "Ограда удалена")
}
