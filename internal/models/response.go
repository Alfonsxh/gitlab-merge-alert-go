package models

import (
	"net/http"
)

// ResponseBaseModel 统一API响应格式
type ResponseBaseModel struct {
	Code    int         `json:"code"`    // 正常 0，错误 1
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 返回数据，错误时为 nil
	Success bool        `json:"success"` // 是否成功
}

// SuccessResponse 创建成功响应
func SuccessResponse(data interface{}) ResponseBaseModel {
	return ResponseBaseModel{
		Code:    0,
		Message: "操作成功",
		Data:    data,
		Success: true,
	}
}

// SuccessResponseWithMessage 创建带自定义消息的成功响应
func SuccessResponseWithMessage(data interface{}, message string) ResponseBaseModel {
	return ResponseBaseModel{
		Code:    0,
		Message: message,
		Data:    data,
		Success: true,
	}
}

// ErrorResponse 创建错误响应
func ErrorResponse(message string) ResponseBaseModel {
	return ResponseBaseModel{
		Code:    1,
		Message: message,
		Data:    nil,
		Success: false,
	}
}

// ErrorResponseWithCode 创建带自定义错误码的错误响应
func ErrorResponseWithCode(code int, message string) ResponseBaseModel {
	return ResponseBaseModel{
		Code:    code,
		Message: message,
		Data:    nil,
		Success: false,
	}
}

// ValidationErrorResponse 创建验证错误响应
func ValidationErrorResponse(message string) ResponseBaseModel {
	return ResponseBaseModel{
		Code:    400,
		Message: "请求参数验证失败: " + message,
		Data:    nil,
		Success: false,
	}
}

// NotFoundResponse 创建资源未找到响应
func NotFoundResponse(resource string) ResponseBaseModel {
	return ResponseBaseModel{
		Code:    404,
		Message: resource + "不存在",
		Data:    nil,
		Success: false,
	}
}

// ConflictResponse 创建资源冲突响应
func ConflictResponse(message string) ResponseBaseModel {
	return ResponseBaseModel{
		Code:    409,
		Message: message,
		Data:    nil,
		Success: false,
	}
}

// InternalErrorResponse 创建内部服务器错误响应
func InternalErrorResponse(message string) ResponseBaseModel {
	return ResponseBaseModel{
		Code:    500,
		Message: "服务器内部错误: " + message,
		Data:    nil,
		Success: false,
	}
}

// BatchOperationResponse 批量操作的特殊响应格式
type BatchOperationResponse struct {
	SuccessCount int               `json:"success_count"`
	FailureCount int               `json:"failure_count"`
	Results      []BatchResultItem `json:"results"`
	TotalCount   int               `json:"total_count"`
}

// BatchResultItem 批量操作中单个项目的结果
type BatchResultItem struct {
	ID      interface{} `json:"id"`
	Name    string      `json:"name"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// CreateBatchResponse 创建批量操作响应
func CreateBatchResponse(successCount, failureCount int, results []BatchResultItem) ResponseBaseModel {
	batchData := BatchOperationResponse{
		SuccessCount: successCount,
		FailureCount: failureCount,
		Results:      results,
		TotalCount:   successCount + failureCount,
	}

	var code int
	var message string
	var success bool

	if successCount == 0 {
		// 全部失败
		code = 1
		message = "批量操作失败，所有项目都未能成功处理"
		success = false
	} else if failureCount > 0 {
		// 部分成功
		code = 207 // Multi-Status
		message = "批量操作部分成功"
		success = true
	} else {
		// 全部成功
		code = 0
		message = "批量操作成功"
		success = true
	}

	return ResponseBaseModel{
		Code:    code,
		Message: message,
		Data:    batchData,
		Success: success,
	}
}

// GetHTTPStatusFromResponseCode 根据响应码获取HTTP状态码
func GetHTTPStatusFromResponseCode(code int) int {
	switch code {
	case 0:
		return http.StatusOK
	case 1:
		return http.StatusBadRequest
	case 207:
		return http.StatusMultiStatus
	case 400:
		return http.StatusBadRequest
	case 404:
		return http.StatusNotFound
	case 409:
		return http.StatusConflict
	case 500:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}
