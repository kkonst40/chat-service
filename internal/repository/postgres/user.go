package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/apperror"
	"github.com/kkonst40/ichat/internal/logger"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
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
	log := logger.FromContext(ctx)
	const query = `
		SELECT id, chat_id, role
		FROM users
		WHERE id = $1 AND chat_id = $2
	`

	log.Debug("getting chat user from DB", "chatID", chatID, "userID", userID)

	var user model.User
	err := r.db.QueryRowContext(ctx, query, userID, chatID).Scan(
		&user.ID,
		&user.ChatID,
		&user.Role,
	)

	if err == sql.ErrNoRows {
		return nil, &apperror.NotFoundError{Msg: fmt.Sprintf("user (%v) in chat (%v) not found", userID, chatID)}
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetChatUsers(ctx context.Context, chatID uuid.UUID) ([]model.User, error) {
	log := logger.FromContext(ctx)
	const query = `
		SELECT id, chat_id, role
		FROM users
		WHERE chat_id = $1
	`

	log.Debug("getting chat users from DB", "chatID", chatID)

	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, &apperror.DBError{Msg: err.Error()}
	}

	return users, nil
}

func (r *UserRepository) AddChatUsers(ctx context.Context, chatID uuid.UUID, userIDs []uuid.UUID) error {
	log := logger.FromContext(ctx)
	if len(userIDs) == 0 {
		return nil
	}

	log.Debug("adding chat users in DB", "chatID", chatID, "userIDs", userIDs)

	var queryBuilder strings.Builder
	queryBuilder.WriteString("INSERT INTO users (id, chat_id, role) VALUES ")
	args := make([]any, 0, len(userIDs)*3)

	for i, userID := range userIDs {
		if i > 0 {
			queryBuilder.WriteString(", ")
		}

		n := i * 3
		fmt.Fprintf(&queryBuilder, "($%d, $%d, $%d)", n+1, n+2, n+3)
		args = append(args, userID, chatID, model.Common)
	}

	_, err := r.db.ExecContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return &apperror.DBError{Msg: err.Error()}
	}

	return nil
}

func (r *UserRepository) DeleteChatUser(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) error {
	log := logger.FromContext(ctx)
	const query = `
		DELETE FROM users
		WHERE id = $1 AND chat_id = $2
	`

	log.Debug("deleting chat user in DB", "chatID", chatID, "userID", userID)

	_, err := r.db.ExecContext(ctx, query, userID, chatID)
	if err != nil {
		return &apperror.DBError{Msg: err.Error()}
	}

	return nil
}

func (r *UserRepository) UpdateUserRole(ctx context.Context, chatID, userID uuid.UUID, newRole model.Role) error {
	log := logger.FromContext(ctx)
	const query = `
		UPDATE users
		SET role = $1
		WHERE id = $2 AND chat_id = $3
	`

	log.Debug("updating chat user role in DB", "chatID", chatID, "userID", userID, "role", newRole)

	_, err := r.db.ExecContext(ctx, query, newRole, userID, chatID)
	if err != nil {
		return &apperror.DBError{Msg: err.Error()}
	}

	return nil
}

func (r *UserRepository) IsUserInChat(ctx context.Context, chatID, userID uuid.UUID) (bool, error) {
	log := logger.FromContext(ctx)
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE id = $1 AND chat_id = $2
		)
	`

	log.Debug("checking chat user existance in DB", "chatID", chatID, "userID", userID)

	var exists bool

	err := r.db.QueryRowContext(ctx, query, userID, chatID).Scan(
		&exists,
	)

	if err != nil {
		return false, &apperror.DBError{Msg: err.Error()}
	}

	return exists, nil
}

var _ repository.UserRepository = (*UserRepository)(nil)
