package vehiclemotion

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kellydunn/golang-geo"

	"go.viam.com/rdk/components/base"
	"go.viam.com/rdk/components/movementsensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/motion"
)

var (
	OutdoorMotionServiceModel = resource.NewModel("erh", "vehicle-motion", "outdoor-motion-service")
)

func init() {
	resource.RegisterService(motion.API, OutdoorMotionServiceModel,
		resource.Registration[motion.Service, *Config]{
			Constructor: newVehicleMotionOutdoorMotionService,
		},
	)
}

type Config struct {
	Base           string
	MovementSensor string `json:"movement_sensor"`

	DeviationMeters float64 `json:"deviation_meters"`
	SpeedKMH        float64 `json:"speed_kmh"`
	SpeedAngular    float64 `json:"speed_degrees_per_second"`

	defaultConfig *motion.MotionConfiguration
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	if cfg.Base == "" {
		return nil, nil, fmt.Errorf("need a base")
	}
	if cfg.MovementSensor == "" {
		return nil, nil, fmt.Errorf("need a movement_sensor")
	}
	return []string{cfg.Base, cfg.MovementSensor}, nil, nil
}

func (cfg *Config) fixMotionCfg(c *motion.MotionConfiguration) *motion.MotionConfiguration {

	if c == nil {
		if cfg.defaultConfig == nil {
			cfg.defaultConfig = &motion.MotionConfiguration{}
		}
		c = cfg.defaultConfig
	}

	if c.PlanDeviationMM <= 0 {
		c.PlanDeviationMM = max(1000, cfg.DeviationMeters*1000)
	}

	if c.LinearMPerSec <= 0 {
		c.LinearMPerSec = max(1, cfg.SpeedKMH*1000/3600)
	}

	if c.AngularDegsPerSec <= 0 {
		c.AngularDegsPerSec = max(10, cfg.SpeedAngular)
	}

	return c
}

type vehicleMotionOutdoorMotionService struct {
	resource.AlwaysRebuild

	name resource.Name

	logger logging.Logger
	cfg    *Config

	base base.Base
	ms   movementsensor.MovementSensor

	cancelFunc func()

	dataLock      sync.Mutex
	lastRequest   *motion.MoveOnGlobeReq
	currentStatus motion.PlanStatusWithID
}

