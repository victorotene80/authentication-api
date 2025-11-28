package dtos

import (
	"authentication/internal/domain/aggregates"
	"authentication/internal/domain/valueobjects"
)

type LoginResult struct {
	User         *aggregates.UserAggregate
	AccessToken  *valueobjects.Token
	RefreshToken *valueobjects.Token
	SessionID    string
}