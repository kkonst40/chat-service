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
	if _, err := s.GetChatUser(ctx, chatID, requesterID); err != nil {
		return nil, fmt.Errorf("user is not in the chat")
	}

	return s.userRepository.GetChatUsers(ctx, chatID)
}

func (s *UserService) GetUserChatIds(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return s.userRepository.GetUserChatIds(ctx, userID)
}

func (s *UserService) InitialAddChatUser(ctx context.Context, chatID, userID uuid.UUID) error {
	err := s.userRepository.AddChatUsers(ctx, chatID, []uuid.UUID{userID})
	if err != nil {
		return err
	}

	return s.userRepository.SetUserRole(ctx, chatID, userID, model.Admin)
}

func (s *UserService) AddChatUsers(ctx context.Context, chatID uuid.UUID, userIds []uuid.UUID, requesterID uuid.UUID) error {
	if _, err := s.GetChatUser(ctx, chatID, requesterID); err != nil {
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

	if requester.Role > user.Role || requesterID == userID {
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

	if requester.Role <= user.Role {
		return fmt.Errorf("user has no permission")
	}
	return s.userRepository.SetUserRole(ctx, chatID, userID, newRole)
}
