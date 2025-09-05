// Package integration_test はエンドツーエンドワークフローの統合テストを提供する
package integration_test

import (
	"net/http"
	"testing"

	"backend/internal/testutil"

	"github.com/stretchr/testify/suite"
)

// WorkflowIntegrationTestSuite はエンドツーエンドワークフローの統合テストスイート
type WorkflowIntegrationTestSuite struct {
	suite.Suite
	server *testutil.TestServer
	token  string
}

// SetupSuite はテストスイートのセットアップを行う
func (suite *WorkflowIntegrationTestSuite) SetupSuite() {
	suite.server = testutil.SetupTestServer(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
	suite.token = suite.server.LoginAndGetToken(suite.T())
}

// TearDownSuite はテストスイートのクリーンアップを行う
func (suite *WorkflowIntegrationTestSuite) TearDownSuite() {
	suite.server.TeardownTestServer(suite.T())
}

// SetupTest は各テストの前に実行される
func (suite *WorkflowIntegrationTestSuite) SetupTest() {
	// 各テスト前にデータをクリーンアップしてシード
	suite.server.TestDB.CleanupTestData(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
}

// TestCompleteAuthenticationFlow は完全な認証フローをテストする
func (suite *WorkflowIntegrationTestSuite) TestCompleteAuthenticationFlow() {
	// 要件1.1, 1.2, 1.3, 1.4の統合テスト

	// 1. 無効な認証情報でログイン試行
	invalidLogin := map[string]string{
		"username": "admin",
		"password": "wrongpassword",
	}
	w1 := suite.server.MakeRequest(suite.T(), "POST", "/api/auth/login", invalidLogin, nil)
	suite.server.AssertErrorResponse(suite.T(), w1, http.StatusUnauthorized, "認証に失敗しました")

	// 2. 正しい認証情報でログイン
	validLogin := map[string]string{
		"username": "admin",
		"password": "password",
	}
	w2 := suite.server.MakeRequest(suite.T(), "POST", "/api/auth/login", validLogin, nil)
	suite.Equal(http.StatusOK, w2.Code)

	var loginResponse map[string]interface{}
	err := w2.Body.UnmarshalJSON(&loginResponse)
	suite.NoError(err)
	token := loginResponse["token"].(string)

	// 3. トークンを使用して保護されたエンドポイントにアクセス
	w3 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments", nil, token)
	suite.Equal(http.StatusOK, w3.Code)

	// 4. トークンをリフレッシュ
	refreshRequest := map[string]string{
		"token": token,
	}
	w4 := suite.server.MakeRequest(suite.T(), "POST", "/api/auth/refresh", refreshRequest, nil)
	suite.Equal(http.StatusOK, w4.Code)

	var refreshResponse map[string]interface{}
	err = w4.Body.UnmarshalJSON(&refreshResponse)
	suite.NoError(err)
	newToken := refreshResponse["token"].(string)

	// 5. 新しいトークンで保護されたエンドポイントにアクセス
	w5 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments", nil, newToken)
	suite.Equal(http.StatusOK, w5.Code)

	// 6. 無効なトークンでアクセス試行
	w6 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments", nil, "invalid.token")
	suite.server.AssertErrorResponse(suite.T(), w6, http.StatusUnauthorized, "無効または期限切れのトークンです")
}

// TestCompleteTournamentManagementFlow は完全なトーナメント管理フローをテストする
func (suite *WorkflowIntegrationTestSuite) TestCompleteTournamentManagementFlow() {
	// 要件2.1, 2.2, 2.4の統合テスト

	// 1. 全トーナメントを取得
	w1 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments", nil, suite.token)
	suite.Equal(http.StatusOK, w1.Code)

	var tournamentsResponse map[string]interface{}
	err := w1.Body.UnmarshalJSON(&tournamentsResponse)
	suite.NoError(err)
	tournaments := tournamentsResponse["data"].([]interface{})
	suite.GreaterOrEqual(len(tournaments), 3, "3つのスポーツのトーナメントが必要です")

	// 2. バレーボールトーナメントの詳細を取得
	w2 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/volleyball", nil, suite.token)
	suite.Equal(http.StatusOK, w2.Code)

	// 3. バレーボールトーナメントのブラケットを取得
	w3 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/volleyball/bracket", nil, suite.token)
	suite.Equal(http.StatusOK, w3.Code)

	var bracketResponse map[string]interface{}
	err = w3.Body.UnmarshalJSON(&bracketResponse)
	suite.NoError(err)
	initialBracket := bracketResponse["data"]

	// 4. 試合結果を提出してブラケットを更新
	matchID := suite.getFirstPendingMatchID("volleyball")
	resultRequest := map[string]interface{}{
		"score1": 3,
		"score2": 1,
		"winner": "チームA",
	}

	w4 := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/matches/"+matchID+"/result", resultRequest, suite.token)
	suite.Equal(http.StatusOK, w4.Code)

	// 5. 更新されたブラケットを取得
	w5 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/volleyball/bracket", nil, suite.token)
	suite.Equal(http.StatusOK, w5.Code)

	var updatedBracketResponse map[string]interface{}
	err = w5.Body.UnmarshalJSON(&updatedBracketResponse)
	suite.NoError(err)
	updatedBracket := updatedBracketResponse["data"]

	// ブラケットが更新されていることを確認
	suite.NotEqual(initialBracket, updatedBracket, "ブラケットが更新されていません")
}

// TestTableTennisFormatSwitchingFlow は卓球の形式切り替えフローをテストする
func (suite *WorkflowIntegrationTestSuite) TestTableTennisFormatSwitchingFlow() {
	// 要件5.2, 5.5の統合テスト

	// 1. 初期の卓球トーナメント形式を確認
	w1 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/table_tennis", nil, suite.token)
	suite.Equal(http.StatusOK, w1.Code)

	var initialResponse map[string]interface{}
	err := w1.Body.UnmarshalJSON(&initialResponse)
	suite.NoError(err)
	initialTournament := initialResponse["data"].(map[string]interface{})
	suite.Equal("standard", initialTournament["format"])

	// 2. 初期ブラケットを取得
	w2 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/table_tennis/bracket", nil, suite.token)
	suite.Equal(http.StatusOK, w2.Code)

	// 3. 雨天時形式に切り替え
	formatRequest := map[string]interface{}{
		"format": "rainy",
	}

	w3 := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/tournaments/table_tennis/format", formatRequest, suite.token)
	suite.Equal(http.StatusOK, w3.Code)

	// 4. 形式が変更されたことを確認
	w4 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/table_tennis", nil, suite.token)
	suite.Equal(http.StatusOK, w4.Code)

	var updatedResponse map[string]interface{}
	err = w4.Body.UnmarshalJSON(&updatedResponse)
	suite.NoError(err)
	updatedTournament := updatedResponse["data"].(map[string]interface{})
	suite.Equal("rainy", updatedTournament["format"])

	// 5. 更新されたブラケットを取得
	w5 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/table_tennis/bracket", nil, suite.token)
	suite.Equal(http.StatusOK, w5.Code)

	// 6. 標準形式に戻す
	standardFormatRequest := map[string]interface{}{
		"format": "standard",
	}

	w6 := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/tournaments/table_tennis/format", standardFormatRequest, suite.token)
	suite.Equal(http.StatusOK, w6.Code)

	// 7. 形式が戻されたことを確認
	w7 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/table_tennis", nil, suite.token)
	suite.Equal(http.StatusOK, w7.Code)

	var finalResponse map[string]interface{}
	err = w7.Body.UnmarshalJSON(&finalResponse)
	suite.NoError(err)
	finalTournament := finalResponse["data"].(map[string]interface{})
	suite.Equal("standard", finalTournament["format"])
}

// TestCompleteMatchManagementFlow は完全な試合管理フローをテストする
func (suite *WorkflowIntegrationTestSuite) TestCompleteMatchManagementFlow() {
	// 要件2.3, 2.4の統合テスト

	// 1. 全試合を取得
	w1 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/matches", nil, suite.token)
	suite.Equal(http.StatusOK, w1.Code)

	// 2. バレーボールの試合のみを取得
	w2 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/matches/volleyball", nil, suite.token)
	suite.Equal(http.StatusOK, w2.Code)

	var volleyballMatches map[string]interface{}
	err := w2.Body.UnmarshalJSON(&volleyballMatches)
	suite.NoError(err)
	matches := volleyballMatches["data"].([]interface{})

	// 3. 未完了の試合を見つける
	var pendingMatchID string
	for _, match := range matches {
		matchData := match.(map[string]interface{})
		if matchData["status"] == "pending" {
			pendingMatchID = string(rune(int(matchData["id"].(float64))))
			break
		}
	}
	suite.NotEmpty(pendingMatchID, "未完了の試合が見つかりません")

	// 4. 試合の詳細を取得
	w3 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/matches/match/"+pendingMatchID, nil, suite.token)
	suite.Equal(http.StatusOK, w3.Code)

	var matchDetail map[string]interface{}
	err = w3.Body.UnmarshalJSON(&matchDetail)
	suite.NoError(err)
	matchData := matchDetail["data"].(map[string]interface{})
	suite.Equal("pending", matchData["status"])

	// 5. 試合結果を提出
	resultRequest := map[string]interface{}{
		"score1": 3,
		"score2": 2,
		"winner": matchData["team1"].(string),
	}

	w4 := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/matches/"+pendingMatchID+"/result", resultRequest, suite.token)
	suite.Equal(http.StatusOK, w4.Code)

	// 6. 試合が完了状態になったことを確認
	w5 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/matches/match/"+pendingMatchID, nil, suite.token)
	suite.Equal(http.StatusOK, w5.Code)

	var completedMatchDetail map[string]interface{}
	err = w5.Body.UnmarshalJSON(&completedMatchDetail)
	suite.NoError(err)
	completedMatchData := completedMatchDetail["data"].(map[string]interface{})
	suite.Equal("completed", completedMatchData["status"])
	suite.Equal(float64(3), completedMatchData["score1"])
	suite.Equal(float64(2), completedMatchData["score2"])
	suite.Equal(matchData["team1"].(string), completedMatchData["winner"])

	// 7. 同じ試合に再度結果を提出しようとしてエラーになることを確認
	w6 := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/matches/"+pendingMatchID+"/result", resultRequest, suite.token)
	suite.server.AssertErrorResponse(suite.T(), w6, http.StatusBadRequest, "この試合は既に完了しています")
}

// TestMultiSportTournamentFlow は複数スポーツのトーナメントフローをテストする
func (suite *WorkflowIntegrationTestSuite) TestMultiSportTournamentFlow() {
	// 要件5.1, 5.2, 5.3の統合テスト

	sports := []string{"volleyball", "table_tennis", "soccer"}

	for _, sport := range sports {
		suite.Run("スポーツ_"+sport, func() {
			// 1. スポーツ別トーナメントを取得
			w1 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/"+sport, nil, suite.token)
			suite.Equal(http.StatusOK, w1.Code)

			var tournamentResponse map[string]interface{}
			err := w1.Body.UnmarshalJSON(&tournamentResponse)
			suite.NoError(err)
			tournament := tournamentResponse["data"].(map[string]interface{})
			suite.Equal(sport, tournament["sport"])

			// 2. スポーツ別ブラケットを取得
			w2 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/"+sport+"/bracket", nil, suite.token)
			suite.Equal(http.StatusOK, w2.Code)

			var bracketResponse map[string]interface{}
			err = w2.Body.UnmarshalJSON(&bracketResponse)
			suite.NoError(err)
			bracket := bracketResponse["data"].(map[string]interface{})
			suite.Equal(sport, bracket["sport"])

			// 3. スポーツ別試合を取得
			w3 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/matches/"+sport, nil, suite.token)
			suite.Equal(http.StatusOK, w3.Code)

			var matchesResponse map[string]interface{}
			err = w3.Body.UnmarshalJSON(&matchesResponse)
			suite.NoError(err)
			matches := matchesResponse["data"].([]interface{})
			suite.GreaterOrEqual(len(matches), 1, sport+"の試合が見つかりません")

			// 4. 各スポーツで試合結果を提出
			matchID := suite.getFirstPendingMatchID(sport)
			if matchID != "" {
				resultRequest := map[string]interface{}{
					"score1": 2,
					"score2": 1,
					"winner": "チームA",
				}

				w4 := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/matches/"+matchID+"/result", resultRequest, suite.token)
				suite.Equal(http.StatusOK, w4.Code)
			}
		})
	}
}

// TestErrorHandlingFlow はエラーハンドリングフローをテストする
func (suite *WorkflowIntegrationTestSuite) TestErrorHandlingFlow() {
	// 要件7.2, 7.4, 7.5の統合テスト

	// 1. 存在しないエンドポイントへのアクセス
	w1 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/nonexistent", nil, suite.token)
	suite.Equal(http.StatusNotFound, w1.Code)

	// 2. 無効なJSONでのリクエスト
	w2 := suite.server.MakeRequest(suite.T(), "POST", "/api/auth/login", "invalid json", nil)
	suite.server.AssertErrorResponse(suite.T(), w2, http.StatusBadRequest, "無効なリクエスト形式です")

	// 3. 存在しないリソースへのアクセス
	w3 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/nonexistent_sport", nil, suite.token)
	suite.server.AssertErrorResponse(suite.T(), w3, http.StatusNotFound, "トーナメントが見つかりません")

	// 4. 権限のないアクションの実行
	w4 := suite.server.MakeRequest(suite.T(), "POST", "/api/tournaments", map[string]interface{}{
		"sport":  "volleyball",
		"format": "standard",
	}, nil)
	suite.server.AssertErrorResponse(suite.T(), w4, http.StatusUnauthorized, "認証が必要です")

	// 5. 無効なデータでのリクエスト
	w5 := suite.server.MakeAuthenticatedRequest(suite.T(), "POST", "/api/matches", map[string]interface{}{
		"tournament_id": 99999, // 存在しないトーナメントID
		"round":         "quarterfinal",
		"team1":         "チームA",
		"team2":         "チームB",
	}, suite.token)
	suite.server.AssertErrorResponse(suite.T(), w5, http.StatusBadRequest, "無効な試合データです")
}

// TestHealthCheckAndSystemStatus はヘルスチェックとシステム状態をテストする
func (suite *WorkflowIntegrationTestSuite) TestHealthCheckAndSystemStatus() {
	// ヘルスチェックエンドポイント
	w := suite.server.MakeRequest(suite.T(), "GET", "/health", nil, nil)
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := w.Body.UnmarshalJSON(&response)
	suite.NoError(err)

	suite.Equal("ok", response["status"])
	suite.Contains(response["message"], "サーバーは正常に動作しています")
}

// getFirstPendingMatchID は指定されたスポーツの最初の未完了試合IDを取得する
func (suite *WorkflowIntegrationTestSuite) getFirstPendingMatchID(sport string) string {
	row := suite.server.TestDB.QueryRow(suite.T(), `
		SELECT m.id FROM matches m 
		JOIN tournaments t ON m.tournament_id = t.id 
		WHERE t.sport = ? AND m.status = 'pending' 
		LIMIT 1
	`, sport)
	var id int
	err := row.Scan(&id)
	if err != nil {
		return "" // 未完了の試合がない場合
	}
	return string(rune(id))
}

// TestWorkflowIntegrationTestSuite はワークフロー統合テストスイートを実行する
func TestWorkflowIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowIntegrationTestSuite))
}