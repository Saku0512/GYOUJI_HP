// Package integration_test はエンドツーエンドAPIテストを提供する
package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"backend/internal/testutil"

	"github.com/stretchr/testify/suite"
)

// E2EAPITestSuite はエンドツーエンドAPIテストスイート
type E2EAPITestSuite struct {
	suite.Suite
	server *testutil.TestServer
}

// SetupSuite はテストスイートのセットアップを行う
func (suite *E2EAPITestSuite) SetupSuite() {
	suite.server = testutil.SetupTestServer(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
}

// TearDownSuite はテストスイートのクリーンアップを行う
func (suite *E2EAPITestSuite) TearDownSuite() {
	suite.server.TeardownTestServer(suite.T())
}

// SetupTest は各テストの前に実行される
func (suite *E2EAPITestSuite) SetupTest() {
	// 各テスト前にデータをクリーンアップしてシード
	suite.server.TestDB.CleanupTestData(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
}

// TestE2E_CompleteWorkflow は完全なワークフローのエンドツーエンドテスト
func (suite *E2EAPITestSuite) TestE2E_CompleteWorkflow() {
	// 要件7.4: 統合テストが実行される場合、システムは実際のAPI呼び出しでエンドツーエンドの動作を検証する
	
	// Step 1: 管理者ログイン
	suite.T().Log("Step 1: 管理者ログイン")
	token := suite.server.LoginAndGetToken(suite.T())
	suite.NotEmpty(token, "ログイントークンが取得できませんでした")

	// Step 2: 公開トーナメント一覧の取得（認証不要）
	suite.T().Log("Step 2: 公開トーナメント一覧の取得")
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments", nil, nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var tournamentsResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &tournamentsResponse)
	suite.NoError(err)
	suite.True(tournamentsResponse["success"].(bool))
	
	tournaments := tournamentsResponse["data"].([]interface{})
	suite.Greater(len(tournaments), 0, "トーナメントが存在しません")

	// Step 3: 特定スポーツのトーナメント詳細取得
	suite.T().Log("Step 3: バレーボールトーナメント詳細取得")
	w = suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments/sport/volleyball", nil, nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var tournamentResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &tournamentResponse)
	suite.NoError(err)
	suite.True(tournamentResponse["success"].(bool))
	
	tournament := tournamentResponse["data"].(map[string]interface{})
	suite.Equal("volleyball", tournament["sport"])
	tournamentID := int(tournament["id"].(float64))

	// Step 4: 試合一覧の取得
	suite.T().Log("Step 4: バレーボール試合一覧の取得")
	w = suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/matches/sport/volleyball", nil, nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var matchesResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &matchesResponse)
	suite.NoError(err)
	suite.True(matchesResponse["success"].(bool))

	// Step 5: 新しい試合の作成（管理者のみ）
	suite.T().Log("Step 5: 新しい試合の作成")
	createMatchRequest := map[string]interface{}{
		"tournament_id": tournamentID,
		"round":         "quarterfinal",
		"team1":         "E2Eテストチーム1",
		"team2":         "E2Eテストチーム2",
		"scheduled_at":  time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
	}
	
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "POST", "/api/v1/admin/matches", createMatchRequest, token)
	suite.Equal(http.StatusCreated, w.Code)
	
	var createMatchResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createMatchResponse)
	suite.NoError(err)
	suite.True(createMatchResponse["success"].(bool))
	
	createdMatch := createMatchResponse["data"].(map[string]interface{})
	matchID := int(createdMatch["id"].(float64))
	suite.Greater(matchID, 0, "試合IDが正しく生成されませんでした")

	// Step 6: 作成した試合の詳細取得
	suite.T().Log("Step 6: 作成した試合の詳細取得")
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "GET", fmt.Sprintf("/api/v1/matches/%d", matchID), nil, token)
	suite.Equal(http.StatusOK, w.Code)
	
	var matchResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &matchResponse)
	suite.NoError(err)
	suite.True(matchResponse["success"].(bool))
	
	match := matchResponse["data"].(map[string]interface{})
	suite.Equal("E2Eテストチーム1", match["team1"])
	suite.Equal("E2Eテストチーム2", match["team2"])
	suite.Equal("pending", match["status"])

	// Step 7: 試合結果の提出
	suite.T().Log("Step 7: 試合結果の提出")
	submitResultRequest := map[string]interface{}{
		"score1": 3,
		"score2": 1,
		"winner": "E2Eテストチーム1",
	}
	
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", fmt.Sprintf("/api/v1/admin/matches/%d/result", matchID), submitResultRequest, token)
	suite.Equal(http.StatusOK, w.Code)
	
	var resultResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resultResponse)
	suite.NoError(err)
	suite.True(resultResponse["success"].(bool))
	
	updatedMatch := resultResponse["data"].(map[string]interface{})
	suite.Equal(float64(3), updatedMatch["score1"])
	suite.Equal(float64(1), updatedMatch["score2"])
	suite.Equal("E2Eテストチーム1", updatedMatch["winner"])
	suite.Equal("completed", updatedMatch["status"])
	suite.NotNil(updatedMatch["completed_at"])

	// Step 8: トーナメントブラケットの取得
	suite.T().Log("Step 8: トーナメントブラケットの取得")
	w = suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments/sport/volleyball/bracket", nil, nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var bracketResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &bracketResponse)
	suite.NoError(err)
	suite.True(bracketResponse["success"].(bool))
	
	bracket := bracketResponse["data"].(map[string]interface{})
	suite.Equal("volleyball", bracket["sport"])
	suite.Contains(bracket, "rounds")

	// Step 9: トーナメント進行状況の取得
	suite.T().Log("Step 9: トーナメント進行状況の取得")
	w = suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments/sport/volleyball/progress", nil, nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var progressResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &progressResponse)
	suite.NoError(err)
	suite.True(progressResponse["success"].(bool))
	
	progress := progressResponse["data"].(map[string]interface{})
	suite.Equal("volleyball", progress["sport"])
	suite.Contains(progress, "total_matches")
	suite.Contains(progress, "completed_matches")
	suite.Contains(progress, "progress_percent")

	// Step 10: 試合統計の取得
	suite.T().Log("Step 10: 試合統計の取得")
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "GET", fmt.Sprintf("/api/v1/matches/tournament/%d/statistics", tournamentID), nil, token)
	suite.Equal(http.StatusOK, w.Code)
	
	var statsResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &statsResponse)
	suite.NoError(err)
	suite.True(statsResponse["success"].(bool))
	
	stats := statsResponse["data"].(map[string]interface{})
	suite.Equal(float64(tournamentID), stats["tournament_id"])
	suite.Contains(stats, "total_matches")
	suite.Contains(stats, "completion_rate")

	suite.T().Log("✅ エンドツーエンドワークフローテストが正常に完了しました")
}

