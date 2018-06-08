package postgres

import (
	kitlog "github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/thingful/iotdevicereg/pkg/crypto"
)

// User is our exported local type that allows us to maniuplate User records
// stored in the database. Represents an individual DECODE user. A single User
// may register multiple devices.
type User struct {
	ID         int    `db:"id"`
	UID        string `db:"uid"`
	PrivateKey string `db:"private_key"`
	PublicKey  string `db:"public_key"`
}

// Device is our exported local type that allows us to maniuplate Device records
// that are stored in the DB. Represents an individual device that may be
// registered. A single user may register multiple devices.
type Device struct {
	ID          int     `db:"id"`
	Token       string  `db:"token"`
	PrivateKey  string  `db:"private_key"`
	PublicKey   string  `db:"public_key"`
	Longitude   float64 `db:"longitude"`
	Latitude    float64 `db:"latitude"`
	Disposition string  `db:"disposition"`

	User *User
}

// DB is our interface to Postgres. Exposes methods for inserting a new Device
// (and associated Stream), listing all Devices, getting an individual Device,
// and deleting a Stream
type DB interface {
	// RegisterDevice takes as input a pointer to an instantiated Device instance
	// populated from the incoming public ClaimDeviceRequest message. We persist
	// the associated user into the DB if they aren't already registered, and then
	// attempt to persist the Device. Attempting to register an already registered
	// device will return an error.
	//
	// TODO: what should we do if the device is
	// already registered by the same user? what should we do if the device is
	// already registered for another user?
	RegisterDevice(device *Device) (*Device, error)

	// DeleteDevice attempts to delete a device identified by its token and the
	// public key of the owning user. If after deleting a device successfully the
	// user has no remaining registered devices, we also delete the user record.
	DeleteDevice(token, publicKey string) error

	// MigrateUp is a helper method that attempts to run all up migrations against
	// the underlying Postgres DB or returns an error.
	MigrateUp() error

	// MigrateDownAll is a helper method that attempts to run all down migrations
	// against the underlying Postgres DB or returns an error.
	MigrateDownAll() error
}

// Open is a helper function that takes as input a connection string for a DB,
// and returns either a sqlx.DB instance or an error. This function is separated
// out to help with CLI tasks for managing migrations.
func Open(connStr string) (*sqlx.DB, error) {
	return sqlx.Open("postgres", connStr)
}

// db is our type that wraps an sqlx.DB instance and provides an API for the
// data access functions we require.
type db struct {
	connStr            string
	encryptionPassword []byte
	DB                 *sqlx.DB
	logger             kitlog.Logger
}

// Config is used to carry package local configuration for Postgres DB module.
type Config struct {
	ConnStr            string
	EncryptionPassword string
}

// NewDB creates a new DB instance with the given connection string. We also
// pass in a logger.
func NewDB(config *Config, logger kitlog.Logger) DB {
	logger = kitlog.With(logger, "module", "postgres")

	logger.Log("msg", "creating DB instance")

	return &db{
		connStr:            config.ConnStr,
		encryptionPassword: []byte(config.EncryptionPassword),
		logger:             logger,
	}
}

// Start creates our DB connection pool running returning an error if any
// failure occurs.
func (d *db) Start() error {
	d.logger.Log("msg", "starting postgres")

	db, err := Open(d.connStr)
	if err != nil {
		return errors.Wrap(err, "opening db connection failed")
	}

	d.DB = db

	return nil
}

// Stop closes the DB connection pool.
func (d *db) Stop() error {
	d.logger.Log("msg", "stopping postgres")

	return d.DB.Close()
}

