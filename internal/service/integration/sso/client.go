package sso

import (
	"context"

	"github.com/google/uuid"
	"github.com/kkonst40/chat-service/internal/domain/model"
	pb "github.com/kkonst40/chat-service/internal/gen/user"
)

type Service struct {
	client pb.UserServiceClient
}

func NewClient(client pb.UserServiceClient) *Service {
	return &Service{
		client: client,
	}
}

func (c *Service) ExistMany(ctx context.Context, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	idsStrings := make([]string, len(userIDs))
	for i, id := range userIDs {
		idsStrings[i] = id.String()
	}

	resp, err := c.client.Exist(ctx, &pb.ExistRequest{
		Ids: idsStrings,
	})

	if err != nil {
		return nil, err
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

func (c *Service) GetUsersLogins(ctx context.Context, userIDs []uuid.UUID) ([]model.UserInfo, error) {
	idsStrings := make([]string, len(userIDs))
	for i, id := range userIDs {
		idsStrings[i] = id.String()
	}

	resp, err := c.client.GetUsersLogins(ctx, &pb.GetUsersLoginsRequest{
		Ids: idsStrings,
	})

	if err != nil {
		return nil, err
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

func (c *Service) GetUsersIDs(ctx context.Context, userLogins []string) ([]model.UserInfo, error) {
	resp, err := c.client.GetUsersIDs(ctx, &pb.GetUsersIDsRequest{
		Logins: userLogins,
	})

	if err != nil {
		return nil, err
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