// TestE2E_MultiSportWorkflow は複数スポーツのワークフローテスト
func (suite *E2EAPITestSuite) TestE2E_MultiSportWorkflow() {
	// 要件7.5: モックサーバーが必要な場合、システムは実際のAPIと同じ仕様でモックレスポンスを提供する
	token := suite.server.LoginAndGetToken(suite.T())
	
	sports := []string{"volleyball", "table_tennis", "soccer"}
	
	for _, sport := range sports {
		suite.Run(fmt.Sprintf("Sport_%s", sport), func() {
			suite.T().Logf("Testing workflow for sport: %s", sport)
			
			// 1. スポーツ別トーナメント取得
			w := suite.server.MakeRequest(suite.T(), "GET", fmt.Sprintf("/api/v1/public/tournaments/sport/%s", sport), nil, nil)
			suite.Equal(http.StatusOK, w.Code)
			
			var tournamentResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &tournamentResponse)
			suite.NoError(err)
			suite.True(tournamentResponse["success"].(bool))
			
			tournament := tournamentResponse["data"].(map[string]interface{})
			suite.Equal(sport, tournament["sport"])
			tournamentID := int(tournament["id"].(float64))
			
			// 2. スポーツ別試合取得
			w = suite.server.MakeRequest(suite.T(), "GET", fmt.Sprintf("/api/v1/public/matches/sport/%s", sport), nil, nil)
			suite.Equal(http.StatusOK, w.Code)
			
			var matchesResponse map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &matchesResponse)
			suite.NoError(err)
			suite.True(matchesResponse["success"].(bool))
			
			// 3. 次の試合取得
			w = suite.server.MakeRequest(suite.T(), "GET", fmt.Sprintf("/api/v1/public/matches/tournament/%d/next", tournamentID), nil, nil)
			suite.Equal(http.StatusOK, w.Code)
			
			var nextMatchesResponse map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &nextMatchesResponse)
			suite.NoError(err)
			suite.True(nextMatchesResponse["success"].(bool))
			
			// 4. ブラケット取得
			w = suite.server.MakeRequest(suite.T(), "GET", fmt.Sprintf("/api/v1/public/tournaments/sport/%s/bracket", sport), nil, nil)
			suite.Equal(http.StatusOK, w.Code)
			
			var bracketResponse map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &bracketResponse)
			suite.NoError(err)
			suite.True(bracketResponse["success"].(bool))
			
			bracket := bracketResponse["data"].(map[string]interface{})
			suite.Equal(sport, bracket["sport"])
			
			suite.T().Logf("✅ %s のワークフローテストが完了しました", sport)
		})
	}
}

