package healthz

import (
	"github.com/alibaba/aigateway-group/aigateway-console/backend/api/healthz"
	platformsvc "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/service/platform"
)

func NewV1(platformService *platformsvc.Service) healthz.IHealthzV1 {
	return &ControllerV1{
		platformService: platformService,
	}
}
