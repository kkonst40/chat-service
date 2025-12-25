package service

import (
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

func (s *UserService) GetChatUser(chatID uuid.UUID, userID uuid.UUID) (*model.User, error) {
	return s.userRepository.GetChatUser(chatID, userID)
}

func (s *UserService) GetChatUsers(chatID uuid.UUID, requesterID uuid.UUID) ([]*model.User, error) {
	if _, err := s.GetChatUser(chatID, requesterID); err != nil {
		return nil, fmt.Errorf("user is not in the chat")
	}

	return s.userRepository.GetChatUsers(chatID)
}

func (s *UserService) GetUserChatIds(userID uuid.UUID) ([]uuid.UUID, error) {
	return s.userRepository.GetUserChatIds(userID)
}

func (s *UserService) InitialAddChatUser(chatID, userID uuid.UUID) error {
	err := s.userRepository.AddChatUsers(chatID, []uuid.UUID{userID})
	if err != nil {
		return err
	}

	return s.userRepository.SetUserRole(chatID, userID, model.Admin)
}

func (s *UserService) AddChatUsers(chatID uuid.UUID, userIds []uuid.UUID, requesterID uuid.UUID) error {
	if _, err := s.GetChatUser(chatID, requesterID); err != nil {
		return fmt.Errorf("user is not in the chat")
	}

	return s.userRepository.AddChatUsers(chatID, userIds)
}

func (s *UserService) DeleteChatUser(chatID uuid.UUID, userID uuid.UUID, requesterID uuid.UUID) error {
	requester, err := s.userRepository.GetChatUser(chatID, requesterID)
	if err != nil {
		return err
	}

	user, err := s.userRepository.GetChatUser(chatID, userID)
	if err != nil {
		return err
	}

	if requester.Role > user.Role || requesterID == userID {
		return s.userRepository.DeleteChatUser(chatID, userID)
	} else {
		return fmt.Errorf("user has no permission")
	}
}

func (s *UserService) SetUserRole(chatID, userID uuid.UUID, newRole model.Role, requesterID uuid.UUID) error {
	user, err := s.userRepository.GetChatUser(chatID, userID)
	if err != nil {
		return err
	}
	requester, err := s.userRepository.GetChatUser(chatID, requesterID)
	if err != nil {
		return err
	}

	if requester.Role <= user.Role {
		return fmt.Errorf("user has no permission")
	}
	return s.userRepository.SetUserRole(chatID, userID, newRole)
}
