package service

import (
	"fmt"
	"strconv"

	"github.com/saku0512/GYOUJI_HP/backend/internal/repository"
)

// SettingService は設定に関するビジネスロジックのインターフェースです。
type SettingService interface {
	GetVisibility() (map[string]bool, error)
	UpdateVisibility(key string, value bool) error
	GetWeather() (string, error)
	UpdateWeather(value string) error
}

// settingService は SettingService の実装です。
type settingService struct {
	repo repository.SettingRepository
}

// NewSettingService は settingService の新しいインスタンスを生成します。
func NewSettingService(repo repository.SettingRepository) SettingService {
	return &settingService{repo: repo}
}

// GetVisibility はスコア表示設定を取得します。
func (s *settingService) GetVisibility() (map[string]bool, error) {
	settings := make(map[string]bool)

	keys := []string{"showTotalScores", "showQuestionnaireButton"}

	for _, key := range keys {
		value, err := s.repo.GetSettingVisibility(key)
		if err != nil {
			// キーが見つからない場合、デフォルト値として false を使用
			settings[key] = false
		} else {
			boolValue, err := strconv.ParseBool(value)
			if err != nil {
				// パースに失敗した場合もデフォルト値として false を使用
				settings[key] = false
			} else {
				settings[key] = boolValue
			}
		}
	}

	return settings, nil
}

// UpdateVisibility はスコア表示設定を更新します。
func (s *settingService) UpdateVisibility(key string, value bool) error {
	return s.repo.UpdateSettingVisibility(key, strconv.FormatBool(value))
}

// GetWeather は天候設定を取得します。
func (s *settingService) GetWeather() (string, error) {
	return s.repo.GetSettingWeather("tableTennisWeather")
}

// UpdateWeather は天候設定を更新します。
func (s *settingService) UpdateWeather(value string) error {
	// 簡単なバリデーション
	if value != "sunny" && value != "rainy" {
		return fmt.Errorf("invalid weather value: %s", value)
	}
	return s.repo.UpdateSettingWeather("tableTennisWeather", value)
}
