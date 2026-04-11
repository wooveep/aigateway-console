package system

import (
	"github.com/alibaba/aigateway-group/aigateway-console/backend/api/system"
	platformsvc "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/service/platform"
)

func NewV1(platformService *platformsvc.Service) system.ISystemV1 {
	return &ControllerV1{
		platformService: platformService,
	}
}
