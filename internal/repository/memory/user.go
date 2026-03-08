package memory

import (
	"context"

	"github.com/google/uuid"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
	"github.com/kkonst40/ichat/internal/domain/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type key struct {
	UserID uuid.UUID
	ChatID uuid.UUID
}

type UserRepository struct {
	db *MemoryDB
}

func NewUserRepository(db *MemoryDB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetChatUser(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) (*model.User, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	key := key{UserID: userID, ChatID: chatID}
	user, ok := r.db.users[key]
	if !ok {
		return nil, errs.ErrUserNotFound
	}

	return user, nil
}

func (r *UserRepository) GetChatUserIDs(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	result := make([]uuid.UUID, 0)
	for _, user := range r.db.users {
		if user.ChatID == chatID {
			result = append(result, user.ID)
		}
	}

	return result, nil
}

func (r *UserRepository) GetChatUsers(ctx context.Context, chatID uuid.UUID) ([]model.User, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	result := make([]model.User, 0)
	for _, user := range r.db.users {
		if user.ChatID == chatID {
			result = append(result, *user)
		}
	}

	return result, nil
}

func (r *UserRepository) GetPersonalChatsInterlocutors(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	return make(map[uuid.UUID]uuid.UUID), nil
}

func (r *UserRepository) AddChatUsers(ctx context.Context, chatID uuid.UUID, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	for _, userID := range userIDs {
		key := key{UserID: userID, ChatID: chatID}
		if _, ok := r.db.users[key]; !ok {
			user := &model.User{
				ID:     userID,
				ChatID: chatID,
				Role:   model.Common,
			}
			r.db.users[key] = user
		}
	}

	return userIDs, nil
}

func (r *UserRepository) DeleteChatUser(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	key := key{UserID: userID, ChatID: chatID}
	delete(r.db.users, key)

	return nil
}

func (r *UserRepository) UpdateUserRole(ctx context.Context, chatID, userID uuid.UUID, newRole model.Role) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	key := key{UserID: userID, ChatID: chatID}
	if _, ok := r.db.users[key]; ok {
		r.db.users[key].Role = newRole
	} else {
		return errs.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) UserInChat(ctx context.Context, chatID, userID uuid.UUID) (bool, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	key := key{UserID: userID, ChatID: chatID}
	if _, ok := r.db.users[key]; ok {
		return true, nil
	}

	return false, nil
}

var _ repository.UserRepository = (*UserRepository)(nil)
