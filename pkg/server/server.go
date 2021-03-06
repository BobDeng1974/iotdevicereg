package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	kitlog "github.com/go-kit/kit/log"
	twrpprom "github.com/joneskoo/twirp-serverhook-prometheus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	devicereg "github.com/thingful/twirp-devicereg-go"
	encoder "github.com/thingful/twirp-encoder-go"

	"github.com/thingful/iotdevicereg/pkg/postgres"
	"github.com/thingful/iotdevicereg/pkg/rpc"
	"github.com/thingful/iotdevicereg/pkg/system"
)

// Config is a top level config object. Populated by viper in the command setup,
// we then pass down config to the right places.
type Config struct {
	ListenAddr         string
	ConnStr            string
	EncryptionPassword string
	EncoderAddr        string
	Verbose            bool
}

// Server is our top level type, contains all other components, is responsible
// for starting and stopping them in the correct order.
type Server struct {
	srv    *http.Server
	db     postgres.DB
	logger kitlog.Logger
}

// PulseHandler is the simplest possible handler function - used to expose an
// endpoint which a load balancer can ping to verify that a node is running and
// accepting connections.
func PulseHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

// NewServer returns a new HTTP server. Is also responsible for constructing all
// components, and injecting them into the right place. This perhaps belongs
// elsewhere, but leaving here for now.
func NewServer(config *Config, logger kitlog.Logger) *Server {
	db := postgres.NewDB(&postgres.Config{
		ConnStr:            config.ConnStr,
		EncryptionPassword: config.EncryptionPassword,
	}, logger)

	encoderClient := encoder.NewEncoderProtobufClient(
		config.EncoderAddr,
		&http.Client{
			Timeout: time.Second * 10,
		},
	)

	deviceReg := rpc.NewDeviceReg(&rpc.Config{
		DB:            db,
		EncoderClient: encoderClient,
		Verbose:       config.Verbose,
	}, logger)

	hooks := twrpprom.NewServerHooks(nil)

	logger = kitlog.With(logger, "module", "server")
	logger.Log("msg", "creating server", "encoder", config.EncoderAddr)

	twirpHandler := devicereg.NewDeviceRegistrationServer(deviceReg, hooks)

	// multiplex twirp handler into a mux with our other handlers
	mux := http.NewServeMux()
	mux.Handle(devicereg.DeviceRegistrationPathPrefix, twirpHandler)
	mux.HandleFunc("/pulse", PulseHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// create our http.Server instance
	srv := &http.Server{
		Addr:    config.ListenAddr,
		Handler: mux,
	}

	// return the instantiated server
	return &Server{
		srv:    srv,
		db:     db,
		logger: logger,
	}
}

// Start starts the server running. This is responsible for starting components
// in the correct order, and in addition we attempt to run all up migrations as
// we start.
//
// We also create a channel listening for interrupt signals before gracefully
// shutting down.
func (s *Server) Start() error {
	// start the postgres connection pool
	err := s.db.(system.Startable).Start()
	if err != nil {
		return errors.Wrap(err, "failed to start db")
	}

	// migrate up the database
	err = s.db.MigrateUp()
	if err != nil {
		return errors.Wrap(err, "failed to migrate the database")
	}

	// add signal handling stuff to shutdown gracefully
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	go func() {
		s.logger.Log("listenAddr", s.srv.Addr, "msg", "starting server")
		if err := s.srv.ListenAndServe(); err != nil {
			s.logger.Log("err", err)
			os.Exit(1)
		}
	}()

	<-stopChan
	return s.Stop()
}

func (s *Server) Stop() error {
	s.logger.Log("msg", "stopping")
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	err := s.db.(system.Stoppable).Stop()
	if err != nil {
		return err
	}

	return s.srv.Shutdown(ctx)
}
