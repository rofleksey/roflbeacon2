package offline

import (
	"context"
	"fmt"
	"log/slog"
	"roflbeacon2/app/service/alert"
	"roflbeacon2/pkg/config"
	"roflbeacon2/pkg/database"
	"time"

	"github.com/samber/do"
)

const offlineThresholdMinutes = 30

type Service struct {
	cfg          *config.Config
	queries      *database.Queries
	alertService *alert.Service
}

func New(di *do.Injector) (*Service, error) {
	return &Service{
		cfg:          do.MustInvoke[*config.Config](di),
		queries:      do.MustInvoke[*database.Queries](di),
		alertService: do.MustInvoke[*alert.Service](di),
	}, nil
}

func (s *Service) RunBackgroundChecks(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.CheckOfflineAccounts(ctx)
		}
	}
}

func (s *Service) CheckOfflineAccounts(ctx context.Context) {
	accounts, err := s.queries.GetAllAccounts(ctx)
	if err != nil {
		slog.Error("Get all accounts failed", slog.Any("error", err))
		return
	}

	for _, a := range accounts {
		if a.Status.Offline {
			continue
		}

		lastUpdates, err := s.queries.GetLastUpdateByAccountID(ctx, a.ID)
		if err != nil {
			slog.Error("Get last updates failed", slog.Any("error", err))
			return
		}

		if len(lastUpdates) < 1 {
			continue
		}

		lastUpdate := lastUpdates[0]
		if time.Since(lastUpdate.Created).Minutes() < offlineThresholdMinutes {
			continue
		}

		newStatus := a.Status
		newStatus.Offline = true

		if err = s.queries.UpdateAccountStatus(ctx, database.UpdateAccountStatusParams{
			ID:     a.ID,
			Status: newStatus,
		}); err != nil {
			slog.Error("Update account status failed", slog.Any("error", err))
			return
		}

		ruText := fmt.Sprintf("🚨 %s перестал присылать обновления", a.Name)
		s.alertService.Alert(ruText, nil)
	}
}
