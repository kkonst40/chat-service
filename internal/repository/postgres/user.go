package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kkonst40/chat-service/internal/domain/model"
	"github.com/kkonst40/chat-service/internal/repository"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetChatUser(ctx context.Context, chatID, userID uuid.UUID) (*model.User, error) {
	const query = `
		SELECT id, chat_id, role
		FROM users
		WHERE id = $1 AND chat_id = $2
	`

	slog.DebugContext(ctx, "getting chat user from DB", "chatID", chatID, "userID", userID)

	var user model.User
	err := r.db.QueryRowContext(ctx, query, userID, chatID).Scan(
		&user.ID,
		&user.ChatID,
		&user.Role,
	)

	if err == sql.ErrNoRows {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", repository.ErrDatabase, err)
	}

	return &user, nil
}

func (r *UserRepository) GetChatUserIDs(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error) {
	const query = `
		SELECT id
		FROM users
		WHERE chat_id = $1
	`

	slog.DebugContext(ctx, "getting chat userIDs from DB", "chatID", chatID)

	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", repository.ErrDatabase, err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("%w: %w", repository.ErrDatabase, err)
		}

		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", repository.ErrDatabase, err)
	}

	return userIDs, nil
}

func (r *UserRepository) GetChatUsers(ctx context.Context, chatID uuid.UUID) ([]model.User, error) {
	const query = `
		SELECT id, chat_id, role
		FROM users
		WHERE chat_id = $1
	`

	slog.DebugContext(ctx, "getting chat users from DB", "chatID", chatID)

	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", repository.ErrDatabase, err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(
			&user.ID,
			&user.ChatID,
			&user.Role,
		); err != nil {
			return nil, fmt.Errorf("%w: %w", repository.ErrDatabase, err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", repository.ErrDatabase, err)
	}

	return users, nil
}

func (r *UserRepository) GetPersonalChatsInterlocutors(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	const query = `
		SELECT c.id, u2.id
		FROM users u1
		JOIN chats c ON u1.chat_id = c.id
		JOIN users u2 ON u1.chat_id = u2.chat_id
		WHERE u1.id = $1
			AND c.is_group = FALSE
			AND u2.id != $1;
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", repository.ErrDatabase, err)
	}
	defer rows.Close()

	chatsInterlocutors := make(map[uuid.UUID]uuid.UUID)
	for rows.Next() {
		var chatID uuid.UUID
		var interlocutorID uuid.UUID
		if err := rows.Scan(&chatID, &interlocutorID); err != nil {
			return nil, fmt.Errorf("%w: %w", repository.ErrDatabase, err)
		}

		chatsInterlocutors[chatID] = interlocutorID
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", repository.ErrDatabase, err)
	}

	return chatsInterlocutors, nil

}

func (r *UserRepository) AddChatUsers(ctx context.Context, chatID uuid.UUID, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(userIDs) == 0 {
		return make([]uuid.UUID, 0), nil
	}

	slog.DebugContext(ctx, "adding chat users in DB", "chatID", chatID, "userIDs", userIDs)

	const query = `
		INSERT INTO users (id, chat_id, role)
		SELECT u_id, $1, $2
		FROM UNNEST($3::uuid[]) AS u_id
		ON CONFLICT (id, chat_id) DO NOTHING
		RETURNING id
	`

	rows, err := r.db.QueryContext(ctx, query, chatID, model.Common, userIDs)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" && pgErr.ConstraintName == "fk_users_chat" {
				return nil, repository.ErrNotFound
			}
		}
		return nil, fmt.Errorf("%w: %w", repository.ErrDatabase, err)
	}
	defer rows.Close()

	insertedIDs := make([]uuid.UUID, 0, len(userIDs))
	for rows.Next() {
		var returnedID uuid.UUID
		if err := rows.Scan(&returnedID); err != nil {
			return nil, fmt.Errorf("%w: %w", repository.ErrDatabase, err)
		}
		insertedIDs = append(insertedIDs, returnedID)
	}

	return insertedIDs, nil
}

func (r *UserRepository) DeleteChatUser(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) error {
	const query = `
		DELETE FROM users
		WHERE id = $1 AND chat_id = $2
	`

	slog.DebugContext(ctx, "deleting chat user in DB", "chatID", chatID, "userID", userID)

	if _, err := r.db.ExecContext(ctx, query, userID, chatID); err != nil {
		return fmt.Errorf("%w: %w", repository.ErrDatabase, err)
	}

	return nil
}

func (r *UserRepository) UpdateUserRole(ctx context.Context, chatID, userID uuid.UUID, newRole model.Role) error {
	const query = `
		UPDATE users
		SET role = $1
		WHERE id = $2 AND chat_id = $3
	`

	slog.DebugContext(ctx, "updating chat user role in DB", "chatID", chatID, "userID", userID, "role", newRole)

	res, err := r.db.ExecContext(ctx, query, newRole, userID, chatID)
	if err != nil {
		return fmt.Errorf("%w: %w", repository.ErrDatabase, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %w", repository.ErrDatabase, err)
	}

	if rowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *UserRepository) UserInChat(ctx context.Context, chatID, userID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE id = $1 AND chat_id = $2
		)
	`

	slog.DebugContext(ctx, "checking chat user existance in DB", "chatID", chatID, "userID", userID)

	var exists bool

	err := r.db.QueryRowContext(ctx, query, userID, chatID).Scan(
		&exists,
	)

	if err != nil {
		return false, fmt.Errorf("%w: %w", repository.ErrDatabase, err)
	}

	return exists, nil
}
