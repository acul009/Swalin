package pki

const (
	passwordFilePath = "password.pem"
	hostFilname      = "host"
)

var ErrNotInitialized = hostNotInitializedError{}

type hostNotInitializedError struct {
}

func (e hostNotInitializedError) Error() string {
	return "server not yet initialized"
}

func (e hostNotInitializedError) Is(target error) bool {
	_, ok := target.(hostNotInitializedError)
	return ok
}
