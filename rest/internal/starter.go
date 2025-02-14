package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/shippomx/zard/core/logx"
	"github.com/shippomx/zard/core/proc"
	"github.com/shippomx/zard/internal/health"
)

const probeNamePrefix = "rest"

// StartOption defines the method to customize http.Server.
type StartOption func(svr *http.Server)

// StartHttp starts a http server.
// nolint revive
func StartHttp(host string, port int, handler http.Handler, probe health.Probe, opts ...StartOption) error {
	return start(host, port, handler, probe, func(svr *http.Server) error {
		return svr.ListenAndServe()
	}, opts...)
}

// StartHttps starts a https server.
// nolint revive
func StartHttps(host string, port int, certFile, keyFile string, handler http.Handler, probe health.Probe,
	opts ...StartOption,
) error {
	return start(host, port, handler, probe, func(svr *http.Server) error {
		// certFile and keyFile are set in buildHttpsServer
		return svr.ListenAndServeTLS(certFile, keyFile)
	}, opts...)
}

func start(host string, port int, handler http.Handler, probe health.Probe, run func(svr *http.Server) error,
	opts ...StartOption,
) (err error) {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: handler,
	}
	for _, opt := range opts {
		opt(server)
	}

	waitForCalled := proc.AddShutdownListener(func() {
		if e := server.Shutdown(context.Background()); e != nil {
			logx.Error(e)
		}
	})

	defer func() {
		if errors.Is(err, http.ErrServerClosed) {
			probe.MarkNotReady()
			waitForCalled()
		}
	}()
	probe.MarkReady()
	return run(server)
}
