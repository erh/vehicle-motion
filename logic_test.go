package vehiclemotion

import (
	"testing"

	"github.com/golang/geo/r3"
	"github.com/kellydunn/golang-geo"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/services/motion"
	"go.viam.com/test"
)

func TestComputeSetVelocity1(t *testing.T) {
	logger := logging.NewTestLogger(t)

	cfg := &motion.MotionConfiguration{
		LinearMPerSec:     100,
		AngularDegsPerSec: 30,
	}

	// we're there, don't move
	linear, angular := computeSetVelocity(geo.NewPoint(50, 50), geo.NewPoint(50, 50), 0, cfg, logger)
	test.That(t, linear, test.ShouldResemble, r3.Vector{})
	test.That(t, angular, test.ShouldResemble, r3.Vector{})

	// heading in the right direction, just go fast
	linear, angular = computeSetVelocity(geo.NewPoint(50, 50), geo.NewPoint(51, 50), 0, cfg, logger)
	test.That(t, linear, test.ShouldResemble, r3.Vector{Y: 100})
	test.That(t, angular, test.ShouldResemble, r3.Vector{})

	// need to turn hard, go slow
	linear, angular = computeSetVelocity(geo.NewPoint(50, 50), geo.NewPoint(51, 50), 90, cfg, logger)
	test.That(t, linear, test.ShouldResemble, r3.Vector{Y: 20})
	test.That(t, angular, test.ShouldResemble, r3.Vector{Z: -30})

	linear, angular = computeSetVelocity(geo.NewPoint(50, 50), geo.NewPoint(51, 50), -90, cfg, logger)
	test.That(t, linear, test.ShouldResemble, r3.Vector{Y: 20})
	test.That(t, angular, test.ShouldResemble, r3.Vector{Z: 30})

	linear, angular = computeSetVelocity(geo.NewPoint(50, 50), geo.NewPoint(51, 50), 270, cfg, logger)
	test.That(t, linear, test.ShouldResemble, r3.Vector{Y: 20})
	test.That(t, angular, test.ShouldResemble, r3.Vector{Z: 30})

	// need to turn a little, go fast
	linear, angular = computeSetVelocity(geo.NewPoint(50, 50), geo.NewPoint(51, 50), 5, cfg, logger)
	test.That(t, linear, test.ShouldResemble, r3.Vector{Y: 100})
	test.That(t, angular, test.ShouldResemble, r3.Vector{Z: -7.5})

	// need to turn medium, go a little slower
	linear, angular = computeSetVelocity(geo.NewPoint(50, 50), geo.NewPoint(51, 50), 10, cfg, logger)
	test.That(t, linear, test.ShouldResemble, r3.Vector{Y: 100})
	test.That(t, angular, test.ShouldResemble, r3.Vector{Z: -15})

}
