package rpc_test

import (
	"context"
	"os"
	"testing"

	"github.com/thingful/twirp-devicereg-go"

	kitlog "github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/thingful/iotdevicereg/pkg/postgres"
	"github.com/thingful/iotdevicereg/pkg/rpc"
	"github.com/thingful/iotdevicereg/pkg/system"
)

type DeviceRegistrationSuite struct {
	suite.Suite

	db     postgres.DB
	logger kitlog.Logger
}

func (s *DeviceRegistrationSuite) SetupTest() {
	connStr := os.Getenv("DEVICEREG_DATABASE_URL")

	s.logger = kitlog.NewNopLogger()

	s.db = postgres.NewDB(
		&postgres.Config{
			ConnStr:            connStr,
			EncryptionPassword: "password",
		},
		s.logger,
	)

	err := s.db.(system.Startable).Start()
	if err != nil {
		s.T().Fatalf("Failed to start db: %v", err)
	}

	err = s.db.MigrateDownAll()
	if err != nil {
		s.T().Fatalf("Failed to migrate db down: %v", err)
	}

	err = s.db.MigrateUp()
	if err != nil {
		s.T().Fatalf("Failed to migrate db up: %v", err)
	}
}

func (s *DeviceRegistrationSuite) TearDownTest() {
	err := s.db.(system.Stoppable).Stop()
	if err != nil {
		s.T().Fatalf("Failed to stop db: %v", err)
	}
}

func (s *DeviceRegistrationSuite) TestFullLifecycle() {
	dr := rpc.NewDeviceReg(s.db, s.logger)
	err := dr.(system.Startable).Start()
	assert.Nil(s.T(), err, "starting devicereg")
	defer func() {
		err := dr.(system.Stoppable).Stop()
		assert.Nil(s.T(), err, "stopping devicereg")
	}()

	claimResp, err := dr.ClaimDevice(context.Background(), &devicereg.ClaimDeviceRequest{
		DeviceToken: "abc123",
		UserUid:     "alice",
		Location: &devicereg.ClaimDeviceRequest_Location{
			Longitude: 12.2,
			Latitude:  32.1,
		},
		Disposition: devicereg.ClaimDeviceRequest_INDOOR,
	})

	assert.Nil(s.T(), err, "claiming device")
	assert.NotEqual(s.T(), "", claimResp.DevicePublicKey)
	assert.NotEqual(s.T(), "", claimResp.UserPrivateKey)
	assert.NotEqual(s.T(), "", claimResp.UserPublicKey)

	_, err = dr.RevokeDevice(context.Background(), &devicereg.RevokeDeviceRequest{
		DeviceToken:   "abc123",
		UserPublicKey: claimResp.UserPublicKey,
	})
	assert.Nil(s.T(), err)
}

