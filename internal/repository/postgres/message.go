package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
	"github.com/kkonst40/ichat/internal/domain/model"
	"github.com/kkonst40/ichat/internal/logger"
	"github.com/kkonst40/ichat/internal/repository"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

func (r *MessageRepository) GetMessage(ctx context.Context, msgID uuid.UUID) (*model.Message, error) {
	log := logger.FromContext(ctx)

	const query = `
		SELECT id, user_id, chat_id, text, created_at
		FROM messages
		WHERE id = $1
	`

	log.Debug("getting message from DB", "msgID", msgID)

	var msg model.Message
	err := r.db.QueryRowContext(ctx, query, msgID).Scan(
		&msg.ID,
		&msg.UserID,
		&msg.ChatID,
		&msg.Text,
		&msg.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errs.ErrMsgNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return &msg, nil
}

func (r *MessageRepository) GetChatMessages(ctx context.Context, chatID uuid.UUID, from uuid.UUID, count int64) ([]model.Message, error) {
	log := logger.FromContext(ctx)
	const queryStart = `
        SELECT id, user_id, chat_id, text, created_at
        WHERE chat_id = $1
        ORDER BY id DESC
        LIMIT $2
	`

	const query = `
		SELECT id, user_id, chat_id, text, created_at
		FROM messages
		WHERE chat_id = $1
  			AND id < $2
		ORDER BY id DESC
		LIMIT $3;
	`

	log.Debug("getting chat messages from DB", "chatID", chatID)

	var rows *sql.Rows
	var err error
	if from == uuid.Nil {
		rows, err = r.db.QueryContext(ctx, queryStart, chatID, count)
	} else {
		rows, err = r.db.QueryContext(ctx, query, chatID, from, count)
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}
	defer rows.Close()

	var messages []model.Message

	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.UserID,
			&msg.ChatID,
			&msg.Text,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return messages, nil
}

func (r *MessageRepository) CreateMessage(ctx context.Context, msg *model.Message) error {
	log := logger.FromContext(ctx)
	const msgQuery = `
		INSERT INTO messages (id, user_id, chat_id, text, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	const chatQuery = `
		UPDATE chats
		SET last_message_at = $1
		WHERE id = $2
	`

	log.Debug("creating new message in DB", "msgID", msg.ID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	defer tx.Rollback()

	if _, err = tx.ExecContext(
		ctx,
		msgQuery,
		msg.ID,
		msg.UserID,
		msg.ChatID,
		msg.Text,
		msg.CreatedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				switch pgErr.ConstraintName {
				case "fk_messages_chat":
					return errs.ErrChatNotFound
				case "fk_messages_user":
					return errs.ErrUserNotFound
				}
			}
		}

		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	res, err := tx.ExecContext(ctx, chatQuery, msg.CreatedAt, msg.ChatID)
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	if rowsAffected == 0 {
		return errs.ErrChatNotFound
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return nil
}

func (r *MessageRepository) UpdateMessage(ctx context.Context, msg *model.Message) error {
	log := logger.FromContext(ctx)
	const query = `
		UPDATE messages
		SET text = $1
		WHERE id = $2
	`

	log.Debug("updating message in DB", "msgID", msg.ID)

	res, err := r.db.ExecContext(ctx, query, msg.Text, msg.ID)
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	if rowsAffected == 0 {
		return errs.ErrMsgNotFound
	}

	return nil
}

func (r *MessageRepository) DeleteMessage(ctx context.Context, msgID uuid.UUID) error {
	log := logger.FromContext(ctx)
	const query = `
		DELETE FROM messages
		WHERE id = $1
	`

	log.Debug("deleting message in DB", "msgID", msgID)

	if _, err := r.db.ExecContext(ctx, query, msgID); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return nil
}

var _ repository.MessageRepository = (*MessageRepository)(nil)
