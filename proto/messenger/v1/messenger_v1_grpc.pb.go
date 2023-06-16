// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: proto/messenger/v1/messenger_v1.proto

package messenger_v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MessengerServiceClient is the client API for MessengerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MessengerServiceClient interface {
	StreamEnvelopes(ctx context.Context, in *Conversation, opts ...grpc.CallOption) (MessengerService_StreamEnvelopesClient, error)
	SendEnvelope(ctx context.Context, opts ...grpc.CallOption) (MessengerService_SendEnvelopeClient, error)
}

type messengerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMessengerServiceClient(cc grpc.ClientConnInterface) MessengerServiceClient {
	return &messengerServiceClient{cc}
}

func (c *messengerServiceClient) StreamEnvelopes(ctx context.Context, in *Conversation, opts ...grpc.CallOption) (MessengerService_StreamEnvelopesClient, error) {
	stream, err := c.cc.NewStream(ctx, &MessengerService_ServiceDesc.Streams[0], "/messenger.MessengerService/StreamEnvelopes", opts...)
	if err != nil {
		return nil, err
	}
	x := &messengerServiceStreamEnvelopesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type MessengerService_StreamEnvelopesClient interface {
	Recv() (*Envelope, error)
	grpc.ClientStream
}

type messengerServiceStreamEnvelopesClient struct {
	grpc.ClientStream
}

func (x *messengerServiceStreamEnvelopesClient) Recv() (*Envelope, error) {
	m := new(Envelope)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *messengerServiceClient) SendEnvelope(ctx context.Context, opts ...grpc.CallOption) (MessengerService_SendEnvelopeClient, error) {
	stream, err := c.cc.NewStream(ctx, &MessengerService_ServiceDesc.Streams[1], "/messenger.MessengerService/SendEnvelope", opts...)
	if err != nil {
		return nil, err
	}
	x := &messengerServiceSendEnvelopeClient{stream}
	return x, nil
}

type MessengerService_SendEnvelopeClient interface {
	Send(*NewEnvelope) error
	CloseAndRecv() (*Envelope, error)
	grpc.ClientStream
}

type messengerServiceSendEnvelopeClient struct {
	grpc.ClientStream
}

func (x *messengerServiceSendEnvelopeClient) Send(m *NewEnvelope) error {
	return x.ClientStream.SendMsg(m)
}

func (x *messengerServiceSendEnvelopeClient) CloseAndRecv() (*Envelope, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(Envelope)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// MessengerServiceServer is the server API for MessengerService service.
// All implementations must embed UnimplementedMessengerServiceServer
// for forward compatibility
type MessengerServiceServer interface {
	StreamEnvelopes(*Conversation, MessengerService_StreamEnvelopesServer) error
	SendEnvelope(MessengerService_SendEnvelopeServer) error
	mustEmbedUnimplementedMessengerServiceServer()
}

// UnimplementedMessengerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedMessengerServiceServer struct {
}

func (UnimplementedMessengerServiceServer) StreamEnvelopes(*Conversation, MessengerService_StreamEnvelopesServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamEnvelopes not implemented")
}
func (UnimplementedMessengerServiceServer) SendEnvelope(MessengerService_SendEnvelopeServer) error {
	return status.Errorf(codes.Unimplemented, "method SendEnvelope not implemented")
}
func (UnimplementedMessengerServiceServer) mustEmbedUnimplementedMessengerServiceServer() {}

// UnsafeMessengerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MessengerServiceServer will
// result in compilation errors.
type UnsafeMessengerServiceServer interface {
	mustEmbedUnimplementedMessengerServiceServer()
}

func RegisterMessengerServiceServer(s grpc.ServiceRegistrar, srv MessengerServiceServer) {
	s.RegisterService(&MessengerService_ServiceDesc, srv)
}

func _MessengerService_StreamEnvelopes_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Conversation)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(MessengerServiceServer).StreamEnvelopes(m, &messengerServiceStreamEnvelopesServer{stream})
}

type MessengerService_StreamEnvelopesServer interface {
	Send(*Envelope) error
	grpc.ServerStream
}

type messengerServiceStreamEnvelopesServer struct {
	grpc.ServerStream
}

func (x *messengerServiceStreamEnvelopesServer) Send(m *Envelope) error {
	return x.ServerStream.SendMsg(m)
}

func _MessengerService_SendEnvelope_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MessengerServiceServer).SendEnvelope(&messengerServiceSendEnvelopeServer{stream})
}

type MessengerService_SendEnvelopeServer interface {
	SendAndClose(*Envelope) error
	Recv() (*NewEnvelope, error)
	grpc.ServerStream
}

type messengerServiceSendEnvelopeServer struct {
	grpc.ServerStream
}

func (x *messengerServiceSendEnvelopeServer) SendAndClose(m *Envelope) error {
	return x.ServerStream.SendMsg(m)
}

func (x *messengerServiceSendEnvelopeServer) Recv() (*NewEnvelope, error) {
	m := new(NewEnvelope)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// MessengerService_ServiceDesc is the grpc.ServiceDesc for MessengerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MessengerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "messenger.MessengerService",
	HandlerType: (*MessengerServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamEnvelopes",
			Handler:       _MessengerService_StreamEnvelopes_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "SendEnvelope",
			Handler:       _MessengerService_SendEnvelope_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "proto/messenger/v1/messenger_v1.proto",
}