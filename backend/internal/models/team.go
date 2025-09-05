package models

import (
	"time"
)

// Team represents a team in the tournament system
type Team struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null;uniqueIndex"`
	Description  string    `json:"description"`
	TournamentID uint      `json:"tournament_id" gorm:"not null;index"`
	Tournament   Tournament `json:"tournament" gorm:"foreignKey:TournamentID"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}