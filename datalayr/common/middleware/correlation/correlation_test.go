package correlation

import (
	"context"
	"fmt"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func setupSuite(t *testing.T) func(t *testing.T) {
	return func(t *testing.T) {
		fmt.Println("Tearing down suite")
	}
}

// TODO: Was able to unit test the unary server interceptor. Testing the client
// interceptor requires more work such as having a mock server and client.
func TestCorrelationIdUnaryServerInterceptor(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	// Set up a request context with some metadata.
	md := metadata.New(map[string]string{
		"foo": "bar",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	// Set up a mock request and handler.
	req := struct{}{}
	info := &grpc.UnaryServerInfo{}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return ctx, nil
	}

	// Call the interceptor with the mock request and handler.
	resp, err := correlationIdUnaryServerInterceptor(ctx, req, info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Our handler was setup so that it returns the new context
	newCtx := resp.(context.Context)

	md, ok := metadata.FromIncomingContext(newCtx)
	if !ok {
		t.Fatal("missing metadata")
	}

	// Verify that the interceptor added a correlation ID to the request metadata.
	correlationIDs := md.Get(CorrelationIDKey)
	if len(correlationIDs) == 0 {
		t.Fatal("missing correlation ID")
	}

	// Verify that the interceptor injected the correlation ID into the context.
	fields := logging.ExtractFields(newCtx)
	var correlationID string
	for i := 0; i < len(fields); i += 2 {
		if fields[i] == CorrelationIDKey {
			correlationID = fields[i+1]
			break
		}
	}
	if correlationID == "" {
		// Generate a new correlation ID if none is present in the fields.
		t.Fatal("could not find correlation ID")
	}
}