func newVehicleMotionOutdoorMotionService(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (motion.Service, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewOutdoorMotionService(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewOutdoorMotionService(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (motion.Service, error) {

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &vehicleMotionOutdoorMotionService{
		name:       name,
		logger:     logger,
		cfg:        conf,
		cancelFunc: cancelFunc,
	}

	var err error

	s.base, err = base.FromDependencies(deps, conf.Base)
	if err != nil {
		return nil, fmt.Errorf("cannot load base (%s): %w", conf.Base, err)
	}

	s.ms, err = movementsensor.FromDependencies(deps, conf.MovementSensor)
	if err != nil {
		return nil, err
	}

	prop, err := s.ms.Properties(ctx, nil)
	if err != nil {
		return nil, err
	}

	if !prop.PositionSupported {
		return nil, fmt.Errorf("movementsensor needs PositionSupported")
	}
	if !prop.CompassHeadingSupported {
		return nil, fmt.Errorf("movementsensor needs CompassHeadingSupported")
	}
	if !prop.LinearVelocitySupported {
		return nil, fmt.Errorf("movementsensor needs LinearVelocitySupported")
	}
	if !prop.AngularVelocitySupported {
		return nil, fmt.Errorf("movementsensor needs AngularVelocitySupported")
	}

	go s.run(cancelCtx)

	return s, nil
}

func (s *vehicleMotionOutdoorMotionService) Name() resource.Name {
	return s.name
}

func (s *vehicleMotionOutdoorMotionService) Move(ctx context.Context, req motion.MoveReq) (bool, error) {
	return false, fmt.Errorf("Move not supported by %v", OutdoorMotionServiceModel)
}

func (s *vehicleMotionOutdoorMotionService) MoveOnMap(ctx context.Context, req motion.MoveOnMapReq) (motion.ExecutionID, error) {
	id := uuid.New()
	return id, fmt.Errorf("MoveOnMap not supported by %v", OutdoorMotionServiceModel)
}

func (s *vehicleMotionOutdoorMotionService) MoveOnGlobe(ctx context.Context, req motion.MoveOnGlobeReq) (motion.ExecutionID, error) {
	if req.ComponentName.ShortName() != s.cfg.Base {
		return uuid.Nil, fmt.Errorf("req had name %v but configured %s", req.ComponentName.ShortName(), s.cfg.Base)
	}

	if req.MovementSensorName.ShortName() != s.cfg.MovementSensor {
		return uuid.Nil, fmt.Errorf("req had name %v but configured %s", req.MovementSensorName.ShortName(), s.cfg.MovementSensor)
	}

	s.logger.Infof("new location to go to: %v cfg: %v", req.Destination, req.MotionCfg)

	req.MotionCfg = s.cfg.fixMotionCfg(req.MotionCfg)

	id := uuid.New()
	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	s.lastRequest = &req
	s.currentStatus.ExecutionID = id
	s.currentStatus.PlanID = id
	s.currentStatus.ComponentName = s.base.Name()
	s.currentStatus.Status.Timestamp = time.Now()
	s.currentStatus.Status.State = motion.PlanStateInProgress

	return id, nil
}

func (s *vehicleMotionOutdoorMotionService) GetPose(ctx context.Context, componentName resource.Name, destinationFrame string, supplementalTransforms []*referenceframe.LinkInFrame, extra map[string]interface{}) (*referenceframe.PoseInFrame, error) {
	return nil, fmt.Errorf("GetPose not supported by %v", OutdoorMotionServiceModel)
}

func (s *vehicleMotionOutdoorMotionService) StopPlan(ctx context.Context, req motion.StopPlanReq) error {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()
	s.lastRequest = nil
	s.currentStatus.Status.Timestamp = time.Now()
	s.currentStatus.Status.State = motion.PlanStateStopped
	return s.base.Stop(ctx, nil)
}

func (s *vehicleMotionOutdoorMotionService) ListPlanStatuses(ctx context.Context, req motion.ListPlanStatusesReq) ([]motion.PlanStatusWithID, error) {
	s.dataLock.Lock()
	defer s.dataLock.Unlock()

	if req.OnlyActivePlans && s.currentStatus.Status.State != motion.PlanStateInProgress {
		return []motion.PlanStatusWithID{}, nil
	}
	return []motion.PlanStatusWithID{s.currentStatus}, nil
}

func (s *vehicleMotionOutdoorMotionService) PlanHistory(ctx context.Context, req motion.PlanHistoryReq) ([]motion.PlanWithStatus, error) {
	return nil, fmt.Errorf("eliot finish PlanHistory")
}

func (s *vehicleMotionOutdoorMotionService) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (s *vehicleMotionOutdoorMotionService) Close(context.Context) error {
	s.logger.Infof("close called")
	s.cancelFunc()
	return nil
}

func (s *vehicleMotionOutdoorMotionService) run(ctx context.Context) {
	for ctx.Err() == nil {
		err := s.doLoop(ctx)
		if err != nil {
			s.logger.Errorf("doLoop error: %v", err)
		}
		time.Sleep(time.Second)
	}
	s.logger.Infof("vehicleMotionOutdoorMotionService shutting down, stopping base")
	err := s.base.Stop(ctx, nil)
	if err != nil {
		s.logger.Errorf("can't stop base: %v", err)
	}
}

func (s *vehicleMotionOutdoorMotionService) doLoop(ctx context.Context) error {
	pos, _, err := s.ms.Position(ctx, nil)
	if err != nil {
		return fmt.Errorf("can't get position: %v", err)
	}

	heading, err := s.ms.CompassHeading(ctx, nil)
	if err != nil {
		return fmt.Errorf("can't get compass heading: %v", err)
	}

	s.logger.Debugf("current pos: %v heading: %v", pos, heading)

	var goal *geo.Point
	var cfg *motion.MotionConfiguration

	s.dataLock.Lock()
	if s.lastRequest != nil {
		goal = s.lastRequest.Destination
		cfg = s.lastRequest.MotionCfg
	}
	s.dataLock.Unlock()

	if goal == nil {
		return nil
	}

	linear, angular := computeSetVelocity(pos, goal, heading, cfg, s.logger)

	str := fmt.Sprintf("setting linear: %0.2f mm/sec %0.2f kmh angular: %0.2f degs / sec", linear.Y, linear.Y/277.778, angular.Z)

	s.logger.Debug(str)

	if linear.Y == 0 && angular.Z == 0 {
		s.dataLock.Lock()
		if s.lastRequest != nil {
			s.lastRequest = nil
		}
		s.currentStatus.Status.Timestamp = time.Now()
		s.currentStatus.Status.State = motion.PlanStateSucceeded
		s.dataLock.Unlock()
		return s.base.Stop(ctx, nil)
	}

	s.dataLock.Lock()
	s.currentStatus.Status.Reason = &str
	s.currentStatus.Status.Timestamp = time.Now()
	s.dataLock.Unlock()

	return s.base.SetVelocity(ctx, linear, angular, nil)
}
