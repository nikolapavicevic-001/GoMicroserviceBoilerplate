package grpc

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const tracerName = "grpc-server"

// TracingInterceptor returns a gRPC unary server interceptor that adds OpenTelemetry tracing
func TracingInterceptor() grpc.UnaryServerInterceptor {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract trace context from incoming metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ctx = propagator.Extract(ctx, metadataCarrier(md))
		}

		// Start a new span
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", info.FullMethod),
			),
		)
		defer span.End()

		// Call the handler
		resp, err := handler(ctx, req)

		// Record error if any
		if err != nil {
			span.RecordError(err)
			if st, ok := status.FromError(err); ok {
				span.SetStatus(codes.Error, st.Message())
				span.SetAttributes(attribute.String("rpc.grpc.status_code", st.Code().String()))
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return resp, err
	}
}

// StreamTracingInterceptor returns a gRPC stream server interceptor that adds OpenTelemetry tracing
func StreamTracingInterceptor() grpc.StreamServerInterceptor {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		// Extract trace context from incoming metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ctx = propagator.Extract(ctx, metadataCarrier(md))
		}

		// Start a new span
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", info.FullMethod),
				attribute.Bool("rpc.grpc.is_client_stream", info.IsClientStream),
				attribute.Bool("rpc.grpc.is_server_stream", info.IsServerStream),
			),
		)
		defer span.End()

		// Wrap the stream with the new context
		wrappedStream := &tracingServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// Call the handler
		err := handler(srv, wrappedStream)

		// Record error if any
		if err != nil {
			span.RecordError(err)
			if st, ok := status.FromError(err); ok {
				span.SetStatus(codes.Error, st.Message())
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// ClientTracingInterceptor returns a gRPC unary client interceptor that adds OpenTelemetry tracing
func ClientTracingInterceptor() grpc.UnaryClientInterceptor {
	tracer := otel.Tracer("grpc-client")
	propagator := otel.GetTextMapPropagator()

	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Start a new span
		ctx, span := tracer.Start(ctx, method,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", method),
				attribute.String("net.peer.name", cc.Target()),
			),
		)
		defer span.End()

		// Inject trace context into outgoing metadata
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		propagator.Inject(ctx, metadataCarrier(md))
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Call the invoker
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Record error if any
		if err != nil {
			span.RecordError(err)
			if st, ok := status.FromError(err); ok {
				span.SetStatus(codes.Error, st.Message())
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// metadataCarrier is an adapter to use gRPC metadata as a TextMapCarrier
type metadataCarrier metadata.MD

func (mc metadataCarrier) Get(key string) string {
	vals := metadata.MD(mc).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (mc metadataCarrier) Set(key, value string) {
	metadata.MD(mc).Set(key, value)
}

func (mc metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(mc))
	for k := range mc {
		keys = append(keys, k)
	}
	return keys
}

// tracingServerStream wraps grpc.ServerStream with a custom context
type tracingServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *tracingServerStream) Context() context.Context {
	return s.ctx
}

