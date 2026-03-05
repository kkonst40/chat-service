package sso

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
	"github.com/kkonst40/ichat/internal/domain/model"
	pb "github.com/kkonst40/ichat/internal/gen/user"
)

type SSOService struct {
	client pb.UserServiceClient
}

func NewSSOClient(client pb.UserServiceClient) *SSOService {
	return &SSOService{
		client: client,
	}
}

func (c *SSOService) ExistMany(ctx context.Context, userIDs []uuid.UUID) ([]uuid.UUID, error) {
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
			"%w: sso service (ExistMany): %w",
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

func (c *SSOService) GetUsersLogins(ctx context.Context, userIDs []uuid.UUID) ([]model.UserInfo, error) {
	idsStrings := make([]string, len(userIDs))
	for i, id := range userIDs {
		idsStrings[i] = id.String()
	}

	ssoCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	resp, err := c.client.GetUsersLogins(ssoCtx, &pb.GetUsersLoginsRequest{
		Ids: idsStrings,
	})

	if err != nil {
		return nil, fmt.Errorf("%w: sso service: %w", errs.ErrExternalService, err)
	}

	result := make([]model.UserInfo, 0, len(resp.GetUsers()))
	for _, u := range resp.GetUsers() {
		parsedID, err := uuid.Parse(u.GetId())
		if err != nil {
			continue
		}

		result = append(result, model.UserInfo{
			ID:    parsedID,
			Login: u.GetLogin(),
		})
	}

	return result, nil
}
