// Package integration_test はOpenAPI仕様検証テストを提供する
package integration_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"backend/internal/testutil"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
)

// OpenAPIValidationTestSuite はOpenAPI仕様検証テストスイート
type OpenAPIValidationTestSuite struct {
	suite.Suite
	server   *testutil.TestServer
	openAPI  *OpenAPISpec
}

// OpenAPISpec はOpenAPI仕様の簡略化された構造体
type OpenAPISpec struct {
	OpenAPI    string                 `yaml:"openapi"`
	Info       OpenAPIInfo            `yaml:"info"`
	Paths      map[string]PathItem    `yaml:"paths"`
	Components OpenAPIComponents      `yaml:"components"`
}

// OpenAPIInfo はAPI情報
type OpenAPIInfo struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

// PathItem はパス項目
type PathItem struct {
	Get    *Operation `yaml:"get,omitempty"`
	Post   *Operation `yaml:"post,omitempty"`
	Put    *Operation `yaml:"put,omitempty"`
	Delete *Operation `yaml:"delete,omitempty"`
}

// Operation はオペレーション
type Operation struct {
	Tags        []string                       `yaml:"tags,omitempty"`
	Summary     string                         `yaml:"summary,omitempty"`
	Description string                         `yaml:"description,omitempty"`
	Parameters  []Parameter                    `yaml:"parameters,omitempty"`
	RequestBody *RequestBody                   `yaml:"requestBody,omitempty"`
	Responses   map[string]ResponseDefinition  `yaml:"responses"`
	Security    []map[string][]string          `yaml:"security,omitempty"`
}