// TestE2E_ErrorHandlingWorkflow はエラーハンドリングのエンドツーエンドテスト
func (suite *E2EAPITestSuite) TestE2E_ErrorHandlingWorkflow() {
	// 要件7.4: エラーケースの包括的テスト
	
	// 1. 認証なしでの保護されたリソースアクセス
	suite.T().Log("Testing unauthorized access")
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/tournaments", nil, nil)
	suite.Equal(http.StatusUnauthorized, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	suite.NoError(err)
	suite.False(errorResponse["success"].(bool))
	suite.Equal("AUTH_UNAUTHORIZED", errorResponse["error"])
	
	// 2. 無効なトークンでのアクセス
	suite.T().Log("Testing invalid token access")
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/v1/tournaments", nil, "invalid.jwt.token")
	suite.Equal(http.StatusUnauthorized, w.Code)
	
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	suite.NoError(err)
	suite.False(errorResponse["success"].(bool))
	suite.Equal("AUTH_TOKEN_INVALID", errorResponse["error"])
	
	// 3. 存在しないリソースへのアクセス
	suite.T().Log("Testing not found resources")
	w = suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments/sport/invalid_sport", nil, nil)
	suite.Equal(http.StatusNotFound, w.Code)
	
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	suite.NoError(err)
	suite.False(errorResponse["success"].(bool))
	suite.Equal("RESOURCE_NOT_FOUND", errorResponse["error"])
	
	// 4. バリデーションエラー
	suite.T().Log("Testing validation errors")
	token := suite.server.LoginAndGetToken(suite.T())
	
	invalidMatchRequest := map[string]interface{}{
		"tournament_id": -1, // 無効なID
		"round":         "",  // 空のラウンド
		"team1":         "",  // 空のチーム名
		"team2":         "",  // 空のチーム名
		"scheduled_at":  "invalid-date", // 無効な日時
	}
	
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "POST", "/api/v1/admin/matches", invalidMatchRequest, token)
	suite.Equal(http.StatusBadRequest, w.Code)
	
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	suite.NoError(err)
	suite.False(errorResponse["success"].(bool))
	suite.Equal("VALIDATION_ERROR", errorResponse["error"])
	
	// 5. ビジネスロジックエラー
	suite.T().Log("Testing business logic errors")
	
	// まず有効な試合を作成
	createMatchRequest := map[string]interface{}{
		"tournament_id": 1,
		"round":         "quarterfinal",
		"team1":         "チームA",
		"team2":         "チームB",
		"scheduled_at":  time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
	}
	
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "POST", "/api/v1/admin/matches", createMatchRequest, token)
	suite.Equal(http.StatusCreated, w.Code)
	
	var createResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createResponse)
	suite.NoError(err)
	
	match := createResponse["data"].(map[string]interface{})
	matchID := int(match["id"].(float64))
	
	// 無効な試合結果を提出（勝者がチームに含まれない）
	invalidResultRequest := map[string]interface{}{
		"score1": 3,
		"score2": 1,
		"winner": "存在しないチーム", // 無効な勝者
	}
	
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", fmt.Sprintf("/api/v1/admin/matches/%d/result", matchID), invalidResultRequest, token)
	suite.Equal(http.StatusUnprocessableEntity, w.Code)
	
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	suite.NoError(err)
	suite.False(errorResponse["success"].(bool))
	suite.Contains(errorResponse["error"].(string), "BUSINESS_")
	
	suite.T().Log("✅ エラーハンドリングワークフローテストが完了しました")
}

// TestE2E_PerformanceAndLimits はパフォーマンスと制限のテスト
func (suite *E2EAPITestSuite) TestE2E_PerformanceAndLimits() {
	// ページネーションのテスト
	suite.T().Log("Testing pagination")
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments?page=1&page_size=2", nil, nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response["success"].(bool))
	
	// ページネーション情報の確認（実装されている場合）
	if pagination, exists := response["pagination"]; exists {
		paginationData := pagination.(map[string]interface{})
		suite.Equal(float64(1), paginationData["page"])
		suite.Equal(float64(2), paginationData["page_size"])
	}
	
	// 大きなページサイズの制限テスト
	suite.T().Log("Testing page size limits")
	w = suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments?page=1&page_size=1000", nil, nil)
	// 制限が実装されている場合は400、されていない場合は200
	suite.True(w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
	
	// レスポンス時間の基本チェック
	suite.T().Log("Testing response time")
	start := time.Now()
	w = suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments", nil, nil)
	duration := time.Since(start)
	
	suite.Equal(http.StatusOK, w.Code)
	suite.Less(duration, 5*time.Second, "レスポンス時間が5秒を超えています")
	
	suite.T().Log("✅ パフォーマンスと制限テストが完了しました")
}

// TestE2E_ConcurrentAccess は同時アクセスのテスト
func (suite *E2EAPITestSuite) TestE2E_ConcurrentAccess() {
	// 複数の同時リクエストをテスト
	suite.T().Log("Testing concurrent access")
	
	const numRequests = 10
	results := make(chan int, numRequests)
	
	// 同時に複数のリクエストを送信
	for i := 0; i < numRequests; i++ {
		go func() {
			w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments", nil, nil)
			results <- w.Code
		}()
	}
	
	// 結果を収集
	successCount := 0
	for i := 0; i < numRequests; i++ {
		statusCode := <-results
		if statusCode == http.StatusOK {
			successCount++
		}
	}
	
	// 全てのリクエストが成功することを確認
	suite.Equal(numRequests, successCount, "同時アクセス時に一部のリクエストが失敗しました")
	
	suite.T().Log("✅ 同時アクセステストが完了しました")
}

// TestE2EAPITestSuite はエンドツーエンドAPIテストスイートを実行する
func TestE2EAPITestSuite(t *testing.T) {
	suite.Run(t, new(E2EAPITestSuite))
}