package main

import (
	"fmt"
	rcon_hub "github.com/itzg/rcon-hub"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"strings"
)

var (
	version = ""
	commit = ""
	date = ""
	builtBy = ""
)

const (
	ArgBind = "bind"
	ArgUser = "user"
	ArgPassword = "password"
	ArgHostKeyFile = "host-key-file"
	ArgConnection = "connection"
)

var rootCmd = &cobra.Command{
	Use: "rcon-hub",
	Short: "An ssh server that can connect out to rcon enabled games",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := rcon_hub.LoadConfig()
		if err != nil {
			log.Fatal(err)
		}

		sshServer := rcon_hub.NewSshServer(config)
		log.Fatal(sshServer.ListenAndServe())
	},
}

var versionCmd = &cobra.Command{
	Use: "version",
	Short: "Show version and build info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("version=%s commit=%s, buildDate=%s, builtBy=%s\n",
			version, commit, date, builtBy)
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String(ArgBind, ":2222", "host:port to bind ssh server")
	if err := viper.BindPFlag(rcon_hub.ConfigBind, rootCmd.PersistentFlags().Lookup(ArgBind)); err != nil {
		log.Fatal(err)
	}

	rootCmd.PersistentFlags().String(ArgHostKeyFile, "", "PEM file containing SSH host key")
	if err := viper.BindPFlag(rcon_hub.ConfigHostKeyFile, rootCmd.PersistentFlags().Lookup(ArgHostKeyFile)); err != nil {
		log.Fatal(err)
	}

	rootCmd.PersistentFlags().String(ArgUser, "user", "A user to register for ssh authentication")
	if err := viper.BindPFlag(rcon_hub.ConfigUser, rootCmd.PersistentFlags().Lookup(ArgUser)); err != nil {
		log.Fatal(err)
	}

	rootCmd.PersistentFlags().String(ArgPassword, "", "If given, a user with this password will be registered for ssh authentication")
	if err := viper.BindPFlag(rcon_hub.ConfigPassword, rootCmd.PersistentFlags().Lookup(ArgPassword)); err != nil {
		log.Fatal(err)
	}

	rootCmd.PersistentFlags().StringSlice(ArgConnection, []string{}, "Register a connection formatted as "+rcon_hub.ConnectionFormat)
	if err := viper.BindPFlag(rcon_hub.ConfigConnection, rootCmd.PersistentFlags().Lookup(ArgConnection)); err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(versionCmd)
}

func initConfig() {

	viper.SetEnvPrefix("rh")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/rcon-hub")
	viper.AddConfigPath("$HOME/.rcon-hub")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("W: Unable to load config file: %s\n", err)
	}
}
