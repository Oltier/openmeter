package streaming

import (
	"context"
	"time"

	"github.com/openmeterio/openmeter/pkg/models"
)

type QueryParams struct {
	From           *time.Time
	To             *time.Time
	Subject        []string
	GroupBySubject bool
	GroupBy        []string
	Aggregation    *models.MeterAggregation
	WindowSize     *models.WindowSize
}

type Connector interface {
	CreateMeter(ctx context.Context, namespace string, meter *models.Meter) error
	DeleteMeter(ctx context.Context, namespace string, meterSlug string) error
	QueryMeter(ctx context.Context, namespace string, meterSlug string, params *QueryParams) ([]*models.MeterValue, *models.WindowSize, error)
	ListMeterSubjects(ctx context.Context, namespace string, meterSlug string) ([]string, error)
	// Add more methods as needed ...
}
