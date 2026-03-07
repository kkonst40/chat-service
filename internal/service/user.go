package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/dispatcher"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
	"github.com/kkonst40/ichat/internal/domain/model"
	"github.com/kkonst40/ichat/internal/integration/sso"
	"github.com/kkonst40/ichat/internal/repository"
)

type UserLoginCache interface {
	GetUserLogins(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error)
	SetUserLogins(ctx context.Context, logins map[uuid.UUID]string) error
}

type UserService struct {
	userRepository repository.UserRepository
	dispatcher     *dispatcher.Dispatcher
	ssoClient      *sso.SSOService
	loginCache     UserLoginCache
}

func NewUserService(
	userRepository repository.UserRepository,
	dispatcher *dispatcher.Dispatcher,
	ssoClient *sso.SSOService,
	loginCache UserLoginCache,
) *UserService {
	return &UserService{
		userRepository: userRepository,
		dispatcher:     dispatcher,
		ssoClient:      ssoClient,
		loginCache:     loginCache,
	}
}

func (s *UserService) GetChatUser(ctx context.Context, chatID, userID uuid.UUID) (*model.User, error) {
	//??
	return s.userRepository.GetChatUser(ctx, chatID, userID)
}

func (s *UserService) GetChatUsers(ctx context.Context, chatID uuid.UUID, requesterID uuid.UUID) ([]model.User, error) {
	slog.DebugContext(ctx, "userService.GetChatUsers", "chatID", chatID)

	if !s.userInChat(ctx, chatID, requesterID) {
		return nil, fmt.Errorf(
			"%w: user %v is not in chat %v",
			errs.ErrForbidden,
			requesterID,
			chatID,
		)
	}

	user, err := s.userRepository.GetChatUsers(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get chat %v users: %w", chatID, err)
	}
	slog.DebugContext(ctx, "chat users retrieved")

	return user, nil
}

func (s *UserService) GetChatUserIDs(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error) {
	slog.DebugContext(ctx, "userService.GetChatUserIDs", "chatID", chatID)

	userIDs, err := s.userRepository.GetChatUserIDs(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get chat %v user ids: %w", chatID, err)
	}
	slog.DebugContext(ctx, "chat user ids retrieved")

	return userIDs, nil
}

func (s *UserService) AddChatUsers(ctx context.Context, chatID uuid.UUID, userIDs []uuid.UUID, requesterID uuid.UUID) error {
	slog.DebugContext(ctx, "userService.AddChatUsers", "chatID", chatID)

	if !s.userInChat(ctx, chatID, requesterID) {
		return fmt.Errorf(
			"%w: user %v is not in chat %v",
			errs.ErrForbidden,
			requesterID,
			chatID,
		)
	}

	existingUserIDs, err := s.existMany(ctx, userIDs)
	if err != nil {
		return fmt.Errorf("check users existence before add to chat %v: %w", chatID, err)
	}

	err = s.userRepository.AddChatUsers(ctx, chatID, existingUserIDs)
	if err != nil {
		if errors.Is(err, errs.ErrChatNotFound) {
			return fmt.Errorf("%w: ID %v", err, chatID)
		}
		return fmt.Errorf("add chat %v users: %w", chatID, err)
	}
	slog.DebugContext(ctx, "chat users added")

	return nil
}

func (s *UserService) DeleteChatUser(ctx context.Context, chatID uuid.UUID, userID uuid.UUID, requesterID uuid.UUID) error {
	slog.DebugContext(ctx, "userService.DeleteChatUser", "chatID", chatID)

	if userID == requesterID {
		err := s.userRepository.DeleteChatUser(ctx, chatID, userID)
		if err != nil {
			return fmt.Errorf("delete user %v from chat %v: %w", userID, chatID, err)
		}
		slog.DebugContext(ctx, "user deleted")

		return nil
	}

	requester, err := s.userRepository.GetChatUser(ctx, chatID, requesterID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return fmt.Errorf(
				"%w: user %v is not in chat %v",
				errs.ErrForbidden,
				requesterID,
				chatID,
			)
		}
		return fmt.Errorf("get requester %v to perform delete: %w", requesterID, err)
	}

	user, err := s.userRepository.GetChatUser(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return fmt.Errorf("%w: user ID %v, chat ID %v", err, userID, chatID)
		}
		return fmt.Errorf("get user %v to delete: %w", userID, err)
	}

	if !(model.UserPriority[requester.Role] > model.UserPriority[user.Role]) {
		return fmt.Errorf(
			"%w: user %v has no permission to delete user %v from chat %v",
			errs.ErrForbidden,
			requesterID,
			userID,
			chatID,
		)
	}

	err = s.userRepository.DeleteChatUser(ctx, chatID, userID)
	if err != nil {
		return fmt.Errorf("delete user %v from chat %v: %w", userID, chatID, err)
	}
	slog.DebugContext(ctx, "user deleted")

	return nil
}

