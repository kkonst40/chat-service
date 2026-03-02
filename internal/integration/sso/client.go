package sso

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
	pb "github.com/kkonst40/ichat/internal/gen/user"
)

type SSOClient struct {
	client pb.UserServiceClient
}

func NewSSOClient(client pb.UserServiceClient) *SSOClient {
	return &SSOClient{
		client: client,
	}
}

func (c *SSOClient) ExistMany(ctx context.Context, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	idsStrings := make([]string, len(userIDs))
	for i, id := range userIDs {
		idsStrings[i] = id.String()
	}

	ssoCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	resp, err := c.client.Exist(ssoCtx, &pb.ExistRequest{
		Ids: idsStrings,
	})

	if err != nil {
		return nil, fmt.Errorf(
			"%w: sso service: %w",
			errs.ErrExternalService,
			err,
		)
	}

	existingIDs := make([]uuid.UUID, 0, len(resp.GetExistingIds()))
	for _, idStr := range resp.GetExistingIds() {
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		existingIDs = append(existingIDs, parsedID)
	}

	return existingIDs, nil
}
