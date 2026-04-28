package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	errs "github.com/kkonst40/chat-service/internal/domain/errors"
	"github.com/kkonst40/chat-service/internal/domain/event"
	"github.com/kkonst40/chat-service/internal/domain/model"
	"github.com/kkonst40/chat-service/internal/repository"
	"github.com/kkonst40/chat-service/internal/service/dispatcher"
)

type ChatService struct {
	chatRepository ChatRepository
	userService    *UserService
	dispatcher     *dispatcher.Dispatcher
}

type ChatRepository interface {
	GetChat(ctx context.Context, chatID uuid.UUID) (model.Chat, error)
	GetUserChats(ctx context.Context, userID uuid.UUID, filter model.ChatFilter) ([]model.Chat, error)
	CreateGroupChat(ctx context.Context, chat *model.Chat, creatorID uuid.UUID, userIDs []uuid.UUID) error
	CreatePersonalChat(ctx context.Context, chat *model.Chat, userID1, userID2 uuid.UUID) error
	UpdateChatName(ctx context.Context, chatID uuid.UUID, name string) error
	DeleteChat(ctx context.Context, chatID uuid.UUID) error
	DeletePersonalChat(ctx context.Context, userID1, userID2 uuid.UUID) error
	ChatExists(ctx context.Context, chatID uuid.UUID) (bool, error)
}

func NewChatService(
	chatRepository ChatRepository,
	userService *UserService,
	dispatcher *dispatcher.Dispatcher,
) *ChatService {
	service := ChatService{
		chatRepository: chatRepository,
		userService:    userService,
		dispatcher:     dispatcher,
	}

	return &service
}

func (s *ChatService) GetChat(ctx context.Context, chatID uuid.UUID) (model.Chat, error) {
	slog.DebugContext(ctx, "chatService.GetChat", "chatID", chatID)

	chat, err := s.chatRepository.GetChat(ctx, chatID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Chat{}, fmt.Errorf("%w: id %v", errs.ErrChatNotFound, chatID)
		}
		return model.Chat{}, fmt.Errorf("get chat %v: %w", chatID, err)
	}
	slog.DebugContext(ctx, "chat retrieved")

	return chat, nil
}

func (s *ChatService) GetUserChats(ctx context.Context, userID uuid.UUID, filter model.ChatFilter) ([]model.Chat, error) {
	slog.DebugContext(ctx, "chatService.GetUserChats")

	chats, err := s.chatRepository.GetUserChats(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("get user %v chats: %w", userID, err)
	}
	slog.DebugContext(ctx, "chats retrieved")

	if filter == model.GroupChats {
		return chats, nil
	}

	chatsInterlocutors, err := s.userService.getPersonalChatsInterlocutors(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user %v chats: %w", userID, err)
	}

	userIDs := make([]uuid.UUID, 0, len(chatsInterlocutors))
	for _, interlocutorID := range chatsInterlocutors {
		userIDs = append(userIDs, interlocutorID)
	}

	logins, err := s.userService.getUserLogins(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("get user %v chats: %w", userID, err)
	}

	for i := range chats {
		if !chats[i].IsGroup {
			chats[i].Name = logins[chatsInterlocutors[chats[i].ID]]
		}
	}

	return chats, nil
}

