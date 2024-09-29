package cmd

import (
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "heu",
	Short: "Hue Entertainment Utility",
	Long: `HEU (Hue Entertainment Utility) is a command-line utility for interacting with
the Philips Hue Streaming API. You provide channel and colour information, and
it handles the details.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "alternate config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show more logs")
}

func initConfig() {
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	cobra.CheckErr(godotenv.Load())

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		userConfigDir, err := os.UserConfigDir()
		cobra.CheckErr(err)

		executablePath, err := os.Executable()
		cobra.CheckErr(err)

		viper.AddConfigPath(path.Join(userConfigDir, "heu"))
		viper.AddConfigPath(filepath.Dir(executablePath))
		viper.SetConfigType("toml")
		viper.SetConfigName("heu")
	}

	viper.SetEnvPrefix("heu")
	viper.SetEnvKeyReplacer(strings.NewReplacer())
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		slog.Debug("Loaded config", "file", viper.ConfigFileUsed())
	}
}
