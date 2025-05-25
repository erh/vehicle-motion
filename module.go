package vehiclemotion

import (
	"context"
	"fmt"

	"github.com/google/uuid"

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
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	return nil, nil, nil
}

type vehicleMotionOutdoorMotionService struct {
	resource.AlwaysRebuild

	name resource.Name

	logger logging.Logger
	cfg    *Config

	cancelCtx  context.Context
	cancelFunc func()
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
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
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
	id := uuid.New()
	return id, fmt.Errorf("eliot finish MoveOnGlobeReq")
}

func (s *vehicleMotionOutdoorMotionService) GetPose(ctx context.Context, componentName resource.Name, destinationFrame string, supplementalTransforms []*referenceframe.LinkInFrame, extra map[string]interface{}) (*referenceframe.PoseInFrame, error) {
	return nil, fmt.Errorf("GetPose not supported by %v", OutdoorMotionServiceModel)
}

func (s *vehicleMotionOutdoorMotionService) StopPlan(ctx context.Context, req motion.StopPlanReq) error {
	return fmt.Errorf("eliot finish StopPlan")
}

func (s *vehicleMotionOutdoorMotionService) ListPlanStatuses(ctx context.Context, req motion.ListPlanStatusesReq) ([]motion.PlanStatusWithID, error) {
	return nil, fmt.Errorf("eliot finish ListPlanStatuses")
}

func (s *vehicleMotionOutdoorMotionService) PlanHistory(ctx context.Context, req motion.PlanHistoryReq) ([]motion.PlanWithStatus, error) {
	return nil, fmt.Errorf("eliot finish PlanHistory")
}

func (s *vehicleMotionOutdoorMotionService) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (s *vehicleMotionOutdoorMotionService) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
