package tasks

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/thingful/iotdevicereg/pkg/logger"
	"github.com/thingful/iotdevicereg/pkg/server"
)

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringP("addr", "a", "0.0.0.0:8080", "Address to which the HTTP server binds")
	serverCmd.Flags().StringP("encoder", "e", "", "Address at which the encoder is listening")
	serverCmd.Flags().Bool("verbose", false, "Enable verbose output")

	viper.BindPFlag("addr", serverCmd.Flags().Lookup("addr"))
	viper.BindPFlag("encoder", serverCmd.Flags().Lookup("encoder"))
	viper.BindPFlag("verbose", serverCmd.Flags().Lookup("verbose"))
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts datastore listening for requests",
	Long: `
Starts our implementation of the DECODE device registration RPC interface, which is
designed to expose a simple API to claim and revoke ownership of devices, as well as
eventually being able to create entitlements for owned devices which will allow data
to be shared under certain conditions with specific parties.

The server uses Twirp to expose both a JSON API along with a more performant
Protocol Buffer API. The JSON API is not intended for use other than for
clients unable to use the Protocol Buffer API.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := viper.GetString("addr")
		if addr == "" {
			return errors.New("Must provide a bind address")
		}

		encoderAddr := viper.GetString("encoder")
		if encoderAddr == "" {
			return errors.New("Must provide encoder address")
		}

		connStr := viper.GetString("database_url")
		if connStr == "" {
			return errors.New("Missing required environment variable: $DEVICEREG_DATABASE_URL")
		}

		encryptionPassword := viper.GetString("encryption_password")
		if encryptionPassword == "" {
			return errors.New("Missing required environment variable: $DEVICEREG_ENCRYPTION_PASSWORD")
		}

		logger := logger.NewLogger()

		config := &server.Config{
			ListenAddr:         addr,
			ConnStr:            connStr,
			EncryptionPassword: encryptionPassword,
			EncoderAddr:        encoderAddr,
			Verbose:            viper.GetBool("verbose"),
		}

		s := server.NewServer(config, logger)

		return s.Start()
	},
}
