package main

import (
	"flag"
	"os"
	"time"

	"alertika/config"

	"alertika/cwlog"
	"alertika/db"
	"alertika/model"
	"alertika/notify"
	"alertika/process"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	//runtime.GOMAXPROCS(runtime.NumCPU())
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	configfile := flag.String("config", "alertika.toml", "config file location in toml format")
	debug := flag.Bool("debug", false, "sets log level to debug")
	logjson := flag.Bool("logjson", false, "logs in json")
	flag.Parse()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if !*logjson {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
	}

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if *configfile == "" {
		log.Fatal().Msg("please specify config file")
	}

	log.Info().Msgf("using configfile %s", *configfile)
	config, err := config.ParseConfig(*configfile)

	if err != nil {
		log.Fatal().Msgf("cannot parse configfile %s: %s:", *configfile, err)
	}
	log.Debug().Msgf("config: %+v", config)

	db, err := db.New()
	duration, _ := time.ParseDuration(config.General.RunInterval)

	for {
		results := make(chan *model.ResultItem, len(config.Rules))
		notifications := make(chan *model.NotificationItem, len(config.Rules)) // big enough for all rules

		cwlog.ProcessQueries(config.Rules, results)
		close(results)

		process.Results(results, notifications, db)
		close(notifications)

		notify.ProcessNotifications(notifications, config.General.Template)

		log.Info().Msgf("waiting %s for next collection interval", config.General.RunInterval)
		time.Sleep(duration)
	}
}
