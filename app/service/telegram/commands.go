package telegram

import (
	"context"
	"encoding/json"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"roflbeacon2/pkg/database"
	"strings"
)

func (s *Service) handleList(ctx context.Context, selfAcc *database.Account) {
	accounts, err := s.queries.GetAllAccounts(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get all accounts",
			slog.Any("error", err),
		)
		return
	}

	myLastUpdates, err := s.queries.GetLastUpdateByAccountID(ctx, selfAcc.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get last update",
			slog.Any("error", err),
		)
		return
	}

	if len(myLastUpdates) == 0 {
		slog.ErrorContext(ctx, "Myself doesn't have any updates")
		return
	}

	myLastUpdate := myLastUpdates[0]
	myLastLocation := myLastUpdate.Data.Location

	var result []string

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

		result = append(result, s.formatUpdate(selfAcc, lastUpdates[0], myLastLocation))
	}

	s.SendMessage(ctx, *selfAcc.ChatID, strings.Join(result, "\n\n"))
}

func (s *Service) handleHistory(ctx context.Context, selfAcc *database.Account) {
	accounts, err := s.queries.GetAllAccounts(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get all accounts",
			slog.Any("error", err),
		)
		return
	}

	var buttons []models.InlineKeyboardButton

	for _, acc := range accounts {
		callbackDTO := HistoryCallbackDTO{
			Type: "history",
			ID:   acc.ID,
		}

		callbackBytes, _ := json.Marshal(&callbackDTO)

		buttons = append(buttons, models.InlineKeyboardButton{
			Text:         acc.Name,
			CallbackData: string(callbackBytes),
		})
	}

	if _, err = s.tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: selfAcc.ID,
		Text:   "Выберите пользователя",
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{buttons},
		},
	}); err != nil {
		slog.ErrorContext(ctx, "Failed to send message",
			slog.Any("error", err),
		)
	}
}
