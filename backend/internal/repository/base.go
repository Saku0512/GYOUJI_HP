// Package repository はデータアクセス層の実装を提供する
package repository

import (
	"database/sql"
	"log"
	"time"

	"backend/internal/database"
)

// BaseRepository は共通のデータベース操作を提供するベースリポジトリインターフェース
type BaseRepository interface {
	// データベース接続管理
	GetDB() *database.DB
	Ping() error
	Close() error
	
	// トランザクション管理
	BeginTx() (*sql.Tx, error)
	CommitTx(tx *sql.Tx) error
	RollbackTx(tx *sql.Tx) error
	
	// 共通クエリ操作
	ExecQuery(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	
	// トランザクション内でのクエリ操作
	ExecQueryTx(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error)
	QueryRowTx(tx *sql.Tx, query string, args ...interface{}) *sql.Row
	QueryTx(tx *sql.Tx, query string, args ...interface{}) (*sql.Rows, error)
}

// baseRepositoryImpl はBaseRepositoryの実装
type baseRepositoryImpl struct {
	db *database.DB
}

// NewBaseRepository は新しいベースリポジトリインスタンスを作成する
func NewBaseRepository(db *database.DB) BaseRepository {
	if db == nil {
		log.Fatal("データベース接続がnilです")
	}
	return &baseRepositoryImpl{
		db: db,
	}
}

// GetDB はデータベース接続を取得する
func (r *baseRepositoryImpl) GetDB() *database.DB {
	return r.db
}

// Ping はデータベース接続の健全性をチェックする
func (r *baseRepositoryImpl) Ping() error {
	if r.db == nil {
		return NewRepositoryError(ErrTypeConnection, "データベース接続が初期化されていません", nil)
	}
	
	if err := r.db.Ping(); err != nil {
		return NewRepositoryError(ErrTypeConnection, "データベース接続の確認に失敗しました", err)
	}
	
	return nil
}

// Close はデータベース接続を閉じる
func (r *baseRepositoryImpl) Close() error {
	if r.db == nil {
		return nil
	}
	
	if err := r.db.Close(); err != nil {
		return NewRepositoryError(ErrTypeConnection, "データベース接続の終了に失敗しました", err)
	}
	
	return nil
}

// BeginTx はトランザクションを開始する
func (r *baseRepositoryImpl) BeginTx() (*sql.Tx, error) {
	if r.db == nil {
		return nil, NewRepositoryError(ErrTypeConnection, "データベース接続が初期化されていません", nil)
	}
	
	tx, err := r.db.BeginTx()
	if err != nil {
		return nil, NewRepositoryError(ErrTypeTransaction, "トランザクションの開始に失敗しました", err)
	}
	
	log.Println("トランザクションを開始しました")
	return tx, nil
}

// CommitTx はトランザクションをコミットする
func (r *baseRepositoryImpl) CommitTx(tx *sql.Tx) error {
	if tx == nil {
		return NewRepositoryError(ErrTypeTransaction, "トランザクションがnilです", nil)
	}
	
	if err := tx.Commit(); err != nil {
		return NewRepositoryError(ErrTypeTransaction, "トランザクションのコミットに失敗しました", err)
	}
	
	log.Println("トランザクションをコミットしました")
	return nil
}

// RollbackTx はトランザクションをロールバックする
func (r *baseRepositoryImpl) RollbackTx(tx *sql.Tx) error {
	if tx == nil {
		return NewRepositoryError(ErrTypeTransaction, "トランザクションがnilです", nil)
	}
	
	if err := tx.Rollback(); err != nil {
		return NewRepositoryError(ErrTypeTransaction, "トランザクションのロールバックに失敗しました", err)
	}
	
	log.Println("トランザクションをロールバックしました")
	return nil
}

// ExecQuery はクエリを実行し、結果を返す
func (r *baseRepositoryImpl) ExecQuery(query string, args ...interface{}) (sql.Result, error) {
	if r.db == nil {
		return nil, NewRepositoryError(ErrTypeConnection, "データベース接続が初期化されていません", nil)
	}
	
	start := time.Now()
	result, err := r.db.Exec(query, args...)
	duration := time.Since(start)
	
	if err != nil {
		log.Printf("クエリ実行エラー [%v]: %s", duration, query)
		return nil, NewRepositoryError(ErrTypeQuery, "クエリの実行に失敗しました", err)
	}
	
	log.Printf("クエリ実行完了 [%v]: %s", duration, query)
	return result, nil
}

// QueryRow は単一行を取得するクエリを実行する
func (r *baseRepositoryImpl) QueryRow(query string, args ...interface{}) *sql.Row {
	if r.db == nil {
		log.Printf("データベース接続が初期化されていません: %s", query)
		return nil
	}
	
	start := time.Now()
	row := r.db.QueryRow(query, args...)
	duration := time.Since(start)
	
	log.Printf("QueryRow実行完了 [%v]: %s", duration, query)
	return row
}

// Query は複数行を取得するクエリを実行する
func (r *baseRepositoryImpl) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if r.db == nil {
		return nil, NewRepositoryError(ErrTypeConnection, "データベース接続が初期化されていません", nil)
	}
	
	start := time.Now()
	rows, err := r.db.Query(query, args...)
	duration := time.Since(start)
	
	if err != nil {
		log.Printf("クエリ実行エラー [%v]: %s", duration, query)
		return nil, NewRepositoryError(ErrTypeQuery, "クエリの実行に失敗しました", err)
	}
	
	log.Printf("Query実行完了 [%v]: %s", duration, query)
	return rows, nil
}

// ExecQueryTx はトランザクション内でクエリを実行し、結果を返す
func (r *baseRepositoryImpl) ExecQueryTx(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	if tx == nil {
		return nil, NewRepositoryError(ErrTypeTransaction, "トランザクションがnilです", nil)
	}
	
	start := time.Now()
	result, err := tx.Exec(query, args...)
	duration := time.Since(start)
	
	if err != nil {
		log.Printf("トランザクション内クエリ実行エラー [%v]: %s", duration, query)
		return nil, NewRepositoryError(ErrTypeQuery, "トランザクション内でのクエリ実行に失敗しました", err)
	}
	
	log.Printf("トランザクション内クエリ実行完了 [%v]: %s", duration, query)
	return result, nil
}

// QueryRowTx はトランザクション内で単一行を取得するクエリを実行する
func (r *baseRepositoryImpl) QueryRowTx(tx *sql.Tx, query string, args ...interface{}) *sql.Row {
	if tx == nil {
		log.Printf("トランザクションがnilです: %s", query)
		return nil
	}
	
	start := time.Now()
	row := tx.QueryRow(query, args...)
	duration := time.Since(start)
	
	log.Printf("トランザクション内QueryRow実行完了 [%v]: %s", duration, query)
	return row
}

// QueryTx はトランザクション内で複数行を取得するクエリを実行する
func (r *baseRepositoryImpl) QueryTx(tx *sql.Tx, query string, args ...interface{}) (*sql.Rows, error) {
	if tx == nil {
		return nil, NewRepositoryError(ErrTypeTransaction, "トランザクションがnilです", nil)
	}
	
	start := time.Now()
	rows, err := tx.Query(query, args...)
	duration := time.Since(start)
	
	if err != nil {
		log.Printf("トランザクション内クエリ実行エラー [%v]: %s", duration, query)
		return nil, NewRepositoryError(ErrTypeQuery, "トランザクション内でのクエリ実行に失敗しました", err)
	}
	
	log.Printf("トランザクション内Query実行完了 [%v]: %s", duration, query)
	return rows, nil
}