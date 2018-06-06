package rpc

import (
	"context"

	kitlog "github.com/go-kit/kit/log"
	devicereg "github.com/thingful/twirp-devicereg-go"
)

type deviceRegImpl struct {
	logger  kitlog.Logger
	verbose bool
}

func NewDeviceReg(logger kitlog.Logger) devicereg.DeviceRegistration {
	logger = kitlog.With(logger, "module", "rpc")
	logger.Log("msg", "creating devicereg")

	return &deviceRegImpl{
		logger:  logger,
		verbose: true,
	}
}

func (d *deviceRegImpl) ClaimDevice(ctx context.Context, req *devicereg.ClaimDeviceRequest) (*devicereg.ClaimDeviceResponse, error) {
	return nil, nil
}

func (d *deviceRegImpl) RevokeDevice(ctx context.Context, req *devicereg.RevokeDeviceRequest) (*devicereg.RevokeDeviceResponse, error) {
	return nil, nil
}
