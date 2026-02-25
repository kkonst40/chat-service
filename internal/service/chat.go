package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	errs "github.com/kkonst40/ichat/internal/errors"
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
		if errors.Is(err, errs.ErrChatNotFound) {
			return nil, fmt.Errorf("%w: id %v", err, chatID)
		}
		return nil, fmt.Errorf("get chat %v: %w", chatID, err)
	}
	log.Debug("chat retrieved")

	return chat, nil
}

func (s *ChatService) GetUserChats(ctx context.Context, userID uuid.UUID) ([]model.Chat, error) {
	log := logger.FromContext(ctx)
	log.Debug("chatService.GetUserChats")

	chats, err := s.chatRepository.GetUserChats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user %v chats: %w", userID, err)
	}
	log.Debug("chats retrieved")

	return chats, nil
}

func (s *ChatService) CreateChat(ctx context.Context, name string, userIDs []uuid.UUID, requesterID uuid.UUID) (*model.Chat, error) {
	log := logger.FromContext(ctx)
	log.Debug("chatService.CreateChat")

	newID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%w: generating uuid: %w", errs.ErrInternal, err)
	}

	chat := &model.Chat{
		ID:   newID,
		Name: name,
	}

	err = s.chatRepository.CreateChat(ctx, chat, requesterID)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	log.Debug("chat created")

	err = s.userService.AddChatUsers(ctx, chat.ID, userIDs, requesterID)
	if err != nil {
		return nil, fmt.Errorf("add users to new chat: %w", err)
	}
	log.Debug("chat users added")

	return chat, nil
}

func (s *ChatService) UpdateChatName(ctx context.Context, chatID uuid.UUID, name string, requesterID uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("chatService.UpdateChatName", "chatID", chatID)

	if !s.userService.hasPermission(ctx, chatID, requesterID, model.Admin) {
		return fmt.Errorf(
			"%w: user %v has no permission to update chat %v name",
			errs.ErrForbidden,
			requesterID,
			chatID,
		)
	}

	err := s.chatRepository.UpdateChatName(ctx, chatID, name)
	if err != nil {
		return fmt.Errorf("update chat %v name: %w", chatID, err)
	}
	log.Debug("chat name updated")

	return nil
}

func (s *ChatService) DeleteChat(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) error {
	log := logger.FromContext(ctx)
	log.Debug("chatService.DeleteChat", "chatID", chatID)

	if !s.userService.hasPermission(ctx, chatID, requesterID, model.Owner) {
		return fmt.Errorf(
			"%w: user %v has no permission to delete chat %v",
			errs.ErrForbidden,
			requesterID,
			chatID,
		)
	}

	err := s.chatRepository.DeleteChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("delete chat %v: %w", chatID, err)
	}
	log.Debug("chat deleted")

	return nil
}

func (s *ChatService) ChatExists(ctx context.Context, chatID uuid.UUID) bool {
	exists, err := s.chatRepository.ChatExists(ctx, chatID)
	if err != nil {
		return false
	}

	return exists
}

func (s *ChatService) AllowedToConnect(ctx context.Context, chatID, userID uuid.UUID) bool {
	return s.userService.userInChat(ctx, chatID, userID)
}
