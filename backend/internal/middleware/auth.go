package middleware

import (
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/alibaba/aigateway-group/aigateway-console/backend/internal/consts"
	"github.com/alibaba/aigateway-group/aigateway-console/backend/internal/service/platform"
)

func Auth(platformService *platform.Service) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		if isAnonymousPath(r.URL.Path) {
			r.Middleware.Next()
			return
		}

		token := ""
		if cookie := r.Cookie.Get(consts.DefaultAdminCookieName); cookie != nil {
			token = cookie.String()
		}
		if token == "" {
			header := strings.TrimSpace(r.Header.Get(consts.AuthorizationHeader))
			if strings.HasPrefix(strings.ToLower(header), "basic ") {
				token = strings.TrimSpace(header[6:])
			} else {
				token = header
			}
		}

		user, err := platformService.ValidateSessionToken(r.Context(), token)
		if err != nil {
			g.Log().Warningf(r.Context(), "auth rejected path=%s err=%v", r.URL.Path, err)
			r.Response.WriteStatusExit(401, g.Map{
				"success": false,
				"message": "unauthorized",
			})
			return
		}
		r.SetCtxVar(consts.CtxKeyCurrentUser, user)
		r.Middleware.Next()
	}
}

func isAnonymousPath(path string) bool {
	if path == "/healthz" || path == "/healthz/ready" || path == "/landing" {
		return true
	}
	if strings.HasPrefix(path, "/swagger") || path == "/api.json" || strings.HasPrefix(path, "/mcp-templates/") {
		return true
	}
	switch path {
	case "/session/login", "/system/info", "/system/config", "/system/init":
		return true
	default:
		return false
	}
}
