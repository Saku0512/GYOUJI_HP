package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"backend/internal/models"
)

// TestPaginationHelper はページネーションヘルパーの機能をテストする
func TestPaginationHelper(t *testing.T) {
	helper := NewPaginationHelper()

	t.Run("ApplyDefaults", func(t *testing.T) {
		tests := []struct {
			name     string
			input    *models.PaginationRequest
			expected *models.PaginationRequest
		}{
			{
				name:  "nil入力",
				input: nil,
				expected: &models.PaginationRequest{
					Page:     1,
					PageSize: 20,
				},
			},
			{
				name: "無効なページ番号",
				input: &models.PaginationRequest{
					Page:     0,
					PageSize: 10,
				},
				expected: &models.PaginationRequest{
					Page:     1,
					PageSize: 10,
				},
			},
			{
				name: "無効なページサイズ",
				input: &models.PaginationRequest{
					Page:     2,
					PageSize: 0,
				},
				expected: &models.PaginationRequest{
					Page:     2,
					PageSize: 20,
				},
			},
			{
				name: "ページサイズ上限超過",
				input: &models.PaginationRequest{
					Page:     1,
					PageSize: 150,
				},
				expected: &models.PaginationRequest{
					Page:     1,
					PageSize: 100,
				},
			},
			{
				name: "正常な値",
				input: &models.PaginationRequest{
					Page:     3,
					PageSize: 50,
				},
				expected: &models.PaginationRequest{
					Page:     3,
					PageSize: 50,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := helper.ApplyDefaults(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("ApplyPagination", func(t *testing.T) {
		tests := []struct {
			name     string
			query    string
			req      *models.PaginationRequest
			expected string
		}{
			{
				name:  "nil リクエスト",
				query: "SELECT * FROM tournaments",
				req:   nil,
				expected: "SELECT * FROM tournaments",
			},
			{
				name:  "基本的なページネーション",
				query: "SELECT * FROM tournaments ORDER BY id",
				req: &models.PaginationRequest{
					Page:     2,
					PageSize: 10,
				},
				expected: "SELECT * FROM tournaments ORDER BY id LIMIT 10 OFFSET 10",
			},
			{
				name:  "最初のページ",
				query: "SELECT * FROM matches",
				req: &models.PaginationRequest{
					Page:     1,
					PageSize: 20,
				},
				expected: "SELECT * FROM matches LIMIT 20 OFFSET 0",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := helper.ApplyPagination(tt.query, tt.req)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("BuildCountQuery", func(t *testing.T) {
		tests := []struct {
			name     string
			query    string
			expected string
		}{
			{
				name:     "基本的なSELECTクエリ",
				query:    "SELECT id, name FROM tournaments WHERE sport = 'volleyball'",
				expected: "SELECT COUNT(*) FROM tournaments WHERE sport = 'volleyball'",
			},
			{
				name:     "ORDER BY付きクエリ",
				query:    "SELECT * FROM matches ORDER BY created_at DESC",
				expected: "SELECT COUNT(*) FROM matches",
			},
			{
				name:     "LIMIT付きクエリ",
				query:    "SELECT * FROM tournaments LIMIT 10 OFFSET 5",
				expected: "SELECT COUNT(*) FROM tournaments",
			},
			{
				name:     "複雑なクエリ",
				query:    "SELECT t.*, m.count FROM tournaments t JOIN (SELECT tournament_id, COUNT(*) as count FROM matches GROUP BY tournament_id) m ON t.id = m.tournament_id ORDER BY t.created_at DESC LIMIT 20",
				expected: "SELECT COUNT(*) FROM tournaments t JOIN (SELECT tournament_id, COUNT(*) as count FROM matches GROUP BY tournament_id) m ON t.id = m.tournament_id",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := helper.BuildCountQuery(tt.query)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("CreatePaginationResponse", func(t *testing.T) {
		tests := []struct {
			name       string
			req        *models.PaginationRequest
			totalItems int
			expected   *models.PaginationResponse
		}{
			{
				name: "基本的なページネーション",
				req: &models.PaginationRequest{
					Page:     2,
					PageSize: 10,
				},
				totalItems: 25,
				expected: &models.PaginationResponse{
					Page:       2,
					PageSize:   10,
					TotalItems: 25,
					TotalPages: 3,
					HasNext:    true,
					HasPrev:    true,
				},
			},
			{
				name: "最初のページ",
				req: &models.PaginationRequest{
					Page:     1,
					PageSize: 20,
				},
				totalItems: 50,
				expected: &models.PaginationResponse{
					Page:       1,
					PageSize:   20,
					TotalItems: 50,
					TotalPages: 3,
					HasNext:    true,
					HasPrev:    false,
				},
			},
			{
				name: "最後のページ",
				req: &models.PaginationRequest{
					Page:     3,
					PageSize: 10,
				},
				totalItems: 25,
				expected: &models.PaginationResponse{
					Page:       3,
					PageSize:   10,
					TotalItems: 25,
					TotalPages: 3,
					HasNext:    false,
					HasPrev:    true,
				},
			},
			{
				name: "データなし",
				req: &models.PaginationRequest{
					Page:     1,
					PageSize: 10,
				},
				totalItems: 0,
				expected: &models.PaginationResponse{
					Page:       1,
					PageSize:   10,
					TotalItems: 0,
					TotalPages: 1,
					HasNext:    false,
					HasPrev:    false,
				},
			},
			{
				name:       "nil リクエスト",
				req:        nil,
				totalItems: 100,
				expected: &models.PaginationResponse{
					Page:       1,
					PageSize:   20,
					TotalItems: 100,
					TotalPages: 5,
					HasNext:    true,
					HasPrev:    false,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := helper.CreatePaginationResponse(tt.req, tt.totalItems)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// TestPaginationRequest はPaginationRequestの機能をテストする
func TestPaginationRequest(t *testing.T) {
	t.Run("GetOffset", func(t *testing.T) {
		tests := []struct {
			name     string
			req      *models.PaginationRequest
			expected int
		}{
			{
				name: "最初のページ",
				req: &models.PaginationRequest{
					Page:     1,
					PageSize: 10,
				},
				expected: 0,
			},
			{
				name: "2ページ目",
				req: &models.PaginationRequest{
					Page:     2,
					PageSize: 10,
				},
				expected: 10,
			},
			{
				name: "5ページ目、ページサイズ20",
				req: &models.PaginationRequest{
					Page:     5,
					PageSize: 20,
				},
				expected: 80,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.req.GetOffset()
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("GetLimit", func(t *testing.T) {
		tests := []struct {
			name     string
			req      *models.PaginationRequest
			expected int
		}{
			{
				name: "ページサイズ10",
				req: &models.PaginationRequest{
					Page:     1,
					PageSize: 10,
				},
				expected: 10,
			},
			{
				name: "ページサイズ50",
				req: &models.PaginationRequest{
					Page:     2,
					PageSize: 50,
				},
				expected: 50,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.req.GetLimit()
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("Validate", func(t *testing.T) {
		tests := []struct {
			name      string
			req       *models.PaginationRequest
			expectErr bool
		}{
			{
				name: "正常な値",
				req: &models.PaginationRequest{
					Page:     1,
					PageSize: 20,
				},
				expectErr: false,
			},
			{
				name: "無効なページ番号",
				req: &models.PaginationRequest{
					Page:     0,
					PageSize: 20,
				},
				expectErr: true,
			},
			{
				name: "無効なページサイズ（小さすぎる）",
				req: &models.PaginationRequest{
					Page:     1,
					PageSize: 0,
				},
				expectErr: true,
			},
			{
				name: "無効なページサイズ（大きすぎる）",
				req: &models.PaginationRequest{
					Page:     1,
					PageSize: 101,
				},
				expectErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.req.Validate()
				if tt.expectErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}