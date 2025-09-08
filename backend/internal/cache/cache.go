// Package cache はRedis統合キャッシュシステムを提供する
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"backend/internal/config"
	"backend/internal/models"
)

// CacheManager はRedis統合キャッシュマネージャーのインターフェース
type CacheManager interface {
	// 基本操作
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	
	// トーナメント関連キャッシュ
	SetTournament(ctx context.Context, sport string, tournament *models.Tournament) error
	GetTournament(ctx context.Context, sport string) (*models.Tournament, error)
	DeleteTournament(ctx context.Context, sport string) error
	
	// 試合関連キャッシュ
	SetMatches(ctx context.Context, sport string, matches []models.Match) error
	GetMatches(ctx context.Context, sport string) ([]models.Match, error)
	DeleteMatches(ctx context.Context, sport string) error
	
	// ブラケット関連キャッシュ
	SetBracket(ctx context.Context, sport string, bracket *models.Bracket) error
	GetBracket(ctx context.Context, sport string) (*models.Bracket, error)
	DeleteBracket(ctx context.Context, sport string) error
	
	// 統計関連キャッシュ
	SetStatistics(ctx context.Context, key string, stats interface{}) error
	GetStatistics(ctx context.Context, key string, dest interface{}) error
	DeleteStatistics(ctx context.Context, key string) error
	
	// キャッシュ無効化
	InvalidateTournamentCache(ctx context.Context, sport string) error
	InvalidateAllCache(ctx context.Context) error
	
	// 接続管理
	Ping(ctx context.Context) error
	Close() error
}

// cacheManagerImpl はCacheManagerの実装
type cacheManagerImpl struct {
	client  *redis.Client
	enabled bool
}

// NewCacheManager は新しいキャッシュマネージャーインスタンスを作成する
func NewCacheManager(cfg *config.Config) CacheManager {
	if !cfg.Redis.Enabled {
		log.Println("Redis キャッシュが無効化されています")
		return &cacheManagerImpl{
			client:  nil,
			enabled: false,
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddress(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 接続テスト
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Redis接続に失敗しました: %v", err)
		log.Println("キャッシュ機能を無効化します")
		return &cacheManagerImpl{
			client:  nil,
			enabled: false,
		}
	}

	log.Printf("Redis接続が確立されました: %s", cfg.GetRedisAddress())
	return &cacheManagerImpl{
		client:  client,
		enabled: true,
	}
}

// キーの生成関数
func (c *cacheManagerImpl) generateKey(prefix, identifier string) string {
	return fmt.Sprintf("tournament:%s:%s", prefix, identifier)
}

// Set は指定されたキーに値を設定する
func (c *cacheManagerImpl) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !c.enabled {
		return nil // キャッシュが無効の場合は何もしない
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("値のシリアライズに失敗しました: %w", err)
	}

	if err := c.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("キャッシュの設定に失敗しました: %w", err)
	}

	log.Printf("キャッシュを設定しました: %s (有効期限: %v)", key, expiration)
	return nil
}

// Get は指定されたキーから値を取得する
func (c *cacheManagerImpl) Get(ctx context.Context, key string, dest interface{}) error {
	if !c.enabled {
		return redis.Nil // キャッシュが無効の場合はキャッシュミス扱い
	}

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("キャッシュミス: %s", key)
		} else {
			log.Printf("キャッシュ取得エラー: %s - %v", key, err)
		}
		return err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("値のデシリアライズに失敗しました: %w", err)
	}

	log.Printf("キャッシュヒット: %s", key)
	return nil
}

// Delete は指定されたキーを削除する
func (c *cacheManagerImpl) Delete(ctx context.Context, key string) error {
	if !c.enabled {
		return nil
	}

	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("キャッシュの削除に失敗しました: %w", err)
	}

	log.Printf("キャッシュを削除しました: %s", key)
	return nil
}

// Exists は指定されたキーが存在するかチェックする
func (c *cacheManagerImpl) Exists(ctx context.Context, key string) (bool, error) {
	if !c.enabled {
		return false, nil
	}

	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("キーの存在確認に失敗しました: %w", err)
	}

	return count > 0, nil
}

// SetTournament はトーナメントデータをキャッシュに設定する
func (c *cacheManagerImpl) SetTournament(ctx context.Context, sport string, tournament *models.Tournament) error {
	key := c.generateKey("tournament", sport)
	return c.Set(ctx, key, tournament, 5*time.Minute)
}

// GetTournament はトーナメントデータをキャッシュから取得する
func (c *cacheManagerImpl) GetTournament(ctx context.Context, sport string) (*models.Tournament, error) {
	key := c.generateKey("tournament", sport)
	var tournament models.Tournament
	
	if err := c.Get(ctx, key, &tournament); err != nil {
		return nil, err
	}
	
	return &tournament, nil
}

