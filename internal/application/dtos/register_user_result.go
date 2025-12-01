package dtos

type RegisterUserResult struct {
    UserID       string
    Email        string
    AccessToken  string
    RefreshToken string
    ExpiresAt    int64
}