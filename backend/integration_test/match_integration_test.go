// Package integration_test は試合機能の統合テストを提供する
package integration_test

import (
	"net/http"
	"testing"

	"backend/internal/testutil"

	"github.com/stretchr/testify/suite"
)

// MatchIntegrationTestSuite は試合機能の統合テストスイート
type MatchIntegrationTestSuite struct {
	suite.Suite
	server *testutil.TestServer
	token  string
}

// SetupSuite はテストスイートのセットアップを行う
func (suite *MatchIntegrationTestSuite) SetupSuite() {
	suite.server = testutil.SetupTestServer(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
	suite.token = suite.server.LoginAndGetToken(suite.T())
}

// TearDownSuite はテストスイートのクリーンアップを行う
func (suite *MatchIntegrationTestSuite) TearDownSuite() {
	suite.server.TeardownTestServer(suite.T())
}

// SetupTest は各テストの前に実行される
func (suite *MatchIntegrationTestSuite) SetupTest() {
	// 各テスト前にデータをクリーンアップしてシード
	suite.server.TestDB.CleanupTestData(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
}

// TestGetMatches は全試合取得をテストする
func (suite *MatchIntegrationTestSuite) TestGetMatches() {
	w := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/matches", nil, suite.token)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := w.Body.UnmarshalJSON(&response)
	suite.NoError(err)

	suite.Contains(response, "success")
	suite.Equal(true, response["success"])
	suite.Contains(response, "data")

	// データが配列であることを確認
	data, ok := response["data"].([]interface{})
	suite.True(ok, "データが配列ではありません")
	suite.GreaterOrEqual(len(data), 4, "シードされた試合が含まれていません")
}

// TestGetMatchesBySport はスポーツ別試合取得をテストする
func (suite *MatchIntegrationTestSuite) TestGetMatchesBySport() {
	// 要件2.3: 試合結果を入力する場合、システムはチーム名とそれぞれのスコアの両方を要求する
	sports := []string{"volleyball", "table_tennis", "soccer"}

	for _, sport := range sports {
		suite.Run("スポーツ_"+sport, func() {
			w := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/matches/"+sport, nil, suite.token)

			suite.Equal(http.StatusOK, w.Code)

			var response map[string]interface{}
			err := w.Body.UnmarshalJSON(&response)
			suite.NoError(err)

			suite.Contains(response, "success")
			suite.Equal(true, response["success"])
			suite.Contains(response, "data")

			// スポーツ別の試合データの検証
			data, ok := response["data"].([]interface{})
			suite.True(ok, "データが配列ではありません")

			// 各試合がそのスポーツのものであることを確認
			for _, match := range data {
				matchData, ok := match.(map[string]interface{})
				suite.True(ok, "試合データがオブジェクトではありません")
				suite.Contains(matchData, "tournament_id")
				suite.Contains(matchData, "team1")
				suite.Contains(matchData, "team2")
			}
		})
	}
}

// TestGetMatch は特定試合取得をテストする
func (suite *MatchIntegrationTestSuite) TestGetMatch() {
	// 既存の試合IDを取得
	matchID := suite.getFirstMatchID()

	w := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/matches/match/"+matchID, nil, suite.token)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := w.Body.UnmarshalJSON(&response)
	suite.NoError(err)

	suite.Contains(response, "success")
	suite.Equal(true, response["success"])
	suite.Contains(response, "data")

	// 試合データの検証
	data, ok := response["data"].(map[string]interface{})
	suite.True(ok, "試合データがオブジェクトではありません")
	suite.Contains(data, "id")
	suite.Contains(data, "team1")
	suite.Contains(data, "team2")
	suite.Contains(data, "status")
}

// TestGetMatch_NotFound は存在しない試合の取得をテストする
func (suite *MatchIntegrationTestSuite) TestGetMatch_NotFound() {
	w := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/matches/match/99999", nil, suite.token)

	suite.server.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "試合が見つかりません")
}

// TestCreateMatch は新規試合作成をテストする（管理者のみ）
func (suite *MatchIntegrationTestSuite) TestCreateMatch() {
	// 既存のトーナメントIDを取得
	tournamentID := suite.getTournamentID("volleyball")

	matchRequest := map[string]interface{}{
		"tournament_id": tournamentID,
		"round":         "quarterfinal",
		"team1":         "新チームA",
		"team2":         "新チームB",
		"scheduled_at":  "2024-12-01T10:00:00Z",
	}

	w := suite.server.MakeAuthenticatedRequest(suite.T(), "POST", "/api/matches", matchRequest, suite.token)

	suite.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := w.Body.UnmarshalJSON(&response)
	suite.NoError(err)

	suite.Contains(response, "success")
	suite.Equal(true, response["success"])
	suite.Contains(response, "data")

	// 作成された試合データの検証
	data, ok := response["data"].(map[string]interface{})
	suite.True(ok, "作成された試合データがオブジェクトではありません")
	suite.Contains(data, "id")
	suite.Equal("新チームA", data["team1"])
	suite.Equal("新チームB", data["team2"])
}

// TestCreateMatch_Unauthorized は認証なしでの試合作成をテストする
func (suite *MatchIntegrationTestSuite) TestCreateMatch_Unauthorized() {
	matchRequest := map[string]interface{}{
		"tournament_id": 1,
		"round":         "quarterfinal",
		"team1":         "チームA",
		"team2":         "チームB",
		"scheduled_at":  "2024-12-01T10:00:00Z",
	}

	w := suite.server.MakeRequest(suite.T(), "POST", "/api/matches", matchRequest, nil)

	suite.server.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "認証が必要です")
}

