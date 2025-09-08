// Package integration_test はAPI契約テストを提供する
package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"backend/internal/testutil"

	"github.com/stretchr/testify/suite"
)

// ContractTestSuite はAPI契約テストスイート
type ContractTestSuite struct {
	suite.Suite
	server *testutil.TestServer
}

// SetupSuite はテストスイートのセットアップを行う
func (suite *ContractTestSuite) SetupSuite() {
	suite.server = testutil.SetupTestServer(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
}

// TearDownSuite はテストスイートのクリーンアップを行う
func (suite *ContractTestSuite) TearDownSuite() {
	suite.server.TeardownTestServer(suite.T())
}

// SetupTest は各テストの前に実行される
func (suite *ContractTestSuite) SetupTest() {
	// 各テスト前にデータをクリーンアップしてシード
	suite.server.TestDB.CleanupTestData(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
}

// APIResponse は統一されたAPIレスポンス構造体
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Message   string      `json:"message"`
	Code      int         `json:"code"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// LoginResponseData はログインレスポンスのデータ部分
type LoginResponseData struct {
	Token     string `json:"token"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	ExpiresAt string `json:"expires_at"`
}

// TournamentData はトーナメントデータ構造体
type TournamentData struct {
	ID        int    `json:"id"`
	Sport     string `json:"sport"`
	Format    string `json:"format"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// MatchData は試合データ構造体
type MatchData struct {
	ID           int     `json:"id"`
	TournamentID int     `json:"tournament_id"`
	Round        string  `json:"round"`
	Team1        string  `json:"team1"`
	Team2        string  `json:"team2"`
	Score1       *int    `json:"score1"`
	Score2       *int    `json:"score2"`
	Winner       *string `json:"winner"`
	Status       string  `json:"status"`
	ScheduledAt  string  `json:"scheduled_at"`
	CompletedAt  *string `json:"completed_at"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// validateAPIResponse は統一されたAPIレスポンス形式を検証する
func (suite *ContractTestSuite) validateAPIResponse(response map[string]interface{}, expectedSuccess bool) {
	// 必須フィールドの存在確認
	suite.Contains(response, "success", "レスポンスにsuccessフィールドが含まれていません")
	suite.Contains(response, "message", "レスポンスにmessageフィールドが含まれていません")
	suite.Contains(response, "code", "レスポンスにcodeフィールドが含まれていません")
	suite.Contains(response, "timestamp", "レスポンスにtimestampフィールドが含まれていません")

	// 型の検証
	suite.IsType(expectedSuccess, response["success"], "successフィールドの型が正しくありません")
	suite.IsType("", response["message"], "messageフィールドの型が正しくありません")
	suite.IsType(float64(0), response["code"], "codeフィールドの型が正しくありません")
	suite.IsType("", response["timestamp"], "timestampフィールドの型が正しくありません")

	// 値の検証
	suite.Equal(expectedSuccess, response["success"], "successフィールドの値が期待値と異なります")

	// タイムスタンプ形式の検証（ISO 8601）
	timestamp, ok := response["timestamp"].(string)
	suite.True(ok, "timestampフィールドが文字列ではありません")
	_, err := time.Parse(time.RFC3339, timestamp)
	suite.NoError(err, "timestampがISO 8601形式ではありません: %s", timestamp)

	// 成功時とエラー時の条件付きフィールド検証
	if expectedSuccess {
		suite.Contains(response, "data", "成功レスポンスにdataフィールドが含まれていません")
		suite.NotContains(response, "error", "成功レスポンスにerrorフィールドが含まれています")
	} else {
		suite.Contains(response, "error", "エラーレスポンスにerrorフィールドが含まれていません")
		suite.IsType("", response["error"], "errorフィールドの型が正しくありません")
		suite.NotEmpty(response["error"], "errorフィールドが空です")
	}
}

// validateDateTimeFormat は日時フィールドのISO 8601形式を検証する
func (suite *ContractTestSuite) validateDateTimeFormat(dateTime string, fieldName string) {
	_, err := time.Parse(time.RFC3339, dateTime)
	suite.NoError(err, "%sフィールドがISO 8601形式ではありません: %s", fieldName, dateTime)
}

// validateSportType はスポーツ種目の有効性を検証する
func (suite *ContractTestSuite) validateSportType(sport string) {
	validSports := []string{"volleyball", "table_tennis", "soccer"}
	suite.Contains(validSports, sport, "無効なスポーツ種目です: %s", sport)
}

// validateTournamentStatus はトーナメントステータスの有効性を検証する
func (suite *ContractTestSuite) validateTournamentStatus(status string) {
	validStatuses := []string{"registration", "active", "completed", "cancelled"}
	suite.Contains(validStatuses, status, "無効なトーナメントステータスです: %s", status)
}

// validateMatchStatus は試合ステータスの有効性を検証する
func (suite *ContractTestSuite) validateMatchStatus(status string) {
	validStatuses := []string{"pending", "in_progress", "completed", "cancelled"}
	suite.Contains(validStatuses, status, "無効な試合ステータスです: %s", status)
}

// TestContract_AuthLogin_Success は認証ログインの契約テスト
func (suite *ContractTestSuite) TestContract_AuthLogin_Success() {
	// 要件7.1: APIスキーマが定義された場合、システムはリクエスト・レスポンスの自動検証を行う
	loginRequest := map[string]string{
		"username": "admin",
		"password": "password",
	}

	w := suite.server.MakeRequest(suite.T(), "POST", "/api/v1/auth/login", loginRequest, nil)
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err, "レスポンスのJSONパースに失敗しました")

	// 統一レスポンス形式の検証
	suite.validateAPIResponse(response, true)

	// ログイン固有のデータ構造検証
	data, ok := response["data"].(map[string]interface{})
	suite.True(ok, "dataフィールドがオブジェクトではありません")

	// 必須フィールドの存在確認
	suite.Contains(data, "token", "ログインレスポンスにtokenフィールドが含まれていません")
	suite.Contains(data, "username", "ログインレスポンスにusernameフィールドが含まれていません")
	suite.Contains(data, "role", "ログインレスポンスにroleフィールドが含まれていません")
	suite.Contains(data, "expires_at", "ログインレスポンスにexpires_atフィールドが含まれていません")

	// 型の検証
	suite.IsType("", data["token"], "tokenフィールドの型が正しくありません")
	suite.IsType("", data["username"], "usernameフィールドの型が正しくありません")
	suite.IsType("", data["role"], "roleフィールドの型が正しくありません")
	suite.IsType("", data["expires_at"], "expires_atフィールドの型が正しくありません")

	// 値の検証
	suite.NotEmpty(data["token"], "tokenフィールドが空です")
	suite.Equal("admin", data["username"], "usernameフィールドの値が正しくありません")
	suite.Equal("admin", data["role"], "roleフィールドの値が正しくありません")

	// JWTトークン形式の検証（簡易）
	token := data["token"].(string)
	parts := strings.Split(token, ".")
	suite.Equal(3, len(parts), "JWTトークンの形式が正しくありません")

	// 有効期限の日時形式検証
	expiresAt := data["expires_at"].(string)
	suite.validateDateTimeFormat(expiresAt, "expires_at")
}

// TestContract_AuthLogin_ValidationError は認証ログインのバリデーションエラー契約テスト
func (suite *ContractTestSuite) TestContract_AuthLogin_ValidationError() {
	// 要件7.2: 契約テストが実行される場合、システムはフロントエンドとバックエンドの仕様整合性を検証する
	loginRequest := map[string]string{
		"username": "", // 空のユーザー名
		"password": "password",
	}

	w := suite.server.MakeRequest(suite.T(), "POST", "/api/v1/auth/login", loginRequest, nil)
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err, "レスポンスのJSONパースに失敗しました")

	// 統一エラーレスポンス形式の検証
	suite.validateAPIResponse(response, false)

	// エラーコードの検証
	errorCode := response["error"].(string)
	suite.Equal("VALIDATION_ERROR", errorCode, "エラーコードが正しくありません")
}

