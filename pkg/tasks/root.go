package tasks

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/thingful/iotdevicereg/pkg/version"
)

func init() {
	viper.SetEnvPrefix("devicereg")
	viper.AutomaticEnv()
}

var rootCmd = &cobra.Command{
	Use:   version.BinaryName,
	Short: "Device registration service for the DECODE IoT Pilot",
	Long: `This tool is an implementation of the device registration service being
developed as part of the IoT Pilot for DECODE (https://decodeproject.eu/).

This components exposes a simple RPC API implemented using a library called
Twirp, that exposes either a JSON or Protocol Buffer API over HTTP 1.1.

Data is persisted locally to PostgreSQL, and we call down to the other components
being developed for DECODE which provide encryption and storage capabilities.
`,
	Version: version.VersionString(),
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("FATAL: %+v\n", err)
		os.Exit(1)
	}
}
