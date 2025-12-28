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
	chat, err := s.chatRepository.GetChat(ctx, chatID)
	return chat, err
}

func (s *ChatService) GetChats(ctx context.Context, userID uuid.UUID) ([]*model.Chat, error) {
	chatIDs, err := s.userService.GetUserChatIds(ctx, userID)
	if err != nil {
		return nil, err
	}
	chats, err := s.chatRepository.GetChats(ctx, chatIDs)
	return chats, err
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

	err = s.chatRepository.CreateChat(ctx, chat)
	if err != nil {
		return nil, err
	}

	err = s.userService.InitialAddChatUser(ctx, chat.ID, requesterID)
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
	if !s.hasPermission(ctx, chatID, requesterID, model.Admin) {
		return fmt.Errorf("user has no permission")
	}

	err := s.chatRepository.UpdateChatName(ctx, chatID, name)
	return err
}

func (s *ChatService) DeleteChat(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) error {
	if !s.hasPermission(ctx, chatID, requesterID, model.Admin) {
		return fmt.Errorf("user has no permission")
	}

	err := s.chatRepository.DeleteChat(ctx, chatID)
	return err
}

func (s *ChatService) hasPermission(ctx context.Context, chatID, requesterID uuid.UUID, role model.Role) bool {
	requester, err := s.userService.GetChatUser(ctx, chatID, requesterID)
	if err != nil {
		return false
	}

	if requester.Role < role {
		return false
	}

	return true
}