// TestContract_TournamentList は全トーナメント取得の契約テスト
func (suite *ContractTestSuite) TestContract_TournamentList() {
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments", nil, nil)
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err, "レスポンスのJSONパースに失敗しました")

	// 統一レスポンス形式の検証
	suite.validateAPIResponse(response, true)

	// リストレスポンス固有の検証
	data, ok := response["data"].([]interface{})
	suite.True(ok, "dataフィールドが配列ではありません")

	suite.Contains(response, "count", "リストレスポンスにcountフィールドが含まれていません")
	count, ok := response["count"].(float64)
	suite.True(ok, "countフィールドが数値ではありません")
	suite.Equal(float64(len(data)), count, "countフィールドの値がデータ配列の長さと一致しません")

	// 各トーナメントデータの構造検証
	if len(data) > 0 {
		tournament, ok := data[0].(map[string]interface{})
		suite.True(ok, "トーナメントデータがオブジェクトではありません")

		// 必須フィールドの存在確認
		requiredFields := []string{"id", "sport", "format", "status", "created_at", "updated_at"}
		for _, field := range requiredFields {
			suite.Contains(tournament, field, "トーナメントデータに%sフィールドが含まれていません", field)
		}

		// 型の検証
		suite.IsType(float64(0), tournament["id"], "idフィールドの型が正しくありません")
		suite.IsType("", tournament["sport"], "sportフィールドの型が正しくありません")
		suite.IsType("", tournament["format"], "formatフィールドの型が正しくありません")
		suite.IsType("", tournament["status"], "statusフィールドの型が正しくありません")
		suite.IsType("", tournament["created_at"], "created_atフィールドの型が正しくありません")
		suite.IsType("", tournament["updated_at"], "updated_atフィールドの型が正しくありません")

		// 値の検証
		suite.Greater(tournament["id"].(float64), float64(0), "idフィールドが正の値ではありません")
		suite.validateSportType(tournament["sport"].(string))
		suite.validateTournamentStatus(tournament["status"].(string))
		suite.validateDateTimeFormat(tournament["created_at"].(string), "created_at")
		suite.validateDateTimeFormat(tournament["updated_at"].(string), "updated_at")
	}
}

