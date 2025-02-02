package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	grpcapi "github.com/jamesread/OliveTin/internal/grpcapi"
	updatecheck "github.com/jamesread/OliveTin/internal/updatecheck"

	"github.com/jamesread/OliveTin/internal/httpservers"

	"github.com/fsnotify/fsnotify"
	config "github.com/jamesread/OliveTin/internal/config"
	"github.com/spf13/viper"
	"os"
	"path"
)

var (
	cfg     *config.Config
	version = "dev"
	commit  = "nocommit"
	date    = "nodate"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceQuote:       true,
		DisableTimestamp: true,
	})

	log.WithFields(log.Fields{
		"version": version,
		"commit":  commit,
		"date":    date,
	}).Info("OliveTin initializing")

	log.SetLevel(log.DebugLevel) // Default to debug, to catch cfg issues

	var configDir string
	flag.StringVar(&configDir, "configdir", ".", "Config directory path")
	flag.Parse()

	log.WithFields(log.Fields{
		"value": configDir,
	}).Debugf("Value of -configdir flag")

	viper.AutomaticEnv()
	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath("/config") // For containers.
	viper.AddConfigPath("/etc/OliveTin/")

	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("Config file error at startup. %s", err)
		os.Exit(1)
	}

	cfg = config.DefaultConfig()

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		if e.Op == fsnotify.Write {
			log.Info("Config file changed:", e.String())

			reloadConfig()
		}
	})

	reloadConfig()
	log.Info("Init complete")
}

func reloadConfig() {
	if err := viper.UnmarshalExact(&cfg); err != nil {
		log.Errorf("Config unmarshal error %+v", err)
		os.Exit(1)
	}

	cfg.Sanitize()
}

func main() {
	configDir := path.Dir(viper.ConfigFileUsed())

	log.WithFields(log.Fields{
		"configDir": configDir,
	}).Infof("OliveTin started")

	log.Debugf("Config: %+v", cfg)

	go updatecheck.StartUpdateChecker(version, commit, cfg, configDir)

	go grpcapi.Start(cfg)

	httpservers.StartServers(cfg)
}
