package router

import (
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/render"

	"github.com/openmeterio/openmeter/api"
	"github.com/openmeterio/openmeter/internal/streaming"
	"github.com/openmeterio/openmeter/pkg/contextx"
	"github.com/openmeterio/openmeter/pkg/filter"
	"github.com/openmeterio/openmeter/pkg/models"
)

// QueryMeter queries the values stored for a meter.
func (a *Router) QueryMeter(w http.ResponseWriter, r *http.Request, meterIDOrSlug string, params api.QueryMeterParams) {
	ctx := contextx.WithAttr(r.Context(), "operation", "queryMeter")
	ctx = contextx.WithAttr(ctx, "id", meterIDOrSlug)
	ctx = contextx.WithAttr(ctx, "params", params) // TODO: we should probable NOT add this to the context

	namespace := a.config.NamespaceManager.GetDefaultNamespace()

	// Get meter
	meter, err := a.config.Meters.GetMeterByIDOrSlug(ctx, namespace, meterIDOrSlug)
	if err != nil {
		if _, ok := err.(*models.MeterNotFoundError); ok {
			err := fmt.Errorf("meter not found: %w", err)

			models.NewStatusProblem(ctx, err, http.StatusNotFound).Respond(w, r)

			return
		}

		err := fmt.Errorf("get meter: %w", err)

		a.config.ErrorHandler.HandleContext(ctx, err)
		models.NewStatusProblem(ctx, err, http.StatusInternalServerError).Respond(w, r)

		return
	}

	a.QueryMeterWithMeter(ctx, w, r, slog.Default(), meter, params)
}

// QueryMeter queries the values stored for a meter.
func (a *Router) QueryMeterWithMeter(ctx context.Context, w http.ResponseWriter, r *http.Request, logger *slog.Logger, meter models.Meter, params api.QueryMeterParams) {
	// Query Params
	queryParams := &streaming.QueryParams{
		From:        params.From,
		To:          params.To,
		WindowSize:  params.WindowSize,
		Aggregation: meter.Aggregation,
	}

	if params.GroupBy != nil {
		for _, groupBy := range *params.GroupBy {
			// Validate group by, `subject` is a special group by
			if ok := groupBy == "subject" || meter.GroupBy[groupBy] != ""; !ok {
				err := fmt.Errorf("invalid group by: %s", groupBy)

				models.NewStatusProblem(ctx, err, http.StatusBadRequest).Respond(w, r)

				return
			}

			queryParams.GroupBy = append(queryParams.GroupBy, groupBy)
		}
	}

	// Subject is a special query parameter which both filters and groups by subject(s)
	if params.Subject != nil {
		subjects := []string{}
		for _, subject := range *params.Subject {
			subjects = append(subjects, fmt.Sprintf(`"%s"`, subject))
		}

		f, _ := filter.ToFilter(fmt.Sprintf(`{"$in": [%s]}`, strings.Join(subjects, ", ")))
		queryParams.FilterSubject = &f

		// Add subject to group by if not already present
		if !slices.Contains(queryParams.GroupBy, "subject") {
			queryParams.GroupBy = append(queryParams.GroupBy, "subject")
		}
	}

	if params.WindowTimeZone != nil {
		tz, err := time.LoadLocation(*params.WindowTimeZone)
		if err != nil {
			err := fmt.Errorf("invalid time zone: %w", err)

			models.NewStatusProblem(ctx, err, http.StatusBadRequest).Respond(w, r)

			return
		}
		queryParams.WindowTimeZone = tz
	}

	if params.Filter != nil {
		for k, paramFilter := range *params.Filter {
			// TODO: ideally `paramFilter` would be `filter.Filter` type but the OpenAPI parser
			// doesn't support complext objects in query parameters so we have to parse it manually from string.
			// With this we also loose the ability to validate the filter in the OpenAPI schema and we have to do it manually here.
			f, err := filter.ToFilter(paramFilter)
			if err != nil {
				err := fmt.Errorf(`invalid "%s" filter (%s): %w`, k, paramFilter, err)
				models.NewStatusProblem(ctx, err, http.StatusBadRequest).Respond(w, r)
				return
			}

			err = filter.Validate(f)
			if err != nil {
				err := fmt.Errorf("invalid %s filter (%s): %w", k, paramFilter, err)
				models.NewStatusProblem(ctx, err, http.StatusBadRequest).Respond(w, r)
				return
			}

			// Subject filters
			if k == "subject" {
				queryParams.FilterSubject = &f
				continue
			}

			// GroupBy filters
			if _, ok := meter.GroupBy[k]; ok {
				if queryParams.FilterGroupBy == nil {
					queryParams.FilterGroupBy = map[string]filter.Filter{}
				}

				queryParams.FilterGroupBy[k] = f
				continue
			} else {
				err := fmt.Errorf("invalid group by filter: %s", k)
				models.NewStatusProblem(ctx, err, http.StatusBadRequest).Respond(w, r)
				return
			}
		}
	}

	if err := queryParams.Validate(meter.WindowSize); err != nil {
		err := fmt.Errorf("invalid query parameters: %w", err)

		models.NewStatusProblem(ctx, err, http.StatusBadRequest).Respond(w, r)

		return
	}

	// Query connector
	data, err := a.config.StreamingConnector.QueryMeter(ctx, meter.Namespace, meter.Slug, queryParams)
	if err != nil {
		err := fmt.Errorf("query meter: %w", err)

		a.config.ErrorHandler.HandleContext(ctx, err)
		models.NewStatusProblem(ctx, err, http.StatusInternalServerError).Respond(w, r)

		return
	}

	resp := &QueryMeterResponse{
		WindowSize: params.WindowSize,
		From:       params.From,
		To:         params.To,
		Data:       data,
	}

	// If total data is queried for a period, replace the window start and end with the period for each row
	if params.WindowSize == nil {
		for i := range resp.Data {
			if params.From != nil {
				resp.Data[i].WindowStart = *params.From
			}
			if params.To != nil {
				resp.Data[i].WindowEnd = *params.To
			}
		}
	}

	// Parse media type
	accept := r.Header.Get("Accept")
	if accept == "" {
		accept = "application/json"
	}
	mediatype, _, err := mime.ParseMediaType(accept)
	// Browser can send back media type Go marks as invalid
	// If that happens, default to JSON
	if err != nil {
		logger.Debug("invalid media type, default to json", "error", err)
		mediatype = "application/json"
	}

	if mediatype == "text/csv" {
		resp.RenderCSV(w, r, queryParams.GroupBy, meter.Slug)
	} else {
		_ = render.Render(w, r, resp)
	}
}