// TestContract_MatchList は試合一覧取得の契約テスト
func (suite *ContractTestSuite) TestContract_MatchList() {
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/matches/sport/volleyball", nil, nil)
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err, "レスポンスのJSONパースに失敗しました")

	// 統一レスポンス形式の検証
	suite.validateAPIResponse(response, true)

	// リストレスポンス固有の検証
	data, ok := response["data"].([]interface{})
	suite.True(ok, "dataフィールドが配列ではありません")

	// 各試合データの構造検証
	if len(data) > 0 {
		match, ok := data[0].(map[string]interface{})
		suite.True(ok, "試合データがオブジェクトではありません")

		// 必須フィールドの存在確認
		requiredFields := []string{"id", "tournament_id", "round", "team1", "team2", "status", "scheduled_at", "created_at", "updated_at"}
		for _, field := range requiredFields {
			suite.Contains(match, field, "試合データに%sフィールドが含まれていません", field)
		}

		// 型の検証
		suite.IsType(float64(0), match["id"], "idフィールドの型が正しくありません")
		suite.IsType(float64(0), match["tournament_id"], "tournament_idフィールドの型が正しくありません")
		suite.IsType("", match["round"], "roundフィールドの型が正しくありません")
		suite.IsType("", match["team1"], "team1フィールドの型が正しくありません")
		suite.IsType("", match["team2"], "team2フィールドの型が正しくありません")
		suite.IsType("", match["status"], "statusフィールドの型が正しくありません")
		suite.IsType("", match["scheduled_at"], "scheduled_atフィールドの型が正しくありません")
		suite.IsType("", match["created_at"], "created_atフィールドの型が正しくありません")
		suite.IsType("", match["updated_at"], "updated_atフィールドの型が正しくありません")

		// 値の検証
		suite.Greater(match["id"].(float64), float64(0), "idフィールドが正の値ではありません")
		suite.Greater(match["tournament_id"].(float64), float64(0), "tournament_idフィールドが正の値ではありません")
		suite.NotEmpty(match["team1"], "team1フィールドが空です")
		suite.NotEmpty(match["team2"], "team2フィールドが空です")
		suite.validateMatchStatus(match["status"].(string))
		suite.validateDateTimeFormat(match["scheduled_at"].(string), "scheduled_at")
		suite.validateDateTimeFormat(match["created_at"].(string), "created_at")
		suite.validateDateTimeFormat(match["updated_at"].(string), "updated_at")

		// null許可フィールドの検証
		if match["score1"] != nil {
			suite.IsType(float64(0), match["score1"], "score1フィールドの型が正しくありません")
			suite.GreaterOrEqual(match["score1"].(float64), float64(0), "score1フィールドが負の値です")
		}

		if match["score2"] != nil {
			suite.IsType(float64(0), match["score2"], "score2フィールドの型が正しくありません")
			suite.GreaterOrEqual(match["score2"].(float64), float64(0), "score2フィールドが負の値です")
		}

		if match["winner"] != nil {
			suite.IsType("", match["winner"], "winnerフィールドの型が正しくありません")
			suite.NotEmpty(match["winner"], "winnerフィールドが空です")
		}

		if match["completed_at"] != nil {
			suite.IsType("", match["completed_at"], "completed_atフィールドの型が正しくありません")
			suite.validateDateTimeFormat(match["completed_at"].(string), "completed_at")
		}
	}
}

