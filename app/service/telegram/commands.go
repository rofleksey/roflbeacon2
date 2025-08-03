package telegram

import (
	"context"
	"encoding/json"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"roflbeacon2/pkg/database"
	"strconv"
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

		result = append(result, s.formatUpdate(&acc, lastUpdates[0], myLastLocation))
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

	cancelDTO := GenericCallbackDTO{
		Type: "cancel",
	}

	cancelBytes, _ := json.Marshal(&cancelDTO)

	buttons = append(buttons, models.InlineKeyboardButton{
		Text:         "Отмена",
		CallbackData: string(cancelBytes),
	})

	if _, err = s.tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: selfAcc.ChatID,
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

func (s *Service) handleCancel(ctx context.Context, selfAcc *database.Account) {
	if selfAcc.ChatID != &s.cfg.Telegram.AdminChatID {
		s.SendMessage(ctx, *selfAcc.ChatID, "Вы не можете использовать данную команду")
		return
	}

	s.m.Lock()
	defer s.m.Unlock()

	s.resetState()

	s.SendMessage(ctx, *selfAcc.ChatID, "ОК")
}

func (s *Service) handleAddFence(ctx context.Context, selfAcc *database.Account) {
	if selfAcc.ChatID != &s.cfg.Telegram.AdminChatID {
		s.SendMessage(ctx, *selfAcc.ChatID, "Вы не можете использовать данную команду")
		return
	}

	s.m.Lock()
	defer s.m.Unlock()

	s.resetState()
	s.state.Stage = "add_fence_name"

	s.SendMessage(ctx, *selfAcc.ChatID, "Введите имя новой ограды:")
}

func (s *Service) handleDeleteFence(ctx context.Context, selfAcc *database.Account) {
	fences, err := s.queries.GetAllFences(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get all fences",
			slog.Any("error", err),
		)
		return
	}

	var buttons []models.InlineKeyboardButton

	for _, f := range fences {
		callbackDTO := DeleteFenceCallbackDTO{
			Type: "delete_fence",
			ID:   f.ID,
		}

		callbackBytes, _ := json.Marshal(&callbackDTO)

		buttons = append(buttons, models.InlineKeyboardButton{
			Text:         f.Name,
			CallbackData: string(callbackBytes),
		})
	}

	cancelDTO := GenericCallbackDTO{
		Type: "cancel",
	}

	cancelBytes, _ := json.Marshal(&cancelDTO)

	buttons = append(buttons, models.InlineKeyboardButton{
		Text:         "Отмена",
		CallbackData: string(cancelBytes),
	})

	if _, err = s.tgBot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: selfAcc.ChatID,
		Text:   "Выберите ограду для удаления",
		ReplyMarkup: models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{buttons},
		},
	}); err != nil {
		slog.ErrorContext(ctx, "Failed to send message",
			slog.Any("error", err),
		)
	}
}

func (s *Service) handleUnknownMessage(ctx context.Context, selfAcc *database.Account, text string) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.state.Stage == "idle" || selfAcc.ChatID != &s.cfg.Telegram.AdminChatID {
		s.SendMessage(ctx, *selfAcc.ChatID, "Неизвестная команда")
		return
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	switch s.state.Stage {
	case "add_fence_name":
		s.state.FenceParams.Name = text
		s.state.Stage = "add_fence_latitude"
		s.SendMessage(ctx, *selfAcc.ChatID, "Введите широту:")
	case "add_fence_latitude":
		value, err := strconv.ParseFloat(text, 64)
		if value == 0 || err != nil {
			s.SendMessage(ctx, *selfAcc.ChatID, "Неверное число, попробуйте еще раз")
			return
		}

		s.state.FenceParams.Latitude = value
		s.state.Stage = "add_fence_longitude"
		s.SendMessage(ctx, *selfAcc.ChatID, "Введите долготу:")
	case "add_fence_longitude":
		value, err := strconv.ParseFloat(text, 64)
		if value == 0 || err != nil {
			s.SendMessage(ctx, *selfAcc.ChatID, "Неверное число, попробуйте еще раз")
			return
		}

		s.state.FenceParams.Longitude = value
		s.state.Stage = "add_fence_radius"
		s.SendMessage(ctx, *selfAcc.ChatID, "Введите радиус (в метрах):")
	case "add_fence_radius":
		value, err := strconv.ParseFloat(text, 64)
		if value == 0 || err != nil {
			s.SendMessage(ctx, *selfAcc.ChatID, "Неверное число, попробуйте еще раз")
			return
		}

		s.state.FenceParams.Radius = value

		jsonBytes, _ := json.Marshal(&s.state.FenceParams)

		if _, err = s.queries.CreateFence(ctx, s.state.FenceParams); err != nil {
			slog.ErrorContext(ctx, "Failed to create fence",
				slog.Any("error", err),
			)
			return
		}

		s.SendMessage(ctx, *selfAcc.ChatID, string(jsonBytes))
		s.SendMessage(ctx, *selfAcc.ChatID, "Ограда успешно создана")

		s.state.Stage = "idle"
	}
}
