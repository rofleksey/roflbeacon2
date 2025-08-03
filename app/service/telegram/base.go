package telegram

import (
	"context"
	"encoding/json"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"strings"
)

func (s *Service) handleUpdates(ctx context.Context, _ *bot.Bot, update *models.Update) {
	if update.Message != nil {
		s.handleMessage(ctx, update.Message)
	}

	if update.CallbackQuery != nil {
		s.handleCallbackQuery(ctx, update.CallbackQuery)
	}
}

func (s *Service) handleMessage(ctx context.Context, msg *models.Message) {
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
	case "/history":
		s.handleHistory(ctx, &acc)
	}
}

func (s *Service) handleCallbackQuery(ctx context.Context, query *models.CallbackQuery) {
	if query.Message.Message == nil {
		return
	}

	acc, err := s.queries.GetAccountByChatID(ctx, &query.Message.Message.Chat.ID)
	if err != nil {
		return
	}

	if acc.ChatID == nil {
		return
	}

	var genericDTO GenericCallbackDTO
	_ = json.Unmarshal([]byte(query.Data), &genericDTO)

	switch genericDTO.Type {
	case "history":
		var historyDTO HistoryCallbackDTO
		_ = json.Unmarshal([]byte(query.Data), &historyDTO)

		s.handleHistoryCallback(ctx, &acc, historyDTO, query)
	default:
		slog.ErrorContext(ctx, "Unknown callback type",
			slog.String("type", genericDTO.Type),
		)
	}
}
