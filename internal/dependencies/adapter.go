package dependencies

import (
	"github.com/warehouse/ai-service/internal/adapter/auth"
	"github.com/warehouse/ai-service/internal/adapter/mail"
	"github.com/warehouse/ai-service/internal/adapter/random"
	"github.com/warehouse/ai-service/internal/adapter/time"

	"go.uber.org/zap"
)

func (d *dependencies) TimeAdapter() time.Adapter {
	if d.timeAdapter == nil {
		d.timeAdapter = time.NewAdapter(
			d.cfg.Time.Locale,
		)
	}

	return d.timeAdapter
}

func (d *dependencies) RandomAdapter() random.Adapter {
	if d.randomAdapter == nil {
		d.randomAdapter = random.NewAdapter()
	}
	return d.randomAdapter
}

func (d *dependencies) AuthAdapter() auth.Adapter {
	if d.authAdapter == nil {
		var err error
		if d.authAdapter, err = auth.NewAdapter(d.cfg.Grpc); err != nil {
			d.log.Zap().Panic("create auth grpc adapter", zap.Error(err))
		}
	}

	return d.authAdapter
}

func (d *dependencies) MailAdapter() mail.Adapter {
	if d.mailAdapter == nil {
		var err error
		if d.mailAdapter, err = mail.NewAdapter(d.cfg.Rabbit.MailQueue, d.RabbitClient()); err != nil {
			d.log.Zap().Panic("create mail broker adapter", zap.Error(err))
		}
	}

	return d.mailAdapter
}