// RegisterDevice is our implementation of the RegisterDevice method defined in
// our interface.
func (d *db) RegisterDevice(device *Device) (_ *Device, err error) {
	// user upsert sql, note we encrypt the private using postgres native symmetric
	// encryption
	sql := `INSERT INTO users
		(uid, private_key, public_key)
		VALUES
			(:uid,
			 pgp_sym_encrypt(:private_key, :encryption_password),
			 :public_key
			)
		ON CONFLICT (uid) DO UPDATE
		SET updated_at = NOW()
		RETURNING id, public_key, pgp_sym_decrypt(private_key, :encryption_password) AS private_key`

	userKeyPair, err := crypto.NewKeyPair()
	if err != nil {
		return nil, err
	}

	mapArgs := map[string]interface{}{
		"uid":                 device.User.UID,
		"private_key":         userKeyPair.PrivateKey,
		"public_key":          userKeyPair.PublicKey,
		"encryption_password": d.encryptionPassword,
	}

	tx, err := BeginTX(d.DB)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start transaction when inserting device")
	}

	defer func() {
		if cerr := tx.CommitOrRollback(); err == nil && cerr != nil {
			err = cerr
		}
	}()

	var user User

	// we use a Get for the upsert so we get back the user id
	err = tx.Get(&user, sql, mapArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to upsert user")
	}

	// now attempt to insert the device
	sql = `INSERT INTO devices
			(token, user_id, private_key, public_key, longitude, latitude, disposition)
		VALUES (
			:token,
			:user_id,
			pgp_sym_encrypt(:private_key, :encryption_password),
			:public_key,
			:longitude,
			:latitude,
			:disposition
	)`

	deviceKeyPair, err := crypto.NewKeyPair()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate device key pair")
	}

	mapArgs = map[string]interface{}{
		"token":               device.Token,
		"user_id":             user.ID,
		"private_key":         deviceKeyPair.PrivateKey,
		"public_key":          deviceKeyPair.PublicKey,
		"longitude":           device.Longitude,
		"latitude":            device.Latitude,
		"disposition":         device.Disposition,
		"encryption_password": d.encryptionPassword,
	}

	err = tx.Exec(sql, mapArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert device")
	}

	return &Device{
		PublicKey: deviceKeyPair.PublicKey,
		User: &User{
			PrivateKey: user.PrivateKey,
			PublicKey:  user.PublicKey,
		},
	}, err
}

// DeleteDevice finds the device identified by the given token, and deletes it
// from the database. We also delete the associated user if they have no other
// devices currently registered in the database.
func (d *db) DeleteDevice(token, publicKey string) (err error) {
	sql := `DELETE FROM devices d
		USING users u
		WHERE u.id = d.user_id
		AND d.token = :token
		AND u.public_key = :public_key
		RETURNING u.id`

	mapArgs := map[string]interface{}{
		"token":      token,
		"public_key": publicKey,
	}

	tx, err := BeginTX(d.DB)
	if err != nil {
		return errors.Wrap(err, "failed to start transaction when deleting device")
	}

	defer func() {
		if cerr := tx.CommitOrRollback(); err == nil && cerr != nil {
			err = cerr
		}
	}()

	var userID int

	err = tx.Get(&userID, sql, mapArgs)
	if err != nil {
		return errors.Wrap(err, "failed to delete device")
	}

	// now count devices for the user
	sql = `SELECT COUNT(*) FROM devices WHERE user_id = :user_id`

	mapArgs = map[string]interface{}{
		"user_id": userID,
	}

	var deviceCount int

	err = tx.Get(&deviceCount, sql, mapArgs)
	if err != nil {
		return errors.Wrap(err, "failed to count remaining devices")
	}

	if deviceCount == 0 {
		sql = `DELETE FROM users WHERE id = :id`

		mapArgs = map[string]interface{}{
			"id": userID,
		}

		err = tx.Exec(sql, mapArgs)
		if err != nil {
			return errors.Wrap(err, "failed to delete user")
		}
	}

	return nil
}

// MigrateUp is a convenience function to run all up migrations in the context
// of an instantiated DB instance.
func (d *db) MigrateUp() error {
	return MigrateUp(d.DB.DB, d.logger)
}

// MigrateDownAll is a convenience function to run all down migrations in the
// context of an instantiated DB instance.
func (d *db) MigrateDownAll() error {
	return MigrateDownAll(d.DB.DB, d.logger)
}
