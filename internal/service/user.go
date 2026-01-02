package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

func (s *UserService) GetChatUser(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) (*model.User, error) {
	return s.userRepository.GetChatUser(ctx, chatID, userID)
}

func (s *UserService) GetChatUsers(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) ([]*model.User, error) {
	if user, err := s.GetChatUser(ctx, chatID, requesterID); user == nil && err == nil {
		return nil, fmt.Errorf("user is not in the chat")
	}

	return s.userRepository.GetChatUsers(ctx, chatID)
}

func (s *UserService) InitialAddChatUser(ctx context.Context, chatID, userID uuid.UUID) error {
	err := s.userRepository.AddChatUsers(ctx, chatID, []uuid.UUID{userID})
	if err != nil {
		return err
	}

	return s.userRepository.UpdateUserRole(ctx, chatID, userID, model.Admin)
}

func (s *UserService) AddChatUsers(ctx context.Context, chatID uuid.UUID, userIds []uuid.UUID, requesterID uuid.UUID) error {
	if user, err := s.GetChatUser(ctx, chatID, requesterID); user == nil && err == nil {
		return fmt.Errorf("user is not in the chat")
	}

	return s.userRepository.AddChatUsers(ctx, chatID, userIds)
}

func (s *UserService) DeleteChatUser(ctx context.Context, chatID uuid.UUID, userID uuid.UUID, requesterID uuid.UUID) error {
	requester, err := s.userRepository.GetChatUser(ctx, chatID, requesterID)
	if err != nil {
		return err
	}

	user, err := s.userRepository.GetChatUser(ctx, chatID, userID)
	if err != nil {
		return err
	}

	if requester.Role == model.Admin && user.Role == model.Common || requesterID == userID {
		return s.userRepository.DeleteChatUser(ctx, chatID, userID)
	} else {
		return fmt.Errorf("user has no permission")
	}
}

func (s *UserService) SetUserRole(ctx context.Context, chatID, userID uuid.UUID, newRole model.Role, requesterID uuid.UUID) error {
	user, err := s.userRepository.GetChatUser(ctx, chatID, userID)
	if err != nil {
		return err
	}
	requester, err := s.userRepository.GetChatUser(ctx, chatID, requesterID)
	if err != nil {
		return err
	}

	if !(requester.Role == model.Admin && user.Role == model.Common) {
		return fmt.Errorf("user has no permission")
	}
	return s.userRepository.UpdateUserRole(ctx, chatID, userID, newRole)
}
