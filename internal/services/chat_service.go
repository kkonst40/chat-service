package services

import (
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/domain/models"
	"github.com/kkonst40/ichat/internal/repositories"
)

type ChatService struct {
	chatRepository repositories.ChatRepository
}

func NewChatService() *ChatService {
	repo := repositories.NewInMemoryChatRepository()
	service := ChatService{
		chatRepository: repo,
	}

	return &service
}

func (s *ChatService) GetChat(id uuid.UUID) (*models.Chat, error) {
	chat, err := s.chatRepository.GetChat(id)
	return chat, err
}

func (s *ChatService) GetChats() ([]models.Chat, error) {
	chats, err := s.chatRepository.GetChats()
	return chats, err
}

func (s *ChatService) CreateChat(name string, userIds []uuid.UUID) (*models.Chat, error) {
	newId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	chat := &models.Chat{
		ID:      newId,
		Name:    name,
		UserIDs: userIds,
	}

	err = s.chatRepository.CreateChat(chat)
	return chat, err
}

func (s *ChatService) UpdateChatName(id uuid.UUID, name string) error {
	err := s.chatRepository.UpdateChatName(id, name)
	return err
}

func (s *ChatService) AddChatUser(id, userId uuid.UUID) error {
	err := s.chatRepository.AddChatUser(id, userId)
	return err
}

func (s *ChatService) DeleteChat(id uuid.UUID) error {
	err := s.chatRepository.DeleteChat(id)
	return err
}
