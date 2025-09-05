package verification

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ApplyVerificationCode 申请验证码（公开接口）
// 注意：这个函数现在由路由层直接调用verificationService的方法，此文件可以删除或保留作为参考
func ApplyVerificationCode(c *gin.Context) {
	var request struct {
		Name    string `json:"name" binding:"required"`
		Channel string `json:"channel" binding:"required,oneof=qq phone"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 注意：这里应该使用注入的verificationService实例
	// 由于路由层已经处理了具体逻辑，此函数可以删除
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "此接口已废弃，请使用路由层直接调用",
	})
}
