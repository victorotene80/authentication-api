package dtos

import (
	//"authentication/internal/domain/aggregates"
	//"authentication/internal/domain/valueobjects"
)

/*type LoginResult struct {
	User         *aggregates.UserAggregate
	AccessToken  *valueobjects.Token
	RefreshToken *valueobjects.Token
	SessionID    string
}*/

type LoginResult struct {
	UserID       string
	Email        string
	Username     string
	FirstName    string
	LastName     string
	Role         string
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
	IsNewUser    bool 
}