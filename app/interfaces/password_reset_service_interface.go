package interfaces

type PasswordResetService interface {
	RequestReset(email string) (rawToken string, err error)
	ResetPassword(rawToken string, newPassword string) error
}
