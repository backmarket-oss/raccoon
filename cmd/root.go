package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	envPrefix = "RACCOON"
)

var (
	logLevel string
	port     string
	rootCmd  = &cobra.Command{
		Use:   "raccoon",
		Short: "Raccoon mark and delete resources based on their age",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(cmd)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "level", "info", "set log level")
	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "2112", "set HTTP port")
}

// Execute start the cli execution.
func Execute(ctx context.Context) error {
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Debugf("HTTP server starting and listening on port %s", port)

		// nosemgrep: go.lang.security.audit.net.use-tls.use-tls
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("error starting http server: %s\n", err)
		}
	}()

	return rootCmd.ExecuteContext(ctx)
}

func initConfig(cmd *cobra.Command) error {
	v := viper.New()
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	if err := bindFlags(cmd, v); err != nil {
		return err
	}

	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)

	return nil
}

// Bind each cobra flag to its associated viper environment variable.
func bindFlags(cmd *cobra.Command, v *viper.Viper) error {
	var err error
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			err = v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix))
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			err = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})

	return err
}
