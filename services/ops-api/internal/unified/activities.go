package unified

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.temporal.io/sdk/client"
)

type Activities struct {
	temporalClient client.Client
	unifiedSDK     *Service
	dbpool         *pgxpool.Pool
}

func NewActivities(temporalClient client.Client, unifiedService *Service, dbpool *pgxpool.Pool) *Activities {
	return &Activities{
		temporalClient: temporalClient,
		unifiedSDK:     unifiedService,
		dbpool:         dbpool,
	}
}
