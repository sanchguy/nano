package main

import (
	"fmt"
	"github.com/sanchguy/nano"
	"github.com/sanchguy/nano/games/truco/game"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	"os"
	"runtime/pprof"
	"sync"
	"time"
)

func main() {
	app := cli.NewApp()

	// base application info
	app.Name = "truco server"
	app.Author = "sanchguy"
	app.Version = "0.0.1"
	app.Copyright = "nebula team reserved"
	app.Usage = "huya games"

	// flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "../configs/config.toml",
			Usage: "load configuration from `FILE`",
		},
		cli.BoolFlag{
			Name:  "cpuprofile",
			Usage: "enable cpu profile",
		},
	}

	app.Action = serve
	app.Run(os.Args)
}

func serve(c *cli.Context) error {
	viper.SetConfigType("toml")
	viper.SetConfigFile(c.String("config"))
	if err := viper.ReadInConfig() ; err != nil {
		log.Infof("load toml faile error")
		log.Error(err.Error())
	}

	log.SetFormatter(&log.TextFormatter{DisableColors: true})
	nano.EnableDebug()
	if viper.GetBool("core.debug") {
		log.SetLevel(log.DebugLevel)
	}

	if c.Bool("cpuprofile") {
		filename := fmt.Sprintf("cpuprofile-%d.pprof", time.Now().Unix())
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() { defer wg.Done(); game.Startup() }() // 开启游戏服

	wg.Wait()
	return nil
}
