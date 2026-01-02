package service

import (
	"context"
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

func (s *ChatService) GetChat(ctx context.Context, chatID uuid.UUID) (*model.Chat, error) {
	return s.chatRepository.GetChat(ctx, chatID)
}

func (s *ChatService) GetUserChats(ctx context.Context, userID uuid.UUID) ([]*model.Chat, error) {
	return s.chatRepository.GetUserChats(ctx, userID)
}

func (s *ChatService) CreateChat(ctx context.Context, name string, userIDs []uuid.UUID, requesterID uuid.UUID) (*model.Chat, error) {
	newID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	chat := &model.Chat{
		ID:   newID,
		Name: name,
	}

	err = s.chatRepository.CreateChat(ctx, chat, requesterID)
	if err != nil {
		return nil, err
	}

	err = s.userService.AddChatUsers(ctx, chat.ID, userIDs, requesterID)
	if err != nil {
		return nil, err
	}

	return chat, nil
}

func (s *ChatService) UpdateChatName(ctx context.Context, chatID uuid.UUID, name string, requesterID uuid.UUID) error {
	if !s.userService.hasPermission(ctx, chatID, requesterID, model.Admin) {
		return fmt.Errorf("user has no permission")
	}

	err := s.chatRepository.UpdateChatName(ctx, chatID, name)
	return err
}

func (s *ChatService) DeleteChat(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) error {
	if !s.userService.hasPermission(ctx, chatID, requesterID, model.Owner) {
		return fmt.Errorf("user has no permission")
	}

	err := s.chatRepository.DeleteChat(ctx, chatID)
	return err
}

func (s *ChatService) doesChatExist(ctx context.Context, chatID uuid.UUID) bool {
	exists, err := s.chatRepository.DoesChatExist(ctx, chatID)
	if err != nil {
		return false
	}

	return exists
}
