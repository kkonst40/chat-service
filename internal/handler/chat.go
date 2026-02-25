package handler

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/dto"
	errs "github.com/kkonst40/ichat/internal/errors"
	"github.com/kkonst40/ichat/internal/logger"
	"github.com/kkonst40/ichat/internal/service"
	"github.com/kkonst40/ichat/internal/ws"
)

type ChatHandler struct {
	chatService *service.ChatService
	validate    *validator.Validate
}

func NewChatHandler(newChatService *service.ChatService, validate *validator.Validate) *ChatHandler {
	return &ChatHandler{
		chatService: newChatService,
		validate:    validate,
	}
}

func (h *ChatHandler) GetChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest), log)
		return
	}

	chat, err := h.chatService.GetChat(ctx, chatID)
	if err != nil {
		WriteError(w, err, log)
		return
	}

	resp := dto.GetChatResponse{
		ID:   chat.ID,
		Name: chat.Name,
	}

	WriteJSON(w, http.StatusOK, resp, log)
}

func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := getUserID(ctx)
	log := logger.FromContext(ctx)

	chats, err := h.chatService.GetUserChats(ctx, requesterID)
	if err != nil {
		WriteError(w, err, log)
		return
	}

	log.Info("user chats retrieved")

	resp := dto.GetChatsResponse{
		Chats: make([]dto.GetChatResponse, 0, len(chats)),
	}

	for _, chat := range chats {
		resp.Chats = append(resp.Chats, dto.GetChatResponse{
			ID:   chat.ID,
			Name: chat.Name,
		})
	}

	WriteJSON(w, http.StatusOK, resp, log)
}

func (h *ChatHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := getUserID(ctx)
	log := logger.FromContext(ctx)

	var req dto.CreateChatRequest
	if err := bindJSON(r, &req, h.validate); err != nil {
		WriteError(w, err, log)
		return
	}

	chat, err := h.chatService.CreateChat(ctx, req.Name, req.UserIDs, requesterID)
	if err != nil {
		WriteError(w, err, log)
		return
	}

	log.Info("chat created", "chatID", chat.ID)
	location := fmt.Sprintf("/chats/%s", chat.ID.String())

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusCreated)
}

func (h *ChatHandler) UpdateChatName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := getUserID(ctx)
	log := logger.FromContext(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest), log)
		return
	}

	var req dto.UpdateChatNameRequest
	if err := bindJSON(r, &req, h.validate); err != nil {
		WriteError(w, handleValidationErr(err), log)
		return
	}

	err = h.chatService.UpdateChatName(ctx, chatID, req.Name, requesterID)
	if err != nil {
		WriteError(w, err, log)
		return
	}

	log.Info("chat name updated", "chatID", chatID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ChatHandler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requesterID := getUserID(ctx)
	log := logger.FromContext(ctx)

	chatID, err := uuid.Parse(r.PathValue("chatId"))
	if err != nil {
		WriteError(w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest), log)
		return
	}

	err = h.chatService.DeleteChat(ctx, chatID, requesterID)
	if err != nil {
		WriteError(w, err, log)
		return
	}

	log.Info("chat deleted", "chatID", chatID)

	w.WriteHeader(http.StatusNoContent)
}

func (h *ChatHandler) ConnectToChat(wsServer *ws.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requesterID := getUserID(ctx)
		log := logger.FromContext(ctx)

		chatID, err := uuid.Parse(r.PathValue("chatId"))
		if err != nil {
			WriteError(w, fmt.Errorf("%w: chat ID format", errs.ErrInvalidRequest), log)
			return
		}

		if !h.chatService.AllowedToConnect(ctx, chatID, requesterID) {
			WriteError(
				w,
				fmt.Errorf(
					"%w: user %v is not in chat %v",
					errs.ErrForbidden,
					requesterID,
					chatID,
				),
				log,
			)
			return
		}

		err = wsServer.Connect(w, r, requesterID, chatID)
		if err != nil {
			WriteError(w, err, log)
			return
		}

		log.Info("connected to chat", "chatID", chatID)
	}
}
