package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/pkg/response"
	"github.com/imkerbos/mxcmdb/internal/service"
)

// RequireRole 角色权限中间件（保留向后兼容）
func RequireRole(roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		roleSet[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusForbidden, 40300, "无权访问")
			c.Abort()
			return
		}

		if _, ok := roleSet[role.(string)]; !ok {
			response.Error(c, http.StatusForbidden, 40300, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission 细粒度权限中间件
func RequirePermission(permSvc *service.PermissionService, permCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusForbidden, 40300, "无权访问")
			c.Abort()
			return
		}

		if !permSvc.HasPermission(role.(string), permCode) {
			response.Error(c, http.StatusForbidden, 40300, "权限不足: "+permCode)
			c.Abort()
			return
		}

		c.Next()
	}
}
