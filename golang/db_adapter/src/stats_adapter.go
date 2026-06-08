package db_adapter

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type GlobalStats struct {
	TotalUsers      int `db:"total_users"`
	TotalRooms      int `db:"total_rooms"`
	TotalCharacters int `db:"total_characters"`
}

type TopCharacter struct {
	CharUUID uuid.UUID `db:"char_uuid"`
	Name     string    `db:"name"`
	Level    int16     `db:"level"`
}

type TopRoom struct {
	RoomUUID    uuid.UUID `db:"room_uuid"`
	Name        string    `db:"name"`
	MemberCount int       `db:"member_count"`
}

type UserStats struct {
	CreatedAt        time.Time     `db:"created_at"`
	OwnedCharsCount  int           `db:"owned_chars_count"`
	OwnedRoomsCount  int           `db:"owned_rooms_count"`
	JoinedRoomsCount int           `db:"joined_rooms_count"`
	TopCharacter     *TopCharacter // Будет null в JSON, если nil
	TopRoom          *TopRoom      // Будет null в JSON, если nil
}

func (adapter *PostgresAdapter) GetGlobalStats() (GlobalStats, error) {
	var stats GlobalStats
	
	// Можно выполнить одним запросом для эффективности
	query := `
		SELECT 
			(SELECT COUNT(*) FROM users) AS total_users,
			(SELECT COUNT(*) FROM rooms) AS total_rooms,
			(SELECT COUNT(*) FROM characters) AS total_characters
	`
	err := adapter.db.QueryRow(query).Scan(&stats.TotalUsers, &stats.TotalRooms, &stats.TotalCharacters)
	if err != nil {
		return stats, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return stats, nil
}

func (adapter *PostgresAdapter) GetUserStats(userUUID uuid.UUID) (UserStats, error) {
	var stats UserStats

	// 1. Получаем дату создания и счетчики
	countsQuery := `
		SELECT 
			u.created_at,
			(SELECT COUNT(*) FROM characters WHERE owner_uuid = $1) AS owned_chars_count,
			(SELECT COUNT(*) FROM rooms WHERE owner_uuid = $1) AS owned_rooms_count,
			(SELECT COUNT(*) FROM rooms_users_relations WHERE member_uuid = $1) AS joined_rooms_count
		FROM users u WHERE u.uuid = $1
	`
	err := adapter.db.QueryRow(countsQuery, userUUID).Scan(
		&stats.CreatedAt, &stats.OwnedCharsCount, &stats.OwnedRoomsCount, &stats.JoinedRoomsCount,
	)
	if err != nil {
		return stats, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// 2. Получаем самого прокачанного персонажа (если есть)
	topCharQuery := `
		SELECT char_uuid, name, level 
		FROM characters 
		WHERE owner_uuid = $1 
		ORDER BY level DESC, created_at ASC 
		LIMIT 1
	`
	var topChar TopCharacter
	err = adapter.db.QueryRow(topCharQuery, userUUID).Scan(&topChar.CharUUID, &topChar.Name, &topChar.Level)
	if err == nil { // Если нашли, присваиваем указатель
		stats.TopCharacter = &topChar
	}

	// 3. Получаем комнату с наибольшим количеством участников (если есть)
	topRoomQuery := `
		SELECT r.uuid, r.name, COUNT(rur.member_uuid) AS member_count
		FROM rooms r
		LEFT JOIN rooms_users_relations rur ON r.uuid = rur.room_uuid
		WHERE r.owner_uuid = $1
		GROUP BY r.uuid, r.name
		ORDER BY member_count DESC, r.created_at ASC
		LIMIT 1
	`
	var topRoom TopRoom
	err = adapter.db.QueryRow(topRoomQuery, userUUID).Scan(&topRoom.RoomUUID, &topRoom.Name, &topRoom.MemberCount)
	if err == nil { // Если нашли, присваиваем указатель
		stats.TopRoom = &topRoom
	}

	return stats, nil
}
