package utils

import (
	"github.com/go-playground/validator/v10"
	"sync"
)

var (
	once      sync.Once
	validateInstance *validator.Validate
)

func NewValidator() *validator.Validate {
	once.Do(func() {
		validateInstance = validator.New()
	})
	return validateInstance
}
