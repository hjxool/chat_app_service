package profile

import (
	"user_service/common"

	"github.com/gin-gonic/gin"
)

// ProfileHandler 受保护的获取个人资料接口
func ProfileHandler(c *gin.Context) {
	// 从上下文中取出中间件注入的用户信息
	userID, _ := c.Get("userID") // 第二个返回值是exist
	username, _ := c.Get("username")

	resp := ProfileResponse{
		UserID:   userID,
		Username: username,
		Remark:   "这是只有登录后才能看到的绝密数据",
	}

	common.SuccessResponse(c, resp)
}
