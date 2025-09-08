package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"backend/internal/config"
	"backend/internal/models"
)

// MockCacheManager はテスト用のモックキャッシュマネージャー
type MockCacheManager struct {
	mock.Mock
}

func (m *MockCacheManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCacheManager) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheManager) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheManager) SetTournament(ctx context.Context, sport string, tournament *models.Tournament) error {
	args := m.Called(ctx, sport, tournament)
	return args.Error(0)
}

func (m *MockCacheManager) GetTournament(ctx context.Context, sport string) (*models.Tournament, error) {
	args := m.Called(ctx, sport)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tournament), args.Error(1)
}

func (m *MockCacheManager) DeleteTournament(ctx context.Context, sport string) error {
	args := m.Called(ctx, sport)
	return args.Error(0)
}

func (m *MockCacheManager) SetMatches(ctx context.Context, sport string, matches []models.Match) error {
	args := m.Called(ctx, sport, matches)
	return args.Error(0)
}

func (m *MockCacheManager) GetMatches(ctx context.Context, sport string) ([]models.Match, error) {
	args := m.Called(ctx, sport)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Match), args.Error(1)
}

func (m *MockCacheManager) DeleteMatches(ctx context.Context, sport string) error {
	args := m.Called(ctx, sport)
	return args.Error(0)
}

func (m *MockCacheManager) SetBracket(ctx context.Context, sport string, bracket *models.Bracket) error {
	args := m.Called(ctx, sport, bracket)
	return args.Error(0)
}

func (m *MockCacheManager) GetBracket(ctx context.Context, sport string) (*models.Bracket, error) {
	args := m.Called(ctx, sport)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bracket), args.Error(1)
}

func (m *MockCacheManager) DeleteBracket(ctx context.Context, sport string) error {
	args := m.Called(ctx, sport)
	return args.Error(0)
}

func (m *MockCacheManager) SetStatistics(ctx context.Context, key string, stats interface{}) error {
	args := m.Called(ctx, key, stats)
	return args.Error(0)
}

func (m *MockCacheManager) GetStatistics(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCacheManager) DeleteStatistics(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheManager) InvalidateTournamentCache(ctx context.Context, sport string) error {
	args := m.Called(ctx, sport)
	return args.Error(0)
}

func (m *MockCacheManager) InvalidateAllCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCacheManager) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCacheManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestNewCacheManager はキャッシュマネージャーの作成をテストする
func TestNewCacheManager(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.Config
		expected bool // enabled状態の期待値
	}{
		{
			name: "Redis有効設定",
			config: &config.Config{
				Redis: config.RedisConfig{
					Host:    "localhost",
					Port:    6379,
					Enabled: true,
				},
			},
			expected: false, // 実際のRedis接続がないためfalse
		},
		{
			name: "Redis無効設定",
			config: &config.Config{
				Redis: config.RedisConfig{
					Enabled: false,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewCacheManager(tt.config)
			assert.NotNil(t, cache)
			
			// 無効化されたキャッシュでも基本操作は動作する
			ctx := context.Background()
			err := cache.Set(ctx, "test", "value", time.Minute)
			assert.NoError(t, err)
		})
	}
}

// TestCacheManagerOperations はキャッシュマネージャーの基本操作をテストする
func TestCacheManagerOperations(t *testing.T) {
	// 無効化されたキャッシュマネージャーでテスト
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Enabled: false,
		},
	}
	cache := NewCacheManager(cfg)
	ctx := context.Background()

	t.Run("Set操作", func(t *testing.T) {
		err := cache.Set(ctx, "test_key", "test_value", time.Minute)
		assert.NoError(t, err)
	})

	t.Run("Get操作", func(t *testing.T) {
		var result string
		err := cache.Get(ctx, "test_key", &result)
		// 無効化されたキャッシュではキャッシュミス
		assert.Error(t, err)
	})

	t.Run("Delete操作", func(t *testing.T) {
		err := cache.Delete(ctx, "test_key")
		assert.NoError(t, err)
	})

	t.Run("Exists操作", func(t *testing.T) {
		exists, err := cache.Exists(ctx, "test_key")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestTournamentCacheOperations はトーナメントキャッシュ操作をテストする
func TestTournamentCacheOperations(t *testing.T) {
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Enabled: false,
		},
	}
	cache := NewCacheManager(cfg)
	ctx := context.Background()

	tournament := &models.Tournament{
		ID:     1,
		Sport:  "volleyball",
		Format: "standard",
		Status: "active",
	}

	t.Run("SetTournament", func(t *testing.T) {
		err := cache.SetTournament(ctx, "volleyball", tournament)
		assert.NoError(t, err)
	})

	t.Run("GetTournament", func(t *testing.T) {
		result, err := cache.GetTournament(ctx, "volleyball")
		assert.Error(t, err) // 無効化されたキャッシュではエラー
		assert.Nil(t, result)
	})

	t.Run("DeleteTournament", func(t *testing.T) {
		err := cache.DeleteTournament(ctx, "volleyball")
		assert.NoError(t, err)
	})
}

// TestMatchCacheOperations は試合キャッシュ操作をテストする
func TestMatchCacheOperations(t *testing.T) {
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Enabled: false,
		},
	}
	cache := NewCacheManager(cfg)
	ctx := context.Background()

	matches := []models.Match{
		{
			ID:           1,
			TournamentID: 1,
			Round:        "1st_round",
			Team1:        "チームA",
			Team2:        "チームB",
			Status:       "pending",
		},
	}

	t.Run("SetMatches", func(t *testing.T) {
		err := cache.SetMatches(ctx, "volleyball", matches)
		assert.NoError(t, err)
	})

	t.Run("GetMatches", func(t *testing.T) {
		result, err := cache.GetMatches(ctx, "volleyball")
		assert.Error(t, err) // 無効化されたキャッシュではエラー
		assert.Nil(t, result)
	})

	t.Run("DeleteMatches", func(t *testing.T) {
		err := cache.DeleteMatches(ctx, "volleyball")
		assert.NoError(t, err)
	})
}

// TestCacheInvalidation はキャッシュ無効化をテストする
func TestCacheInvalidation(t *testing.T) {
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Enabled: false,
		},
	}
	cache := NewCacheManager(cfg)
	ctx := context.Background()

	t.Run("InvalidateTournamentCache", func(t *testing.T) {
		err := cache.InvalidateTournamentCache(ctx, "volleyball")
		assert.NoError(t, err)
	})

	t.Run("InvalidateAllCache", func(t *testing.T) {
		err := cache.InvalidateAllCache(ctx)
		assert.NoError(t, err)
	})
}

// TestCacheConnection は接続管理をテストする
func TestCacheConnection(t *testing.T) {
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Enabled: false,
		},
	}
	cache := NewCacheManager(cfg)
	ctx := context.Background()

	t.Run("Ping", func(t *testing.T) {
		err := cache.Ping(ctx)
		assert.NoError(t, err) // 無効化されたキャッシュでは常に成功
	})

	t.Run("Close", func(t *testing.T) {
		err := cache.Close()
		assert.NoError(t, err)
	})
}