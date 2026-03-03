package srverr

type ServerError interface {
	Error() string
	SetError(err string) ServerError

	GetServerError() Error

	SetDetails(details string) ServerError
	GetDetails() string
}

type Error interface {
	String() string
}
