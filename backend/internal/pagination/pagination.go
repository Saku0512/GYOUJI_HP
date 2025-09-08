// Package pagination はページネーション機能の統一実装を提供する
package pagination

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"backend/internal/models"
)

// PaginationHelper はページネーション処理のヘルパー
type PaginationHelper interface {
	// SQLクエリにページネーションを適用
	ApplyPagination(query string, req *models.PaginationRequest) string
	
	// 総件数を取得するクエリを生成
	BuildCountQuery(baseQuery string) string
	
	// ページネーションレスポンスを作成
	CreatePaginationResponse(req *models.PaginationRequest, totalItems int) *models.PaginationResponse
	
	// デフォルトページネーション設定を適用
	ApplyDefaults(req *models.PaginationRequest) *models.PaginationRequest
}

// paginationHelperImpl はPaginationHelperの実装
type paginationHelperImpl struct {
	defaultPageSize int
	maxPageSize     int
}

// NewPaginationHelper は新しいページネーションヘルパーを作成する
func NewPaginationHelper() PaginationHelper {
	return &paginationHelperImpl{
		defaultPageSize: 20,
		maxPageSize:     100,
	}
}

// ApplyPagination はSQLクエリにLIMITとOFFSETを追加する
func (h *paginationHelperImpl) ApplyPagination(query string, req *models.PaginationRequest) string {
	if req == nil {
		return query
	}

	// デフォルト値を適用
	req = h.ApplyDefaults(req)

	// LIMIT と OFFSET を追加
	offset := req.GetOffset()
	limit := req.GetLimit()

	return fmt.Sprintf("%s LIMIT %d OFFSET %d", query, limit, offset)
}

// BuildCountQuery は総件数取得用のクエリを生成する
func (h *paginationHelperImpl) BuildCountQuery(baseQuery string) string {
	// SELECT句をCOUNT(*)に置換
	query := strings.TrimSpace(baseQuery)
	
	// ORDER BY句を除去（COUNT クエリでは不要）
	if orderByIndex := strings.Index(strings.ToUpper(query), "ORDER BY"); orderByIndex != -1 {
		query = query[:orderByIndex]
	}
	
	// LIMIT句を除去
	if limitIndex := strings.Index(strings.ToUpper(query), "LIMIT"); limitIndex != -1 {
		query = query[:limitIndex]
	}
	
	// SELECT句を見つけてCOUNT(*)に置換
	selectIndex := strings.Index(strings.ToUpper(query), "SELECT")
	fromIndex := strings.Index(strings.ToUpper(query), "FROM")
	
	if selectIndex == -1 || fromIndex == -1 {
		log.Printf("無効なクエリ形式: %s", query)
		return "SELECT COUNT(*) " + query
	}
	
	return "SELECT COUNT(*) " + query[fromIndex:]
}

// CreatePaginationResponse はページネーション情報を作成する
func (h *paginationHelperImpl) CreatePaginationResponse(req *models.PaginationRequest, totalItems int) *models.PaginationResponse {
	if req == nil {
		req = &models.PaginationRequest{
			Page:     1,
			PageSize: h.defaultPageSize,
		}
	}

	req = h.ApplyDefaults(req)
	return models.NewPaginationResponse(req.Page, req.PageSize, totalItems)
}

// ApplyDefaults はデフォルト値を適用する
func (h *paginationHelperImpl) ApplyDefaults(req *models.PaginationRequest) *models.PaginationRequest {
	if req == nil {
		return &models.PaginationRequest{
			Page:     1,
			PageSize: h.defaultPageSize,
		}
	}

	result := *req // コピーを作成

	if result.Page < 1 {
		result.Page = 1
	}

	if result.PageSize < 1 {
		result.PageSize = h.defaultPageSize
	}

	if result.PageSize > h.maxPageSize {
		result.PageSize = h.maxPageSize
	}

	return &result
}

// PaginatedQuery はページネーション付きクエリの実行結果
type PaginatedQuery[T any] struct {
	Data       []T                        `json:"data"`
	Pagination *models.PaginationResponse `json:"pagination"`
}

// QueryExecutor はページネーション付きクエリの実行インターフェース
type QueryExecutor interface {
	// ページネーション付きクエリを実行
	ExecutePaginatedQuery(
		ctx context.Context,
		baseQuery string,
		countQuery string,
		req *models.PaginationRequest,
		scanner func(*sql.Rows) (interface{}, error),
		args ...interface{},
	) (*PaginatedQuery[interface{}], error)
}

// queryExecutorImpl はQueryExecutorの実装
type queryExecutorImpl struct {
	db     interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
		QueryRow(query string, args ...interface{}) *sql.Row
	}
	helper PaginationHelper
}

// NewQueryExecutor は新しいクエリ実行器を作成する
func NewQueryExecutor(db interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}) QueryExecutor {
	return &queryExecutorImpl{
		db:     db,
		helper: NewPaginationHelper(),
	}
}

