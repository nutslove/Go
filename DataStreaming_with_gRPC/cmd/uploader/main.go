package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"protobuf/pb"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	_ "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	_ "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedStreamServiceServer
}

func (*server) Upload(ctx context.Context, req *pb.StreamRequest) (*pb.StreamResponse, error) {
	// ここではグローバルなTracerProviderを用いてSpanを開始
	tr := otel.Tracer("uploader")
	_, span := tr.Start(ctx, "uploader started")
	defer span.End()

	span.AddEvent("upload executed")
	fmt.Println("trace-id:", span.SpanContext().TraceID().String())
	fmt.Println("span-id:", span.SpanContext().SpanID().String())

	fmt.Printf("Received request: %v\n", req)

	dataSize := len([]byte(req.GetData()))
	return &pb.StreamResponse{Size: int32(dataSize)}, nil
}

func main() {
	ctxInit := context.Background()
	exporter, err := otlptracegrpc.New(ctxInit,
		otlptracegrpc.WithEndpoint("10.111.1.179:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalln("Failed to set exporter for otlp/grpc:", err)
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("uploader"),
		semconv.ServiceVersionKey.String("1.0.1"),
	)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			trace.WithBatchTimeout(1*time.Second),
			trace.WithMaxExportBatchSize(128),
		),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(resource),
	)
	defer tp.Shutdown(ctxInit)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	pb.RegisterStreamServiceServer(s, &server{})

	fmt.Println("server is running...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
