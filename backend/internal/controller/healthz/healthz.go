package healthz

import platformsvc "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/service/platform"

type ControllerV1 struct {
	platformService *platformsvc.Service
}
