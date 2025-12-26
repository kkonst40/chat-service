package memory

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type key struct {
	UserID uuid.UUID
	ChatID uuid.UUID
}

type UserRepository struct {
	users map[key]*model.User
	mu    sync.RWMutex
}

func NewUserRepository() *UserRepository {
	userID1 := uuid.MustParse("018f95a5-bc7c-7e5c-9a4a-12c5d7316c3e")
	userID2 := uuid.MustParse("018f95a5-bc7d-7500-9b03-3e2a1c5d0f4a")

	ID0 := uuid.MustParse("018f95a6-8d27-7e03-822c-6a81ce0d1f4b")
	ID1 := uuid.MustParse("018f95a6-8d28-7b41-8d9f-12e8b341a6c5")
	ID2 := uuid.MustParse("018f95a6-8d28-7c8a-9123-4f67d890a12e")
	ID3 := uuid.MustParse("018f95a6-8d28-7de9-b8a4-5c32f1e09d76")
	ID4 := uuid.MustParse("018f95a6-8d29-7023-84d1-9a0b2c3e4f5a")
	ID5 := uuid.MustParse("018f95a6-8d29-71bc-9a8b-76e54d32c10f")
	ID6 := uuid.MustParse("018f95a6-8d2a-72f4-a123-b456c789d0e1")
	ID7 := uuid.MustParse("018f95a6-8d2a-743d-b987-6543210fedcb")
	ID8 := uuid.MustParse("018f95a6-8d2b-7586-c246-8ace13579bdf")
	ID9 := uuid.MustParse("018f95a6-8d2b-76cf-d369-147f258b036a")

	users := make(map[key]*model.User)

	users[key{UserID: userID1, ChatID: ID0}] = &model.User{ID: userID1, ChatID: ID0, Role: model.Common}
	users[key{UserID: userID1, ChatID: ID1}] = &model.User{ID: userID1, ChatID: ID1, Role: model.Common}
	users[key{UserID: userID1, ChatID: ID3}] = &model.User{ID: userID1, ChatID: ID3, Role: model.Common}
	users[key{UserID: userID1, ChatID: ID5}] = &model.User{ID: userID1, ChatID: ID5, Role: model.Common}
	users[key{UserID: userID1, ChatID: ID7}] = &model.User{ID: userID1, ChatID: ID7, Role: model.Common}
	users[key{UserID: userID1, ChatID: ID8}] = &model.User{ID: userID1, ChatID: ID8, Role: model.Common}
	users[key{UserID: userID1, ChatID: ID9}] = &model.User{ID: userID1, ChatID: ID9, Role: model.Common}

	users[key{UserID: userID2, ChatID: ID0}] = &model.User{ID: userID2, ChatID: ID0, Role: model.Common}
	users[key{UserID: userID2, ChatID: ID1}] = &model.User{ID: userID2, ChatID: ID1, Role: model.Common}
	users[key{UserID: userID2, ChatID: ID2}] = &model.User{ID: userID2, ChatID: ID2, Role: model.Common}
	users[key{UserID: userID2, ChatID: ID3}] = &model.User{ID: userID2, ChatID: ID3, Role: model.Common}
	users[key{UserID: userID2, ChatID: ID4}] = &model.User{ID: userID2, ChatID: ID4, Role: model.Common}
	users[key{UserID: userID2, ChatID: ID6}] = &model.User{ID: userID2, ChatID: ID6, Role: model.Common}
	users[key{UserID: userID2, ChatID: ID7}] = &model.User{ID: userID2, ChatID: ID7, Role: model.Common}
	users[key{UserID: userID2, ChatID: ID9}] = &model.User{ID: userID2, ChatID: ID9, Role: model.Common}

	return &UserRepository{
		users: users,
	}
}

func (r *UserRepository) GetChatUser(chatID uuid.UUID, userID uuid.UUID) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := key{UserID: userID, ChatID: chatID}
	user, ok := r.users[key]
	if !ok {
		return nil, fmt.Errorf("user with ID %v is not in the chat ID %v", userID, chatID)
	}

	return user, nil
}

func (r *UserRepository) GetChatUsers(chatID uuid.UUID) ([]*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*model.User, 0)
	for _, user := range r.users {
		if user.ChatID == chatID {
			result = append(result, user)
		}
	}

	return result, nil
}

func (r *UserRepository) GetUserChatIds(userID uuid.UUID) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]uuid.UUID, 0)
	for _, user := range r.users {
		if user.ID == userID {
			result = append(result, user.ChatID)
		}
	}

	return result, nil
}

func (r *UserRepository) AddChatUsers(chatID uuid.UUID, userIDs []uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, userID := range userIDs {
		key := key{UserID: userID, ChatID: chatID}
		if _, ok := r.users[key]; !ok {
			user := &model.User{
				ID:     userID,
				ChatID: chatID,
				Role:   model.Common,
			}
			r.users[key] = user
		}
	}

	return nil
}

func (r *UserRepository) DeleteChatUser(chatID uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := key{UserID: userID, ChatID: chatID}
	delete(r.users, key)

	return nil
}

func (r *UserRepository) SetUserRole(chatID, userID uuid.UUID, newRole model.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := key{UserID: userID, ChatID: chatID}
	if _, ok := r.users[key]; ok {
		r.users[key].Role = newRole
	} else {
		return fmt.Errorf("user with ID %v is not in chat with ID %v", userID, chatID)
	}

	return nil
}

func (r *UserRepository) IsUserInChat(chatID, userID uuid.UUID) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.ID == userID && user.ChatID == chatID {
			return true
		}
	}

	return false
}

var _ repository.UserRepository = (*UserRepository)(nil)
