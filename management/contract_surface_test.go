package management_test

import (
	"reflect"
	"testing"

	"github.com/owlvigil/owlvigil-go/management"
)

func TestClient_ExposesOnlyPublishedOpenAPIManagementMethods(t *testing.T) {
	const (
		publishedOperations = 130
		convenienceMethods  = 1 // CreatePaymentMethodSetupIntentForWorkspace
		clientMethods       = 1 // BaseURL
	)

	want := publishedOperations + convenienceMethods + clientMethods
	if got := reflect.TypeOf(&management.Client{}).NumMethod(); got != want {
		t.Fatalf("published management client method count = %d, want %d", got, want)
	}
}
