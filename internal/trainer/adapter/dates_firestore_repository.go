package adapter

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/quangbach27/wild-workout/internal/trainer/app/query"
	"github.com/quangbach27/wild-workout/internal/trainer/domain/hour"
	"google.golang.org/api/iterator"
)

type DateModel struct {
	Date         time.Time   `firestore:"Date"`
	HasFreeHours bool        `firestore:"HasFreeHours"`
	Hours        []HourModel `firestore:"Hours"`
}

type HourModel struct {
	Available            bool      `firestore:"Available"`
	HasTrainingScheduled bool      `firestore:"HasTrainingScheduled"`
	Hour                 time.Time `firestore:"Hour"`
}

type DatesFirestoreRepository struct {
	firestoreClient *firestore.Client
	factoryConfig   hour.FactoryConfig
}

func (repo DatesFirestoreRepository) getTrainerHoursCollection() *firestore.CollectionRef {
	return repo.firestoreClient.Collection("trainer-hours")
}

func (repo DatesFirestoreRepository) AvailableHours(ctx context.Context, from time.Time, to time.Time) ([]query.Date, error) {
	iter := repo.
		getTrainerHoursCollection().
		Where("Date", ">=", from).
		Where("Date", "<=", to).
		Documents(ctx)

	var dates []query.Date

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		dateModel := DateModel{}
		if err := doc.DataTo(&dateModel); err != nil {
			return nil, err
		}
		dates = append(dates, dateModelToApp(dateModel))
	}

	dates = repo.addMissingDates(dates, from, to)

	return dates, nil
}

func (repo DatesFirestoreRepository) setDefaultAvailability(date query.Date, isMissingDate bool) query.Date {
	existingHours := make(map[int]bool)
	if !isMissingDate {
		for _, h := range date.Hours {
			existingHours[h.Hour.Hour()] = true
		}
	}

	for h := repo.factoryConfig.MinUtcHour; h <= repo.factoryConfig.MaxUtcHour; h++ {
		if !isMissingDate && existingHours[h] {
			continue
		}

		h := time.Date(date.Date.Year(), date.Date.Month(), date.Date.Day(), h, 0, 0, 0, time.UTC)

		date.Hours = append(date.Hours, query.Hour{
			Hour:                 h,
			Available:            false,
			HasTrainingScheduled: false,
		})
	}

	return date
}

func (repo DatesFirestoreRepository) addMissingDates(dates []query.Date, from, to time.Time) []query.Date {
	for day := from.UTC(); !day.After(to); day.AddDate(0, 0, 1) {
		found := false
		for _, date := range dates {
			if date.Date.Equal(day) {
				repo.setDefaultAvailability(date, false)
				found = true
				break
			}
		}

		if !found {
			date := query.Date{
				Date: day,
			}
			date = repo.setDefaultAvailability(date, true)
			dates = append(dates, date)
		}
	}

	return dates
}

func dateModelToApp(dateModel DateModel) query.Date {
	var hours []query.Hour
	for _, hourModel := range dateModel.Hours {
		hours = append(hours, query.Hour{
			Available:            hourModel.Available,
			HasTrainingScheduled: hourModel.HasTrainingScheduled,
			Hour:                 hourModel.Hour,
		})

	}

	return query.Date{
		HasFreeHours: dateModel.HasFreeHours,
		Date:         dateModel.Date,
		Hours:        hours,
	}
}
