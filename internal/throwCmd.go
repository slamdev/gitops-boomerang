package internal

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitops-boomerang/pkg/boomerang"
	"time"
)

var throwCmd = &cobra.Command{
	Use:   "throw",
	Short: "Throw a boomerang and wait until it comes back",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var cfg boomerang.Config
		var err error
		if cfg.Application, err = cmd.Flags().GetString("application"); err != nil {
			return fmt.Errorf("failed to get %v flag value. %w", "application", err)
		}
		if cfg.Image, err = cmd.Flags().GetString("image"); err != nil {
			return fmt.Errorf("failed to get %v flag value. %w", "image", err)
		}
		if cfg.Namespace, err = cmd.Flags().GetString("namespace"); err != nil {
			return fmt.Errorf("failed to get %v flag value. %w", "namespace", err)
		}
		if cfg.Timeout, err = cmd.Flags().GetDuration("timeout"); err != nil {
			return fmt.Errorf("failed to get %v flag value. %w", "timeout", err)
		}
		return boomerang.Throw(context.Background(), cmd.OutOrStdout(), cfg)
	},
}

func init() {
	throwCmd.PersistentFlags().StringP("application", "a", "", "application in a form of kind/name, e.g. deploy/nginx")
	if err := throwCmd.MarkPersistentFlagRequired("application"); err != nil {
		logrus.WithError(err).Fatal("failed mark flag as required")
	}
	throwCmd.PersistentFlags().StringP("image", "i", "", "docker image to poll for update")
	if err := throwCmd.MarkPersistentFlagRequired("image"); err != nil {
		logrus.WithError(err).Fatal("failed mark flag as required")
	}
	throwCmd.PersistentFlags().StringP("namespace", "n", "default", "namespace where to search")
	throwCmd.PersistentFlags().DurationP("timeout", "t", 90*time.Second, "timeout for image update polling")
	cobra.OnInitialize(func() {
		fillWithEnvVars(throwCmd.Flags())
	})
	rootCmd.AddCommand(throwCmd)
}
