package extensions

import (
	sdk "agones.dev/agones/sdks/go"
	"context"
	quakenet "github.com/criticalstack/quake-kube/internal/quake/net"
	quakeserver "github.com/criticalstack/quake-kube/internal/quake/server"
	"github.com/pkg/errors"
	"log"
	"time"
)

type Agones struct {
	Server *quakeserver.Server
	sdk    *sdk.SDK
}

func (a *Agones) Start(ctx context.Context, server *quakeserver.Server) error {
	if err := a.InitSdk(); err != nil {
		return err
	}
	a.Server = server

	go a.StartHeathCheck(ctx)
	go a.TrackStatus(ctx)

	if err := server.Start(ctx); err != nil {
		return errors.Wrap(err, "failed to start server")
	}

	return nil
}

func (a *Agones) InitSdk() error {
	log.Println("[Agones] starting Agones SDK client")
	s, err := sdk.NewSDK()
	if err != nil {
		return err
	}

	a.sdk = s

	return nil
}

func (a *Agones) StartHeathCheck(ctx context.Context) {
	tick := time.Tick(2 * time.Second)
	maxAttempts := 0
	for {
		if err := a.sdk.Health(); err != nil {
			if maxAttempts > 5 {
				log.Fatalf("[Agones] could not send health ping: %v", err)
			}
			maxAttempts++
		} else {
			maxAttempts = 0
		}

		select {
		case <-ctx.Done():
			log.Print("[Agones] stopped health pings")
			return
		case <-tick:
		}
	}
}

func (a *Agones) TrackStatus(ctx context.Context) {
	tick := time.Tick(5 * time.Second)
	for {
		if status, err := quakenet.GetStatus(a.Server.Addr); err == nil {
			if err != nil {
				log.Printf("[Agones] failed to get status from server: %s", err.Error())
				continue
			}
			for _, p := range status.Players {
				log.Printf("[Agones] player: %s", p.Name)
				if _, err := a.sdk.Alpha().PlayerConnect(p.Name); err != nil {
					log.Printf("[Agones] failed to register player: %s", err.Error())
				}
			}
			log.Println("[Agones] status checked")
		}

		select {
		case <-ctx.Done():
			log.Print("[Agones] stopped status checks")
			return
		case <-tick:
		}
	}
}
