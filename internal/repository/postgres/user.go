package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
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
	const query = `
		SELECT id, chat_id, role
		FROM users
		WHERE id = $1 AND chat_id = $2
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, userID, chatID).Scan(
		&user.ID,
		&user.ChatID,
		&user.Role,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetChatUsers(ctx context.Context, chatID uuid.UUID) ([]*model.User, error) {
	const query = `
		SELECT id, chat_id, role
		FROM users
		WHERE chat_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(
			&user.ID,
			&user.ChatID,
			&user.Role,
		); err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) AddChatUsers(ctx context.Context, chatID uuid.UUID, userIDs []uuid.UUID) error {
	if len(userIDs) == 0 {
		return nil
	}

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
	return err
}

func (r *UserRepository) DeleteChatUser(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) error {
	const query = `
		DELETE FROM users
		WHERE id = $1 AND chat_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, userID, chatID)
	return err
}

func (r *UserRepository) UpdateUserRole(ctx context.Context, chatID, userID uuid.UUID, newRole model.Role) error {
	const query = `
		UPDATE users
		SET role = $1
		WHERE id = $2 AND chat_id = $3
	`

	_, err := r.db.ExecContext(ctx, query, newRole, userID, chatID)
	return err
}

func (r *UserRepository) IsUserInChat(ctx context.Context, chatID, userID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE id = $1 AND chat_id = $2
		)
	`

	var exists bool

	err := r.db.QueryRowContext(ctx, query, userID, chatID).Scan(
		&exists,
	)

	if err != nil {
		return false, err
	}

	return exists, nil
}

var _ repository.UserRepository = (*UserRepository)(nil)
