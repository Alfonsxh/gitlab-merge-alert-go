package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab-merge-alert-go/internal/models"
)

// ResponseHelper 响应辅助器
type ResponseHelper struct{}

// NewResponseHelper 创建响应辅助器
func NewResponseHelper() *ResponseHelper {
	return &ResponseHelper{}
}

// Success 返回成功响应
func (rh *ResponseHelper) Success(c *gin.Context, data interface{}) {
	response := models.SuccessResponse(data)
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// SuccessWithMessage 返回带自定义消息的成功响应
func (rh *ResponseHelper) SuccessWithMessage(c *gin.Context, data interface{}, message string) {
	response := models.SuccessResponseWithMessage(data, message)
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// Error 返回错误响应
func (rh *ResponseHelper) Error(c *gin.Context, message string) {
	response := models.ErrorResponse(message)
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// ValidationError 返回验证错误响应
func (rh *ResponseHelper) ValidationError(c *gin.Context, message string) {
	response := models.ValidationErrorResponse(message)
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// NotFound 返回资源未找到响应
func (rh *ResponseHelper) NotFound(c *gin.Context, resource string) {
	response := models.NotFoundResponse(resource)
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

// BatchOperation 返回批量操作响应
func (rh *ResponseHelper) BatchOperation(c *gin.Context, successCount, failureCount int, results []models.BatchResultItem) {
	response := models.CreateBatchResponse(successCount, failureCount, results)
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// Created 返回资源创建成功响应
func (rh *ResponseHelper) Created(c *gin.Context, data interface{}) {
	response := models.SuccessResponseWithMessage(data, "创建成功")
	c.JSON(201, response) // 创建成功始终返回201
}

// Updated 返回资源更新成功响应
func (rh *ResponseHelper) Updated(c *gin.Context, data interface{}) {
	response := models.SuccessResponseWithMessage(data, "更新成功")
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}

// Deleted 返回资源删除成功响应
func (rh *ResponseHelper) Deleted(c *gin.Context) {
	response := models.SuccessResponseWithMessage(nil, "删除成功")
	c.JSON(models.GetHTTPStatusFromResponseCode(response.Code), response)
}
