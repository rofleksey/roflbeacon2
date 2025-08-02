package account

import (
	"context"
	_ "embed"
	"github.com/samber/do"
	"roflbeacon2/pkg/config"
	"roflbeacon2/pkg/database"
)

type Service struct {
	cfg     *config.Config
	queries *database.Queries
}

func New(di *do.Injector) (*Service, error) {
	return &Service{
		cfg:     do.MustInvoke[*config.Config](di),
		queries: do.MustInvoke[*database.Queries](di),
	}, nil
}

func (s *Service) ExtractCtxAccount(ctx context.Context) *database.Account {
	accountOpt := ctx.Value("account")
	if accountOpt == nil {
		return nil
	}

	account, ok := accountOpt.(*database.Account)
	if !ok {
		return nil
	}

	return account
}
