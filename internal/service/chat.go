package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type ChatService struct {
	chatRepository repository.ChatRepository
	userService    *UserService
}

func NewChatService(newChatRepository repository.ChatRepository, newUserService *UserService) *ChatService {
	service := ChatService{
		chatRepository: newChatRepository,
		userService:    newUserService,
	}

	return &service
}

func (s *ChatService) GetChat(chatID uuid.UUID) (*model.Chat, error) {
	chat, err := s.chatRepository.GetChat(chatID)
	return chat, err
}

func (s *ChatService) GetChats(userID uuid.UUID) ([]*model.Chat, error) {
	chatIDs, err := s.userService.GetUserChatIds(userID)
	if err != nil {
		return nil, err
	}
	chats, err := s.chatRepository.GetChats(chatIDs)
	return chats, err
}

func (s *ChatService) CreateChat(name string, userIDs []uuid.UUID, requesterID uuid.UUID) (*model.Chat, error) {
	newID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	chat := &model.Chat{
		ID:   newID,
		Name: name,
	}

	err = s.chatRepository.CreateChat(chat)
	if err != nil {
		return nil, err
	}

	err = s.userService.InitialAddChatUser(chat.ID, requesterID)
	if err != nil {
		return nil, err
	}

	err = s.userService.AddChatUsers(chat.ID, userIDs, requesterID)
	if err != nil {
		return nil, err
	}

	return chat, nil
}

func (s *ChatService) UpdateChatName(chatID uuid.UUID, name string, requesterID uuid.UUID) error {
	if !s.hasPermission(chatID, requesterID, model.Admin) {
		return fmt.Errorf("user has no permission")
	}

	err := s.chatRepository.UpdateChatName(chatID, name)
	return err
}

func (s *ChatService) DeleteChat(chatID uuid.UUID, requesterID uuid.UUID) error {
	if !s.hasPermission(chatID, requesterID, model.Admin) {
		return fmt.Errorf("user has no permission")
	}

	err := s.chatRepository.DeleteChat(chatID)
	return err
}

func (s *ChatService) hasPermission(chatID, requesterID uuid.UUID, role model.Role) bool {
	requester, err := s.userService.GetChatUser(chatID, requesterID)
	if err != nil {
		return false
	}

	if requester.Role < role {
		return false
	}

	return true
}
