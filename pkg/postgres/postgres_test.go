package postgres_test

import (
	"os"
	"testing"

	kitlog "github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/thingful/iotdevicereg/pkg/postgres"
	"github.com/thingful/iotencoder/pkg/system"
)

type PostgresSuite struct {
	suite.Suite
	db postgres.DB
}

func (s *PostgresSuite) SetupTest() {
	logger := kitlog.NewNopLogger()
	connStr := os.Getenv("DEVICEREG_DATABASE_URL")

	s.db = postgres.NewDB(
		&postgres.Config{
			ConnStr:            connStr,
			EncryptionPassword: "password",
		},
		logger,
	)

	s.db.(system.Startable).Start()

	err := s.db.MigrateDownAll()
	if err != nil {
		s.T().Fatalf("Failed to migrate db down: %v", err)
	}

	err = s.db.MigrateUp()
	if err != nil {
		s.T().Fatalf("Failed to migrate db up: %v", err)
	}
}

func (s *PostgresSuite) TearDownTest() {
	err := s.db.(system.Stoppable).Stop()
	if err != nil {
		s.T().Fatalf("failed to stop db component: %v", err)
	}
}

func (s *PostgresSuite) TestRoundTrip() {
	tx, err := s.db.BeginTX()
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), tx)

	device1, err := s.db.RegisterDevice(tx,
		&postgres.Device{
			Token:       "abc123",
			Longitude:   2.3,
			Latitude:    23.3,
			Disposition: "indoor",
			User: &postgres.User{
				UID: "alice",
			},
		})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), device1)

	assert.NotEqual(s.T(), 0, device1.ID)
	assert.NotEqual(s.T(), "", device1.PublicKey)
	assert.NotEqual(s.T(), "", device1.User.PrivateKey)
	assert.NotEqual(s.T(), "", device1.User.PublicKey)
	assert.NotEqual(s.T(), device1.PublicKey, device1.User.PublicKey)

	var count int
	err = tx.Get(&count, `SELECT COUNT(*) FROM users`)
	assert.Nil(s.T(), err)

	err = tx.Get(&count, `SELECT COUNT(*) FROM users`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, count)

	err = tx.Get(&count, `SELECT COUNT(*) FROM devices`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, count)

	device2, err := s.db.RegisterDevice(tx, &postgres.Device{
		Token:       "hij567",
		Longitude:   2.3,
		Latitude:    23.3,
		Disposition: "indoor",
		User: &postgres.User{
			UID: "alice",
		},
	})

	assert.Nil(s.T(), err)

	assert.NotEqual(s.T(), "", device2.PublicKey)
	assert.NotEqual(s.T(), "", device2.User.PrivateKey)
	assert.NotEqual(s.T(), "", device2.User.PublicKey)

	err = tx.Get(&count, `SELECT COUNT(*) FROM users`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, count)

	err = tx.Get(&count, `SELECT COUNT(*) FROM devices`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 2, count)

	err = s.db.DeleteDevice(tx, "abc123", device1.User.PublicKey)
	assert.Nil(s.T(), err)

	err = tx.Get(&count, `SELECT COUNT(*) FROM users`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, count)

	err = tx.Get(&count, `SELECT COUNT(*) FROM devices`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, count)

	err = s.db.DeleteDevice(tx, "hij567", device2.User.PublicKey)
	assert.Nil(s.T(), err)

	err = tx.Get(&count, `SELECT COUNT(*) FROM users`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, count)

	err = tx.Get(&count, `SELECT COUNT(*) FROM devices`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, count)

	err = tx.Rollback()
	assert.Nil(s.T(), err)
}

func (s *PostgresSuite) TestDuplicateTokenError() {
	tx, err := s.db.BeginTX()
	assert.Nil(s.T(), err)

	_, err = s.db.RegisterDevice(tx, &postgres.Device{
		Token:       "abc123",
		Longitude:   2.3,
		Latitude:    23.3,
		Disposition: "indoor",
		User: &postgres.User{
			UID: "alice",
		},
	})

	assert.Nil(s.T(), err)

	_, err = s.db.RegisterDevice(tx, &postgres.Device{
		Token:       "abc123",
		Longitude:   2.3,
		Latitude:    23.3,
		Disposition: "indoor",
		User: &postgres.User{
			UID: "alice",
		},
	})

	assert.NotNil(s.T(), err)

	tx.Rollback()
}

func (s *PostgresSuite) TestErrorDeletingDevices() {
	tx, err := s.db.BeginTX()
	assert.Nil(s.T(), err)

	_, err = s.db.RegisterDevice(tx, &postgres.Device{
		Token:       "abc123",
		Longitude:   2.3,
		Latitude:    23.3,
		Disposition: "indoor",
		User: &postgres.User{
			UID: "alice",
		},
	})
	assert.Nil(s.T(), err)

	err = s.db.DeleteDevice(tx, "hij567", "u_privkey")
	assert.NotNil(s.T(), err)

	err = s.db.DeleteDevice(tx, "abc123", "not_u_privkey")
	assert.NotNil(s.T(), err)

	tx.Rollback()
}

func (s *PostgresSuite) TestCreateStreams() {
	tx, err := s.db.BeginTX()
	assert.Nil(s.T(), err)

	device, err := s.db.RegisterDevice(tx, &postgres.Device{
		Token:       "abc123",
		Longitude:   2.3,
		Latitude:    23.3,
		Disposition: "indoor",
		User: &postgres.User{
			UID: "alice",
		},
	})

	assert.Nil(s.T(), err)

	err = s.db.CreateStream(tx, device.ID, "abc")
	assert.Nil(s.T(), err)

	err = s.db.CreateStream(tx, device.ID, "hij")
	assert.Nil(s.T(), err)

	var count int
	err = tx.Get(&count, `SELECT COUNT(*) FROM streams`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 2, count)

	err = s.db.CreateStream(tx, device.ID, "hij")
	assert.NotNil(s.T(), err)

	tx.Rollback()
}

func TestRunPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}
