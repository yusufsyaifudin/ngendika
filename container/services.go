package container

import (
	"fmt"
	"github.com/sony/sonyflake"
	"github.com/yusufsyaifudin/ngendika/backend"
	"github.com/yusufsyaifudin/ngendika/internal/svc/appsvc"
	"github.com/yusufsyaifudin/ngendika/internal/svc/msgsvc"
	"github.com/yusufsyaifudin/ngendika/internal/svc/pnpsvc"
	"github.com/yusufsyaifudin/ngendika/pkg/uid"
	"time"
)

type Services interface {
	UIDGen() uid.UID
	App() appsvc.Service
	PushNotificationProvider() pnpsvc.Service
	Message() msgsvc.Service
}

type ServicesImpl struct {
	uidGen uid.UID
	app    appsvc.Service
	pnp    pnpsvc.Service
	msg    msgsvc.Service
}

var _ Services = (*ServicesImpl)(nil)

func SetupServices(svcCfg ConfigServices, repos Repositories) (svc *ServicesImpl, err error) {
	if repos == nil {
		err = fmt.Errorf("nil repositories on services preparation")
	}

	uidGen := sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime:      time.Date(2022, 9, 23, 0, 0, 0, 0, time.UTC),
		MachineID:      nil,
		CheckMachineID: nil,
	})

	if uidGen == nil {
		err = fmt.Errorf("uid generator is nil")
		return
	}

	// ** Prepare app service at once
	appRepo, err := repos.AppRepo(svcCfg.App.DBLabel)
	if err != nil {
		err = fmt.Errorf("services cannot get app repo: %w", err)
		return
	}

	appService, err := appsvc.New(appsvc.DefaultServiceConfig{
		UIDGen:  uidGen,
		AppRepo: appRepo,
	})
	if err != nil {
		err = fmt.Errorf("services cannot get prepare app service: %w", err)
		return
	}

	// ** Prepare push notification provider service at once
	pnpRepo, err := repos.PNProviderRepo(svcCfg.ServiceProvider.DBLabel)
	if err != nil {
		err = fmt.Errorf("services cannot get pnp repo: %w", err)
		return
	}

	pnpSvc, err := pnpsvc.New(pnpsvc.Config{
		UIDGen:  uidGen,
		PnpRepo: pnpRepo,
	})
	if err != nil {
		err = fmt.Errorf("services cannot get prepare pnp service: %w", err)
		return
	}

	// ** prepare message service
	msgSvc, err := msgsvc.New(msgsvc.SvcSyncConfig{
		AppSvc:        appService,
		PNProviderSvc: pnpSvc,
		PNSender:      backend.MuxBackend(),
		MaxBuffer:     svcCfg.Messaging.MaxBuffer,
		MaxWorker:     svcCfg.Messaging.MaxParallel,
	})
	if err != nil {
		err = fmt.Errorf("services cannot get prepare messaging service: %w", err)
		return
	}

	svc = &ServicesImpl{
		uidGen: uidGen,
		app:    appService,
		pnp:    pnpSvc,
		msg:    msgSvc,
	}

	return svc, nil
}

func (s *ServicesImpl) UIDGen() uid.UID {
	return s.uidGen
}

func (s *ServicesImpl) App() appsvc.Service {
	return s.app
}

func (s *ServicesImpl) PushNotificationProvider() pnpsvc.Service {
	return s.pnp
}

func (s *ServicesImpl) Message() msgsvc.Service {
	return s.msg
}
