package auth

import (
	"context"

	"github.com/warehouse/ai-service/internal/config"
	"github.com/warehouse/ai-service/internal/domain"
	"github.com/warehouse/ai-service/internal/pkg/errors"
	"github.com/warehouse/ai-service/internal/warehousepb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type (
	Adapter interface {
		Authenticate(ctx context.Context, request *warehousepb.AuthRequest) (domain.Account, int64, *errors.Error)
	}

	adapter struct {
		client warehousepb.AuthClient
		config config.Grpc
	}
)

func NewAdapter(
	config config.Grpc,
) (Adapter, error) {
	conn, err := grpc.Dial(config.AuthAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := warehousepb.NewAuthClient(conn)

	return &adapter{
		client: client,
		config: config,
	}, nil
}

func (a *adapter) Authenticate(ctx context.Context, request *warehousepb.AuthRequest) (domain.Account, int64, *errors.Error) {
	resp, err := a.client.Authenticate(ctx, request)
	if err != nil {
		return domain.Account{}, 0, errors.GrpcError(err)
	}

	return domain.Account{
		Role:      domain.Role(resp.User.Role),
		Username:  resp.User.Username,
		Firstname: resp.User.Firstname,
		Email:     resp.User.Email,
		Verified:  resp.User.Verified,
		Id:        resp.User.UserId,
	}, resp.Number, nil
}
