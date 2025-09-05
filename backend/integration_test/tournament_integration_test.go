// Package integration_test はトーナメント機能の統合テストを提供する
package integration_test

import (
	"net/http"
	"testing"

	"backend/internal/testutil"

	"github.com/stretchr/testify/suite"
)

// TournamentIntegrationTestSuite はトーナメント機能の統合テストスイート
type TournamentIntegrationTestSuite struct {
	suite.Suite
	server *testutil.TestServer
	token  string
}

// SetupSuite はテストスイートのセットアップを行う
func (suite *TournamentIntegrationTestSuite) SetupSuite() {
	suite.server = testutil.SetupTestServer(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
	suite.token = suite.server.LoginAndGetToken(suite.T())
}

// TearDownSuite はテストスイートのクリーンアップを行う
func (suite *TournamentIntegrationTestSuite) TearDownSuite() {
	suite.server.TeardownTestServer(suite.T())
}

// SetupTest は各テストの前に実行される
func (suite *TournamentIntegrationTestSuite) SetupTest() {
	// 各テスト前にデータをクリーンアップしてシード
	suite.server.TestDB.CleanupTestData(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
}

// TestGetTournaments は全トーナメント取得をテストする
func (suite *TournamentIntegrationTestSuite) TestGetTournaments() {
	// 要件4.2: ユーザーがトーナメントデータを要求した場合、システムはすべてのスポーツの現在のブラケット状況を返す
	w := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments", nil, suite.token)

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
	suite.GreaterOrEqual(len(data), 3, "3つのスポーツのトーナメントが含まれていません")
}

// TestGetTournamentBySport はスポーツ別トーナメント取得をテストする
func (suite *TournamentIntegrationTestSuite) TestGetTournamentBySport() {
	// 要件4.4: 特定のスポーツのトーナメントデータが要求された場合、システムは関連するトーナメント情報のみを返す
	sports := []string{"volleyball", "table_tennis", "soccer"}

	for _, sport := range sports {
		suite.Run("スポーツ_"+sport, func() {
			w := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/"+sport, nil, suite.token)

			suite.Equal(http.StatusOK, w.Code)

			var response map[string]interface{}
			err := w.Body.UnmarshalJSON(&response)
			suite.NoError(err)

			suite.Contains(response, "success")
			suite.Equal(true, response["success"])
			suite.Contains(response, "data")

			// トーナメントデータの検証
			data, ok := response["data"].(map[string]interface{})
			suite.True(ok, "データがオブジェクトではありません")
			suite.Contains(data, "sport")
			suite.Equal(sport, data["sport"])
		})
	}
}

// TestGetTournamentBySport_InvalidSport は無効なスポーツでのトーナメント取得をテストする
func (suite *TournamentIntegrationTestSuite) TestGetTournamentBySport_InvalidSport() {
	w := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/invalid_sport", nil, suite.token)

	suite.server.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "トーナメントが見つかりません")
}

// TestGetTournamentBracket はトーナメントブラケット取得をテストする
func (suite *TournamentIntegrationTestSuite) TestGetTournamentBracket() {
	// 要件4.1: トーナメントデータが更新された場合、システムはフロントエンドクライアントに更新されたブラケット情報を提供する
	sports := []string{"volleyball", "table_tennis", "soccer"}

	for _, sport := range sports {
		suite.Run("ブラケット_"+sport, func() {
			w := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/"+sport+"/bracket", nil, suite.token)

			suite.Equal(http.StatusOK, w.Code)

			var response map[string]interface{}
			err := w.Body.UnmarshalJSON(&response)
			suite.NoError(err)

			suite.Contains(response, "success")
			suite.Equal(true, response["success"])
			suite.Contains(response, "data")

			// ブラケットデータの検証
			data, ok := response["data"].(map[string]interface{})
			suite.True(ok, "ブラケットデータがオブジェクトではありません")
			suite.Contains(data, "tournament_id")
			suite.Contains(data, "sport")
			suite.Contains(data, "rounds")
			suite.Equal(sport, data["sport"])
		})
	}
}

// TestCreateTournament は新規トーナメント作成をテストする（管理者のみ）
func (suite *TournamentIntegrationTestSuite) TestCreateTournament() {
	// 要件2.1: ダッシュボードにアクセスした場合、システムはバレーボール、卓球、8人制サッカーの3つのスポーツから選択するオプションを提供する
	tournamentRequest := map[string]interface{}{
		"sport":  "volleyball",
		"format": "standard",
		"status": "active",
	}

	w := suite.server.MakeAuthenticatedRequest(suite.T(), "POST", "/api/tournaments", tournamentRequest, suite.token)

	suite.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := w.Body.UnmarshalJSON(&response)
	suite.NoError(err)

	suite.Contains(response, "success")
	suite.Equal(true, response["success"])
	suite.Contains(response, "data")

	// 作成されたトーナメントデータの検証
	data, ok := response["data"].(map[string]interface{})
	suite.True(ok, "作成されたトーナメントデータがオブジェクトではありません")
	suite.Contains(data, "id")
	suite.Contains(data, "sport")
	suite.Equal("volleyball", data["sport"])
}

// TestCreateTournament_Unauthorized は認証なしでのトーナメント作成をテストする
func (suite *TournamentIntegrationTestSuite) TestCreateTournament_Unauthorized() {
	tournamentRequest := map[string]interface{}{
		"sport":  "volleyball",
		"format": "standard",
		"status": "active",
	}

	w := suite.server.MakeRequest(suite.T(), "POST", "/api/tournaments", tournamentRequest, nil)

	suite.server.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "認証が必要です")
}