// QueryMeterResponse is returned by the QueryMeter endpoint.
type QueryMeterResponse struct {
	WindowSize *models.WindowSize     `json:"windowSize,omitempty"`
	From       *time.Time             `json:"from,omitempty"`
	To         *time.Time             `json:"to,omitempty"`
	Data       []models.MeterQueryRow `json:"data"`
}

// Render implements the chi renderer interface.
func (resp QueryMeterResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// RenderCSV renders the response as CSV.
func (resp QueryMeterResponse) RenderCSV(w http.ResponseWriter, r *http.Request, groupByKeys []string, meterIDOrSlug string) {
	records := [][]string{}

	// Filter out the subject from the group by keys
	dataGroupByKeys := make([]string, 0, len(groupByKeys))
	for _, k := range groupByKeys {
		if k == "subject" {
			continue
		}
		dataGroupByKeys = append(dataGroupByKeys, k)
	}

	// CSV headers
	headers := []string{"window_start", "window_end", "subject"}
	if len(dataGroupByKeys) > 0 {
		headers = append(headers, dataGroupByKeys...)
	}
	headers = append(headers, "value")
	records = append(records, headers)

	// CSV data
	for _, row := range resp.Data {
		data := []string{row.WindowStart.Format(time.RFC3339), row.WindowEnd.Format(time.RFC3339)}
		if row.Subject != nil {
			data = append(data, *row.Subject)
		} else {
			data = append(data, "")
		}
		for _, k := range dataGroupByKeys {
			var groupByValue string

			if row.GroupBy[k] != nil {
				groupByValue = *row.GroupBy[k]
			}
			data = append(data, groupByValue)
		}
		data = append(data, fmt.Sprintf("%f", row.Value))
		records = append(records, data)
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", meterIDOrSlug))

	// Write response
	writer := csv.NewWriter(w)
	err := writer.WriteAll(records)
	if err != nil {
		slog.Error("writing record to csv", "error", err)
	}

	if err := writer.Error(); err != nil {
		slog.Error("writing csv", "error", err)
	}
}
