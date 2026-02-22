package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/apperror"
	"github.com/kkonst40/ichat/internal/integration/sso"
	"github.com/kkonst40/ichat/internal/logger"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type UserService struct {
	userRepository repository.UserRepository
	ssoClient      *sso.SSOClient
}

func NewUserService(
	newUserRepository repository.UserRepository,
	ssoClient *sso.SSOClient,
) *UserService {
	return &UserService{
		userRepository: newUserRepository,
		ssoClient:      ssoClient,
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
	return s.ssoClient.ExistMany(ctx, userIDs)
}
