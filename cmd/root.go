/*
Copyright © 2024 Shono <code@shono.io>
*/
package cmd

import (
    "fmt"
    "github.com/rs/zerolog/log"
    "github.com/shono-io/nexthos-runtime/pkg"
    "os"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
    Use:   "nexthos-runtime",
    Short: "nexthos benthos runtime",
    Long: `Run benthos from NATS
This runtime will connect to a nats environment to retrieve the pipeline configuration`,
    Run: func(cmd *cobra.Command, args []string) {
        if err := pkg.Run(); err != nil {
            panic(err)
        }
    },
}

func Execute() {
    err := rootCmd.Execute()
    if err != nil {
        os.Exit(1)
    }
}

func init() {
    cobra.OnInitialize(initConfig)

    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nexthos-runtime.yaml)")
    rootCmd.PersistentFlags().String("namespace", "", "pipeline namespace")
    rootCmd.PersistentFlags().String("pipeline", "", "pipeline id")
    rootCmd.PersistentFlags().String("version", "", "pipeline version")

    rootCmd.PersistentFlags().String("repo-url", "tls://connect.ngs.global", "repository url")
    rootCmd.PersistentFlags().String("repo-seed", "", "repository user seed")
    rootCmd.PersistentFlags().String("repo-jwt", "", "repository user jwt")
    rootCmd.PersistentFlags().String("repo-creds-file", "", "repository credentials file")
    rootCmd.PersistentFlags().String("repo-kv", "meta", "repository key value bucket")
    rootCmd.PersistentFlags().String("repo-ob", "artifacts", "repository artifact bucket")

    if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
        log.Panic().Err(err).Msg("failed to bind flags")
    }
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
    if cfgFile != "" {
        // Use config file from the flag.
        viper.SetConfigFile(cfgFile)
    } else {
        // Find home directory.
        home, err := os.UserHomeDir()
        cobra.CheckErr(err)

        // Search config in home directory with name ".nexthos-runtime" (without extension).
        viper.AddConfigPath(home)
        viper.SetConfigType("yaml")
        viper.SetConfigName(".nexthos-runtime")
    }

    viper.SetEnvPrefix("NEXTHOS_RUNTIME")
    viper.AutomaticEnv() // read in environment variables that match

    // If a config file is found, read it in.
    if err := viper.ReadInConfig(); err == nil {
        fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
    }
}
