package handlers

// ==========================================
// AUTH MODELS
// ==========================================

// RegisterRequest модель запроса регистрации
type RegisterRequest struct {
	Username string `json:"username" example:"vault_dweller"`
	Email    string `json:"email" example:"dweller@vault-tec.com"`
	Password string `json:"password" example:"SecurePassword123!"`
}

// LoginRequest модель запроса входа
type LoginRequest struct {
	Email    string `json:"email" example:"dweller@vault-tec.com"`
	Password string `json:"password" example:"SecurePassword123!"`
}

// ==========================================
// USER MODELS
// ==========================================

// ChangeUsernameRequest модель запроса смены имени
type ChangeUsernameRequest struct {
	NewUsername string `json:"new_username" example:"new_cool_name"`
}

// ChangePasswordRequest модель запроса смены пароля
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" example:"SecurePassword123!"`
	NewPassword string `json:"new_password" example:"EvenMoreSecure456!"`
}

// DeleteUserRequest модель запроса удаления пользователя
type DeleteUserRequest struct {
	Password string `json:"password" example:"SecurePassword123!"`
}

// ==========================================
// ROOM MODELS
// ==========================================

// CreateRoomRequest модель запроса создания комнаты
type CreateRoomRequest struct {
	Name        string `json:"name" example:"Wasteland Trading Post"`
	Description string `json:"description" example:"A safe place to trade caps."`
}

// RoomMembersRequest модель запроса для добавления/удаления участников
type RoomMembersRequest struct {
	Usernames []string `json:"usernames" example:"trader_joe,ghoul_steve"`
}

// UpdateRoomNameRequest модель запроса смены имени комнаты
type UpdateRoomNameRequest struct {
	NewName string `json:"new_name" example:"Upgraded Trading Post"`
}

// UpdateRoomDescriptionRequest модель запроса смены описания комнаты
type UpdateRoomDescriptionRequest struct {
	NewDescription string `json:"new_description" example:"Now with 20% more radiation shielding!"`
}

// ==========================================
// CHARACTER MODELS
// ==========================================

// CreateCharacterRequest модель запроса создания персонажа
// Используем *int, чтобы Swagger корректно отображал опциональные числовые поля
type CreateCharacterRequest struct {
	Name         string `json:"name" example:"Lone Wanderer"`
	Description  string `json:"description" example:"A survivor from Vault 101."`
	Level        *int   `json:"level" example:"1"`
	Strength     *int   `json:"strength" example:"5"`
	Perception   *int   `json:"perception" example:"6"`
	Endurance    *int   `json:"endurance" example:"4"`
	Charisma     *int   `json:"charisma" example:"7"`
	Intelligence *int   `json:"intelligence" example:"8"`
	Agility      *int   `json:"agility" example:"5"`
	Luck         *int   `json:"luck" example:"5"`
}

// UpdateCharacterNameRequest модель запроса смены имени персонажа
type UpdateCharacterNameRequest struct {
	NewName string `json:"new_name" example:"Hero of the Wasteland"`
}

// UpdateCharacterDescriptionRequest модель запроса смены описания персонажа
type UpdateCharacterDescriptionRequest struct {
	NewDescription string `json:"new_description" example:"Now a legendary vault dweller."`
}

// LevelUpCharacterRequest модель запроса прокачки персонажа
type LevelUpCharacterRequest struct {
	Stat string `json:"stat" example:"strength"`
}

// ==========================================
// ОБЩИЕ МОДЕЛИ
// ==========================================

// ErrorResponse стандартная модель ошибки
type ErrorResponse struct {
	Error string `json:"Error" example:"Invalid request body"`
}

// MessageResponse стандартная модель успешного сообщения
type MessageResponse struct {
	Message string `json:"message" example:"Operation successful"`
}

// ==========================================
// AUTH RESPONSES
// ==========================================

