package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	dashboardcontroller "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/controller/dashboard"
	gatewaycontroller "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/controller/gateway"
	"github.com/alibaba/aigateway-group/aigateway-console/backend/internal/controller/healthz"
	jobscontroller "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/controller/jobs"
	portalcontroller "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/controller/portal"
	sessioncontroller "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/controller/session"
	"github.com/alibaba/aigateway-group/aigateway-console/backend/internal/controller/system"
	usercontroller "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/controller/user"
	"github.com/alibaba/aigateway-group/aigateway-console/backend/internal/middleware"
	gatewaysvc "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/service/gateway"
	jobssvc "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/service/jobs"
	platformsvc "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/service/platform"
	portalsvc "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/service/portal"
	grafanaclient "github.com/alibaba/aigateway-group/aigateway-console/backend/utility/clients/grafana"
	k8sclient "github.com/alibaba/aigateway-group/aigateway-console/backend/utility/clients/k8s"
	portaldbclient "github.com/alibaba/aigateway-group/aigateway-console/backend/utility/clients/portaldb"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			k8sService := k8sclient.New(k8sclient.Config{
				Enabled:        g.Cfg().MustGet(ctx, "clients.k8s.enabled", false).Bool(),
				Namespace:      g.Cfg().MustGet(ctx, "clients.k8s.namespace", "aigateway-system").String(),
				KubectlBin:     g.Cfg().MustGet(ctx, "clients.k8s.kubectlBin", "kubectl").String(),
				KubeconfigPath: g.Cfg().MustGet(ctx, "clients.k8s.kubeconfig", "").String(),
				ResourcePrefix: g.Cfg().MustGet(ctx, "clients.k8s.resourcePrefix", "aigw-console").String(),
			})
			grafanaService := grafanaclient.New(grafanaclient.Config{
				Enabled:  g.Cfg().MustGet(ctx, "clients.grafana.enabled", false).Bool(),
				BaseURL:  g.Cfg().MustGet(ctx, "clients.grafana.baseURL", "").String(),
				Username: g.Cfg().MustGet(ctx, "clients.grafana.username", "").String(),
				Password: g.Cfg().MustGet(ctx, "clients.grafana.password", "").String(),
			})
			portalService := portaldbclient.New(portaldbclient.Config{
				Enabled:     g.Cfg().MustGet(ctx, "clients.portaldb.enabled", false).Bool(),
				Driver:      g.Cfg().MustGet(ctx, "clients.portaldb.driver", "").String(),
				DSN:         g.Cfg().MustGet(ctx, "clients.portaldb.dsn", "").String(),
				AutoMigrate: g.Cfg().MustGet(ctx, "clients.portaldb.autoMigrate", true).Bool(),
			})

			platformService := platformsvc.New(k8sService, grafanaService, portalService)
			gatewayService := gatewaysvc.New(k8sService)
			portalDomainService := portalsvc.New(portalService, k8sService)
			jobsService := jobssvc.New(portalService, portalDomainService, gatewayService, k8sService)
			portalDomainService.SetHook(jobsService)
			if err := jobsService.Start(ctx); err != nil {
				return err
			}
			s := g.Server()
			s.Use(middleware.Trace, middleware.AccessLog, middleware.Auth(platformService))
			s.BindHandler("/landing", func(r *ghttp.Request) {
				r.Response.ServeFile("resource/public/html/index.html")
			})
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					healthz.NewV1(platformService),
					system.NewV1(platformService),
				)
				sessioncontroller.Bind(group, platformService)
				usercontroller.Bind(group, platformService)
				system.BindHTTP(group, platformService)
				dashboardcontroller.Bind(group, platformService)
				gatewaycontroller.Bind(group, gatewayService)
				portalcontroller.Bind(group, portalDomainService)
				jobscontroller.Bind(group, jobsService)
				group.GET("/healthz/ready", func(r *ghttp.Request) {
					r.Response.WriteJsonExit(g.Map{"success": true, "data": "ok"})
				})
			})
			s.Run()
			return nil
		},
	}
)
