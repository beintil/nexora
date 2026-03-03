package srverr

type ErrorTypeInternalServerError string

func (err ErrorTypeInternalServerError) String() string {
	return string(err)
}

const (
	ErrInternalServerError ErrorTypeInternalServerError = "internal server error"
)

type ErrorTypeBadRequest string

func (err ErrorTypeBadRequest) String() string {
	return string(err)
}

type ErrorTypeNotFound string

func (err ErrorTypeNotFound) String() string {
	return string(err)
}

type ErrorTypeConflict string

func (err ErrorTypeConflict) String() string {
	return string(err)
}

type ErrorTypeUnauthorized string

func (err ErrorTypeUnauthorized) String() string {
	return string(err)
}

type ErrorTypeForbidden string

func (err ErrorTypeForbidden) String() string {
	return string(err)
}

const (
	ErrUnauthorized ErrorTypeUnauthorized = "unauthorized"
	ErrForbidden    ErrorTypeForbidden    = "forbidden"
)