// DeleteTournament はトーナメントキャッシュを削除する
func (c *cacheManagerImpl) DeleteTournament(ctx context.Context, sport string) error {
	key := c.generateKey("tournament", sport)
	return c.Delete(ctx, key)
}

// SetMatches は試合データをキャッシュに設定する
func (c *cacheManagerImpl) SetMatches(ctx context.Context, sport string, matches []models.Match) error {
	key := c.generateKey("matches", sport)
	return c.Set(ctx, key, matches, 3*time.Minute)
}

// GetMatches は試合データをキャッシュから取得する
func (c *cacheManagerImpl) GetMatches(ctx context.Context, sport string) ([]models.Match, error) {
	key := c.generateKey("matches", sport)
	var matches []models.Match
	
	if err := c.Get(ctx, key, &matches); err != nil {
		return nil, err
	}
	
	return matches, nil
}

// DeleteMatches は試合キャッシュを削除する
func (c *cacheManagerImpl) DeleteMatches(ctx context.Context, sport string) error {
	key := c.generateKey("matches", sport)
	return c.Delete(ctx, key)
}

// SetBracket はブラケットデータをキャッシュに設定する
func (c *cacheManagerImpl) SetBracket(ctx context.Context, sport string, bracket *models.Bracket) error {
	key := c.generateKey("bracket", sport)
	return c.Set(ctx, key, bracket, 5*time.Minute)
}

// GetBracket はブラケットデータをキャッシュから取得する
func (c *cacheManagerImpl) GetBracket(ctx context.Context, sport string) (*models.Bracket, error) {
	key := c.generateKey("bracket", sport)
	var bracket models.Bracket
	
	if err := c.Get(ctx, key, &bracket); err != nil {
		return nil, err
	}
	
	return &bracket, nil
}

// DeleteBracket はブラケットキャッシュを削除する
func (c *cacheManagerImpl) DeleteBracket(ctx context.Context, sport string) error {
	key := c.generateKey("bracket", sport)
	return c.Delete(ctx, key)
}

// SetStatistics は統計データをキャッシュに設定する
func (c *cacheManagerImpl) SetStatistics(ctx context.Context, key string, stats interface{}) error {
	cacheKey := c.generateKey("stats", key)
	return c.Set(ctx, cacheKey, stats, 10*time.Minute)
}

// GetStatistics は統計データをキャッシュから取得する
func (c *cacheManagerImpl) GetStatistics(ctx context.Context, key string, dest interface{}) error {
	cacheKey := c.generateKey("stats", key)
	return c.Get(ctx, cacheKey, dest)
}

// DeleteStatistics は統計キャッシュを削除する
func (c *cacheManagerImpl) DeleteStatistics(ctx context.Context, key string) error {
	cacheKey := c.generateKey("stats", key)
	return c.Delete(ctx, cacheKey)
}

// InvalidateTournamentCache はスポーツ別のトーナメント関連キャッシュを無効化する
func (c *cacheManagerImpl) InvalidateTournamentCache(ctx context.Context, sport string) error {
	if !c.enabled {
		return nil
	}

	// 関連するキーを削除
	keys := []string{
		c.generateKey("tournament", sport),
		c.generateKey("matches", sport),
		c.generateKey("bracket", sport),
	}

	for _, key := range keys {
		if err := c.Delete(ctx, key); err != nil {
			log.Printf("キャッシュ無効化エラー: %s - %v", key, err)
		}
	}

	log.Printf("トーナメントキャッシュを無効化しました: %s", sport)
	return nil
}

// InvalidateAllCache は全てのキャッシュを無効化する
func (c *cacheManagerImpl) InvalidateAllCache(ctx context.Context) error {
	if !c.enabled {
		return nil
	}

	// tournament:* パターンのキーを全て削除
	keys, err := c.client.Keys(ctx, "tournament:*").Result()
	if err != nil {
		return fmt.Errorf("キー一覧の取得に失敗しました: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("キャッシュの一括削除に失敗しました: %w", err)
		}
	}

	log.Printf("全キャッシュを無効化しました (%d件)", len(keys))
	return nil
}

// Ping はRedis接続の健全性をチェックする
func (c *cacheManagerImpl) Ping(ctx context.Context) error {
	if !c.enabled {
		return nil
	}

	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis接続の確認に失敗しました: %w", err)
	}

	return nil
}

// Close はRedis接続を閉じる
func (c *cacheManagerImpl) Close() error {
	if !c.enabled || c.client == nil {
		return nil
	}

	if err := c.client.Close(); err != nil {
		return fmt.Errorf("Redis接続の終了に失敗しました: %w", err)
	}

	log.Println("Redis接続を終了しました")
	return nil
}