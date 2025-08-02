package controller

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do"
	"github.com/samber/oops"
	"net/http"
	"roflbeacon2/app/api"
	"roflbeacon2/app/service/account"
	"roflbeacon2/app/service/ingest"
	"roflbeacon2/app/service/limits"
	"roflbeacon2/pkg/config"
	"roflbeacon2/pkg/database"
)

var _ api.StrictServerInterface = (*Server)(nil)

type Server struct {
	appCtx         context.Context
	cfg            *config.Config
	dbConn         *pgxpool.Pool
	queries        *database.Queries
	accountService *account.Service
	limitsService  *limits.Service
	ingestService  *ingest.Service
}

func NewStrictServer(di *do.Injector) *Server {
	return &Server{
		appCtx:         do.MustInvoke[context.Context](di),
		cfg:            do.MustInvoke[*config.Config](di),
		dbConn:         do.MustInvoke[*pgxpool.Pool](di),
		queries:        do.MustInvoke[*database.Queries](di),
		accountService: do.MustInvoke[*account.Service](di),
		limitsService:  do.MustInvoke[*limits.Service](di),
		ingestService:  do.MustInvoke[*ingest.Service](di),
	}
}

func (s *Server) IngestUpdate(ctx context.Context, request api.IngestUpdateRequestObject) (api.IngestUpdateResponseObject, error) {
	if !s.limitsService.AllowIpRps(ctx, "ingest_update", 3) {
		return nil, oops.With("statusCode", http.StatusTooManyRequests).New("Too many requests")
	}

	if acc := s.accountService.ExtractCtxAccount(ctx); acc == nil {
		return nil, oops.With("statusCode", http.StatusForbidden).New("Forbidden")
	}

	if err := s.ingestService.Ingest(ctx, *request.Body); err != nil {
		return nil, err
	}

	return api.IngestUpdate200Response{}, nil
}
