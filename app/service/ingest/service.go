package ingest

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"roflbeacon2/app/api"
	"roflbeacon2/app/service/account"
	"roflbeacon2/app/service/alert"
	"roflbeacon2/pkg/config"
	"roflbeacon2/pkg/database"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/elliotchance/pie/v2"
	"github.com/samber/do"
)

const standByRadius = 200

type Service struct {
	cfg            *config.Config
	queries        *database.Queries
	accountService *account.Service
	alertService   *alert.Service
}

func New(di *do.Injector) (*Service, error) {
	return &Service{
		cfg:            do.MustInvoke[*config.Config](di),
		queries:        do.MustInvoke[*database.Queries](di),
		accountService: do.MustInvoke[*account.Service](di),
		alertService:   do.MustInvoke[*alert.Service](di),
	}, nil
}

func (s *Service) alertFenceMovement(acc *database.Account, enteredFences mapset.Set[database.Fence], leftFences mapset.Set[database.Fence]) {
	for fence := range leftFences.Iter() {
		ruText := fmt.Sprintf("🔴 %s покинул %s", acc.Name, fence.Name)

		s.alertService.Alert(ruText, acc.ChatID)
	}

	for fence := range enteredFences.Iter() {
		ruText := fmt.Sprintf("🟢 %s вошел в %s", acc.Name, fence.Name)

		s.alertService.Alert(ruText, acc.ChatID)
	}
}

func (s *Service) handleNewLocation(ctx context.Context, data api.LocationData) error {
	acc := s.accountService.ExtractCtxAccount(ctx)
	if acc == nil {
		return fmt.Errorf("no account in context")
	}

	allFences, err := s.queries.GetAllFences(ctx)
	if err != nil {
		return fmt.Errorf("get all fences: %w", err)
	}

	oldFences := mapset.NewSet(pie.Filter(allFences, func(fence database.Fence) bool {
		return pie.Contains(acc.Status.InsideFences, fence.ID)
	})...)

	newFences := mapset.NewSet[database.Fence]()

	for _, fence := range allFences {
		if fence.Contains(data.Latitude, data.Longitude, 2*data.Accuracy) {
			newFences.Add(fence)
		}
	}

	leftFences := oldFences.Difference(newFences)
	enteredFences := newFences.Difference(oldFences)

	acc.Status.InsideFences = pie.Map(newFences.ToSlice(), func(f database.Fence) int64 {
		return f.ID
	})

	s.alertFenceMovement(acc, enteredFences, leftFences)

	return nil
}

func (s *Service) Ingest(ctx context.Context, data api.UpdateData) error {
	acc := s.accountService.ExtractCtxAccount(ctx)
	if acc == nil {
		return fmt.Errorf("no account in context")
	}

	if data.Location != nil {
		if err := s.handleNewLocation(ctx, *data.Location); err != nil {
			slog.WarnContext(ctx, "Failed to handle location",
				slog.Any("error", err),
			)
		}
	}

	acc.Status.Offline = false

	if err := s.queries.UpdateAccountStatus(ctx, database.UpdateAccountStatusParams{
		ID:     acc.ID,
		Status: acc.Status,
	}); err != nil {
		return fmt.Errorf("update account status: %w", err)
	}

	_, err := s.queries.CreateUpdate(ctx, database.CreateUpdateParams{
		AccountID: acc.ID,
		Created:   time.Now(),
		Data:      data,
	})
	if err != nil {
		return fmt.Errorf("failed to create update in DB: %w", err)
	}

	return nil
}
