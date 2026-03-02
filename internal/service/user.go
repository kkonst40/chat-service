package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/dispatcher"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
	"github.com/kkonst40/ichat/internal/domain/model"
	"github.com/kkonst40/ichat/internal/integration/sso"
	"github.com/kkonst40/ichat/internal/logger"
	"github.com/kkonst40/ichat/internal/repository"
)

type UserService struct {
	userRepository repository.UserRepository
	dispatcher     *dispatcher.Dispatcher
	ssoClient      *sso.SSOClient
}

func NewUserService(
	userRepository repository.UserRepository,
	dispatcher *dispatcher.Dispatcher,
	ssoClient *sso.SSOClient,
) *UserService {
	return &UserService{
		userRepository: userRepository,
		dispatcher:     dispatcher,
		ssoClient:      ssoClient,
	}
}

func (s *UserService) GetChatUser(ctx context.Context, chatID, userID uuid.UUID) (*model.User, error) {
	//??
	return s.userRepository.GetChatUser(ctx, chatID, userID)
}

func (s *UserService) GetChatUsers(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) ([]model.User, error) {
	log := logger.FromContext(ctx)
	log.Debug("userService.GetChatUsers", "chatID", chatID)

	if !s.userInChat(ctx, chatID, requesterID) {
		return nil, fmt.Errorf(
			"%w: user %v is not in chat %v",
			errs.ErrForbidden,
			requesterID,
			chatID,
		)
	}

	user, err := s.userRepository.GetChatUsers(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get chat %v users: %w", chatID, err)
	}
	log.Debug("chat users retrieved")

	return user, nil
}

func (s *UserService) AddChatUsers(ctx context.Context, chatID uuid.UUID, userIDs []uuid.UUID, requesterID uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("userService.AddChatUsers", "chatID", chatID)

	if !s.userInChat(ctx, chatID, requesterID) {
		return fmt.Errorf(
			"%w: user %v is not in chat %v",
			errs.ErrForbidden,
			requesterID,
			chatID,
		)
	}

	existingUserIDs, err := s.existMany(ctx, userIDs)
	if err != nil {
		return fmt.Errorf("check users existence before add to chat %v: %w", chatID, err)
	}

	err = s.userRepository.AddChatUsers(ctx, chatID, existingUserIDs)
	if err != nil {
		if errors.Is(err, errs.ErrChatNotFound) {
			return fmt.Errorf("%w: ID %v", err, chatID)
		}
		return fmt.Errorf("add chat %v users: %w", chatID, err)
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
			return fmt.Errorf("delete user %v from chat %v: %w", userID, chatID, err)
		}
		log.Debug("user deleted")

		return nil
	}

	requester, err := s.userRepository.GetChatUser(ctx, chatID, requesterID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return fmt.Errorf(
				"%w: user %v is not in chat %v",
				errs.ErrForbidden,
				requesterID,
				chatID,
			)
		}
		return fmt.Errorf("get requester %v to perform delete: %w", requesterID, err)
	}

	user, err := s.userRepository.GetChatUser(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return fmt.Errorf("%w: user ID %v, chat ID %v", err, userID, chatID)
		}
		return fmt.Errorf("get user %v to delete: %w", userID, err)
	}

	if !(model.UserPriority[requester.Role] > model.UserPriority[user.Role]) {
		return fmt.Errorf(
			"%w: user %v has no permission to delete user %v from chat %v",
			errs.ErrForbidden,
			requesterID,
			userID,
			chatID,
		)
	}

	err = s.userRepository.DeleteChatUser(ctx, chatID, userID)
	if err != nil {
		return fmt.Errorf("delete user %v from chat %v: %w", userID, chatID, err)
	}
	log.Debug("user deleted")

	return nil
}

func (s *UserService) UpdateUserRole(ctx context.Context, chatID, userID uuid.UUID, newRole model.Role, requesterID uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("userService.UpdateUserRole", "chatID", chatID, "chatUserID", userID, "role", newRole)

	user, err := s.userRepository.GetChatUser(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return fmt.Errorf("%w: user ID %v, chat ID %v", err, userID, chatID)
		}
		return fmt.Errorf("get requester %v to perform role update: %w", requesterID, err)
	}

	requester, err := s.userRepository.GetChatUser(ctx, chatID, requesterID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return fmt.Errorf(
				"%w: user %v is not in chat %v",
				errs.ErrForbidden,
				requesterID,
				chatID,
			)
		}
		return fmt.Errorf("get user %v to update:%w", userID, err)
	}

	if !(model.UserPriority[requester.Role] > model.UserPriority[user.Role]) {
		return fmt.Errorf(
			"%w: user %v has no permission to update user %v role",
			errs.ErrForbidden,
			requesterID,
			userID,
		)
	}

	err = s.userRepository.UpdateUserRole(ctx, chatID, userID, newRole)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return fmt.Errorf("%w: user ID %v, chat ID %v", err, userID, chatID)
		}
		return fmt.Errorf("update user %v role (to %v) in chat %v: %w", userID, newRole, chatID, err)
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

func (s *UserService) userInChat(ctx context.Context, chatID, userID uuid.UUID) bool {
	result, err := s.userRepository.UserInChat(ctx, chatID, userID)
	if err != nil {
		return false
	}

	return result
}

func (s *UserService) existMany(ctx context.Context, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	return s.ssoClient.ExistMany(ctx, userIDs)
}