func (s *ChatService) CreateGroupChat(ctx context.Context, name string, userNames []string, requesterID uuid.UUID) (*model.Chat, error) {
	slog.DebugContext(ctx, "chatService.CreateChat")

	newID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%w: generating uuid: %w", errs.ErrInternal, err)
	}

	chat := &model.Chat{
		ID:            newID,
		Name:          name,
		IsGroup:       true,
		LastMessageAt: time.Now(),
	}

	userIDsMap, err := s.userService.getUserIDs(ctx, userNames)
	if err != nil {
		return nil, fmt.Errorf("get user IDs before create chat: %w", err)
	}

	userIDs := make([]uuid.UUID, 0, len(userIDsMap))
	for _, userName := range userNames {
		userID, ok := userIDsMap[userName]
		if ok {
			userIDs = append(userIDs, userID)
		}
	}

	err = s.chatRepository.CreateGroupChat(ctx, chat, requesterID, userIDs)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	slog.DebugContext(ctx, "chat created")

	err = s.userService.AddChatUsers(ctx, chat.ID, userNames, requesterID)
	if err != nil {
		return nil, fmt.Errorf("add users to new chat: %w", err)
	}
	slog.DebugContext(ctx, "chat users added")

	s.dispatcher.Publish(event.Event{
		Type:   event.CreateChat,
		ChatID: newID,
		Payload: event.CreateChatEvent{
			Name: name,
		},
	})

	return chat, nil
}

func (s *ChatService) CreatePersonalChat(ctx context.Context, userID1 uuid.UUID, userName2 string) (*model.Chat, error) {
	slog.DebugContext(ctx, "chatService.CreatePersonalChat", "user1", userID1, "user2", userName2)

	newID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%w: generating uuid: %w", errs.ErrInternal, err)
	}

	chat := &model.Chat{
		ID:            newID,
		Name:          "",
		IsGroup:       false,
		LastMessageAt: time.Now(),
	}

	idMap, err := s.userService.getUserIDs(ctx, []string{userName2})
	if err != nil {
		return nil, fmt.Errorf("create personal chat: %w", err)
	}

	userID2, ok := idMap[userName2]
	if !ok {
		return nil, fmt.Errorf("%w: user name '%s'", errs.ErrUserNotFound, userName2)
	}

	err = s.chatRepository.CreatePersonalChat(ctx, chat, userID1, userID2)
	if err != nil {
		return nil, fmt.Errorf("create personal chat: %w", err)
	}
	slog.DebugContext(ctx, "personal chat created")

	s.dispatcher.Publish(event.Event{
		Type:   event.CreateChat,
		ChatID: newID,
		Payload: event.CreateChatEvent{
			Name: "",
		},
	})

	return chat, nil
}

func (s *ChatService) UpdateChatName(ctx context.Context, chatID uuid.UUID, name string, requesterID uuid.UUID) error {
	slog.DebugContext(ctx, "chatService.UpdateChatName", "chatID", chatID)

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
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("%w: ID %v", errs.ErrChatNotFound, chatID)
		}
		return fmt.Errorf("update chat %v name: %w", chatID, err)
	}
	slog.DebugContext(ctx, "chat name updated")

	s.dispatcher.Publish(event.Event{
		Type:   event.UpdateChat,
		ChatID: chatID,
		Payload: event.UpdateChatEvent{
			Name: name,
		},
	})

	return nil
}

func (s *ChatService) DeleteChat(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) error {
	slog.DebugContext(ctx, "chatService.DeleteChat", "chatID", chatID)

	if !s.userService.hasPermission(ctx, chatID, requesterID, model.Owner) {
		return fmt.Errorf(
			"%w: user %v has no permission to delete chat %v",
			errs.ErrForbidden,
			requesterID,
			chatID,
		)
	}

	userIDs, err := s.userService.GetChatUserIDs(ctx, chatID)
	if err != nil {
		return fmt.Errorf("get chat %v user ids: %w", chatID, err)
	}

	err = s.chatRepository.DeleteChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("delete chat %v: %w", chatID, err)
	}
	slog.DebugContext(ctx, "chat deleted")

	s.dispatcher.Publish(event.Event{
		Type:    event.DeleteChat,
		ChatID:  chatID,
		Payload: event.DeleteChatEvent{},
	}, userIDs...)

	return nil
}

func (s *ChatService) ChatExists(ctx context.Context, chatID uuid.UUID) bool {
	exists, err := s.chatRepository.ChatExists(ctx, chatID)
	if err != nil {
		return false
	}

	return exists
}
