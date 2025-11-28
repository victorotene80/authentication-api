package dtos

type ApiResponse[T any] struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    T      `json:"data,omitempty"`
}

func NewSuccessResponse[T any](message string, data T) ApiResponse[T] {
    return ApiResponse[T]{
        Code:    200,
        Message: message,
        Data:    data,
    }
}

func NeFailedResponse(code int, message string) ApiResponse[interface{}] {
    return ApiResponse[interface{}]{
        Code:    code,
        Message: message,
        Data:    nil,
    }
}

