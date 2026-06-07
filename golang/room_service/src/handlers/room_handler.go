package handler

import (
	"errors"
	"fmt"
	"room_service/src/repository"
	"time"

	db_adapter "db_adapter/src"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	repository repository.RoomRepository
}

func NewHandler(repository repository.RoomRepository) Handler {
	return Handler{repository: repository}
}

func (h *Handler) CreateRoomHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}
	if req.Name == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "name is required"})
	}

	roomUUID, err := h.repository.CreateRoom(token, req.Name, req.Description)
	if err != nil {
		return h.handleError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"room_uuid": roomUUID.String()})
}

func (h *Handler) AddMembersHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	roomUUID, err := uuid.Parse(ctx.Params("uuid"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	var req struct {
		Usernames []string `json:"usernames"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}
	if len(req.Usernames) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "usernames list is required"})
	}

	members, notFound, alreadyMembers, err := h.repository.AddMembers(token, roomUUID, req.Usernames)
	if err != nil {
		return h.handleError(ctx, err)
	}

	response := fiber.Map{"members": formatMembers(members)}
	if len(notFound) > 0 {
		response["not_found"] = notFound
	}
	if len(alreadyMembers) > 0 {
		response["already_members"] = alreadyMembers
	}

	return ctx.JSON(response)
}

func (h *Handler) GetRoomInfoHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	roomUUID, err := uuid.Parse(ctx.Params("uuid"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	info, err := h.repository.GetRoomInfo(token, roomUUID)
	if err != nil {
		return h.handleError(ctx, err)
	}

	return ctx.JSON(fiber.Map{
		"uuid":        info.UUID.String(),
		"name":        info.Name,
		"description": info.Description,
		"owner":       info.OwnerName,
		"members":     formatMembers(info.Members),
		"created_at":  info.CreatedAt.UTC().Format(time.RFC3339),
	})
}

func (h *Handler) RemoveMembersHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	roomUUID, err := uuid.Parse(ctx.Params("uuid"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	var req struct {
		Usernames []string `json:"usernames"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}
	if len(req.Usernames) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "usernames list is required"})
	}

	members, notFound, ownerError, err := h.repository.RemoveMembers(token, roomUUID, req.Usernames)
	if err != nil {
		return h.handleError(ctx, err)
	}

	response := fiber.Map{"members": formatMembers(members)}
	if len(notFound) > 0 {
		response["not_found"] = notFound
	}
	if ownerError != "" {
		response["owner_error"] = ownerError
	}

	return ctx.JSON(response)
}

func (h *Handler) UpdateRoomNameHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	roomUUID, err := uuid.Parse(ctx.Params("uuid"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	var req struct {
		NewName string `json:"new_name"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}
	if req.NewName == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "new_name is required"})
	}

	oldName, newName, err := h.repository.UpdateRoomName(token, roomUUID, req.NewName)
	if err != nil {
		return h.handleError(ctx, err)
	}

	return ctx.JSON(fiber.Map{"old_name": oldName, "new_name": newName})
}

func (h *Handler) UpdateRoomDescriptionHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	roomUUID, err := uuid.Parse(ctx.Params("uuid"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	var req struct {
		NewDescription string `json:"new_description"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}

	oldDesc, newDesc, err := h.repository.UpdateRoomDescription(token, roomUUID, req.NewDescription)
	if err != nil {
		return h.handleError(ctx, err)
	}

	return ctx.JSON(fiber.Map{"old_description": oldDesc, "new_description": newDesc})
}

func (h *Handler) DeleteRoomHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	roomUUID, err := uuid.Parse(ctx.Params("uuid"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	err = h.repository.DeleteRoom(token, roomUUID)
	if err != nil {
		return h.handleError(ctx, err)
	}

	return ctx.JSON(fiber.Map{"message": "Room deleted successfully"})
}

func (h *Handler) GetOwnedRoomsHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	rooms, err := h.repository.GetOwnedRooms(token)
	if err != nil {
		return h.handleError(ctx, err)
	}

	return ctx.JSON(fiber.Map{"rooms": formatRooms(rooms)})
}

func (h *Handler) GetJoinedRoomsHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	rooms, err := h.repository.GetJoinedRooms(token)
	if err != nil {
		return h.handleError(ctx, err)
	}

	return ctx.JSON(fiber.Map{"rooms": formatRooms(rooms)})
}

func (h *Handler) handleError(ctx fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.InvalidToken):
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Invalid token format"})
	case errors.Is(err, repository.TokenExpired):
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "Token expired"})
	case errors.Is(err, repository.NoUserFound):
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No user found for provided token"})
	case errors.Is(err, repository.RoomNotFound):
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"Error": "Room not found"})
	case errors.Is(err, repository.NotRoomMember):
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"Error": "User is not a member of the room"})
	case errors.Is(err, repository.DatabaseError):
		fmt.Printf("Database error: %#v\n", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
	default:
		fmt.Printf("Base error: %#v\n", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"Error": "Internal server error"})
	}
}

func formatMembers(members []db_adapter.MemberInfo) []fiber.Map {
	result := []fiber.Map{}
	for _, m := range members {
		result = append(result, fiber.Map{
			"uuid":      m.UUID.String(),
			"username":  m.Username,
			"joined_at": m.JoinedAt.UTC().Format(time.RFC3339),
		})
	}
	return result
}

func formatRooms(rooms []db_adapter.Room) []fiber.Map {
	result := []fiber.Map{}
	for _, r := range rooms {
		result = append(result, fiber.Map{
			"uuid":        r.UUID.String(),
			"name":        r.Name,
			"description": r.Description.String,
			"created_at":  r.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return result
}
