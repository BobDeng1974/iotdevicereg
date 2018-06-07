package postgres_test

import (
	"os"
	"testing"

	kitlog "github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/thingful/iotdevicereg/pkg/postgres"
	"github.com/thingful/iotencoder/pkg/system"
)

type PostgresSuite struct {
	suite.Suite
	db    postgres.DB
	rawDb *sqlx.DB
}

func (s *PostgresSuite) SetupTest() {
	logger := kitlog.NewNopLogger()
	connStr := os.Getenv("DEVICEREG_DATABASE_URL")

	db, err := postgres.Open(connStr)
	if err != nil {
		s.T().Fatalf("Failed to open connection: %v", err)
	}

	s.rawDb = db

	err = postgres.MigrateDownAll(s.rawDb.DB, logger)
	if err != nil {
		s.T().Fatalf("Failed to migrate down db: %v", err)
	}

	err = postgres.MigrateUp(s.rawDb.DB, logger)
	if err != nil {
		s.T().Fatalf("Failed to migrate up db: %v", err)
	}

	s.db = postgres.NewDB(
		&postgres.Config{
			ConnStr:            connStr,
			EncryptionPassword: "password",
		},
		logger,
	)

	s.db.(system.Startable).Start()
}

func (s *PostgresSuite) TearDownTest() {
	err := s.db.(system.Stoppable).Stop()
	if err != nil {
		s.T().Fatalf("failed to stop db component: %v", err)
	}

	err = s.rawDb.Close()
	if err != nil {
		s.T().Fatalf("Failed to close raw DB: %v", err)
	}

}

func (s *PostgresSuite) TestRoundTrip() {
	device, err := s.db.RegisterDevice(&postgres.Device{
		Token:       "abc123",
		PrivateKey:  "d_privkey",
		PublicKey:   "d_pubkey",
		Longitude:   2.3,
		Latitude:    23.3,
		Disposition: "indoor",
		User: &postgres.User{
			UID:        "alice",
			PrivateKey: "u_privkey",
			PublicKey:  "u_pubkey",
		},
	})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "d_pubkey", device.PublicKey)
	assert.Equal(s.T(), "u_privkey", device.User.PrivateKey)
	assert.Equal(s.T(), "u_pubkey", device.User.PublicKey)

	var count int

	err = s.rawDb.Get(&count, `SELECT COUNT(*) FROM users`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, count)

	err = s.rawDb.Get(&count, `SELECT COUNT(*) FROM devices`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, count)

	_, err = s.db.RegisterDevice(&postgres.Device{
		Token:       "hij567",
		PrivateKey:  "d_privkey",
		PublicKey:   "d_pubkey",
		Longitude:   2.3,
		Latitude:    23.3,
		Disposition: "indoor",
		User: &postgres.User{
			UID:        "alice",
			PrivateKey: "u_privkey",
			PublicKey:  "u_pubkey",
		},
	})

	assert.Nil(s.T(), err)

	err = s.rawDb.Get(&count, `SELECT COUNT(*) FROM users`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, count)

	err = s.rawDb.Get(&count, `SELECT COUNT(*) FROM devices`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 2, count)

	err = s.db.DeleteDevice("abc123", "u_pubkey")
	assert.Nil(s.T(), err)

	err = s.rawDb.Get(&count, `SELECT COUNT(*) FROM users`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, count)

	err = s.rawDb.Get(&count, `SELECT COUNT(*) FROM devices`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, count)

	err = s.db.DeleteDevice("hij567", "u_pubkey")
	assert.Nil(s.T(), err)

	err = s.rawDb.Get(&count, `SELECT COUNT(*) FROM users`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, count)

	err = s.rawDb.Get(&count, `SELECT COUNT(*) FROM devices`)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, count)
}

func (s *PostgresSuite) TestDuplicateTokenError() {
	_, err := s.db.RegisterDevice(&postgres.Device{
		Token:       "abc123",
		PrivateKey:  "d_privkey",
		PublicKey:   "d_pubkey",
		Longitude:   2.3,
		Latitude:    23.3,
		Disposition: "indoor",
		User: &postgres.User{
			UID:        "alice",
			PrivateKey: "u_privkey",
			PublicKey:  "u_pubkey",
		},
	})

	assert.Nil(s.T(), err)

	_, err = s.db.RegisterDevice(&postgres.Device{
		Token:       "abc123",
		PrivateKey:  "d_privkey",
		PublicKey:   "d_pubkey",
		Longitude:   2.3,
		Latitude:    23.3,
		Disposition: "indoor",
		User: &postgres.User{
			UID:        "alice",
			PrivateKey: "u_privkey",
			PublicKey:  "u_pubkey",
		},
	})

	assert.NotNil(s.T(), err)
}

func (s *PostgresSuite) TestErrorDeletingDevices() {
	_, err := s.db.RegisterDevice(&postgres.Device{
		Token:       "abc123",
		PrivateKey:  "d_privkey",
		PublicKey:   "d_pubkey",
		Longitude:   2.3,
		Latitude:    23.3,
		Disposition: "indoor",
		User: &postgres.User{
			UID:        "alice",
			PrivateKey: "u_privkey",
			PublicKey:  "u_pubkey",
		},
	})
	assert.Nil(s.T(), err)

	err = s.db.DeleteDevice("hij567", "u_privkey")
	assert.NotNil(s.T(), err)

	err = s.db.DeleteDevice("abc123", "not_u_privkey")
	assert.NotNil(s.T(), err)
}

func TestRunPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}
