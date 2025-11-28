package dtos

import "authentication/internal/domain/valueobjects"

type TokenPair struct {
    AccessToken  *valueobjects.Token
    RefreshToken *valueobjects.Token
}
