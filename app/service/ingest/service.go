package ingest

import (
	"context"
	_ "embed"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/elliotchance/pie/v2"
	"github.com/samber/do"
	"log/slog"
	"roflbeacon2/app/api"
	"roflbeacon2/app/service/account"
	"roflbeacon2/app/service/alert"
	"roflbeacon2/pkg/config"
	"roflbeacon2/pkg/database"
	"roflbeacon2/pkg/util"
	"time"
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

func (s *Service) handleStillLocation(ctx context.Context, acc *database.Account, newLocation api.LocationData, wasInsideFences bool) error {
	if len(acc.Status.InsideFences) > 0 || wasInsideFences {
		acc.Status.StillLocation = nil

		return nil
	}

	if acc.Status.StillLocation != nil {
		oldLat := acc.Status.StillLocation.Latitude
		oldLon := acc.Status.StillLocation.Longitude
		newLat := newLocation.Latitude
		newLon := newLocation.Longitude

		dist := util.HaversineDistance(oldLat, oldLon, newLat, newLon)
		actualDist := dist - newLocation.Accuracy

		if actualDist <= standByRadius {
			return nil
		}

		acc.Status.StillLocation = nil

		ruText := fmt.Sprintf("▶️ %s снова начал двигаться", acc.Name)
		s.alertService.Alert(ruText, acc.ChatID)

		return nil
	}

	lastUpdates, err := s.queries.GetLastUpdateByAccountID(ctx, acc.ID)
	if err != nil {
		return fmt.Errorf("get last updates: %w", err)
	}

	if len(lastUpdates) == 0 {
		return nil
	}

	lastUpdate := lastUpdates[0]
	lastLocation := lastUpdate.Data.Location

	if lastLocation == nil {
		return nil
	}

	oldLat := lastLocation.Latitude
	oldLon := lastLocation.Longitude
	newLat := newLocation.Latitude
	newLon := newLocation.Longitude

	dist := util.HaversineDistance(oldLat, oldLon, newLat, newLon)
	actualDist := dist + newLocation.Accuracy

	if actualDist > standByRadius {
		return nil
	}

	acc.Status.StillLocation = &newLocation

	ruText := fmt.Sprintf("⏸️ %s остановился", acc.Name)
	s.alertService.Alert(ruText, acc.ChatID)

	return nil
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
	wasInsideFences := leftFences.Cardinality() > 0

	acc.Status.InsideFences = pie.Map(newFences.ToSlice(), func(f database.Fence) int64 {
		return f.ID
	})

	s.alertFenceMovement(acc, enteredFences, leftFences)

	if err = s.handleStillLocation(ctx, acc, data, wasInsideFences); err != nil {
		slog.ErrorContext(ctx, "Failed to handle still location",
			slog.Any("error", err),
		)
	}

	if err = s.queries.UpdateAccountStatus(ctx, database.UpdateAccountStatusParams{
		ID:     acc.ID,
		Status: acc.Status,
	}); err != nil {
		return fmt.Errorf("update account status: %w", err)
	}

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
