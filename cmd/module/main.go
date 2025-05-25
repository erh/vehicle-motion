package main

import (
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/motion"
	"vehiclemotion"
)

func main() {
	module.ModularMain(resource.APIModel{motion.API, vehiclemotion.OutdoorMotionServiceModel})
}
