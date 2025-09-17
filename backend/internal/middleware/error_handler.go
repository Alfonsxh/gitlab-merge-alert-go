package middleware

import (
	"strings"

	"gitlab-merge-alert-go/internal/models"
	"gitlab-merge-alert-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.GetLogger().Errorf("Panic recovered: %v", recovered)

		response := models.InternalErrorResponse("服务器内部错误")
		c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
		c.Abort()
	})
}

// ResponseHelper 统一响应处理器
type ResponseHelper struct{}

// NewResponseHelper 创建响应处理器
func NewResponseHelper() *ResponseHelper {
	return &ResponseHelper{}
}

// Success 返回成功响应
func (rh *ResponseHelper) Success(c *gin.Context, data interface{}) {
	response := models.SuccessResponse(data)
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// Error 处理错误并返回适当响应
func (rh *ResponseHelper) Error(c *gin.Context, err error) {
	var response models.ResponseBaseModel

	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "不存在") || strings.Contains(errMsg, "not found"):
		response = models.NotFoundResponse("资源")
	case strings.Contains(errMsg, "已存在") || strings.Contains(errMsg, "已被占用"):
		response = models.ConflictResponse(errMsg)
	case strings.Contains(errMsg, "验证失败") || strings.Contains(errMsg, "参数无效"):
		response = models.ValidationErrorResponse(errMsg)
	case strings.Contains(errMsg, "权限") || strings.Contains(errMsg, "未授权"):
		response = models.ErrorResponseWithCode(403, errMsg)
	default:
		response = models.ErrorResponse(errMsg)
	}

	logger.GetLogger().Errorf("Business error: %v", err)
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// Created 返回创建成功响应
func (rh *ResponseHelper) Created(c *gin.Context, data interface{}) {
	response := models.SuccessResponseWithMessage(data, "创建成功")
	c.JSON(201, response)
}

// Updated 返回更新成功响应
func (rh *ResponseHelper) Updated(c *gin.Context, data interface{}) {
	response := models.SuccessResponseWithMessage(data, "更新成功")
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// Deleted 返回删除成功响应
func (rh *ResponseHelper) Deleted(c *gin.Context) {
	response := models.SuccessResponseWithMessage(nil, "删除成功")
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// BatchOperation 返回批量操作响应
func (rh *ResponseHelper) BatchOperation(c *gin.Context, successCount, failureCount int, results []models.BatchResultItem) {
	response := models.CreateBatchResponse(successCount, failureCount, results)
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// ValidationError 返回验证错误响应
func (rh *ResponseHelper) ValidationError(c *gin.Context, message string) {
	response := models.ValidationErrorResponse(message)
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// Conflict 返回资源冲突响应
func (rh *ResponseHelper) Conflict(c *gin.Context, message string) {
	response := models.ConflictResponse(message)
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// InternalError 返回内部服务器错误响应
func (rh *ResponseHelper) InternalError(c *gin.Context, message string) {
	response := models.InternalErrorResponse(message)
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// ErrorWithMessage 返回带消息的错误响应
func (rh *ResponseHelper) ErrorWithMessage(c *gin.Context, message string) {
	response := models.ErrorResponse(message)
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}
