package service

import (
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type ChatService struct {
	chatRepository repository.ChatRepository
}

func NewChatService(newChatRepository repository.ChatRepository) *ChatService {
	service := ChatService{
		chatRepository: newChatRepository,
	}

	return &service
}

func (s *ChatService) GetChat(id uuid.UUID) (*model.Chat, error) {
	chat, err := s.chatRepository.GetChat(id)
	return chat, err
}

func (s *ChatService) GetChats(userId uuid.UUID) ([]*model.Chat, error) {
	chats, err := s.chatRepository.GetChats(userId)
	return chats, err
}

func (s *ChatService) CreateChat(name string, userIds []uuid.UUID) (*model.Chat, error) {
	newId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	chat := &model.Chat{
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
