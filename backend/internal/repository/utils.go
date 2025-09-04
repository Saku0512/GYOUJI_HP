package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"backend/internal/database"
)

// ConnectionManager はデータベース接続の管理を行う
type ConnectionManager struct {
	db             *database.DB
	maxRetries     int
	retryInterval  time.Duration
	healthCheckInterval time.Duration
	stopHealthCheck chan bool
}

// NewConnectionManager は新しいConnectionManagerを作成する
func NewConnectionManager(db *database.DB) *ConnectionManager {
	return &ConnectionManager{
		db:                  db,
		maxRetries:          3,
		retryInterval:       time.Second * 2,
		healthCheckInterval: time.Minute * 5,
		stopHealthCheck:     make(chan bool),
	}
}

// SetRetryConfig はリトライ設定を変更する
func (cm *ConnectionManager) SetRetryConfig(maxRetries int, retryInterval time.Duration) {
	cm.maxRetries = maxRetries
	cm.retryInterval = retryInterval
}

// SetHealthCheckInterval はヘルスチェック間隔を設定する
func (cm *ConnectionManager) SetHealthCheckInterval(interval time.Duration) {
	cm.healthCheckInterval = interval
}

// StartHealthCheck はデータベース接続のヘルスチェックを開始する
func (cm *ConnectionManager) StartHealthCheck() {
	go func() {
		ticker := time.NewTicker(cm.healthCheckInterval)
		defer ticker.Stop()
		
		log.Printf("データベースヘルスチェックを開始しました (間隔: %v)", cm.healthCheckInterval)
		
		for {
			select {
			case <-ticker.C:
				if err := cm.db.Ping(); err != nil {
					log.Printf("データベースヘルスチェック失敗: %v", err)
				} else {
					log.Println("データベースヘルスチェック成功")
				}
			case <-cm.stopHealthCheck:
				log.Println("データベースヘルスチェックを停止しました")
				return
			}
		}
	}()
}

// StopHealthCheck はヘルスチェックを停止する
func (cm *ConnectionManager) StopHealthCheck() {
	select {
	case cm.stopHealthCheck <- true:
	default:
	}
}

// GetConnectionStats はデータベース接続の統計情報を取得する
func (cm *ConnectionManager) GetConnectionStats() sql.DBStats {
	return cm.db.GetStats()
}

// LogConnectionStats は接続統計情報をログに出力する
func (cm *ConnectionManager) LogConnectionStats() {
	stats := cm.GetConnectionStats()
	log.Printf("DB接続統計 - Open: %d, InUse: %d, Idle: %d, WaitCount: %d, WaitDuration: %v",
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
		stats.WaitCount,
		stats.WaitDuration,
	)
}

// TransactionManager はトランザクション管理のユーティリティを提供する
type TransactionManager struct {
	baseRepo BaseRepository
}

// NewTransactionManager は新しいTransactionManagerを作成する
func NewTransactionManager(baseRepo BaseRepository) *TransactionManager {
	return &TransactionManager{
		baseRepo: baseRepo,
	}
}

// ExecuteInTransaction はトランザクション内で操作を実行する
func (tm *TransactionManager) ExecuteInTransaction(operation func(*sql.Tx) error) error {
	tx, err := tm.baseRepo.BeginTx()
	if err != nil {
		return fmt.Errorf("トランザクション開始エラー: %w", err)
	}
	
	// パニック時のロールバック処理
	defer func() {
		if r := recover(); r != nil {
			if rollbackErr := tm.baseRepo.RollbackTx(tx); rollbackErr != nil {
				log.Printf("パニック時のロールバックエラー: %v", rollbackErr)
			}
			panic(r)
		}
	}()
	
	// 操作の実行
	if err := operation(tx); err != nil {
		if rollbackErr := tm.baseRepo.RollbackTx(tx); rollbackErr != nil {
			log.Printf("ロールバックエラー: %v", rollbackErr)
		}
		return fmt.Errorf("トランザクション内操作エラー: %w", err)
	}
	
	// コミット
	if err := tm.baseRepo.CommitTx(tx); err != nil {
		if rollbackErr := tm.baseRepo.RollbackTx(tx); rollbackErr != nil {
			log.Printf("コミット失敗後のロールバックエラー: %v", rollbackErr)
		}
		return fmt.Errorf("トランザクションコミットエラー: %w", err)
	}
	
	return nil
}

