package ports

import (
	"net/http"

	"github.com/go-chi/render"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/quangbach27/wild-workout/internal/common/server/httperr"
	"github.com/quangbach27/wild-workout/internal/trainer/app"
	"github.com/quangbach27/wild-workout/internal/trainer/app/query"
)

type HTTPServer struct {
	application app.Application
}

func (httpServer HTTPServer) GetTrainerAvailableHours(w http.ResponseWriter, r *http.Request, params GetTrainerAvailableHoursParams) {
	cmd := query.AvailableHours{
		From: params.DateFrom,
		To:   params.DateTo,
	}
	dateModels, err := httpServer.application.Queries.AvailableHoursHandler.Handle(r.Context(), cmd)
	if err != nil {
		httperr.RespondWithSlugError(err, w, r)
	}
	dates := dateModelToResponse(dateModels)
	render.Respond(w, r, dates)
}

func dateModelToResponse(dateModels []query.Date) []Date {
	var dates []Date
	for _, dateModel := range dateModels {
		var hours []Hour
		for _, hourModel := range dateModel.Hours {
			hours = append(hours, Hour{
				Available:            hourModel.Available,
				HasTrainingScheduled: hourModel.HasTrainingScheduled,
				Hour:                 hourModel.Hour,
			})
		}
		dates = append(dates, Date{
			Date:         openapi_types.Date{Time: dateModel.Date},
			HasFreeHours: dateModel.HasFreeHours,
			Hours:        hours,
		})
	}

	return dates
}
