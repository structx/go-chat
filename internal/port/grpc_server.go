package port

import (
	"fmt"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/trevatk/go-chat/internal/domain"
	pb "github.com/trevatk/go-chat/proto/messenger/v1"
	"github.com/trevatk/go-pkg/logging"
)

// GrpcServer protobuf server implementation
type GrpcServer struct {
	bundle         *domain.Bundle
	conversationCh map[string][]chan *pb.Envelope
	pb.UnimplementedMessengerServiceServer
}

// interface verification
var _ pb.MessengerServiceServer = (*GrpcServer)(nil)

// NewGrpcServer create new grpc server implementation
func NewGrpcServer(bundle *domain.Bundle) *GrpcServer {
	return &GrpcServer{
		bundle:         bundle,
		conversationCh: make(map[string][]chan *pb.Envelope)}
}

// SendEnvelope persist new envelope and pass to conversation uuid channel
func (g *GrpcServer) SendEnvelope(stream pb.MessengerService_SendEnvelopeServer) error {

	ctx := stream.Context()

	gne, e := stream.Recv()
	if e != nil {

		if e == io.EOF {
			return nil
		}

		logging.FromContext(ctx).Errorf("unable to receive new envelope message %v", e)
		return status.Errorf(codes.DataLoss, "unable to receive envelope")
	}

	ne, e := transformNewEnvelope(gne)
	if e != nil {
		return status.Errorf(codes.InvalidArgument, e.Error())
	}

	ev, e := g.bundle.MessengerService.CreateMessage(ctx, ne)
	if e != nil {
		logging.FromContext(ctx).Errorf("unable to send message %v", e)
		return status.Errorf(codes.Internal, "failed to persist new envelope")
	}

	pbEv := transformEnvelope(ev)

	s := g.conversationCh[ev.ConversationUUID.String()]
	for _, ch := range s {
		ch <- pbEv
	}

	e = stream.SendAndClose(pbEv)
	if e != nil {
		logging.FromContext(ctx).Errorf("unable to send envelope %v", e)
		return status.Errorf(codes.Internal, "failed to send envelope")
	}

	return nil
}

// StreamEnvelopes stream new envelopes to client
func (g *GrpcServer) StreamEnvelopes(in *pb.Conversation, stream pb.MessengerService_StreamEnvelopesServer) error {

	ctx := stream.Context()

	ch := make(chan *pb.Envelope)
	g.conversationCh[in.Conversation] = append(g.conversationCh[in.Conversation], ch)

	for {

		select {
		case <-ctx.Done():
			return nil
		case ev := <-ch:

			e := stream.Send(ev)
			if e != nil {
				logging.FromContext(ctx).Errorf("failed to stream envelope %v", e)
				return status.Errorf(codes.Internal, "failed to stream envelopes")
			}

		}
	}
}

func transformNewEnvelope(newEnvelope *pb.NewEnvelope) (*domain.NewEnvelope, error) {

	sID, e := uuid.Parse(newEnvelope.Sender)
	if e != nil {
		return nil, fmt.Errorf("unable to parse sender uuid %v", e)
	}

	cID, e := uuid.Parse(newEnvelope.Conversation)
	if e != nil {
		return nil, fmt.Errorf("unable to parse conversation uuid %v", e)
	}

	return &domain.NewEnvelope{
		Sender:           sID,
		ConversationUUID: cID,
		Message:          newEnvelope.Message,
	}, nil
}

func transformEnvelope(envelope *domain.Envelope) *pb.Envelope {
	return &pb.Envelope{
		Uid:     envelope.UID.String(),
		Sender:  envelope.Sender.String(),
		Message: envelope.Message,
		Status:  pb.SEND_ENVELOPE_STATUS_DELIVERED,
	}
}