// TestUpdateMatch は試合更新をテストする（管理者のみ）
func (suite *MatchIntegrationTestSuite) TestUpdateMatch() {
	matchID := suite.getFirstMatchID()

	updateRequest := map[string]interface{}{
		"team1": "更新チームA",
		"team2": "更新チームB",
		"round": "semifinal",
	}

	w := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/matches/"+matchID, updateRequest, suite.token)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := w.Body.UnmarshalJSON(&response)
	suite.NoError(err)

	suite.Contains(response, "success")
	suite.Equal(true, response["success"])

	// 更新されたことを確認
	w2 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/matches/match/"+matchID, nil, suite.token)
	suite.Equal(http.StatusOK, w2.Code)

	var matchResponse map[string]interface{}
	err = w2.Body.UnmarshalJSON(&matchResponse)
	suite.NoError(err)

	data, ok := matchResponse["data"].(map[string]interface{})
	suite.True(ok)
	suite.Equal("更新チームA", data["team1"])
	suite.Equal("更新チームB", data["team2"])
}

// TestSubmitMatchResult は試合結果提出をテストする
func (suite *MatchIntegrationTestSuite) TestSubmitMatchResult() {
	// 要件2.3: 試合結果を入力する場合、システムはチーム名とそれぞれのスコアの両方を要求する
	// 要件2.4: 試合結果が提出された場合、システムはトーナメントブラケットを更新し、勝者を次のラウンドに進出させる
	matchID := suite.getFirstPendingMatchID()

	resultRequest := map[string]interface{}{
		"score1": 3,
		"score2": 1,
		"winner": "チームA",
	}

	w := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/matches/"+matchID+"/result", resultRequest, suite.token)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := w.Body.UnmarshalJSON(&response)
	suite.NoError(err)

	suite.Contains(response, "success")
	suite.Equal(true, response["success"])

	// 試合結果が更新されたことを確認
	w2 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/matches/match/"+matchID, nil, suite.token)
	suite.Equal(http.StatusOK, w2.Code)

	var matchResponse map[string]interface{}
	err = w2.Body.UnmarshalJSON(&matchResponse)
	suite.NoError(err)

	data, ok := matchResponse["data"].(map[string]interface{})
	suite.True(ok)
	suite.Equal(float64(3), data["score1"]) // JSONでは数値はfloat64になる
	suite.Equal(float64(1), data["score2"])
	suite.Equal("チームA", data["winner"])
	suite.Equal("completed", data["status"])
}

// TestSubmitMatchResult_InvalidData は無効なデータでの試合結果提出をテストする
func (suite *MatchIntegrationTestSuite) TestSubmitMatchResult_InvalidData() {
	// 要件2.5: 試合結果が無効な場合、システムは検証エラーを返す
	matchID := suite.getFirstPendingMatchID()

	testCases := []struct {
		name    string
		request map[string]interface{}
	}{
		{
			name: "スコアなし",
			request: map[string]interface{}{
				"winner": "チームA",
			},
		},
		{
			name: "勝者なし",
			request: map[string]interface{}{
				"score1": 3,
				"score2": 1,
			},
		},
		{
			name: "負のスコア",
			request: map[string]interface{}{
				"score1": -1,
				"score2": 2,
				"winner": "チームA",
			},
		},
		{
			name: "無効な勝者",
			request: map[string]interface{}{
				"score1": 1,
				"score2": 3,
				"winner": "存在しないチーム",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			w := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/matches/"+matchID+"/result", tc.request, suite.token)
			suite.server.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "無効な試合結果です")
		})
	}
}

