package rpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	devicereg "github.com/thingful/twirp-devicereg-go"
	encoder "github.com/thingful/twirp-encoder-go"
	"github.com/twitchtv/twirp"

	"github.com/thingful/iotdevicereg/pkg/postgres"
)

var (
	// encoderWrites is a prometheus histogram recording writes and durations of
	// calls to the encoder keyed by method and status.
	encoderWrites = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "decode_encoder_rpc",
			Help: "Histogram of calls to the encoder RPC endpoint",
		},
		[]string{
			// which method are we calling
			"method",
			// is the status of the request: success or error
			"status",
		},
	)
)

func init() {
	prometheus.MustRegister(encoderWrites)
}

// deviceRegImpl is our implementation of the device registration rpc server
type deviceRegImpl struct {
	logger        kitlog.Logger
	db            postgres.DB
	encoderClient encoder.Encoder
	verbose       bool
}

// Config is a struct used to inject dependencies into our rpc service implementation
type Config struct {
	DB            postgres.DB
	EncoderClient encoder.Encoder
	Verbose       bool
}

// NewDeviceReg constructs a new DeviceRegistration instance. We pass in the
// components needed for this component to operate.
func NewDeviceReg(config *Config, logger kitlog.Logger) devicereg.DeviceRegistration {
	logger = kitlog.With(logger, "module", "rpc")
	logger.Log("msg", "creating devicereg")

	return &deviceRegImpl{
		db:            config.DB,
		encoderClient: config.EncoderClient,
		logger:        logger,
		verbose:       config.Verbose,
	}
}

// Start starts the service, doing any work that needs to be done
func (d *deviceRegImpl) Start() error {
	d.logger.Log("msg", "starting devicereg")
	return nil
}

// Stop stops the service, cleaning up anything that needs to be cleaned up.
func (d *deviceRegImpl) Stop() error {
	d.logger.Log("msg", "stopping devicereg")
	return nil
}

// ClaimDevice is our implementation of the ClaimDevice method defined for our
// Twirp service. As a result of this call the system as a whole should have
// created some key pairs for the user and the device, and created a new
// encrypted stream for the device on the encoder. THis service will maintain a
// store of the generated keys.
func (d *deviceRegImpl) ClaimDevice(ctx context.Context, req *devicereg.ClaimDeviceRequest) (_ *devicereg.ClaimDeviceResponse, err error) {
	device, err := createValidDevice(req)
	if err != nil {
		return nil, err
	}

	if d.verbose {
		d.logger.Log("method", "ClaimDevice", "deviceToken", req.DeviceToken, "broker", req.Broker)
	}

	tx, err := d.db.BeginTX()
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		if cerr := tx.Commit(); cerr != nil {
			err = twirp.InternalErrorWith(cerr)
		}
	}()

	// insert the device in the context of the current transaction
	device, err = d.db.RegisterDevice(tx, device)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	start := time.Now()

	// attempt to create the default stream - if this fails we can roll back the
	// transaction
	resp, err := d.encoderClient.CreateStream(ctx, &encoder.CreateStreamRequest{
		BrokerAddress:      req.Broker,
		DeviceTopic:        fmt.Sprintf("device/sck/%s/readings", req.DeviceToken),
		DevicePrivateKey:   device.PrivateKey,
		RecipientPublicKey: device.User.PublicKey,
		UserUid:            req.UserUid,
		Location: &encoder.CreateStreamRequest_Location{
			Longitude: req.Location.Longitude,
			Latitude:  req.Location.Latitude,
		},
		Disposition: encoder.CreateStreamRequest_Disposition(encoder.CreateStreamRequest_Disposition_value[req.Disposition.String()]),
	})

	duration := time.Since(start)

	if err != nil {
		encoderWrites.WithLabelValues("CreateStream", "error").Observe(duration.Seconds())

		return nil, twirp.InternalErrorWith(err)
	}

	encoderWrites.WithLabelValues("CreateStream", "success").Observe(duration.Seconds())

	err = d.db.CreateStream(tx, device.ID, resp.StreamUid)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	return &devicereg.ClaimDeviceResponse{
		UserPrivateKey:  device.User.PrivateKey,
		UserPublicKey:   device.User.PublicKey,
		DevicePublicKey: device.PublicKey,
	}, err
}

// RevokeDevice is our implementation of the method defined on the
// DeviceRegistration service interface. Calling this causes the specified
// device to be deleted, and the system will also call down to the encoder in
// order to delete associated streams.
func (d *deviceRegImpl) RevokeDevice(ctx context.Context, req *devicereg.RevokeDeviceRequest) (_ *devicereg.RevokeDeviceResponse, err error) {
	err = validateRevokeRequest(req)
	if err != nil {
		return nil, err
	}

	if d.verbose {
		d.logger.Log("method", "RevokeDevice", "deviceToken", req.DeviceToken)
	}

	tx, err := d.db.BeginTX()
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		if cerr := tx.Commit(); cerr != nil {
			err = twirp.InternalErrorWith(cerr)
		}
	}()

	streams, err := d.db.DeleteDevice(tx, req.DeviceToken, req.UserPublicKey)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	for _, stream := range streams {
		start := time.Now()

		_, err = d.encoderClient.DeleteStream(ctx, &encoder.DeleteStreamRequest{
			StreamUid: stream.UID,
		})

		duration := time.Since(start)

		if err != nil {
			encoderWrites.WithLabelValues("DeleteStream", "error").Observe(duration.Seconds())
			return nil, twirp.InternalErrorWith(err)
		}

		encoderWrites.WithLabelValues("DeleteStream", "success").Observe(duration.Seconds())
	}

	return &devicereg.RevokeDeviceResponse{}, err
}

// createValidDevice both validates the incoming request, and returns an
// instantiated Device object ready for saving.
func createValidDevice(req *devicereg.ClaimDeviceRequest) (*postgres.Device, error) {
	if req.DeviceToken == "" {
		return nil, twirp.RequiredArgumentError("device_token")
	}

	if req.Broker == "" {
		return nil, twirp.RequiredArgumentError("broker")
	}

	if req.UserUid == "" {
		return nil, twirp.RequiredArgumentError("user_uid")
	}

	if req.Location == nil {
		return nil, twirp.RequiredArgumentError("location")
	}

	if req.Location.Longitude == 0 {
		return nil, twirp.RequiredArgumentError("longitude")
	}

	if req.Location.Latitude == 0 {
		return nil, twirp.RequiredArgumentError("latitude")
	}

	if req.Location.Longitude < -180 || req.Location.Longitude > 180 {
		return nil, twirp.InternalError("longitude must be between -180 and 180 degrees")
	}

	if req.Location.Latitude < -90 || req.Location.Latitude > 90 {
		return nil, twirp.InternalError("latitude must be between -90 and 90 degrees")
	}

	return &postgres.Device{
		Token:       req.DeviceToken,
		Longitude:   req.Location.Longitude,
		Latitude:    req.Location.Latitude,
		Disposition: strings.ToLower(req.Disposition.String()),
		User: &postgres.User{
			UID: req.UserUid,
		},
	}, nil
}

// validateRevokeRequest validates the incoming request, returning an error if
// any required fields are missing.
func validateRevokeRequest(req *devicereg.RevokeDeviceRequest) error {
	if req.DeviceToken == "" {
		return twirp.RequiredArgumentError("device_token")
	}

	if req.UserPublicKey == "" {
		return twirp.RequiredArgumentError("user_public_key")
	}

	return nil
}
