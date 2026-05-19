package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

// Parse 从请求参数解析分页信息
func Parse(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.Query("page"))
	pageSize, _ = strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = defaultPage
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return
}

// Offset 根据 page 和 pageSize 计算偏移量
func Offset(page, pageSize int) int {
	return (page - 1) * pageSize
}
