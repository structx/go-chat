package port

import (
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/trevatk/go-chat/internal/domain"
	pb "github.com/trevatk/go-chat/proto/messenger/v1"
)

// GrpcServer protobuf server implementation
type GrpcServer struct {
	log            *zap.SugaredLogger
	bundle         *domain.Bundle
	conversationCh map[string][]chan *pb.Envelope
	pb.UnimplementedMessengerServiceServer
}

// interface verification
var _ pb.MessengerServiceServer = (*GrpcServer)(nil)

// NewGrpcServer create new grpc server implementation
func NewGrpcServer(log *zap.Logger, bundle *domain.Bundle) *GrpcServer {
	return &GrpcServer{
		log:            log.Named("grpc server").Sugar(),
		bundle:         bundle,
		conversationCh: make(map[string][]chan *pb.Envelope)}
}

// SendEnvelope persist new envelope and pass to conversation uuid channel
func (g *GrpcServer) SendEnvelope(stream pb.MessengerService_SendEnvelopeServer) error {

	ctx := stream.Context()

	ne, e := stream.Recv()
	if e != nil {

		if e == io.EOF {
			return nil
		}

		g.log.Errorf("unable to receive new envelope message %v", e)
		return status.Errorf(codes.DataLoss, "unable to receive envelope")
	}

	ev, e := g.bundle.MessengerService.CreateMessage(ctx, transformNewEnvelope(ne))
	if e != nil {
		g.log.Errorf("unable to send message %v", e)
		return status.Errorf(codes.Internal, "failed to persist new envelope")
	}

	pbEv := transformEnvelope(ev)

	s := g.conversationCh[ev.ConversationUUID.String()]
	for _, ch := range s {
		ch <- pbEv
	}

	e = stream.SendAndClose(pbEv)
	if e != nil {
		g.log.Errorf("unable to send envelope %v", e)
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
				g.log.Errorf("failed to stream envelope %v", e)
				return status.Errorf(codes.Internal, "failed to stream envelopes")
			}

		}
	}
}

func transformNewEnvelope(newEnvelope *pb.NewEnvelope) *domain.NewEnvelope {
	return &domain.NewEnvelope{
		Sender:           uuid.MustParse(newEnvelope.Sender),
		ConversationUUID: uuid.MustParse(newEnvelope.Conversation),
		Message:          newEnvelope.Message,
	}
}

func transformEnvelope(envelope *domain.Envelope) *pb.Envelope {
	return &pb.Envelope{
		Uid:     envelope.UID.String(),
		Sender:  envelope.Sender.String(),
		Message: envelope.Message,
		Status:  pb.SEND_ENVELOPE_STATUS_DELIVERED,
	}
}
