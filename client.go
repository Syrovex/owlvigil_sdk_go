package owlvigil

// ResponseMeta contains response metadata common to SDK calls.
type ResponseMeta struct {
	RequestID string
	Code      string
	Message   string
}

// Envelope is the standard OwlVigil response envelope.
type Envelope[T any] struct {
	RequestID string `json:"request_id"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Data      T      `json:"data"`
}
