package owlvigil

import "fmt"

// APIError describes a non-2xx response returned by the OwlVigil API.
type APIError struct {
	StatusCode int
	RequestID  string
	Code       string
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.RequestID != "" {
		return fmt.Sprintf("owlvigil: status=%d code=%s request_id=%s message=%s", e.StatusCode, e.Code, e.RequestID, e.Message)
	}
	return fmt.Sprintf("owlvigil: status=%d code=%s message=%s", e.StatusCode, e.Code, e.Message)
}

// OAuthError describes an OAuth2 error response.
type OAuthError struct {
	StatusCode       int
	ErrorCode        string
	ErrorDescription string
	ErrorURI         string
}

func (e *OAuthError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.ErrorDescription != "" {
		return fmt.Sprintf("owlvigil oauth2: status=%d error=%s description=%s", e.StatusCode, e.ErrorCode, e.ErrorDescription)
	}
	return fmt.Sprintf("owlvigil oauth2: status=%d error=%s", e.StatusCode, e.ErrorCode)
}