// QueryBuilder はクエリ構築のユーティリティを提供する
type QueryBuilder struct {
	query string
	args  []interface{}
}

// NewQueryBuilder は新しいQueryBuilderを作成する
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		query: "",
		args:  make([]interface{}, 0),
	}
}

// Select はSELECT句を追加する
func (qb *QueryBuilder) Select(columns string) *QueryBuilder {
	qb.query = "SELECT " + columns
	return qb
}

// From はFROM句を追加する
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.query += " FROM " + table
	return qb
}

// Where はWHERE句を追加する
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	if qb.query != "" {
		qb.query += " WHERE " + condition
		qb.args = append(qb.args, args...)
	}
	return qb
}

// And はAND条件を追加する
func (qb *QueryBuilder) And(condition string, args ...interface{}) *QueryBuilder {
	qb.query += " AND " + condition
	qb.args = append(qb.args, args...)
	return qb
}

// Or はOR条件を追加する
func (qb *QueryBuilder) Or(condition string, args ...interface{}) *QueryBuilder {
	qb.query += " OR " + condition
	qb.args = append(qb.args, args...)
	return qb
}

// OrderBy はORDER BY句を追加する
func (qb *QueryBuilder) OrderBy(column string) *QueryBuilder {
	qb.query += " ORDER BY " + column
	return qb
}

// Limit はLIMIT句を追加する
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.query += " LIMIT ?"
	qb.args = append(qb.args, limit)
	return qb
}

// Build はクエリと引数を返す
func (qb *QueryBuilder) Build() (string, []interface{}) {
	return qb.query, qb.args
}

// PaginationHelper はページネーション処理のユーティリティを提供する
type PaginationHelper struct {
	Page     int
	PageSize int
	Offset   int
}

// NewPaginationHelper は新しいPaginationHelperを作成する
func NewPaginationHelper(page, pageSize int) *PaginationHelper {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	
	offset := (page - 1) * pageSize
	
	return &PaginationHelper{
		Page:     page,
		PageSize: pageSize,
		Offset:   offset,
	}
}

// ApplyPagination はクエリにページネーションを適用する
func (ph *PaginationHelper) ApplyPagination(qb *QueryBuilder) *QueryBuilder {
	qb.query += " LIMIT ? OFFSET ?"
	qb.args = append(qb.args, ph.PageSize, ph.Offset)
	return qb
}

// GetTotalPages は総ページ数を計算する
func (ph *PaginationHelper) GetTotalPages(totalRecords int) int {
	if totalRecords == 0 {
		return 0
	}
	return (totalRecords + ph.PageSize - 1) / ph.PageSize
}

// PaginationResult はページネーション結果を表す
type PaginationResult struct {
	Page         int `json:"page"`
	PageSize     int `json:"page_size"`
	TotalRecords int `json:"total_records"`
	TotalPages   int `json:"total_pages"`
	HasNext      bool `json:"has_next"`
	HasPrevious  bool `json:"has_previous"`
}

// NewPaginationResult は新しいPaginationResultを作成する
func NewPaginationResult(ph *PaginationHelper, totalRecords int) *PaginationResult {
	totalPages := ph.GetTotalPages(totalRecords)
	
	return &PaginationResult{
		Page:         ph.Page,
		PageSize:     ph.PageSize,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		HasNext:      ph.Page < totalPages,
		HasPrevious:  ph.Page > 1,
	}
}