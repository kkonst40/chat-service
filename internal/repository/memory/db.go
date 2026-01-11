package memory

import (
	"sync"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/model"
)

type MemoryDB struct {
	chats    map[uuid.UUID]*model.Chat
	messages map[uuid.UUID]*model.Message
	users    map[key]*model.User
	mu       sync.RWMutex
}

func NewDB() *MemoryDB {
	db := MemoryDB{
		chats:    make(map[uuid.UUID]*model.Chat),
		messages: make(map[uuid.UUID]*model.Message),
		users:    make(map[key]*model.User),
	}

	userID1 := uuid.MustParse("019bad77-a48a-712b-af62-65e0cc331079") //qqq
	userID2 := uuid.MustParse("019baab0-f622-76c6-9a56-d5782ef27693") //pavel

	chatID0 := uuid.MustParse("018f95a6-8d27-7e03-822c-6a81ce0d1f4b")
	chatID1 := uuid.MustParse("018f95a6-8d28-7b41-8d9f-12e8b341a6c5")
	chatID2 := uuid.MustParse("018f95a6-8d28-7c8a-9123-4f67d890a12e")
	chatID3 := uuid.MustParse("018f95a6-8d28-7de9-b8a4-5c32f1e09d76")
	chatID4 := uuid.MustParse("018f95a6-8d29-7023-84d1-9a0b2c3e4f5a")
	chatID5 := uuid.MustParse("018f95a6-8d29-71bc-9a8b-76e54d32c10f")
	chatID6 := uuid.MustParse("018f95a6-8d2a-72f4-a123-b456c789d0e1")
	chatID7 := uuid.MustParse("018f95a6-8d2a-743d-b987-6543210fedcb")
	chatID8 := uuid.MustParse("018f95a6-8d2b-7586-c246-8ace13579bdf")
	chatID9 := uuid.MustParse("018f95a6-8d2b-76cf-d369-147f258b036a")

	db.chats[chatID0] = &model.Chat{ID: chatID0, Name: "Chat0"}
	db.chats[chatID1] = &model.Chat{ID: chatID1, Name: "Chat1"}
	db.chats[chatID2] = &model.Chat{ID: chatID2, Name: "Chat2"}
	db.chats[chatID3] = &model.Chat{ID: chatID3, Name: "Chat3"}
	db.chats[chatID4] = &model.Chat{ID: chatID4, Name: "Chat4"}
	db.chats[chatID5] = &model.Chat{ID: chatID5, Name: "Chat5"}
	db.chats[chatID6] = &model.Chat{ID: chatID6, Name: "Chat6"}
	db.chats[chatID7] = &model.Chat{ID: chatID7, Name: "Chat7"}
	db.chats[chatID8] = &model.Chat{ID: chatID8, Name: "Chat8"}
	db.chats[chatID9] = &model.Chat{ID: chatID9, Name: "Chat9"}

	db.users[key{UserID: userID1, ChatID: chatID0}] = &model.User{ID: userID1, ChatID: chatID0, Role: model.Admin}
	db.users[key{UserID: userID1, ChatID: chatID1}] = &model.User{ID: userID1, ChatID: chatID1, Role: model.Admin}
	db.users[key{UserID: userID1, ChatID: chatID3}] = &model.User{ID: userID1, ChatID: chatID3, Role: model.Admin}
	db.users[key{UserID: userID1, ChatID: chatID5}] = &model.User{ID: userID1, ChatID: chatID5, Role: model.Admin}
	db.users[key{UserID: userID1, ChatID: chatID7}] = &model.User{ID: userID1, ChatID: chatID7, Role: model.Admin}
	db.users[key{UserID: userID1, ChatID: chatID8}] = &model.User{ID: userID1, ChatID: chatID8, Role: model.Admin}
	db.users[key{UserID: userID1, ChatID: chatID9}] = &model.User{ID: userID1, ChatID: chatID9, Role: model.Admin}

	db.users[key{UserID: userID2, ChatID: chatID0}] = &model.User{ID: userID2, ChatID: chatID0, Role: model.Common}
	db.users[key{UserID: userID2, ChatID: chatID1}] = &model.User{ID: userID2, ChatID: chatID1, Role: model.Common}
	db.users[key{UserID: userID2, ChatID: chatID2}] = &model.User{ID: userID2, ChatID: chatID2, Role: model.Common}
	db.users[key{UserID: userID2, ChatID: chatID3}] = &model.User{ID: userID2, ChatID: chatID3, Role: model.Common}
	db.users[key{UserID: userID2, ChatID: chatID4}] = &model.User{ID: userID2, ChatID: chatID4, Role: model.Common}
	db.users[key{UserID: userID2, ChatID: chatID6}] = &model.User{ID: userID2, ChatID: chatID6, Role: model.Common}
	db.users[key{UserID: userID2, ChatID: chatID7}] = &model.User{ID: userID2, ChatID: chatID7, Role: model.Common}
	db.users[key{UserID: userID2, ChatID: chatID9}] = &model.User{ID: userID2, ChatID: chatID9, Role: model.Common}

	return &db
}
