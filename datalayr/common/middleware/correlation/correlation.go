package correlation

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const CorrelationIDKey = "correlation-id"

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return correlationIdUnaryServerInterceptor
}

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return correlationIdUnaryClientInterceptor
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return correlationIdStreamServerInterceptor
}

func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return correlationIdStreamClientInterceptor
}

// correlationIdUnaryClientInterceptor is a gRPC interceptor that injects a
// correlation ID and context into the request metadata.
func correlationIdUnaryClientInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	// Extract the fields from the context.
	fields := logging.ExtractFields(ctx)

	// Extract the correlation ID from the fields, if present.
	var correlationID string
	for i := 0; i < len(fields); i += 2 {
		if fields[i] == CorrelationIDKey {
			correlationID = fields[i+1]
			break
		}
	}
	if correlationID == "" {
		// Generate a new correlation ID if none is present in the fields.
		correlationID = generateCorrelationID()
	}

	// Inject the correlation ID and context into the request metadata.
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	md[CorrelationIDKey] = []string{correlationID}
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Call the invoker with the modified context.
	return invoker(ctx, method, req, reply, cc, opts...)
}

// correlationIdUnaryServerInterceptor is a gRPC interceptor that injects a
// correlation ID and context into the request metadata.
func correlationIdUnaryServerInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}

	var correlationID logging.Fields

	// Extract the correlation ID from the request metadata, if present.
	correlationIDs := md.Get(CorrelationIDKey)
	if len(correlationIDs) == 0 {
		// Generate a new correlation ID if none is present in the request metadata.
		correlationIDValue := generateCorrelationID()
		correlationID = logging.Fields{CorrelationIDKey, correlationIDValue}
		md[CorrelationIDKey] = []string{correlationIDValue}
	} else {
		correlationID = logging.Fields{CorrelationIDKey, correlationIDs[0]}
	}

	// Extract the fields from the context.
	fields := logging.ExtractFields(ctx)
	fields = fields.AppendUnique(correlationID)

	// Inject the fields into the context.
	ctx = logging.InjectFields(ctx, fields)
	newCtx := metadata.NewIncomingContext(ctx, md)

	// Call the handler with the modified context.
	return handler(newCtx, req)
}

// correlationIdStreamServerInterceptor is a gRPC interceptor that injects a
// correlation ID and context into the request metadata for stream server requests.
func correlationIdStreamServerInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		md = metadata.MD{}
	}

	var correlationID logging.Fields

	// Extract the correlation ID from the request metadata, if present.
	correlationIDs := md.Get(CorrelationIDKey)
	if len(correlationIDs) == 0 {
		// Generate a new correlation ID if none is present in the request metadata.
		correlationIDValue := generateCorrelationID()
		correlationID = logging.Fields{CorrelationIDKey, correlationIDValue}
		md[CorrelationIDKey] = []string{correlationIDValue}

	} else {
		correlationID = logging.Fields{CorrelationIDKey, correlationIDs[0]}
	}

	// Extract the fields from the context.
	fields := logging.ExtractFields(stream.Context())
	fields = fields.AppendUnique(correlationID)

	// Inject the fields into the context.
	ctx := logging.InjectFields(stream.Context(), fields)
	newCtx := metadata.NewIncomingContext(ctx, md)

	// Call the handler with the modified stream.
	return handler(srv, &grpc_middleware.WrappedServerStream{
		ServerStream:   stream,
		WrappedContext: newCtx,
	})
}

// correlationIdStreamClientInterceptor is a gRPC interceptor that injects a
// correlation ID and context into the request metadata for stream client requests.
func correlationIdStreamClientInterceptor(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	// Extract the fields from the context.
	fields := logging.ExtractFields(ctx)

	// Extract the correlation ID from the fields, if present.
	var correlationID string
	for i := 0; i < len(fields); i += 2 {
		if fields[i] == CorrelationIDKey {
			correlationID = fields[i+1]
			break
		}
	}

	if correlationID == "" {
		// Generate a new correlation ID if none is present in the fields.
		correlationID = generateCorrelationID()
	}

	// Inject the correlation ID and context into the request metadata.
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	md[CorrelationIDKey] = []string{correlationID}
	newCtx := metadata.NewOutgoingContext(ctx, md)

	// Call the streamer with the modified context.
	return streamer(newCtx, desc, cc, method, opts...)
}

// Generate a random 16 byte id
func generateCorrelationID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}
