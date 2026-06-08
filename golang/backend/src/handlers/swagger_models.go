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
