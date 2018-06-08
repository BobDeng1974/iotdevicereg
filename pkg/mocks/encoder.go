package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	encoder "github.com/thingful/twirp-encoder-go"
)

type Encoder struct {
	mock.Mock
}

func (e *Encoder) CreateStream(ctx context.Context, req *encoder.CreateStreamRequest) (*encoder.CreateStreamResponse, error) {
	args := e.Called(ctx, req)
	return args.Get(0).(*encoder.CreateStreamResponse), args.Error(1)
}

func (e *Encoder) DeleteStream(ctx context.Context, req *encoder.DeleteStreamRequest) (*encoder.DeleteStreamResponse, error) {
	args := e.Called(ctx, req)
	return args.Get(0).(*encoder.DeleteStreamResponse), args.Error(1)
}