// TestSubmitMatchResult_AlreadyCompleted は既に完了した試合への結果提出をテストする
func (suite *MatchIntegrationTestSuite) TestSubmitMatchResult_AlreadyCompleted() {
	matchID := suite.getFirstPendingMatchID()

	// まず試合結果を提出
	resultRequest := map[string]interface{}{
		"score1": 3,
		"score2": 1,
		"winner": "チームA",
	}

	w1 := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/matches/"+matchID+"/result", resultRequest, suite.token)
	suite.Equal(http.StatusOK, w1.Code)

	// 再度同じ試合に結果を提出
	w2 := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/matches/"+matchID+"/result", resultRequest, suite.token)
	suite.server.AssertErrorResponse(suite.T(), w2, http.StatusBadRequest, "この試合は既に完了しています")
}

// TestMatchWorkflow は試合進行ワークフローをテストする
func (suite *MatchIntegrationTestSuite) TestMatchWorkflow() {
	// 1. 未完了の試合を取得
	w1 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/matches/volleyball", nil, suite.token)
	suite.Equal(http.StatusOK, w1.Code)

	var initialMatches map[string]interface{}
	err := w1.Body.UnmarshalJSON(&initialMatches)
	suite.NoError(err)

	// 2. 試合結果を提出
	matchID := suite.getFirstPendingMatchID()
	resultRequest := map[string]interface{}{
		"score1": 3,
		"score2": 1,
		"winner": "チームA",
	}

	w2 := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/matches/"+matchID+"/result", resultRequest, suite.token)
	suite.Equal(http.StatusOK, w2.Code)

	// 3. 試合リストが更新されていることを確認
	w3 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/matches/volleyball", nil, suite.token)
	suite.Equal(http.StatusOK, w3.Code)

	var updatedMatches map[string]interface{}
	err = w3.Body.UnmarshalJSON(&updatedMatches)
	suite.NoError(err)

	// 試合状態が変更されていることを確認
	suite.NotEqual(initialMatches, updatedMatches, "試合リストが更新されていません")
}

// TestMatchValidation は試合データの検証をテストする
func (suite *MatchIntegrationTestSuite) TestMatchValidation() {
	tournamentID := suite.getTournamentID("volleyball")

	testCases := []struct {
		name    string
		request map[string]interface{}
	}{
		{
			name: "チーム名が空",
			request: map[string]interface{}{
				"tournament_id": tournamentID,
				"round":         "quarterfinal",
				"team1":         "",
				"team2":         "チームB",
				"scheduled_at":  "2024-12-01T10:00:00Z",
			},
		},
		{
			name: "同じチーム名",
			request: map[string]interface{}{
				"tournament_id": tournamentID,
				"round":         "quarterfinal",
				"team1":         "チームA",
				"team2":         "チームA",
				"scheduled_at":  "2024-12-01T10:00:00Z",
			},
		},
		{
			name: "無効なトーナメントID",
			request: map[string]interface{}{
				"tournament_id": 99999,
				"round":         "quarterfinal",
				"team1":         "チームA",
				"team2":         "チームB",
				"scheduled_at":  "2024-12-01T10:00:00Z",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			w := suite.server.MakeAuthenticatedRequest(suite.T(), "POST", "/api/matches", tc.request, suite.token)
			suite.server.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "無効な試合データです")
		})
	}
}

// getFirstMatchID は最初の試合IDを取得する
func (suite *MatchIntegrationTestSuite) getFirstMatchID() string {
	row := suite.server.TestDB.QueryRow(suite.T(), "SELECT id FROM matches LIMIT 1")
	var id int
	err := row.Scan(&id)
	suite.NoError(err)
	return string(rune(id))
}

// getFirstPendingMatchID は最初の未完了試合IDを取得する
func (suite *MatchIntegrationTestSuite) getFirstPendingMatchID() string {
	row := suite.server.TestDB.QueryRow(suite.T(), "SELECT id FROM matches WHERE status = 'pending' LIMIT 1")
	var id int
	err := row.Scan(&id)
	suite.NoError(err)
	return string(rune(id))
}

// getTournamentID は指定されたスポーツのトーナメントIDを取得する
func (suite *MatchIntegrationTestSuite) getTournamentID(sport string) int {
	row := suite.server.TestDB.QueryRow(suite.T(), "SELECT id FROM tournaments WHERE sport = ?", sport)
	var id int
	err := row.Scan(&id)
	suite.NoError(err)
	return id
}

// TestMatchIntegrationTestSuite は試合統合テストスイートを実行する
func TestMatchIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(MatchIntegrationTestSuite))
}