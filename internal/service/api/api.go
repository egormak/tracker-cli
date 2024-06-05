package api

type APIClient interface {
	Request() ([]byte, error)
}
