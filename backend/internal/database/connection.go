package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Config はデータベース接続設定を表す
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Charset  string
}

// DB はデータベース接続のラッパー
type DB struct {
	*sql.DB
}

// NewConnection は新しいデータベース接続を作成する
func NewConnection(config Config) (*DB, error) {
	// デフォルト値の設定
	if config.Charset == "" {
		config.Charset = "utf8mb4"
	}
	if config.Port == "" {
		config.Port = "3306"
	}

	// DSN（Data Source Name）の構築
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Charset,
	)

	// データベース接続の開始
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("データベース接続の開始に失敗しました: %w", err)
	}

	// コネクションプールの設定
	db.SetMaxOpenConns(25)                 // 最大オープン接続数
	db.SetMaxIdleConns(5)                  // 最大アイドル接続数
	db.SetConnMaxLifetime(5 * time.Minute) // 接続の最大生存時間

	// 接続テスト
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("データベースへの接続テストに失敗しました: %w", err)
	}

	log.Printf("データベースに正常に接続しました: %s:%s/%s", config.Host, config.Port, config.Database)

	return &DB{db}, nil
}

// Close はデータベース接続を閉じる
func (db *DB) Close() error {
	if db.DB != nil {
		log.Println("データベース接続を閉じています...")
		return db.DB.Close()
	}
	return nil
}

// Ping はデータベース接続の健全性をチェックする
func (db *DB) Ping() error {
	if db.DB == nil {
		return fmt.Errorf("データベース接続が初期化されていません")
	}
	return db.DB.Ping()
}

// BeginTx はトランザクションを開始する
func (db *DB) BeginTx() (*sql.Tx, error) {
	if db.DB == nil {
		return nil, fmt.Errorf("データベース接続が初期化されていません")
	}
	return db.DB.Begin()
}

// GetStats はデータベース接続の統計情報を取得する
func (db *DB) GetStats() sql.DBStats {
	if db.DB == nil {
		return sql.DBStats{}
	}
	return db.DB.Stats()
}