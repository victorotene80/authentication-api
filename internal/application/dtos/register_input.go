package dtos

type RegisterInput struct {
    Username      string
    Email         string
    Password      string // Empty for OAuth
    Phone         string
    FirstName     string
    LastName      string
    Role          string
    OAuthProvider string
    OAuthID       string
    IsOAuth       bool
}