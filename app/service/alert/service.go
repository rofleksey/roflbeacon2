package alert

import (
	"context"
	"github.com/samber/do"
	"log/slog"
	"roflbeacon2/app/service/telegram"
	"roflbeacon2/pkg/config"
	"roflbeacon2/pkg/database"
)

type Service struct {
	appCtx          context.Context
	cfg             *config.Config
	queries         *database.Queries
	telegramService *telegram.Service
}

func New(di *do.Injector) (*Service, error) {
	return &Service{
		appCtx:          do.MustInvoke[context.Context](di),
		cfg:             do.MustInvoke[*config.Config](di),
		queries:         do.MustInvoke[*database.Queries](di),
		telegramService: do.MustInvoke[*telegram.Service](di),
	}, nil
}

func (s *Service) Alert(text string, ignoreChatID *int64) {
	accounts, err := s.queries.GetAllAccounts(s.appCtx)
	if err != nil {
		slog.ErrorContext(s.appCtx, "Failed to get all accounts",
			slog.Any("error", err),
		)
		return
	}

	for _, account := range accounts {
		if account.ChatID == nil || account.ChatID == ignoreChatID {
			continue
		}

		s.telegramService.SendMessage(s.appCtx, *account.ChatID, text)
	}
}
