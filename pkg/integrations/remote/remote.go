package remote

type Integration interface {
	SetSecret(key string, value string) error
}
