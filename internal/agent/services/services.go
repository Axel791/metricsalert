package services

type AuthService interface {
	ComputeHash(body []byte) string
}
