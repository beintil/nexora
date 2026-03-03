package srverr

import "fmt"

type serverError struct {
	error    string
	details  string
	location string

	servError Error
}

func NewServerError(servError Error, location string) ServerError {
	return &serverError{
		servError: servError,
		location:  location,
	}
}

func (s *serverError) Error() string {
	if s.error == "" {
		return fmt.Sprintf("%s: %s", s.location, s.servError.String())
	}
	return fmt.Sprintf("%s: %s", s.location, s.error)
}

func (s *serverError) SetError(err string) ServerError {
	s.error = err
	return s
}

func (s *serverError) SetDetails(details string) ServerError {
	s.details = details
	return s
}

func (s *serverError) GetServerError() Error {
	return s.servError
}

func (s *serverError) GetDetails() string {
	return s.details
}
