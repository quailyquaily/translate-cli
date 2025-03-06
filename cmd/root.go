/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/quailyquaily/translate-cli/cmd/translate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	debugMode bool
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "translate-cli",
	Short: "Translate your locale files",
	Long:  `Translate your locale files with AI`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func ExecuteContext(ctx context.Context) {
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(translate.NewCmd())

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.translate-cli.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "toggle debug mode")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		fullpath := filepath.Join(home, ".config", "translate-cli")
		if _, err := os.Stat(fullpath); os.IsNotExist(err) {
			os.MkdirAll(fullpath, 0755)
		}

		viper.AddConfigPath(path.Join(home, ".config", "translate-cli"))
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")

		cfgFile = filepath.Join(fullpath, "config.yaml")

		viper.SetConfigFile(cfgFile)

		// if the config file is not found, create a new one
		if _, err := os.Stat(viper.ConfigFileUsed()); os.IsNotExist(err) {
			// create the config file
			viper.Set("provider", "openai")
			viper.Set("openai.api_key", "")
			viper.Set("openai.api_base", "https://api.openai.com/v1")
			viper.Set("openai.model", "")
			viper.Set("debug", false)

			os.MkdirAll(filepath.Dir(viper.ConfigFileUsed()), 0755)
			err = viper.WriteConfigAs(path.Join(fullpath, "config.yaml"))
			if err != nil {
				fmt.Println("failed to save config", err, "config", fullpath)
				return
			}
			fmt.Println("created config file:", fullpath)
			fmt.Println("please edit the config file and run the command again")
		}
	}

	// Read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("failed to read config", err, "config", viper.ConfigFileUsed())
		return
	}
}