// AuthResponse модель ответа при регистрации/входе
type AuthResponse struct {
	Token     string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt int64  `json:"expires_at" example:"1735689600"`
	UserID    string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ==========================================
// USER RESPONSES
// ==========================================

// UserInfoResponse модель ответа с информацией о пользователе
type UserInfoResponse struct {
	Username  string `json:"username" example:"vault_dweller"`
	Email     string `json:"email" example:"dweller@vault-tec.com"`
	CreatedAt string `json:"created_at" example:"2024-06-01T12:00:00Z"`
}

// ChangeUsernameResponse модель ответа при смене имени
type ChangeUsernameResponse struct {
	OldUsername string `json:"old_username" example:"vault_dweller"`
	NewUsername string `json:"new_username" example:"lone_wanderer"`
}

// ==========================================
// ROOM RESPONSES
// ==========================================

// CreateRoomResponse модель ответа при создании комнаты
type CreateRoomResponse struct {
	RoomUUID string `json:"room_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// RoomMember модель участника комнаты
type RoomMember struct {
	UUID     string `json:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username string `json:"username" example:"trader_joe"`
	JoinedAt string `json:"joined_at" example:"2024-06-01T12:00:00Z"`
}

// RoomMembersResponse модель ответа при добавлении/удалении участников
type RoomMembersResponse struct {
	Members        []RoomMember `json:"members"`
	NotFound       []string     `json:"not_found,omitempty" example:"unknown_user"`
	AlreadyMembers []string     `json:"already_members,omitempty" example:"existing_user"`
	OwnerError     string       `json:"owner_error,omitempty" example:"user owner is owner and can not be excluded..."`
}

// RoomInfoResponse полная информация о комнате
type RoomInfoResponse struct {
	UUID        string       `json:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string       `json:"name" example:"Trading Post"`
	Description string       `json:"description" example:"A safe place"`
	Owner       string       `json:"owner" example:"vault_dweller"`
	Members     []RoomMember `json:"members"`
	CreatedAt   string       `json:"created_at" example:"2024-06-01T12:00:00Z"`
}

// UpdateRoomNameResponse модель ответа при смене имени комнаты
type UpdateRoomNameResponse struct {
	OldName string `json:"old_name" example:"Old Name"`
	NewName string `json:"new_name" example:"New Name"`
}

// UpdateRoomDescriptionResponse модель ответа при смене описания комнаты
type UpdateRoomDescriptionResponse struct {
	OldDescription string `json:"old_description" example:"Old desc"`
	NewDescription string `json:"new_description" example:"New desc"`
}

// RoomListItem модель комнаты в списке
type RoomListItem struct {
	UUID        string `json:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string `json:"name" example:"My Room"`
	Description string `json:"description" example:"Room desc"`
	CreatedAt   string `json:"created_at" example:"2024-06-01T12:00:00Z"`
}

// RoomsListResponse список комнат
type RoomsListResponse struct {
	Rooms []RoomListItem `json:"rooms"`
}

// ==========================================
// CHARACTER RESPONSES
// ==========================================

// CreateCharacterResponse модель ответа при создании персонажа
type CreateCharacterResponse struct {
	CharUUID string `json:"char_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// CharacterInfoResponse полная информация о персонаже
type CharacterInfoResponse struct {
	CharUUID     string `json:"char_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name         string `json:"name" example:"Lone Wanderer"`
	Description  string `json:"description" example:"A brave survivor"`
	Level        int    `json:"level" example:"5"`
	Strength     int    `json:"strength" example:"5"`
	Perception   int    `json:"perception" example:"6"`
	Endurance    int    `json:"endurance" example:"4"`
	Charisma     int    `json:"charisma" example:"7"`
	Intelligence int    `json:"intelligence" example:"8"`
	Agility      int    `json:"agility" example:"5"`
	Luck         int    `json:"luck" example:"5"`
	CreatedAt    string `json:"created_at" example:"2024-06-01T12:00:00Z"`
}

// UpdateCharacterNameResponse модель ответа при смене имени персонажа
type UpdateCharacterNameResponse struct {
	OldName string `json:"old_name" example:"Old Name"`
	NewName string `json:"new_name" example:"New Name"`
}

// UpdateCharacterDescriptionResponse модель ответа при смене описания персонажа
type UpdateCharacterDescriptionResponse struct {
	OldDescription string `json:"old_description" example:"Old desc"`
	NewDescription string `json:"new_description" example:"New desc"`
}

// ==========================================
// STATS RESPONSES
// ==========================================

// GlobalStatsResponse глобальная статистика
type GlobalStatsResponse struct {
	TotalUsers      int `json:"total_users" example:"150"`
	TotalRooms      int `json:"total_rooms" example:"42"`
	TotalCharacters int `json:"total_characters" example:"310"`
}

// TopCharacterStats лучший персонаж пользователя
type TopCharacterStats struct {
	CharUUID string `json:"char_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name     string `json:"name" example:"Bruiser"`
	Level    int    `json:"level" example:"10"`
}

// TopRoomStats самая популярная комната пользователя
type TopRoomStats struct {
	RoomUUID    string `json:"room_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string `json:"name" example:"Main Hub"`
	MemberCount int    `json:"member_count" example:"15"`
}

// UserStatsResponse статистика пользователя
// Поля TopCharacter и TopRoom являются указателями, поэтому в JSON они будут null, если данных нет
type UserStatsResponse struct {
	CreatedAt        string             `json:"created_at" example:"2024-06-01T12:00:00Z"`
	OwnedCharsCount  int                `json:"owned_chars_count" example:"2"`
	OwnedRoomsCount  int                `json:"owned_rooms_count" example:"1"`
	JoinedRoomsCount int                `json:"joined_rooms_count" example:"3"`
	TopCharacter     *TopCharacterStats `json:"top_character"`
	TopRoom          *TopRoomStats      `json:"top_room"`
}