// ExecutePaginatedQuery はページネーション付きクエリを実行する
func (e *queryExecutorImpl) ExecutePaginatedQuery(
	ctx context.Context,
	baseQuery string,
	countQuery string,
	req *models.PaginationRequest,
	scanner func(*sql.Rows) (interface{}, error),
	args ...interface{},
) (*PaginatedQuery[interface{}], error) {
	
	// デフォルト値を適用
	req = e.helper.ApplyDefaults(req)

	// 総件数を取得
	var totalItems int
	if countQuery == "" {
		countQuery = e.helper.BuildCountQuery(baseQuery)
	}
	
	row := e.db.QueryRow(countQuery, args...)
	if err := row.Scan(&totalItems); err != nil {
		return nil, fmt.Errorf("総件数の取得に失敗しました: %w", err)
	}

	// ページネーション情報を作成
	pagination := e.helper.CreatePaginationResponse(req, totalItems)

	// データが存在しない場合は空の結果を返す
	if totalItems == 0 {
		return &PaginatedQuery[interface{}]{
			Data:       []interface{}{},
			Pagination: pagination,
		}, nil
	}

	// ページネーション付きクエリを実行
	paginatedQuery := e.helper.ApplyPagination(baseQuery, req)
	rows, err := e.db.Query(paginatedQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("ページネーション付きクエリの実行に失敗しました: %w", err)
	}
	defer rows.Close()

	// 結果をスキャン
	var data []interface{}
	for rows.Next() {
		item, err := scanner(rows)
		if err != nil {
			return nil, fmt.Errorf("行のスキャンに失敗しました: %w", err)
		}
		data = append(data, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("行の処理中にエラーが発生しました: %w", err)
	}

	return &PaginatedQuery[interface{}]{
		Data:       data,
		Pagination: pagination,
	}, nil
}

// PaginatedTournamentQuery はトーナメント用のページネーション付きクエリ
func ExecutePaginatedTournamentQuery(
	ctx context.Context,
	db interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
		QueryRow(query string, args ...interface{}) *sql.Row
	},
	baseQuery string,
	req *models.PaginationRequest,
	args ...interface{},
) (*PaginatedQuery[*models.Tournament], error) {
	
	executor := NewQueryExecutor(db)
	helper := NewPaginationHelper()

	// 総件数クエリを生成
	countQuery := helper.BuildCountQuery(baseQuery)

	// スキャナー関数を定義
	scanner := func(rows *sql.Rows) (interface{}, error) {
		tournament := &models.Tournament{}
		err := rows.Scan(
			&tournament.ID,
			&tournament.Sport,
			&tournament.Format,
			&tournament.Status,
			&tournament.CreatedAt,
			&tournament.UpdatedAt,
		)
		return tournament, err
	}

	// クエリを実行
	result, err := executor.ExecutePaginatedQuery(ctx, baseQuery, countQuery, req, scanner, args...)
	if err != nil {
		return nil, err
	}

	// 型変換
	tournaments := make([]*models.Tournament, len(result.Data))
	for i, item := range result.Data {
		tournaments[i] = item.(*models.Tournament)
	}

	return &PaginatedQuery[*models.Tournament]{
		Data:       tournaments,
		Pagination: result.Pagination,
	}, nil
}

// PaginatedMatchQuery は試合用のページネーション付きクエリ
func ExecutePaginatedMatchQuery(
	ctx context.Context,
	db interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
		QueryRow(query string, args ...interface{}) *sql.Row
	},
	baseQuery string,
	req *models.PaginationRequest,
	args ...interface{},
) (*PaginatedQuery[*models.Match], error) {
	
	executor := NewQueryExecutor(db)
	helper := NewPaginationHelper()

	// 総件数クエリを生成
	countQuery := helper.BuildCountQuery(baseQuery)

	// スキャナー関数を定義
	scanner := func(rows *sql.Rows) (interface{}, error) {
		match := &models.Match{}
		err := rows.Scan(
			&match.ID,
			&match.TournamentID,
			&match.Round,
			&match.Team1,
			&match.Team2,
			&match.Score1,
			&match.Score2,
			&match.Winner,
			&match.Status,
			&match.ScheduledAt,
			&match.CompletedAt,
			&match.CreatedAt,
			&match.UpdatedAt,
		)
		return match, err
	}

	// クエリを実行
	result, err := executor.ExecutePaginatedQuery(ctx, baseQuery, countQuery, req, scanner, args...)
	if err != nil {
		return nil, err
	}

	// 型変換
	matches := make([]*models.Match, len(result.Data))
	for i, item := range result.Data {
		matches[i] = item.(*models.Match)
	}

	return &PaginatedQuery[*models.Match]{
		Data:       matches,
		Pagination: result.Pagination,
	}, nil
}