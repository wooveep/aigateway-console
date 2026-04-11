package portal

import (
	"errors"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	portalsvc "github.com/alibaba/aigateway-group/aigateway-console/backend/internal/service/portal"
)

func Bind(group *ghttp.RouterGroup, portalService *portalsvc.Service) {
	group.GET("/v1/consumers", func(r *ghttp.Request) {
		items, err := portalService.ListConsumers(r.Context())
		writeJSON(r, items, err, 200)
	})
	group.POST("/v1/consumers", func(r *ghttp.Request) {
		var req portalsvc.ConsumerMutation
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.SaveConsumer(r.Context(), req, true)
		writeJSON(r, item, err, 200)
	})
	group.GET("/v1/consumers/:name", func(r *ghttp.Request) {
		item, err := portalService.GetConsumer(r.Context(), r.GetRouter("name").String())
		if err == nil && item == nil {
			writeJSON(r, nil, errors.New("consumer not found"), 404)
			return
		}
		writeJSON(r, item, err, 200)
	})
	group.PUT("/v1/consumers/:name", func(r *ghttp.Request) {
		var req portalsvc.ConsumerMutation
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		req.Name = firstNonEmpty(req.Name, r.GetRouter("name").String())
		item, err := portalService.SaveConsumer(r.Context(), req, false)
		writeJSON(r, item, err, 200)
	})
	group.DELETE("/v1/consumers/:name", func(r *ghttp.Request) {
		if err := portalService.DeleteConsumer(r.Context(), r.GetRouter("name").String()); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		r.Response.WriteStatus(204)
		r.ExitAll()
	})
	group.GET("/v1/consumers/departments", func(r *ghttp.Request) {
		items, err := portalService.ListConsumerDepartments(r.Context())
		writeJSON(r, items, err, 200)
	})
	group.POST("/v1/consumers/departments", func(r *ghttp.Request) {
		var req struct {
			Name string `json:"name"`
		}
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		if err := portalService.AddDepartmentCompat(r.Context(), req.Name); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		r.Response.WriteStatus(204)
		r.ExitAll()
	})
	group.PATCH("/v1/consumers/:name/status", func(r *ghttp.Request) {
		var req struct {
			Status string `json:"status"`
		}
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.UpdateConsumerStatus(r.Context(), r.GetRouter("name").String(), req.Status)
		writeJSON(r, item, err, 200)
	})
	group.POST("/v1/consumers/:name/password/reset", func(r *ghttp.Request) {
		item, err := portalService.ResetPassword(r.Context(), r.GetRouter("name").String())
		writeJSON(r, item, err, 200)
	})

	group.GET("/v1/assets/:assetType/:assetId/grants", func(r *ghttp.Request) {
		items, err := portalService.ListAssetGrants(
			r.Context(),
			r.GetRouter("assetType").String(),
			r.GetRouter("assetId").String(),
		)
		writeJSON(r, items, err, 200)
	})
	group.PUT("/v1/assets/:assetType/:assetId/grants", func(r *ghttp.Request) {
		var req struct {
			Grants []portalsvc.AssetGrantRecord `json:"grants"`
		}
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		items, err := portalService.ReplaceAssetGrants(
			r.Context(),
			r.GetRouter("assetType").String(),
			r.GetRouter("assetId").String(),
			req.Grants,
		)
		writeJSON(r, items, err, 200)
	})

	group.GET("/v1/ai/model-assets", func(r *ghttp.Request) {
		items, err := portalService.ListModelAssets(r.Context())
		writeJSON(r, items, err, 200)
	})
	group.GET("/v1/ai/model-assets/options", func(r *ghttp.Request) {
		item, err := portalService.GetModelAssetOptions(r.Context())
		writeJSON(r, item, err, 200)
	})
	group.GET("/v1/ai/model-assets/:assetId", func(r *ghttp.Request) {
		item, err := portalService.GetModelAsset(r.Context(), r.GetRouter("assetId").String())
		if err == nil && item == nil {
			writeJSON(r, nil, errors.New("model asset not found"), 404)
			return
		}
		writeJSON(r, item, err, 200)
	})
	group.POST("/v1/ai/model-assets", func(r *ghttp.Request) {
		var req portalsvc.ModelAsset
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.CreateModelAsset(r.Context(), req)
		writeJSON(r, item, err, 200)
	})
	group.PUT("/v1/ai/model-assets/:assetId", func(r *ghttp.Request) {
		var req portalsvc.ModelAsset
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		req.AssetID = firstNonEmpty(req.AssetID, r.GetRouter("assetId").String())
		item, err := portalService.UpdateModelAsset(r.Context(), r.GetRouter("assetId").String(), req)
		writeJSON(r, item, err, 200)
	})
	group.POST("/v1/ai/model-assets/:assetId/bindings", func(r *ghttp.Request) {
		var req portalsvc.ModelAssetBinding
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.CreateModelBinding(r.Context(), r.GetRouter("assetId").String(), req)
		writeJSON(r, item, err, 200)
	})
	group.PUT("/v1/ai/model-assets/:assetId/bindings/:bindingId", func(r *ghttp.Request) {
		var req portalsvc.ModelAssetBinding
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		req.BindingID = firstNonEmpty(req.BindingID, r.GetRouter("bindingId").String())
		item, err := portalService.UpdateModelBinding(
			r.Context(),
			r.GetRouter("assetId").String(),
			r.GetRouter("bindingId").String(),
			req,
		)
		writeJSON(r, item, err, 200)
	})
	group.POST("/v1/ai/model-assets/:assetId/bindings/:bindingId/publish", func(r *ghttp.Request) {
		item, err := portalService.PublishModelBinding(
			r.Context(),
			r.GetRouter("assetId").String(),
			r.GetRouter("bindingId").String(),
		)
		writeJSON(r, item, err, 200)
	})
	group.POST("/v1/ai/model-assets/:assetId/bindings/:bindingId/unpublish", func(r *ghttp.Request) {
		item, err := portalService.UnpublishModelBinding(
			r.Context(),
			r.GetRouter("assetId").String(),
			r.GetRouter("bindingId").String(),
		)
		writeJSON(r, item, err, 200)
	})
	group.GET("/v1/ai/model-assets/:assetId/bindings/:bindingId/price-versions", func(r *ghttp.Request) {
		items, err := portalService.ListBindingPriceVersions(
			r.Context(),
			r.GetRouter("assetId").String(),
			r.GetRouter("bindingId").String(),
		)
		writeJSON(r, items, err, 200)
	})
	group.POST("/v1/ai/model-assets/:assetId/bindings/:bindingId/price-versions/:versionId/restore", func(r *ghttp.Request) {
		item, err := portalService.RestoreBindingPriceVersion(
			r.Context(),
			r.GetRouter("assetId").String(),
			r.GetRouter("bindingId").String(),
			r.GetRouter("versionId").Int64(),
		)
		writeJSON(r, item, err, 200)
	})

	group.GET("/v1/ai/agent-catalog", func(r *ghttp.Request) {
		items, err := portalService.ListAgentCatalogs(r.Context())
		writeJSON(r, items, err, 200)
	})
	group.GET("/v1/ai/agent-catalog/options", func(r *ghttp.Request) {
		item, err := portalService.GetAgentCatalogOptions(r.Context())
		writeJSON(r, item, err, 200)
	})
	group.GET("/v1/ai/agent-catalog/:agentId", func(r *ghttp.Request) {
		item, err := portalService.GetAgentCatalog(r.Context(), r.GetRouter("agentId").String())
		if err == nil && item == nil {
			writeJSON(r, nil, errors.New("agent catalog not found"), 404)
			return
		}
		writeJSON(r, item, err, 200)
	})
	group.POST("/v1/ai/agent-catalog", func(r *ghttp.Request) {
		var req portalsvc.AgentCatalogRecord
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.CreateAgentCatalog(r.Context(), req)
		writeJSON(r, item, err, 200)
	})
	group.PUT("/v1/ai/agent-catalog/:agentId", func(r *ghttp.Request) {
		var req portalsvc.AgentCatalogRecord
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		req.AgentID = firstNonEmpty(req.AgentID, r.GetRouter("agentId").String())
		item, err := portalService.UpdateAgentCatalog(r.Context(), r.GetRouter("agentId").String(), req)
		writeJSON(r, item, err, 200)
	})
	group.POST("/v1/ai/agent-catalog/:agentId/publish", func(r *ghttp.Request) {
		item, err := portalService.PublishAgentCatalog(r.Context(), r.GetRouter("agentId").String())
		writeJSON(r, item, err, 200)
	})
	group.POST("/v1/ai/agent-catalog/:agentId/unpublish", func(r *ghttp.Request) {
		item, err := portalService.UnpublishAgentCatalog(r.Context(), r.GetRouter("agentId").String())
		writeJSON(r, item, err, 200)
	})

	group.GET("/v1/org/departments/tree", func(r *ghttp.Request) {
		items, err := portalService.ListDepartmentTree(r.Context())
		writeJSON(r, items, err, 200)
	})
	group.POST("/v1/org/departments", func(r *ghttp.Request) {
		var req portalsvc.DepartmentMutation
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.CreateDepartment(r.Context(), req)
		writeJSON(r, item, err, 200)
	})
	group.PUT("/v1/org/departments/:departmentId", func(r *ghttp.Request) {
		var req portalsvc.DepartmentMutation
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.UpdateDepartment(r.Context(), r.GetRouter("departmentId").String(), req)
		writeJSON(r, item, err, 200)
	})
	group.PATCH("/v1/org/departments/:departmentId/move", func(r *ghttp.Request) {
		var req struct {
			ParentDepartmentID string `json:"parentDepartmentId"`
		}
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.MoveDepartment(r.Context(), r.GetRouter("departmentId").String(), req.ParentDepartmentID)
		writeJSON(r, item, err, 200)
	})
	group.DELETE("/v1/org/departments/:departmentId", func(r *ghttp.Request) {
		if err := portalService.DeleteDepartment(r.Context(), r.GetRouter("departmentId").String()); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		r.Response.WriteStatus(204)
		r.ExitAll()
	})

	group.GET("/v1/org/accounts", func(r *ghttp.Request) {
		items, err := portalService.ListAccounts(r.Context())
		writeJSON(r, items, err, 200)
	})
	group.POST("/v1/org/accounts", func(r *ghttp.Request) {
		var req portalsvc.AccountMutation
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.CreateAccount(r.Context(), req)
		writeJSON(r, item, err, 200)
	})
	group.PUT("/v1/org/accounts/:consumerName", func(r *ghttp.Request) {
		var req portalsvc.AccountMutation
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.UpdateAccount(r.Context(), r.GetRouter("consumerName").String(), req)
		writeJSON(r, item, err, 200)
	})
	group.PATCH("/v1/org/accounts/:consumerName/assignment", func(r *ghttp.Request) {
		var req struct {
			DepartmentID       string `json:"departmentId"`
			ParentConsumerName string `json:"parentConsumerName"`
		}
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.UpdateAccountAssignment(r.Context(), r.GetRouter("consumerName").String(), req.DepartmentID, req.ParentConsumerName)
		writeJSON(r, item, err, 200)
	})
	group.PATCH("/v1/org/accounts/:consumerName/status", func(r *ghttp.Request) {
		var req struct {
			Status string `json:"status"`
		}
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.UpdateAccountStatus(r.Context(), r.GetRouter("consumerName").String(), req.Status)
		writeJSON(r, item, err, 200)
	})

	group.POST("/v1/portal/invite-codes", func(r *ghttp.Request) {
		var req struct {
			ExpiresInDays int `json:"expiresInDays"`
		}
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.CreateInviteCode(r.Context(), req.ExpiresInDays)
		writeJSON(r, item, err, 200)
	})
	group.GET("/v1/portal/invite-codes", func(r *ghttp.Request) {
		items, err := portalService.ListInviteCodes(r.Context(), portalsvc.InviteCodeQuery{
			PageNum:  r.GetQuery("pageNum").Int(),
			PageSize: r.GetQuery("pageSize").Int(),
			Status:   r.GetQuery("status").String(),
		})
		writeJSON(r, items, err, 200)
	})
	group.PATCH("/v1/portal/invite-codes/:code", func(r *ghttp.Request) {
		var req struct {
			Status string `json:"status"`
		}
		if err := r.Parse(&req); err != nil {
			writeJSON(r, nil, err, 400)
			return
		}
		item, err := portalService.UpdateInviteCodeStatus(r.Context(), r.GetRouter("code").String(), req.Status)
		writeJSON(r, item, err, 200)
	})
}

func writeJSON(r *ghttp.Request, data any, err error, successStatus int) {
	if err != nil {
		status := 400
		message := err.Error()
		if strings.Contains(strings.ToLower(message), "not found") {
			status = 404
		} else if strings.Contains(strings.ToLower(message), "unavailable") {
			status = 503
		}
		r.Response.WriteStatusExit(status, g.Map{"success": false, "message": message})
		return
	}
	if successStatus > 0 && successStatus != 200 {
		r.Response.WriteStatus(successStatus)
	}
	r.Response.WriteJsonExit(g.Map{"success": true, "data": data})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