func (s *DeviceRegistrationSuite) TestInvalidClaimRequests() {
	dr := rpc.NewDeviceReg(s.db, s.logger)
	err := dr.(system.Startable).Start()
	assert.Nil(s.T(), err, "starting devicereg")
	defer func() {
		err := dr.(system.Stoppable).Stop()
		assert.Nil(s.T(), err, "stopping devicereg")
	}()

	testcases := []struct {
		label       string
		req         *devicereg.ClaimDeviceRequest
		expectedErr string
	}{
		{
			label: "missing device token",
			req: &devicereg.ClaimDeviceRequest{
				//DeviceToken: "abc123",
				UserUid: "alice",
				Location: &devicereg.ClaimDeviceRequest_Location{
					Longitude: 12.2,
					Latitude:  32.1,
				},
				Disposition: devicereg.ClaimDeviceRequest_INDOOR,
			},
			expectedErr: "twirp error invalid_argument: device_token is required",
		},
		{
			label: "missing user_uid",
			req: &devicereg.ClaimDeviceRequest{
				DeviceToken: "abc123",
				//UserUid: "alice",
				Location: &devicereg.ClaimDeviceRequest_Location{
					Longitude: 12.2,
					Latitude:  32.1,
				},
				Disposition: devicereg.ClaimDeviceRequest_INDOOR,
			},
			expectedErr: "twirp error invalid_argument: user_uid is required",
		},
		{
			label: "missing location",
			req: &devicereg.ClaimDeviceRequest{
				DeviceToken: "abc123",
				UserUid:     "alice",
				//Location: &devicereg.ClaimDeviceRequest_Location{
				//	Longitude: 12.2,
				//	Latitude:  32.1,
				//},
				Disposition: devicereg.ClaimDeviceRequest_INDOOR,
			},
			expectedErr: "twirp error invalid_argument: location is required",
		},
		{
			label: "missing longitude",
			req: &devicereg.ClaimDeviceRequest{
				DeviceToken: "abc123",
				UserUid:     "alice",
				Location: &devicereg.ClaimDeviceRequest_Location{
					//	Longitude: 12.2,
					Latitude: 32.1,
				},
				Disposition: devicereg.ClaimDeviceRequest_INDOOR,
			},
			expectedErr: "twirp error invalid_argument: longitude is required",
		},
		{
			label: "missing latitude",
			req: &devicereg.ClaimDeviceRequest{
				DeviceToken: "abc123",
				UserUid:     "alice",
				Location: &devicereg.ClaimDeviceRequest_Location{
					Longitude: 12.2,
					//Latitude:  32.1,
				},
				Disposition: devicereg.ClaimDeviceRequest_INDOOR,
			},
			expectedErr: "twirp error invalid_argument: latitude is required",
		},
		{
			label: "invalid large longitude",
			req: &devicereg.ClaimDeviceRequest{
				DeviceToken: "abc123",
				UserUid:     "alice",
				Location: &devicereg.ClaimDeviceRequest_Location{
					Longitude: 180.1,
					Latitude:  32.1,
				},
				Disposition: devicereg.ClaimDeviceRequest_INDOOR,
			},
			expectedErr: "twirp error internal: longitude must be between -180 and 180 degrees",
		},
		{
			label: "invalid small longitude",
			req: &devicereg.ClaimDeviceRequest{
				DeviceToken: "abc123",
				UserUid:     "alice",
				Location: &devicereg.ClaimDeviceRequest_Location{
					Longitude: -180.1,
					Latitude:  32.1,
				},
				Disposition: devicereg.ClaimDeviceRequest_INDOOR,
			},
			expectedErr: "twirp error internal: longitude must be between -180 and 180 degrees",
		},
		{
			label: "invalid large latitude",
			req: &devicereg.ClaimDeviceRequest{
				DeviceToken: "abc123",
				UserUid:     "alice",
				Location: &devicereg.ClaimDeviceRequest_Location{
					Longitude: 80.1,
					Latitude:  90.1,
				},
				Disposition: devicereg.ClaimDeviceRequest_INDOOR,
			},
			expectedErr: "twirp error internal: latitude must be between -90 and 90 degrees",
		},
		{
			label: "invalid small latitude",
			req: &devicereg.ClaimDeviceRequest{
				DeviceToken: "abc123",
				UserUid:     "alice",
				Location: &devicereg.ClaimDeviceRequest_Location{
					Longitude: 80.1,
					Latitude:  -90.1,
				},
				Disposition: devicereg.ClaimDeviceRequest_INDOOR,
			},
			expectedErr: "twirp error internal: latitude must be between -90 and 90 degrees",
		},
	}

	for _, tc := range testcases {
		s.T().Run(tc.label, func(t *testing.T) {
			_, err := dr.ClaimDevice(context.Background(), tc.req)
			assert.NotNil(t, err)
			assert.Equal(t, tc.expectedErr, err.Error())
		})
	}
}

func (s *DeviceRegistrationSuite) TestInvalidRevokeRequests() {
	dr := rpc.NewDeviceReg(s.db, s.logger)
	err := dr.(system.Startable).Start()
	assert.Nil(s.T(), err, "starting devicereg")
	defer func() {
		err := dr.(system.Stoppable).Stop()
		assert.Nil(s.T(), err, "stopping devicereg")
	}()

	claimResp, err := dr.ClaimDevice(context.Background(), &devicereg.ClaimDeviceRequest{
		DeviceToken: "abc123",
		UserUid:     "alice",
		Location: &devicereg.ClaimDeviceRequest_Location{
			Longitude: 12.2,
			Latitude:  32.1,
		},
		Disposition: devicereg.ClaimDeviceRequest_INDOOR,
	})

	assert.Nil(s.T(), err)

	testcases := []struct {
		label       string
		req         *devicereg.RevokeDeviceRequest
		expectedErr string
	}{
		{
			label: "invalid user public key",
			req: &devicereg.RevokeDeviceRequest{
				DeviceToken:   "abc123",
				UserPublicKey: "foobar",
			},
			expectedErr: "twirp error internal: failed to delete device: sql: no rows in result set",
		},
		{
			label: "invalid device token",
			req: &devicereg.RevokeDeviceRequest{
				DeviceToken:   "foobar",
				UserPublicKey: claimResp.UserPublicKey,
			},
			expectedErr: "twirp error internal: failed to delete device: sql: no rows in result set",
		},
		{
			label: "missing device token",
			req: &devicereg.RevokeDeviceRequest{
				//DeviceToken:   "foobar",
				UserPublicKey: claimResp.UserPublicKey,
			},
			expectedErr: "twirp error invalid_argument: device_token is required",
		},
		{
			label: "missing device token",
			req: &devicereg.RevokeDeviceRequest{
				DeviceToken: "abc123",
				//UserPublicKey: claimResp.UserPublicKey,
			},
			expectedErr: "twirp error invalid_argument: user_public_key is required",
		},
	}

	for _, tc := range testcases {
		s.T().Run(tc.label, func(t *testing.T) {
			_, err := dr.RevokeDevice(context.Background(), tc.req)
			assert.NotNil(t, err)
			assert.Equal(t, tc.expectedErr, err.Error())
		})
	}
}

func TestRunDeviceRegSuite(t *testing.T) {
	suite.Run(t, new(DeviceRegistrationSuite))
}
