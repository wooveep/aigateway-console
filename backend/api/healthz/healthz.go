package healthz

import (
	"context"

	v1 "github.com/alibaba/aigateway-group/aigateway-console/backend/api/healthz/v1"
)

type IHealthzV1 interface {
	Ping(ctx context.Context, req *v1.PingReq) (res *v1.PingRes, err error)
}