func (s *UserService) UpdateUserRole(ctx context.Context, chatID, userID uuid.UUID, newRole model.Role, requesterID uuid.UUID) error {
	slog.DebugContext(ctx, "userService.UpdateUserRole", "chatID", chatID, "chatUserID", userID, "role", newRole)

	user, err := s.userRepository.GetChatUser(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return fmt.Errorf("%w: user ID %v, chat ID %v", err, userID, chatID)
		}
		return fmt.Errorf("get requester %v to perform role update: %w", requesterID, err)
	}

	requester, err := s.userRepository.GetChatUser(ctx, chatID, requesterID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return fmt.Errorf(
				"%w: user %v is not in chat %v",
				errs.ErrForbidden,
				requesterID,
				chatID,
			)
		}
		return fmt.Errorf("get user %v to update:%w", userID, err)
	}

	if !(model.UserPriority[requester.Role] > model.UserPriority[user.Role]) {
		return fmt.Errorf(
			"%w: user %v has no permission to update user %v role",
			errs.ErrForbidden,
			requesterID,
			userID,
		)
	}

	err = s.userRepository.UpdateUserRole(ctx, chatID, userID, newRole)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return fmt.Errorf("%w: user ID %v, chat ID %v", err, userID, chatID)
		}
		return fmt.Errorf("update user %v role (to %v) in chat %v: %w", userID, newRole, chatID, err)
	}
	slog.DebugContext(ctx, "user role updated")

	return nil
}

func (s *UserService) hasPermission(ctx context.Context, chatID, requesterID uuid.UUID, role model.Role) bool {
	requester, err := s.GetChatUser(ctx, chatID, requesterID)
	if err != nil {
		return false
	}

	if model.UserPriority[requester.Role] < model.UserPriority[role] {
		return false
	}

	return true
}

func (s *UserService) userInChat(ctx context.Context, chatID, userID uuid.UUID) bool {
	result, err := s.userRepository.UserInChat(ctx, chatID, userID)
	if err != nil {
		return false
	}

	return result
}

func (s *UserService) getPersonalChatsInterlocutors(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	chatsInterlocutors, err := s.userRepository.GetPersonalChatsInterlocutors(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get personal chats interlocutors IDs of user %v: %w", userID, err)
	}

	return chatsInterlocutors, nil
}

func (s *UserService) existMany(ctx context.Context, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	IDs, err := s.ssoClient.ExistMany(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: sso service (ExistMany): %w",
			errs.ErrExternalService,
			err,
		)
	}

	return IDs, nil
}

func (s *UserService) getUserLogins(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	userIDs = unique(userIDs)
	result := make(map[uuid.UUID]string, len(userIDs))

	if s.loginCache == nil {
		userInfos, err := s.ssoClient.GetUsersLogins(ctx, userIDs)
		if err != nil {
			return nil, fmt.Errorf("%w: sso service (GetUsersLogins): %w", errs.ErrExternalService, err)
		}

		for _, userInfo := range userInfos {
			result[userInfo.ID] = userInfo.Login
		}

		return result, nil
	}

	cachedLogins, err := s.loginCache.GetUserLogins(ctx, userIDs)
	if err != nil {
		slog.ErrorContext(ctx, "user login cache GetUserLogins error", "error", err)
		cachedLogins = map[uuid.UUID]string{}
	}

	slog.DebugContext(ctx, "got logins from cache")

	maps.Copy(result, cachedLogins)

	missingIDs := make([]uuid.UUID, 0, len(userIDs))
	for _, id := range userIDs {
		if _, ok := cachedLogins[id]; !ok {
			missingIDs = append(missingIDs, id)
		}
	}

	if len(missingIDs) == 0 {
		return result, nil
	}

	userInfos, err := s.ssoClient.GetUsersLogins(ctx, missingIDs)
	if err != nil {
		return nil, fmt.Errorf("%w: sso service (GetUsersLogins): %w", errs.ErrExternalService, err)
	}

	toCache := make(map[uuid.UUID]string, len(userInfos))
	for _, userInfo := range userInfos {
		result[userInfo.ID] = userInfo.Login
		toCache[userInfo.ID] = userInfo.Login
	}

	if err := s.loginCache.SetUserLogins(ctx, toCache); err != nil {
		slog.ErrorContext(ctx, "user login cache SetUserLogins error", "error", err)
	}

	return result, nil
}
