package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/logger"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type UserService struct {
	userRepository repository.UserRepository
}

func NewUserService(newUserRepository repository.UserRepository) *UserService {
	return &UserService{
		userRepository: newUserRepository,
	}
}

func (s *UserService) GetChatUser(ctx context.Context, chatID, userID uuid.UUID) (*model.User, error) {
	// ?
	return s.userRepository.GetChatUser(ctx, chatID, userID)
}

func (s *UserService) GetChatUsers(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) ([]*model.User, error) {
	log := logger.FromContext(ctx)
	log.Debug("userService.GetChatUsers", "chatID", chatID)

	if !s.isUserInChat(ctx, chatID, requesterID) {
		return nil, fmt.Errorf("user is not in the chat")
	}

	user, err := s.userRepository.GetChatUsers(ctx, chatID)
	if err != nil {
		return nil, err
	}
	log.Debug("chat users retrieved")

	return user, nil
}

func (s *UserService) AddChatUsers(ctx context.Context, chatID uuid.UUID, userIds []uuid.UUID, requesterID uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("userService.AddChatUsers", "chatID", chatID)

	if !s.isUserInChat(ctx, chatID, requesterID) {
		return fmt.Errorf("user is not in the chat")
	}

	err := s.userRepository.AddChatUsers(ctx, chatID, userIds)
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
		return fmt.Errorf("user has no permission")
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
		return fmt.Errorf("user has no permission")
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
