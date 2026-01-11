package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/apperror"
	"github.com/kkonst40/ichat/internal/logger"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type UserService struct {
	userRepository repository.UserRepository
	client         *http.Client
	ssoURL         string
}

func NewUserService(
	newUserRepository repository.UserRepository,
	ssoURL string,
) *UserService {
	return &UserService{
		userRepository: newUserRepository,
		ssoURL:         ssoURL,
		client:         &http.Client{},
	}
}

func (s *UserService) GetChatUser(ctx context.Context, chatID, userID uuid.UUID) (*model.User, error) {
	// ?
	return s.userRepository.GetChatUser(ctx, chatID, userID)
}

func (s *UserService) GetChatUsers(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) ([]model.User, error) {
	log := logger.FromContext(ctx)
	log.Debug("userService.GetChatUsers", "chatID", chatID)

	if !s.isUserInChat(ctx, chatID, requesterID) {
		return nil, &apperror.ForbiddenError{Msg: fmt.Sprintf("user (%v) is not in the chat (%v)", requesterID, chatID)}
	}

	user, err := s.userRepository.GetChatUsers(ctx, chatID)
	if err != nil {
		return nil, err
	}
	log.Debug("chat users retrieved")

	return user, nil
}

func (s *UserService) AddChatUsers(ctx context.Context, chatID uuid.UUID, userIDs []uuid.UUID, requesterID uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("userService.AddChatUsers", "chatID", chatID)

	if !s.isUserInChat(ctx, chatID, requesterID) {
		return &apperror.ForbiddenError{Msg: fmt.Sprintf("user (%v) is not in the chat (%v)", requesterID, chatID)}
	}

	existingUserIDs, err := s.existMany(ctx, userIDs)
	if err != nil {
		return err
	}

	err = s.userRepository.AddChatUsers(ctx, chatID, existingUserIDs)
	if err != nil {
		return err
	}
	log.Debug("chat users added")

	return nil
}

func (s *UserService) DeleteChatUser(ctx context.Context, chatID uuid.UUID, userID uuid.UUID, requesterID uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("userService.DeleteChatUser", "chatID", chatID)

	if userID == requesterID {
		err := s.userRepository.DeleteChatUser(ctx, chatID, userID)
		if err != nil {
			return err
		}
		log.Debug("user deleted")

		return nil
	}

	requester, err := s.userRepository.GetChatUser(ctx, chatID, requesterID)
	if err != nil {
		return err
	}

	user, err := s.userRepository.GetChatUser(ctx, chatID, userID)
	if err != nil {
		return err
	}

	if !(model.UserPriority[requester.Role] > model.UserPriority[user.Role]) {
		return &apperror.ForbiddenError{Msg: fmt.Sprintf("user (%v) has no permission", requesterID)}
	}

	err = s.userRepository.DeleteChatUser(ctx, chatID, userID)
	if err != nil {
		return err
	}
	log.Debug("user deleted")

	return nil
}

func (s *UserService) UpdateUserRole(ctx context.Context, chatID, userID uuid.UUID, newRole model.Role, requesterID uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("userService.UpdateUserRole", "chatID", chatID, "chatUserID", userID, "role", newRole)

	user, err := s.userRepository.GetChatUser(ctx, chatID, userID)
	if err != nil {
		return err
	}
	requester, err := s.userRepository.GetChatUser(ctx, chatID, requesterID)
	if err != nil {
		return err
	}

	if !(model.UserPriority[requester.Role] > model.UserPriority[user.Role]) {
		return &apperror.ForbiddenError{Msg: fmt.Sprintf("user (%v) has no permission", requesterID)}
	}

	err = s.userRepository.UpdateUserRole(ctx, chatID, userID, newRole)
	if err != nil {
		return err
	}
	log.Debug("user role updated")

	return nil
}

func (s *UserService) hasPermission(ctx context.Context, chatID, requesterID uuid.UUID, role model.Role) bool {
	requester, err := s.GetChatUser(ctx, chatID, requesterID)
	if err != nil {
		return false
	}

	if model.UserPriority[requester.Role] < model.UserPriority[role] {
		return false
	}

	return true
}

func (s *UserService) isUserInChat(ctx context.Context, chatID, userID uuid.UUID) bool {
	result, err := s.userRepository.IsUserInChat(ctx, chatID, userID)
	if err != nil {
		return false
	}

	return result
}

func (s *UserService) existMany(ctx context.Context, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	jsonData, err := json.Marshal(userIDs)
	if err != nil {
		return nil, &apperror.InternalError{
			Msg: fmt.Sprintf("failed to marshal userIDs: %v", err.Error()),
		}
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%v/exist", s.ssoURL),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, &apperror.InternalError{
			Msg: fmt.Sprintf("failed to create request: %v", err.Error()),
		}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, &apperror.ExternalServiceError{
			Msg: fmt.Sprintf("user service unreachable: %v", err.Error()),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &apperror.ExternalServiceError{
			Msg: fmt.Sprintf("user service returned error status: %d", resp.StatusCode),
		}
	}

	var existingIDs []uuid.UUID
	if err := json.NewDecoder(resp.Body).Decode(&existingIDs); err != nil {
		return nil, &apperror.ExternalServiceError{
			Msg: fmt.Sprintf("failed to decode response: %v", err.Error()),
		}
	}

	return existingIDs, nil
}
