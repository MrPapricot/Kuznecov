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
// @Description Создает комнату. Создатель автоматически становится её владельцем и участником.
// @Tags Rooms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body object true "Данные комнаты" example({"name": "Gaming Room", "description": "For weekend sessions"})
// @Success 201 {object} object "UUID созданной комнаты" example({"room_uuid": "550e8400-..."})
// @Failure 400 {object} object "Ошибка валидации" example({"Error": "name is required"})
// @Failure 401 {object} object "Невалидный токен" example({"Error": "Invalid token format"})
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
// @Description Добавляет пользователей по их username. Доступно только владельцу комнаты.
// @Tags Rooms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID комнаты"
// @Param request body object true "Список username для добавления" example({"usernames": ["friend1", "friend2"]})
// @Success 200 {object} object "Список участников и ошибки (если есть)" example({"members": [{"uuid": "...", "username": "friend1", "joined_at": "..."}], "not_found": ["unknown"], "already_members": ["owner"]})
// @Failure 400 {object} object "Неверный формат UUID или тела запроса" example({"Error": "Invalid uuid format"})
// @Failure 401 {object} object "Невалидный токен или нет прав владельца" example({"Error": "Invalid token format"})
// @Failure 404 {object} object "Комната не найдена" example({"Error": "Room not found"})
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
// @Description Возвращает детали комнаты и список всех участников. Доступно любому участнику комнаты.
// @Tags Rooms
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID комнаты"
// @Success 200 {object} object "Полная информация о комнате" example({"uuid": "...", "name": "Gaming Room", "description": "...", "owner": "johndoe", "members": [...], "created_at": "..."})
// @Failure 400 {object} object "Неверный формат UUID" example({"Error": "Invalid uuid format"})
// @Failure 401 {object} object "Невалидный токен" example({"Error": "Invalid token format"})
// @Failure 403 {object} object "Пользователь не является участником комнаты" example({"Error": "User is not a member of the room"})
// @Failure 404 {object} object "Комната не найдена" example({"Error": "Room not found"})
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
// @Description Удаляет пользователей из комнаты. Доступно только владельцу. Владелец не может удалить сам себя.
// @Tags Rooms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID комнаты"
// @Param request body object true "Список username для удаления" example({"usernames": ["friend1", "johndoe"]})
// @Success 200 {object} object "Обновленный список участников и предупреждения" example({"members": [...], "not_found": ["unknown"], "owner_error": "user johndoe is owner and can not be excluded..."})
// @Failure 400 {object} object "Неверный формат запроса" example({"Error": "Invalid uuid format"})
// @Failure 401 {object} object "Невалидный токен или нет прав владельца" example({"Error": "Invalid token format"})
// @Failure 404 {object} object "Комната не найдена" example({"Error": "Room not found"})
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
// @Description Обновляет название комнаты. Доступно только владельцу.
// @Tags Rooms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID комнаты"
// @Param request body object true "Новое имя комнаты" example({"new_name": "Updated Gaming Room"})
// @Success 200 {object} object "Старое и новое имя" example({"old_name": "Gaming Room", "new_name": "Updated Gaming Room"})
// @Failure 400 {object} object "Неверный формат запроса" example({"Error": "new_name is required"})
// @Failure 401 {object} object "Невалидный токен или нет прав владельца" example({"Error": "Invalid token format"})
// @Failure 404 {object} object "Комната не найдена" example({"Error": "Room not found"})
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
// @Description Обновляет описание комнаты. Доступно только владельцу.
// @Tags Rooms
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param uuid path string true "UUID комнаты"
// @Param request body object true "Новое описание комнаты" example({"new_description": "Now with voice chat!"})
// @Success 200 {object} object "Старое и новое описание" example({"old_description": "...", "new_description": "Now with voice chat!"})
// @Failure 400 {object} object "Неверный формат запроса" example({"Error": "Invalid request body"})
// @Failure 401 {object} object "Невалидный токен или нет прав владельца" example({"Error": "Invalid token format"})
// @Failure 404 {object} object "Комната не найдена" example({"Error": "Room not found"})
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
// @Description Полностью удаляет комнату и все связи с участниками. Доступно только владельцу.
// @Tags Rooms
// @Security ApiKeyAuth
// @Param uuid path string true "UUID комнаты"
// @Success 200 {object} object "Сообщение об успехе" example({"message": "Room deleted successfully"})
// @Failure 400 {object} object "Неверный формат UUID" example({"Error": "Invalid uuid format"})
// @Failure 401 {object} object "Невалидный токен или нет прав владельца" example({"Error": "Invalid token format"})
// @Failure 404 {object} object "Комната не найдена" example({"Error": "Room not found"})
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
// @Summary Получить список комнат, где пользователь является владельцем
// @Description Возвращает массив комнат, созданных текущим пользователем
// @Tags Rooms
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} object "Список комнат" example({"rooms": [{"uuid": "...", "name": "My Room", "description": "...", "created_at": "..."}]})
// @Failure 401 {object} object "Невалидный токен" example({"Error": "Invalid token format"})
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
// @Summary Получить список комнат, где пользователь является участником
// @Description Возвращает массив всех комнат, в которых состоит текущий пользователь (включая те, где он владелец)
// @Tags Rooms
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} object "Список комнат" example({"rooms": [{"uuid": "...", "name": "My Room", "description": "...", "created_at": "..."}]})
// @Failure 401 {object} object "Невалидный токен" example({"Error": "Invalid token format"})
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
