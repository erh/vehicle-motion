package vehiclemotion

import (
	"math"

	"github.com/golang/geo/r3"
	"github.com/kellydunn/golang-geo"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/services/motion"
)

// puts between -180 and 180
func normalizeAngleDiff(d float64) float64 {
	for d > 180 {
		d = d - 360
	}
	return d
}

func computeSetVelocity(pos, goal *geo.Point, heading float64, cfg *motion.MotionConfiguration, logger logging.Logger) (r3.Vector, r3.Vector) {
	distanceKm := pos.GreatCircleDistance(goal)
	distanceMm := distanceKm * 1000 * 1000

	bearing := pos.BearingTo(goal)

	degreesOff := normalizeAngleDiff(heading - bearing)

	logger.Debugf("distanceKm: %0.2f distanceMm: %0.2f bearing: %v degreesOff: %v", distanceKm, distanceMm, bearing, degreesOff)

	if distanceMm <= cfg.PlanDeviationMM {
		return r3.Vector{}, r3.Vector{}
	}

	if math.Abs(degreesOff) <= 0.1 {
		return r3.Vector{Y: cfg.LinearMPerSec * 1000}, r3.Vector{}
	}

	const hardTurnThreshold float64 = 40.0

	if math.Abs(degreesOff) > hardTurnThreshold {
		// go slow and turn
		z := cfg.AngularDegsPerSec
		if degreesOff > 0 {
			z *= -1
		}
		return r3.Vector{Y: cfg.LinearMPerSec * 1000 / 4}, r3.Vector{Z: z}
	}

	z := -1 * (degreesOff / hardTurnThreshold) * cfg.AngularDegsPerSec

	return r3.Vector{Y: cfg.LinearMPerSec * 1000}, r3.Vector{Z: z}
}
