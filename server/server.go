// Package main implements a simple gRPC server that demonstrates how to use gRPC-Go libraries
// to perform unary, client streaming, server streaming and full duplex RPCs.
//
// It implements the route guide service whose definition can be found in routeguide/route_guide.proto.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"

	pbt "github.com/golang/protobuf/ptypes"
	pb "github.com/r7wang/auto-chat/grpctest"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	port       = flag.Int("port", 10000, "The server port")
)

type grpcTestServer struct{}

func (s *grpcTestServer) GetMessage(ctx context.Context, userMessage *pb.UserMessage) (*pb.ProcessedMessage, error) {
	curTime := pbt.TimestampNow()
	return &pb.ProcessedMessage{
			Message:      userMessage,
			ReceivedTime: curTime},
		nil
}

func (s *grpcTestServer) GetMultiMessage(userMessage *pb.UserMessage, stream pb.GrpcTest_GetMultiMessageServer) error {
	curTime := time.Now()
	for i := 0; i < 5; i++ {
		addTime := curTime.Add(time.Duration(i) * time.Hour)
		pbTime, err := pbt.TimestampProto(addTime)
		if err != nil {
			return err
		}
		procMessage := &pb.ProcessedMessage{
			Message:      userMessage,
			ReceivedTime: pbTime,
		}
		stream.Send(procMessage)
		time.Sleep(1 * time.Second)
	}
	// Make the client wait additional time, blocking the RPC call from completing.
	time.Sleep(5 * time.Second)
	return nil
}

func (s *grpcTestServer) SendMultiMessage(stream pb.GrpcTest_SendMultiMessageServer) error {
	curTime := pbt.TimestampNow()
	for {
		userMessage, err := stream.Recv()
		if err == io.EOF {
			procMessage := &pb.ProcessedMessage{
				Message:      userMessage,
				ReceivedTime: curTime,
			}
			return stream.SendAndClose(procMessage)
		}
		if err != nil {
			return err
		}
	}
}

func (s *grpcTestServer) GetSendMultiMessage(stream pb.GrpcTest_GetSendMultiMessageServer) error {
	curTime := pbt.TimestampNow()
	for {
		userMessage, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		procMessage := &pb.ProcessedMessage{
			Message:      userMessage,
			ReceivedTime: curTime,
		}
		if err := stream.Send(procMessage); err != nil {
			return err
		}
	}
}

/*
* proto.Equal()
* updates to protobuf interface
 */

func newServer() *grpcTestServer {
	s := &grpcTestServer{}
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if *tls {
		if *certFile == "" {
			*certFile = testdata.Path("server1.pem")
		}
		if *keyFile == "" {
			*keyFile = testdata.Path("server1.key")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterGrpcTestServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
