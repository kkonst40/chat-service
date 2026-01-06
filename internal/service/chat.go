package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/apperror"
	"github.com/kkonst40/ichat/internal/logger"
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
	log := logger.FromContext(ctx)
	log.Debug("chatService.GetChat", "chatID", chatID)

	chat, err := s.chatRepository.GetChat(ctx, chatID)
	if err != nil {
		return nil, err
	}
	log.Debug("chat retrieved")

	return chat, nil
}

func (s *ChatService) GetUserChats(ctx context.Context, userID uuid.UUID) ([]model.Chat, error) {
	log := logger.FromContext(ctx)
	log.Debug("chatService.GetUserChats")

	chats, err := s.chatRepository.GetUserChats(ctx, userID)
	if err != nil {
		return nil, err
	}
	log.Debug("chats retrieved")

	return chats, nil
}

func (s *ChatService) CreateChat(ctx context.Context, name string, userIDs []uuid.UUID, requesterID uuid.UUID) (*model.Chat, error) {
	log := logger.FromContext(ctx)
	log.Debug("chatService.CreateChat")

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
	log.Debug("chat created")

	err = s.userService.AddChatUsers(ctx, chat.ID, userIDs, requesterID)
	if err != nil {
		return nil, err
	}
	log.Debug("chat users added")

	return chat, nil
}

func (s *ChatService) UpdateChatName(ctx context.Context, chatID uuid.UUID, name string, requesterID uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("chatService.UpdateChatName", "chatID", chatID)

	if !s.userService.hasPermission(ctx, chatID, requesterID, model.Admin) {
		return &apperror.ForbiddenError{Msg: fmt.Sprintf("user (%v) has no permission", requesterID)}
	}

	err := s.chatRepository.UpdateChatName(ctx, chatID, name)
	if err != nil {
		return err
	}
	log.Debug("chat name updated")

	return nil
}

func (s *ChatService) DeleteChat(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("chatService.DeleteChat", "chatID", chatID)

	if !s.userService.hasPermission(ctx, chatID, requesterID, model.Owner) {
		return &apperror.ForbiddenError{Msg: fmt.Sprintf("user (%v) has no permission", requesterID)}
	}

	err := s.chatRepository.DeleteChat(ctx, chatID)
	if err != nil {
		return err
	}
	log.Debug("chat deleted")

	return nil
}

func (s *ChatService) doesChatExist(ctx context.Context, chatID uuid.UUID) bool {
	exists, err := s.chatRepository.DoesChatExist(ctx, chatID)
	if err != nil {
		return false
	}

	return exists
}