// TestContract_CreateMatch は試合作成の契約テスト
func (suite *ContractTestSuite) TestContract_CreateMatch() {
	// 要件7.4: 統合テストが実行される場合、システムは実際のAPI呼び出しでエンドツーエンドの動作を検証する
	token := suite.server.LoginAndGetToken(suite.T())

	createRequest := map[string]interface{}{
		"tournament_id": 1,
		"round":         "quarterfinal",
		"team1":         "テストチームA",
		"team2":         "テストチームB",
		"scheduled_at":  "2024-12-01T10:00:00Z",
	}

	w := suite.server.MakeAuthenticatedRequest(suite.T(), "POST", "/api/v1/admin/matches", createRequest, token)
	suite.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err, "レスポンスのJSONパースに失敗しました")

	// 統一レスポンス形式の検証
	suite.validateAPIResponse(response, true)

	// 作成された試合データの検証
	data, ok := response["data"].(map[string]interface{})
	suite.True(ok, "dataフィールドがオブジェクトではありません")

	// 作成されたデータが入力データと一致することを確認
	suite.Equal(float64(1), data["tournament_id"], "tournament_idが正しく設定されていません")
	suite.Equal("quarterfinal", data["round"], "roundが正しく設定されていません")
	suite.Equal("テストチームA", data["team1"], "team1が正しく設定されていません")
	suite.Equal("テストチームB", data["team2"], "team2が正しく設定されていません")
	suite.Equal("pending", data["status"], "初期ステータスがpendingではありません")

	// 自動設定フィールドの検証
	suite.Greater(data["id"].(float64), float64(0), "idが自動生成されていません")
	suite.validateDateTimeFormat(data["created_at"].(string), "created_at")
	suite.validateDateTimeFormat(data["updated_at"].(string), "updated_at")
}

// TestContract_SubmitMatchResult は試合結果提出の契約テスト
func (suite *ContractTestSuite) TestContract_SubmitMatchResult() {
	token := suite.server.LoginAndGetToken(suite.T())

	// まず試合を作成
	createRequest := map[string]interface{}{
		"tournament_id": 1,
		"round":         "quarterfinal",
		"team1":         "チームA",
		"team2":         "チームB",
		"scheduled_at":  "2024-12-01T10:00:00Z",
	}

	createResp := suite.server.MakeAuthenticatedRequest(suite.T(), "POST", "/api/v1/admin/matches", createRequest, token)
	suite.Equal(http.StatusCreated, createResp.Code)

	var createResponse map[string]interface{}
	err := json.Unmarshal(createResp.Body.Bytes(), &createResponse)
	suite.NoError(err)

	matchData := createResponse["data"].(map[string]interface{})
	matchID := int(matchData["id"].(float64))

	// 試合結果を提出
	resultRequest := map[string]interface{}{
		"score1": 3,
		"score2": 1,
		"winner": "チームA",
	}

	w := suite.server.MakeAuthenticatedRequest(suite.T(), "PUT", fmt.Sprintf("/api/v1/admin/matches/%d/result", matchID), resultRequest, token)
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err, "レスポンスのJSONパースに失敗しました")

	// 統一レスポンス形式の検証
	suite.validateAPIResponse(response, true)

	// 更新された試合データの検証
	data, ok := response["data"].(map[string]interface{})
	suite.True(ok, "dataフィールドがオブジェクトではありません")

	// 結果が正しく設定されていることを確認
	suite.Equal(float64(3), data["score1"], "score1が正しく設定されていません")
	suite.Equal(float64(1), data["score2"], "score2が正しく設定されていません")
	suite.Equal("チームA", data["winner"], "winnerが正しく設定されていません")
	suite.Equal("completed", data["status"], "ステータスがcompletedに更新されていません")

	// completed_atが設定されていることを確認
	suite.NotNil(data["completed_at"], "completed_atが設定されていません")
	suite.validateDateTimeFormat(data["completed_at"].(string), "completed_at")
}

