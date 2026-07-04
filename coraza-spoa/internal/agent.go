package internal

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/dropmorepackets/haproxy-go/pkg/encoding"
	"github.com/dropmorepackets/haproxy-go/spop"
	"github.com/rs/zerolog"
)

type Agent struct {
	Context      context.Context
	Applications map[string]*Application
	Logger       zerolog.Logger

	mtx sync.RWMutex
}

func (a *Agent) Serve(l net.Listener) error {
	agent := spop.Agent{
		Handler:     a,
		BaseContext: a.Context,
	}

	return agent.Serve(l)
}

func (a *Agent) ReplaceApplications(newApps map[string]*Application) {
	a.mtx.Lock()
	a.Applications = newApps
	a.mtx.Unlock()
}

func (a *Agent) HandleSPOE(ctx context.Context, writer *encoding.ActionWriter, message *encoding.Message) {
	const (
		messageCorazaRequest  = "coraza-req"
		messageCorazaResponse = "coraza-res"
	)

	var messageHandler func(*Application, context.Context, *encoding.ActionWriter, *encoding.Message) error
	switch name := string(message.NameBytes()); name {
	case messageCorazaRequest:
		messageHandler = (*Application).HandleRequest
	case messageCorazaResponse:
		messageHandler = (*Application).HandleResponse
	default:
		a.Logger.Debug().Str("message", name).Msg("unknown spoe message")
		return
	}

	k := encoding.AcquireKVEntry()
	defer encoding.ReleaseKVEntry(k)
	if !message.KV.Next(k) {
		a.Logger.Error().Msg("failed reading kv entry — SPOE message has no app identifier")
		return
	}

	appName := string(k.ValueBytes())
	if !k.NameEquals("app") {
		// Without knowing the app, we cannot continue. Log the error and return gracefully
		// instead of panicking to avoid dropping the SPOE connection.
		a.Logger.Error().Str("expected", "app").Str("got", string(k.NameBytes())).Msg("unexpected kv entry — first key must be 'app'")
		return
	}

	a.mtx.RLock()
	app := a.Applications[appName]
	a.mtx.RUnlock()
	if app == nil {
		// If we cannot resolve the app, log the error and return gracefully
		// instead of panicking. This can happen during config transitions.
		a.Logger.Error().Str("app", appName).Msg("app not found — skipping message processing")
		return
	}

	err := messageHandler(app, ctx, writer, message)
	if err == nil {
		return
	}

	var interruption ErrInterrupted
	if errors.As(err, &interruption) {
		_ = writer.SetInt64(encoding.VarScopeTransaction, "status", int64(interruption.Interruption.Status))
		_ = writer.SetString(encoding.VarScopeTransaction, "action", interruption.Interruption.Action)
		_ = writer.SetString(encoding.VarScopeTransaction, "data", interruption.Interruption.Data)
		_ = writer.SetInt64(encoding.VarScopeTransaction, "ruleid", int64(interruption.Interruption.RuleID))

		a.Logger.Debug().Err(err).Msg("sending interruption")
		return
	}

	// error is not ErrInterrupted — log and return gracefully instead of
	// panicking, to avoid dropping the entire SPOE connection on transient failures.
	a.Logger.Error().Err(err).Msg("Error handling request")
}