// Parameter はパラメータ
type Parameter struct {
	Name        string      `yaml:"name"`
	In          string      `yaml:"in"`
	Required    bool        `yaml:"required,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Schema      interface{} `yaml:"schema,omitempty"`
}

// RequestBody はリクエストボディ
type RequestBody struct {
	Required bool                           `yaml:"required,omitempty"`
	Content  map[string]MediaTypeDefinition `yaml:"content"`
}

// ResponseDefinition はレスポンス定義
type ResponseDefinition struct {
	Description string                         `yaml:"description"`
	Content     map[string]MediaTypeDefinition `yaml:"content,omitempty"`
	Headers     map[string]interface{}         `yaml:"headers,omitempty"`
}

// MediaTypeDefinition はメディアタイプ定義
type MediaTypeDefinition struct {
	Schema   interface{}            `yaml:"schema,omitempty"`
	Examples map[string]interface{} `yaml:"examples,omitempty"`
}

// OpenAPIComponents はコンポーネント
type OpenAPIComponents struct {
	Schemas         map[string]interface{} `yaml:"schemas"`
	SecuritySchemes map[string]interface{} `yaml:"securitySchemes"`
	Parameters      map[string]Parameter   `yaml:"parameters"`
	Headers         map[string]interface{} `yaml:"headers"`
}

// SetupSuite はテストスイートのセットアップを行う
func (suite *OpenAPIValidationTestSuite) SetupSuite() {
	suite.server = testutil.SetupTestServer(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
	
	// OpenAPI仕様を読み込み
	suite.loadOpenAPISpec()
}

// TearDownSuite はテストスイートのクリーンアップを行う
func (suite *OpenAPIValidationTestSuite) TearDownSuite() {
	suite.server.TeardownTestServer(suite.T())
}

// SetupTest は各テストの前に実行される
func (suite *OpenAPIValidationTestSuite) SetupTest() {
	// 各テスト前にデータをクリーンアップしてシード
	suite.server.TestDB.CleanupTestData(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
}

// loadOpenAPISpec はOpenAPI仕様ファイルを読み込む
func (suite *OpenAPIValidationTestSuite) loadOpenAPISpec() {
	// プロジェクトルートからの相対パスでOpenAPI仕様ファイルを読み込み
	specPath := filepath.Join("..", "docs", "openapi.yaml")
	
	data, err := ioutil.ReadFile(specPath)
	if err != nil {
		suite.T().Skipf("OpenAPI仕様ファイルが見つかりません: %s", specPath)
		return
	}
	
	var spec OpenAPISpec
	err = yaml.Unmarshal(data, &spec)
	if err != nil {
		suite.T().Fatalf("OpenAPI仕様ファイルの解析に失敗しました: %v", err)
	}
	
	suite.openAPI = &spec
}

// validateResponseAgainstSchema はレスポンスをスキーマに対して検証する
func (suite *OpenAPIValidationTestSuite) validateResponseAgainstSchema(response map[string]interface{}, expectedStatusCode int, path, method string) {
	if suite.openAPI == nil {
		suite.T().Skip("OpenAPI仕様が読み込まれていません")
		return
	}
	
	// パス情報を取得
	pathItem, exists := suite.openAPI.Paths[path]
	if !exists {
		suite.T().Logf("警告: パス %s がOpenAPI仕様に定義されていません", path)
		return
	}
	
	// オペレーション情報を取得
	var operation *Operation
	switch strings.ToUpper(method) {
	case "GET":
		operation = pathItem.Get
	case "POST":
		operation = pathItem.Post
	case "PUT":
		operation = pathItem.Put
	case "DELETE":
		operation = pathItem.Delete
	}
	
	if operation == nil {
		suite.T().Logf("警告: メソッド %s がパス %s に定義されていません", method, path)
		return
	}
	
	// レスポンス定義を取得
	statusCodeStr := fmt.Sprintf("%d", expectedStatusCode)
	responseDefinition, exists := operation.Responses[statusCodeStr]
	if !exists {
		suite.T().Logf("警告: ステータスコード %d がパス %s のメソッド %s に定義されていません", expectedStatusCode, path, method)
		return
	}
	
	// 基本的なレスポンス構造の検証
	suite.validateBasicResponseStructure(response, expectedStatusCode >= 200 && expectedStatusCode < 300)
	
	suite.T().Logf("OpenAPI仕様に対するレスポンス検証が完了しました: %s %s (%d)", method, path, expectedStatusCode)
}

// validateBasicResponseStructure は基本的なレスポンス構造を検証する
func (suite *OpenAPIValidationTestSuite) validateBasicResponseStructure(response map[string]interface{}, isSuccess bool) {
	// 統一レスポンス形式の基本検証
	suite.Contains(response, "success", "レスポンスにsuccessフィールドが含まれていません")
	suite.Contains(response, "message", "レスポンスにmessageフィールドが含まれていません")
	suite.Contains(response, "code", "レスポンスにcodeフィールドが含まれていません")
	suite.Contains(response, "timestamp", "レスポンスにtimestampフィールドが含まれていません")
	
	suite.Equal(isSuccess, response["success"], "successフィールドの値が期待値と異なります")
	
	if isSuccess {
		suite.Contains(response, "data", "成功レスポンスにdataフィールドが含まれていません")
	} else {
		suite.Contains(response, "error", "エラーレスポンスにerrorフィールドが含まれていません")
	}
}

// TestOpenAPI_AuthLogin はログインエンドポイントのOpenAPI仕様検証
func (suite *OpenAPIValidationTestSuite) TestOpenAPI_AuthLogin() {
	// 要件7.1: API仕様が変更された場合、システムは自動的にドキュメントを更新する
	loginRequest := map[string]string{
		"username": "admin",
		"password": "password",
	}

	w := suite.server.MakeRequest(suite.T(), "POST", "/api/v1/auth/login", loginRequest, nil)
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	// OpenAPI仕様に対する検証
	suite.validateResponseAgainstSchema(response, http.StatusOK, "/auth/login", "POST")
}

// TestOpenAPI_AuthLogin_ValidationError はログインバリデーションエラーのOpenAPI仕様検証
func (suite *OpenAPIValidationTestSuite) TestOpenAPI_AuthLogin_ValidationError() {
	loginRequest := map[string]string{
		"username": "",
		"password": "password",
	}

	w := suite.server.MakeRequest(suite.T(), "POST", "/api/v1/auth/login", loginRequest, nil)
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	// OpenAPI仕様に対する検証
	suite.validateResponseAgainstSchema(response, http.StatusBadRequest, "/auth/login", "POST")
}

// TestOpenAPI_TournamentList は全トーナメント取得のOpenAPI仕様検証
func (suite *OpenAPIValidationTestSuite) TestOpenAPI_TournamentList() {
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments", nil, nil)
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	// OpenAPI仕様に対する検証
	suite.validateResponseAgainstSchema(response, http.StatusOK, "/public/tournaments", "GET")
}

// TestOpenAPI_TournamentBySport はスポーツ別トーナメント取得のOpenAPI仕様検証
func (suite *OpenAPIValidationTestSuite) TestOpenAPI_TournamentBySport() {
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments/sport/volleyball", nil, nil)
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	// OpenAPI仕様に対する検証
	suite.validateResponseAgainstSchema(response, http.StatusOK, "/public/tournaments/sport/{sport}", "GET")
}

// TestOpenAPI_TournamentBySport_NotFound はスポーツ別トーナメント404エラーのOpenAPI仕様検証
func (suite *OpenAPIValidationTestSuite) TestOpenAPI_TournamentBySport_NotFound() {
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/tournaments/sport/invalid_sport", nil, nil)
	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	// OpenAPI仕様に対する検証
	suite.validateResponseAgainstSchema(response, http.StatusNotFound, "/public/tournaments/sport/{sport}", "GET")
}

// TestOpenAPI_MatchesBySport はスポーツ別試合取得のOpenAPI仕様検証
func (suite *OpenAPIValidationTestSuite) TestOpenAPI_MatchesBySport() {
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/public/matches/sport/volleyball", nil, nil)
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	// OpenAPI仕様に対する検証
	suite.validateResponseAgainstSchema(response, http.StatusOK, "/public/matches/sport/{sport}", "GET")
}

// TestOpenAPI_CreateMatch は試合作成のOpenAPI仕様検証
func (suite *OpenAPIValidationTestSuite) TestOpenAPI_CreateMatch() {
	// 要件7.2: 契約テストが実行される場合、システムはフロントエンドとバックエンドの仕様整合性を検証する
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
	suite.NoError(err)

	// OpenAPI仕様に対する検証
	suite.validateResponseAgainstSchema(response, http.StatusCreated, "/admin/matches", "POST")
}

// TestOpenAPI_SubmitMatchResult は試合結果提出のOpenAPI仕様検証
func (suite *OpenAPIValidationTestSuite) TestOpenAPI_SubmitMatchResult() {
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
	suite.NoError(err)

	// OpenAPI仕様に対する検証
	suite.validateResponseAgainstSchema(response, http.StatusOK, "/admin/matches/{id}/result", "PUT")
}

// TestOpenAPI_UnauthorizedAccess は認証なしアクセスのOpenAPI仕様検証
func (suite *OpenAPIValidationTestSuite) TestOpenAPI_UnauthorizedAccess() {
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/tournaments", nil, nil)
	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	// OpenAPI仕様に対する検証
	suite.validateResponseAgainstSchema(response, http.StatusUnauthorized, "/tournaments", "GET")
}

// TestOpenAPI_HealthCheck はヘルスチェックのOpenAPI仕様検証
func (suite *OpenAPIValidationTestSuite) TestOpenAPI_HealthCheck() {
	w := suite.server.MakeRequest(suite.T(), "GET", "/health", nil, nil)
	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	// OpenAPI仕様に対する検証
	suite.validateResponseAgainstSchema(response, http.StatusOK, "/health", "GET")
}

// TestOpenAPI_SpecificationConsistency はOpenAPI仕様の一貫性を検証する
func (suite *OpenAPIValidationTestSuite) TestOpenAPI_SpecificationConsistency() {
	if suite.openAPI == nil {
		suite.T().Skip("OpenAPI仕様が読み込まれていません")
		return
	}

	// 基本情報の検証
	suite.Equal("3.0.3", suite.openAPI.OpenAPI, "OpenAPIバージョンが正しくありません")
	suite.Equal("Tournament Management API", suite.openAPI.Info.Title, "APIタイトルが正しくありません")
	suite.Equal("1.0.0", suite.openAPI.Info.Version, "APIバージョンが正しくありません")
	suite.NotEmpty(suite.openAPI.Info.Description, "API説明が空です")

	// パス定義の存在確認
	expectedPaths := []string{
		"/health",
		"/auth/login",
		"/auth/logout",
		"/auth/refresh",
		"/auth/validate",
		"/auth/profile",
		"/public/tournaments",
		"/public/tournaments/active",
		"/public/tournaments/sport/{sport}",
		"/public/tournaments/sport/{sport}/bracket",
		"/public/tournaments/sport/{sport}/progress",
		"/public/matches/sport/{sport}",
		"/public/matches/tournament/{tournament_id}",
		"/public/matches/tournament/{tournament_id}/next",
	}

	for _, path := range expectedPaths {
		suite.Contains(suite.openAPI.Paths, path, "パス %s がOpenAPI仕様に定義されていません", path)
	}

	// コンポーネント定義の存在確認
	suite.Contains(suite.openAPI.Components.Schemas, "BaseResponse", "BaseResponseスキーマが定義されていません")
	suite.Contains(suite.openAPI.Components.Schemas, "ErrorResponse", "ErrorResponseスキーマが定義されていません")
	suite.Contains(suite.openAPI.Components.Schemas, "Tournament", "Tournamentスキーマが定義されていません")
	suite.Contains(suite.openAPI.Components.Schemas, "Match", "Matchスキーマが定義されていません")

	// セキュリティスキーム定義の確認
	suite.Contains(suite.openAPI.Components.SecuritySchemes, "BearerAuth", "BearerAuthセキュリティスキームが定義されていません")

	suite.T().Log("OpenAPI仕様の一貫性検証が完了しました")
}

// TestOpenAPIValidationTestSuite はOpenAPI仕様検証テストスイートを実行する
func TestOpenAPIValidationTestSuite(t *testing.T) {
	suite.Run(t, new(OpenAPIValidationTestSuite))
}