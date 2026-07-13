package owlvigil

// Version identifies the SDK version used in diagnostics and User-Agent headers.
const Version = "0.1.0-dev"

// UserAgent returns the default SDK User-Agent.
func UserAgent() string {
	return "owlvigil-go/" + Version
}
