package db_adapter

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Character struct {
	CharUUID             uuid.UUID
	Name                 string
	CharacterDescription sql.NullString
	OwnerUUID            uuid.UUID
	Strength             int16
	Perception           int16
	Endurance            int16
	Charisma             int16
	Intelligence         int16
	Agility              int16
	Luck                 int16
	Level                int16
	CreatedAt            time.Time
}

type CreateCharacterBody struct {
	Name                 string
	CharacterDescription string
	OwnerUUID            uuid.UUID // <-- Это поле отсутствовало
	Level                int16
	Strength             int16
	Perception           int16
	Endurance            int16
	Charisma             int16
	Intelligence         int16
	Agility              int16
	Luck                 int16
}

func (adapter *PostgresAdapter) CreateCharacter(body CreateCharacterBody) (uuid.UUID, error) {
	var charUUID uuid.UUID
	err := adapter.db.QueryRow(`
		INSERT INTO characters (name, character_description, owner_uuid, level, strength, perception, endurance, charisma, intelligence, agility, luck)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING char_uuid
	`, body.Name, nullString(body.CharacterDescription), body.OwnerUUID, body.Level,
		body.Strength, body.Perception, body.Endurance, body.Charisma, body.Intelligence, body.Agility, body.Luck).Scan(&charUUID)

	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return charUUID, nil
}

func (adapter *PostgresAdapter) GetCharacterByUUID(charUUID uuid.UUID) (Character, error) {
	var char Character
	err := adapter.db.QueryRow(`
		SELECT char_uuid, name, character_description, owner_uuid, strength, perception, endurance, charisma, intelligence, agility, luck, level, created_at
		FROM characters WHERE char_uuid = $1
	`, charUUID).Scan(
		&char.CharUUID, &char.Name, &char.CharacterDescription, &char.OwnerUUID,
		&char.Strength, &char.Perception, &char.Endurance, &char.Charisma, &char.Intelligence, &char.Agility, &char.Luck,
		&char.Level, &char.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Character{}, ErrCharacterNotFound
		}
		return Character{}, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return char, nil
}

func (adapter *PostgresAdapter) CheckUserIsCharacterOwner(charUUID, userUUID uuid.UUID) (bool, error) {
	var isOwner bool
	err := adapter.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM characters WHERE char_uuid = $1 AND owner_uuid = $2)`, charUUID, userUUID).Scan(&isOwner)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return isOwner, nil
}

func (adapter *PostgresAdapter) UpdateCharacterName(charUUID uuid.UUID, newName string) (string, error) {
	var oldName string

	// 1. Сначала получаем текущее (старое) имя
	err := adapter.db.QueryRow(`SELECT name FROM characters WHERE char_uuid = $1`, charUUID).Scan(&oldName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrCharacterNotFound
		}
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// 2. Затем выполняем обновление
	_, err = adapter.db.Exec(`UPDATE characters SET name = $1 WHERE char_uuid = $2`, newName, charUUID)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return oldName, nil
}

func (adapter *PostgresAdapter) UpdateCharacterDescription(charUUID uuid.UUID, newDesc string) (string, error) {
	var oldDesc sql.NullString

	// 1. Сначала получаем текущее (старое) описание
	err := adapter.db.QueryRow(`SELECT character_description FROM characters WHERE char_uuid = $1`, charUUID).Scan(&oldDesc)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrCharacterNotFound
		}
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// 2. Затем выполняем обновление
	_, err = adapter.db.Exec(`UPDATE characters SET character_description = $1 WHERE char_uuid = $2`, nullString(newDesc), charUUID)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return oldDesc.String, nil
}

func (adapter *PostgresAdapter) DeleteCharacter(charUUID uuid.UUID) error {
	res, err := adapter.db.Exec(`DELETE FROM characters WHERE char_uuid = $1`, charUUID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrCharacterNotFound
	}
	return nil
}

// LevelUpStat повышает уровень персонажа и выбранную характеристику на 1
func (adapter *PostgresAdapter) LevelUpStat(charUUID uuid.UUID, statName string) (int16, int16, error) {
	var newLevel, newStat int16
	var query string

	switch statName {
	case "strength":
		query = `UPDATE characters SET level = level + 1, strength = strength + 1 WHERE char_uuid = $1 AND level < 21 AND strength < 10 RETURNING level, strength`
	case "perception":
		query = `UPDATE characters SET level = level + 1, perception = perception + 1 WHERE char_uuid = $1 AND level < 21 AND perception < 10 RETURNING level, perception`
	case "endurance":
		query = `UPDATE characters SET level = level + 1, endurance = endurance + 1 WHERE char_uuid = $1 AND level < 21 AND endurance < 10 RETURNING level, endurance`
	case "charisma":
		query = `UPDATE characters SET level = level + 1, charisma = charisma + 1 WHERE char_uuid = $1 AND level < 21 AND charisma < 10 RETURNING level, charisma`
	case "intelligence":
		query = `UPDATE characters SET level = level + 1, intelligence = intelligence + 1 WHERE char_uuid = $1 AND level < 21 AND intelligence < 10 RETURNING level, intelligence`
	case "agility":
		query = `UPDATE characters SET level = level + 1, agility = agility + 1 WHERE char_uuid = $1 AND level < 21 AND agility < 10 RETURNING level, agility`
	case "luck":
		query = `UPDATE characters SET level = level + 1, luck = luck + 1 WHERE char_uuid = $1 AND level < 21 AND luck < 10 RETURNING level, luck`
	default:
		return 0, 0, ErrInvalidStatName
	}

	err := adapter.db.QueryRow(query, charUUID).Scan(&newLevel, &newStat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Либо персонаж не найден, либо уровень уже 21, либо стат уже 10
			return 0, 0, ErrStatMaxedOrLevelMaxed
		}
		return 0, 0, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return newLevel, newStat, nil
}
