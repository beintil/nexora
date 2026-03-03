package cron

import (
	"context"
	"telephony/internal/modules/countries"
	"telephony/pkg/logger"
	"time"

	"github.com/robfig/cron/v3"
)

type updateCountriesCron struct {
	countryService countries.Service

	log logger.Logger
}

func NewUpdateCountriesCron(
	countryService countries.Service,

	log logger.Logger,
) Cron {
	return &updateCountriesCron{
		countryService: countryService,
		log:            log,
	}
}

func (m *updateCountriesCron) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	m.log.Info("update countries")
	err := m.countryService.SaveUpdateCountries(ctx)
	if err != nil {
		m.log.Error(err)
	}
	cancel()
	m.log.Info("update countries finished")

	c := cron.New()

	c.AddFunc("0 */10 * * *", func() {
		m.log.Info("update countries")
		ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)

		err := m.countryService.SaveUpdateCountries(ctx)
		if err != nil {
			m.log.Error(err)
		}
		cancel()

		m.log.Info("update countries finished")
	})
	c.Start()
}
