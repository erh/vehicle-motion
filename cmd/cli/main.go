package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/kellydunn/golang-geo"

	"go.viam.com/rdk/components/base"
	"go.viam.com/rdk/components/movementsensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/services/motion"

	"github.com/erh/vmodutils"

	"vehiclemotion"
)

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}

func realMain() error {
	ctx := context.Background()
	logger := logging.NewLogger("cli")

	configFile := flag.String("config", "", "config file")
	host := flag.String("host", "", "host to connect to")
	debug := flag.Bool("debug", false, "debugging on")

	flag.Parse()

	logger.Infof("using config file [%s] and host [%s]", *configFile, *host)

	if *configFile == "" {
		return fmt.Errorf("need a config file")
	}

	cfg := &vehiclemotion.Config{}

	err := vmodutils.ReadJSONFromFile(*configFile, cfg)
	if err != nil {
		return err
	}

	_, _, err = cfg.Validate("")
	if err != nil {
		return err
	}

	client, err := vmodutils.ConnectToHostFromCLIToken(ctx, *host, logger)
	if err != nil {
		return err
	}
	defer client.Close(ctx)

	deps, err := vmodutils.MachineToDependencies(client)
	if err != nil {
		return err
	}

	svcLogger := logger.Sublogger("module")
	if *debug {
		svcLogger.SetLevel(logging.DEBUG)
	}

	//pos := geo.NewPoint(40.977310, -73.659143)
	//pos := geo.NewPoint(40.977618, -73.659162)

	//pos := geo.NewPoint(40.975170, -73.660791)
	//pos := geo.NewPoint(40.975264, -73.660687)
	pos := geo.NewPoint(40.975156, -73.660606)

	thing, err := vehiclemotion.NewOutdoorMotionService(ctx, deps, motion.Named("foo"), cfg, svcLogger)
	if err != nil {
		return err
	}
	defer thing.Close(ctx)

	if true {
		_, err = thing.MoveOnGlobe(ctx, motion.MoveOnGlobeReq{
			ComponentName:      base.Named(cfg.Base),
			Destination:        pos,
			MovementSensorName: movementsensor.Named(cfg.MovementSensor),
		})
		if err != nil {
			return err
		}
		time.Sleep(time.Minute)
	}

	return nil
}
