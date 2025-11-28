package commands

type RegisterUserCommand struct {
    Username  string
    Email     string
    Password  string // Empty for OAuth registrations
    Phone     string
    FirstName string
    LastName  string
    Role      string
    IPAddress string
    UserAgent string
    
    // OAuth-specific fields
    OAuthProvider string // "google", "github", etc.
    OAuthID       string // External OAuth user ID
    IsOAuth       bool   // Flag to identify OAuth registration
}

func (c RegisterUserCommand) CommandName() string {
    return "RegisterUserCommand"
}

// IsOAuthRegistration checks if this is an OAuth-based registration
func (c RegisterUserCommand) IsOAuthRegistration() bool {
    return c.IsOAuth && c.OAuthProvider != "" && c.OAuthID != ""
}