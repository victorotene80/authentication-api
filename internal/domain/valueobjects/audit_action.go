package valueobjects

type AuditAction string

const (
    AuditActionUserRegistered     AuditAction = "USER_REGISTERED"
    AuditActionUserLogin          AuditAction = "USER_LOGIN"
    AuditActionUserLoginFailed    AuditAction = "USER_LOGIN_FAILED"
    AuditActionUserLogout         AuditAction = "USER_LOGOUT"
    AuditActionUserPasswordChanged AuditAction = "USER_PASSWORD_CHANGED"
    AuditActionUserUpdated        AuditAction = "USER_UPDATED"
    AuditActionUserDeleted        AuditAction = "USER_DELETED"
    AuditActionUserDeactivated    AuditAction = "USER_DEACTIVATED"
    AuditActionUserActivated      AuditAction = "USER_ACTIVATED"
    AuditActionTokenRefreshed     AuditAction = "TOKEN_REFRESHED"
    AuditActionTokenRevoked       AuditAction = "TOKEN_REVOKED"
    AuditActionOAuthLogin         AuditAction = "OAUTH_LOGIN"
    AuditActionOAuthLoginFailed   AuditAction = "OAUTH_LOGIN_FAILED"
)

func (a AuditAction) String() string {
    return string(a)
}

func (a AuditAction) IsValid() bool {
    switch a {
    case AuditActionUserRegistered, AuditActionUserLogin, AuditActionUserLoginFailed,
        AuditActionUserLogout, AuditActionUserPasswordChanged, AuditActionUserUpdated,
        AuditActionUserDeleted, AuditActionUserDeactivated, AuditActionUserActivated,
        AuditActionTokenRefreshed, AuditActionTokenRevoked, AuditActionOAuthLogin,
        AuditActionOAuthLoginFailed:
        return true
    }
    return false
}
