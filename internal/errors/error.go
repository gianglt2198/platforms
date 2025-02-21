package myerrors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

func (r *AppError) IsZero() bool {
	return r.Status == 0
}

func (r *AppError) Error() string {
	return fmt.Sprintf("%s:%s", r.Code, r.Message)
}

func NewAppError(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, Status: status}
}

func (r *AppError) WithData(data string) *AppError {
	r.Data = data
	return r
}

func QueryNotFound(message string) *AppError {
	return NewAppError("query.001", message, http.StatusNotFound)
}

func QueryInvalid(message string) *AppError {
	return NewAppError("query.002", message, http.StatusInternalServerError)
}

func MQTimeout() *AppError {
	return NewAppError(
		"mq.001",
		"response from MQ took more than 2 seconds.",
		http.StatusGatewayTimeout,
	)
}

func MQUnauthorization() *AppError {
	return NewAppError("mq.002", "need authenticated for request.", http.StatusUnauthorized)
}

func MQAccessDenined() *AppError {
	return NewAppError("mq.003", "need specific roles for request", http.StatusForbidden)
}

func MQWrongAccess() *AppError {
	return NewAppError("mq.004", "access with wrong user type", http.StatusForbidden)
}

func IsQueryNotFound(err error) bool {
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	return appErr.Code == "query.001" && appErr.Status == http.StatusNotFound
}