// TestContract_ErrorResponse_NotFound は404エラーレスポンスの契約テスト
func (suite *ContractTestSuite) TestContract_ErrorResponse_NotFound() {
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments/sport/invalid_sport", nil, nil)
	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err, "レスポンスのJSONパースに失敗しました")

	// 統一エラーレスポンス形式の検証
	suite.validateAPIResponse(response, false)

	// 404固有のエラーコード検証
	errorCode := response["error"].(string)
	suite.Equal("RESOURCE_NOT_FOUND", errorCode, "404エラーのエラーコードが正しくありません")
}

// TestContract_ErrorResponse_Unauthorized は401エラーレスポンスの契約テスト
func (suite *ContractTestSuite) TestContract_ErrorResponse_Unauthorized() {
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/tournaments", nil, nil)
	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err, "レスポンスのJSONパースに失敗しました")

	// 統一エラーレスポンス形式の検証
	suite.validateAPIResponse(response, false)

	// 401固有のエラーコード検証
	errorCode := response["error"].(string)
	suite.Equal("AUTH_UNAUTHORIZED", errorCode, "401エラーのエラーコードが正しくありません")
}

// TestContract_Pagination はページネーションの契約テスト
func (suite *ContractTestSuite) TestContract_Pagination() {
	// ページネーション付きリクエスト
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments?page=1&page_size=2", nil, nil)
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err, "レスポンスのJSONパースに失敗しました")

	// 統一レスポンス形式の検証
	suite.validateAPIResponse(response, true)

	// ページネーション情報の検証（実装されている場合）
	if pagination, exists := response["pagination"]; exists {
		paginationData, ok := pagination.(map[string]interface{})
		suite.True(ok, "paginationフィールドがオブジェクトではありません")

		// ページネーション必須フィールドの検証
		requiredPaginationFields := []string{"page", "page_size", "total_items", "total_pages", "has_next", "has_prev"}
		for _, field := range requiredPaginationFields {
			suite.Contains(paginationData, field, "ページネーション情報に%sフィールドが含まれていません", field)
		}

		// 型の検証
		suite.IsType(float64(0), paginationData["page"], "pageフィールドの型が正しくありません")
		suite.IsType(float64(0), paginationData["page_size"], "page_sizeフィールドの型が正しくありません")
		suite.IsType(float64(0), paginationData["total_items"], "total_itemsフィールドの型が正しくありません")
		suite.IsType(float64(0), paginationData["total_pages"], "total_pagesフィールドの型が正しくありません")
		suite.IsType(true, paginationData["has_next"], "has_nextフィールドの型が正しくありません")
		suite.IsType(true, paginationData["has_prev"], "has_prevフィールドの型が正しくありません")

		// 値の検証
		suite.Equal(float64(1), paginationData["page"], "pageフィールドの値が正しくありません")
		suite.Equal(float64(2), paginationData["page_size"], "page_sizeフィールドの値が正しくありません")
	}
}

// TestContract_DeprecatedAPIHeaders は廃止予定APIのヘッダー契約テスト
func (suite *ContractTestSuite) TestContract_DeprecatedAPIHeaders() {
	// 旧API（廃止予定）にアクセス
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/tournaments", nil, nil)

	// 廃止予定ヘッダーの検証
	suite.Equal("true", w.Header().Get("X-API-Deprecated"), "廃止予定ヘッダーが設定されていません")
	suite.Contains(w.Header().Get("X-API-Deprecation-Message"), "/api/v1", "廃止予定メッセージが正しくありません")
}

// TestContractTestSuite は契約テストスイートを実行する
func TestContractTestSuite(t *testing.T) {
	suite.Run(t, new(ContractTestSuite))
}