package structs

import (
	"strconv"

	"github.com/sonm-io/core/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ImagePush struct {
	sonm.Worker_PushTaskServer

	dealId    string
	imageSize int64
}

func requireHeader(md metadata.MD, name string) (string, error) {
	value, ok := md[name]
	if !ok {
		return "", status.Errorf(codes.InvalidArgument, "`%s` required", name)
	}

	return value[len(value)-1], nil
}

func RequireHeaderInt64(md metadata.MD, name string) (int64, error) {
	value, err := requireHeader(md, name)
	if err != nil {
		return 0, err
	}

	valueInt64, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return valueInt64, nil
}

func NewImagePush(stream sonm.Worker_PushTaskServer) (*ImagePush, error) {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "metadata required")
	}

	dealId, err := requireHeader(md, "deal")
	if err != nil {
		return nil, err
	}

	imageSize, err := RequireHeaderInt64(md, "size")
	if err != nil {
		return nil, err
	}

	return &ImagePush{stream, dealId, imageSize}, nil
}

func (p *ImagePush) DealId() string {
	return p.dealId
}

func (p *ImagePush) ImageSize() int64 {
	return p.imageSize
}
