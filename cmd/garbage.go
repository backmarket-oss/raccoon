package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/backmarket-oss/raccoon/internal"
	"github.com/backmarket-oss/raccoon/internal/k8s"
	"github.com/backmarket-oss/raccoon/internal/strategy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	garbageCmd = &cobra.Command{
		Use:   "garbage",
		Short: "Run raccoon daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			strategy, err := provideStrategy(cmd)
			if err != nil {
				return fmt.Errorf("error providing a strategy: %v", err)
			}
			interval, err := cmd.Flags().GetInt("check-interval")
			if err != nil {
				return err
			}
			return internal.RunDaemon(interval, cmd.Context(), strategy)
		},
	}
	defaultSettings *internal.DefaultSettings
)

func init() {
	defaultSettings = &internal.DefaultSettings{}
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(garbageCmd)
	// required flags

	// optional flags
	garbageCmd.Flags().StringVarP(&defaultSettings.Namespace, "namespace", "n", "", "Namespace to raccoon")
	garbageCmd.Flags().StringVarP(&defaultSettings.Selector, "selector", "s", "backmarket.com/raccoon=true",
		"Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	garbageCmd.Flags().String("kube-location", "in", "Connection mode to the kubernetes api (in or out)")
	garbageCmd.Flags().DurationVar(&defaultSettings.TTL, "ttl", 24*time.Hour, "Minimum age by which a pod will be deleted")
	garbageCmd.Flags().Int("check-interval", 120, "Interval between two raccoon check")
	garbageCmd.Flags().Int("randomized-delay", 120, "Delay the deletion by a randomly amount of time [value/2,value]")
	garbageCmd.Flags().BoolVar(&defaultSettings.DryRun, "dry-run", false, "Test process without deletion")
	garbageCmd.Flags().String("kubeconfig", filepath.Join(homedir, ".kube", "config"), "Path to KUBECONFIG file. Ignored if KUBECONFIG envvar is set")
}

func provideStrategy(cmd *cobra.Command) (internal.Strategy, error) {
	maxDelay, err := cmd.Flags().GetInt("randomized-delay")
	if err != nil {
		return nil, err
	}
	k8sLocation, err := cmd.Flags().GetString("kube-location")
	if err != nil {
		return nil, err
	}

	kubeConfig := os.Getenv("KUBECONFIG")
	// If no KUBECONFIG environment variable and we are executing out of the cluster
	if kubeConfig == "" && k8sLocation == "out" {
		kubeConfig, err = cmd.Flags().GetString("kubeconfig")
		if err != nil {
			return nil, err
		}
		_, err = os.Stat(kubeConfig)
		if err != nil {
			return nil, fmt.Errorf("The kubeconfig path you given: %v doesn't exist", kubeConfig)
		}

	}

	k8sClientSet, err := k8s.AuthenticateToCluster(k8sLocation, kubeConfig)
	if err != nil {
		return nil, err
	}
	k8sClient := k8s.InitKubernetesClient(k8sClientSet)
	rndDelayStg := strategy.InitRandomizedDelay(cmd.Context(), maxDelay, defaultSettings, k8sClient)
	return rndDelayStg, nil
}