// TestUpdateTournament はトーナメント更新をテストする（管理者のみ）
func (suite *TournamentIntegrationTestSuite) TestUpdateTournament() {
	// まず既存のトーナメントIDを取得
	tournamentID := suite.getTournamentID("volleyball")

	updateRequest := map[string]interface{}{
		"format": "updated_format",
		"status": "completed",
	}

	w := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/tournaments/"+tournamentID, updateRequest, suite.token)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := w.Body.UnmarshalJSON(&response)
	suite.NoError(err)

	suite.Contains(response, "success")
	suite.Equal(true, response["success"])
}

// TestSwitchTournamentFormat は卓球のトーナメント形式切り替えをテストする
func (suite *TournamentIntegrationTestSuite) TestSwitchTournamentFormat() {
	// 要件5.5: 天候条件が卓球形式に影響する場合、システムはトーナメント形式間の切り替えを許可する
	formatRequest := map[string]interface{}{
		"format": "rainy",
	}

	w := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/tournaments/table_tennis/format", formatRequest, suite.token)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := w.Body.UnmarshalJSON(&response)
	suite.NoError(err)

	suite.Contains(response, "success")
	suite.Equal(true, response["success"])

	// 形式が変更されたことを確認
	w2 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/table_tennis", nil, suite.token)
	suite.Equal(http.StatusOK, w2.Code)

	var tournamentResponse map[string]interface{}
	err = w2.Body.UnmarshalJSON(&tournamentResponse)
	suite.NoError(err)

	data, ok := tournamentResponse["data"].(map[string]interface{})
	suite.True(ok)
	suite.Equal("rainy", data["format"])
}

// TestSwitchTournamentFormat_InvalidSport は無効なスポーツでの形式切り替えをテストする
func (suite *TournamentIntegrationTestSuite) TestSwitchTournamentFormat_InvalidSport() {
	formatRequest := map[string]interface{}{
		"format": "rainy",
	}

	// バレーボールでは形式切り替えはサポートされていない
	w := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/tournaments/volleyball/format", formatRequest, suite.token)

	suite.server.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "このスポーツでは形式切り替えはサポートされていません")
}

// TestTournamentWorkflow はトーナメント進行ワークフローをテストする
func (suite *TournamentIntegrationTestSuite) TestTournamentWorkflow() {
	// 要件2.2: スポーツを選択した場合、システムは適切なトーナメントブラケット構造を表示する
	// 要件2.4: 試合結果が提出された場合、システムはトーナメントブラケットを更新し、勝者を次のラウンドに進出させる

	// 1. 初期ブラケットを取得
	w1 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/volleyball/bracket", nil, suite.token)
	suite.Equal(http.StatusOK, w1.Code)

	var initialBracket map[string]interface{}
	err := w1.Body.UnmarshalJSON(&initialBracket)
	suite.NoError(err)

	// 2. 試合結果を提出
	matchID := suite.getFirstPendingMatchID("volleyball")
	matchResult := map[string]interface{}{
		"score1": 3,
		"score2": 1,
		"winner": "チームA",
	}

	w2 := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", "/api/matches/"+matchID+"/result", matchResult, suite.token)
	suite.Equal(http.StatusOK, w2.Code)

	// 3. 更新されたブラケットを取得
	w3 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments/volleyball/bracket", nil, suite.token)
	suite.Equal(http.StatusOK, w3.Code)

	var updatedBracket map[string]interface{}
	err = w3.Body.UnmarshalJSON(&updatedBracket)
	suite.NoError(err)

	// ブラケットが更新されていることを確認（詳細な検証は実装に依存）
	suite.NotEqual(initialBracket, updatedBracket, "ブラケットが更新されていません")
}

// getTournamentID は指定されたスポーツのトーナメントIDを取得する
func (suite *TournamentIntegrationTestSuite) getTournamentID(sport string) string {
	row := suite.server.TestDB.QueryRow(suite.T(), "SELECT id FROM tournaments WHERE sport = ?", sport)
	var id int
	err := row.Scan(&id)
	suite.NoError(err)
	return string(rune(id))
}

// getFirstPendingMatchID は指定されたスポーツの最初の未完了試合IDを取得する
func (suite *TournamentIntegrationTestSuite) getFirstPendingMatchID(sport string) string {
	row := suite.server.TestDB.QueryRow(suite.T(), `
		SELECT m.id FROM matches m 
		JOIN tournaments t ON m.tournament_id = t.id 
		WHERE t.sport = ? AND m.status = 'pending' 
		LIMIT 1
	`, sport)
	var id int
	err := row.Scan(&id)
	suite.NoError(err)
	return string(rune(id))
}

// TestTournamentIntegrationTestSuite はトーナメント統合テストスイートを実行する
func TestTournamentIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TournamentIntegrationTestSuite))
}