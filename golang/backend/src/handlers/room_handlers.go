package handlers

import (
	client "MainBackend/src/room_client"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type RoomHandlers struct {
	roomServiceClient *client.RoomServiceClient
}

func NewRoomHandlers(roomServiceClient *client.RoomServiceClient) *RoomHandlers {
	return &RoomHandlers{
		roomServiceClient: roomServiceClient,
	}
}

// CreateRoomHandler godoc
// @Summary Создать новую комнату
// @Tags Rooms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body handlers.CreateRoomRequest true "Данные комнаты"
// @Success 201 {object} handlers.CreateRoomResponse "UUID созданной комнаты"
// @Router /api/room/create [post]
func (h *RoomHandlers) CreateRoomHandler(ctx fiber.Ctx) error {
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

	resp, err := h.roomServiceClient.CreateRoom(token, req)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Room service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// AddMembersHandler godoc
// @Summary Добавить участников в комнату
// @Tags Rooms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID комнаты"
// @Param request body handlers.RoomMembersRequest true "Список username"
// @Success 200 {object} handlers.RoomMembersResponse "Результат операции"
// @Failure 404 {object} handlers.ErrorResponse "Комната не найдена"
// @Router /api/room/add-members/{uuid} [post]
func (h *RoomHandlers) AddMembersHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	roomUUID := ctx.Params("uuid")
	if _, err := uuid.Parse(roomUUID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	var req struct {
		Usernames []string `json:"usernames"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}

	resp, err := h.roomServiceClient.AddMembers(token, roomUUID, req)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Room service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// GetRoomInfoHandler godoc
// @Summary Получить информацию о комнате
// @Tags Rooms
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID комнаты"
// @Success 200 {object} handlers.RoomInfoResponse "Полная информация о комнате"
// @Failure 403 {object} handlers.ErrorResponse "Нет доступа"
// @Router /api/room/info/{uuid} [get]
func (h *RoomHandlers) GetRoomInfoHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	roomUUID := ctx.Params("uuid")
	if _, err := uuid.Parse(roomUUID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	resp, err := h.roomServiceClient.GetRoomInfo(token, roomUUID)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Room service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// RemoveMembersHandler godoc
// @Summary Исключить участников из комнаты
// @Tags Rooms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID комнаты"
// @Param request body handlers.RoomMembersRequest true "Список username для удаления"
// @Success 200 {object} handlers.RoomMembersResponse "Обновленный список и предупреждения"
// @Router /api/room/remove-members/{uuid} [post]
func (h *RoomHandlers) RemoveMembersHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	roomUUID := ctx.Params("uuid")
	if _, err := uuid.Parse(roomUUID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	var req struct {
		Usernames []string `json:"usernames"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}

	resp, err := h.roomServiceClient.RemoveMembers(token, roomUUID, req)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Room service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// UpdateRoomNameHandler godoc
// @Summary Изменить имя комнаты
// @Tags Rooms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID комнаты"
// @Param request body handlers.UpdateRoomNameRequest true "Новое имя"
// @Success 200 {object} handlers.UpdateRoomNameResponse "Старое и новое имя"
// @Router /api/room/update-name/{uuid} [patch]
func (h *RoomHandlers) UpdateRoomNameHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	roomUUID := ctx.Params("uuid")
	if _, err := uuid.Parse(roomUUID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	var req struct {
		NewName string `json:"new_name"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}

	resp, err := h.roomServiceClient.UpdateRoomName(token, roomUUID, req)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Room service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// UpdateRoomDescriptionHandler godoc
// @Summary Изменить описание комнаты
// @Tags Rooms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID комнаты"
// @Param request body handlers.UpdateRoomDescriptionRequest true "Новое описание"
// @Success 200 {object} handlers.UpdateRoomDescriptionResponse "Старое и новое описание"
// @Router /api/room/update-description/{uuid} [patch]
func (h *RoomHandlers) UpdateRoomDescriptionHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	roomUUID := ctx.Params("uuid")
	if _, err := uuid.Parse(roomUUID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	var req struct {
		NewDescription string `json:"new_description"`
	}
	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid request body"})
	}

	resp, err := h.roomServiceClient.UpdateRoomDescription(token, roomUUID, req)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Room service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// DeleteRoomHandler godoc
// @Summary Удалить комнату
// @Tags Rooms
// @Security ApiKeyAuth
// @Param uuid path string true "UUID комнаты"
// @Success 200 {object} handlers.MessageResponse "Сообщение об успехе"
// @Router /api/room/{uuid} [delete]
func (h *RoomHandlers) DeleteRoomHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	roomUUID := ctx.Params("uuid")
	if _, err := uuid.Parse(roomUUID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"Error": "Invalid uuid format"})
	}

	resp, err := h.roomServiceClient.DeleteRoom(token, roomUUID)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Room service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// GetOwnedRoomsHandler godoc
// @Summary Получить список комнат владельца
// @Tags Rooms
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handlers.RoomsListResponse "Список комнат"
// @Router /api/room/owned [get]
func (h *RoomHandlers) GetOwnedRoomsHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	resp, err := h.roomServiceClient.GetOwnedRooms(token)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Room service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}

// GetJoinedRoomsHandler godoc
// @Summary Получить список комнат участника
// @Tags Rooms
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handlers.RoomsListResponse "Список комнат"
// @Router /api/room/joined [get]
func (h *RoomHandlers) GetJoinedRoomsHandler(ctx fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"Error": "No Authorization provided"})
	}

	resp, err := h.roomServiceClient.GetJoinedRooms(token)
	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"Error": "Room service unavailable"})
	}
	return ctx.Status(resp.StatusCode).Send(resp.Body)
}
