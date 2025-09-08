# 統一バリデーションシステム使用ガイド

## 概要

このドキュメントは、トーナメント管理システムの統一バリデーションシステムの使用方法を説明します。
新しいバリデーションシステムは、一貫性のあるエラーメッセージ、多言語対応、カスタムバリデーションルールを提供します。

## 主要機能

### 1. 統一されたバリデーションエラー構造

```go
type ValidationError struct {
    Field   string `json:"field"`   // エラーが発生したフィールド名
    Message string `json:"message"` // エラーメッセージ
    Value   string `json:"value"`   // 入力された値
    Code    string `json:"code"`    // エラーコード
    Rule    string `json:"rule"`    // 違反したバリデーションルール
}
```

### 2. 複数エラーの管理

```go
type ValidationErrors []ValidationError

// エラーの追加
var errors ValidationErrors
errors.Add("username", "ユーザー名は必須です", "", "VALIDATION_REQUIRED_FIELD", "required")

// 特定フィールドのエラー取得
usernameErrors := errors.GetFieldErrors("username")

// エラーの存在確認
if errors.HasErrors() {
    // エラー処理
}
```

### 3. バリデーションコンテキスト

```go
// コンテキストの作成
ctx := NewValidationContext()
ctx.SetLanguage("en") // 英語に設定
ctx.SetData("user_id", 123) // 追加データの設定

// バリデーターにコンテキストを設定
validator := NewValidator().WithContext(ctx)
```

## 基本的な使用方法

### 1. 単一フィールドのバリデーション

```go
validator := models.NewValidator()

// 必須チェック
if err := validator.ValidateRequired(username, "username"); err != nil {
    // エラー処理
}

// 文字列長チェック
if err := validator.ValidateStringLength(username, "username", 1, 50); err != nil {
    // エラー処理
}

// メールアドレス形式チェック
if err := validator.ValidateEmail(email, "email"); err != nil {
    // エラー処理
}
```

### 2. 複数フィールドのバリデーション

```go
func ValidateUserRequest(req *UserRequest) ValidationErrors {
    var errors ValidationErrors
    validator := NewValidator()
    
    // ユーザー名の検証
    if err := validator.ValidateRequired(req.Username, "username"); err != nil {
        errors.AddError(*err)
    }
    
    if err := validator.ValidateStringLength(req.Username, "username", 1, 50); err != nil {
        errors.AddError(*err)
    }
    
    if err := validator.ValidateAlphanumeric(req.Username, "username"); err != nil {
        errors.AddError(*err)
    }
    
    // パスワードの検証
    if err := validator.ValidatePassword(req.Password, "password"); err != nil {
        errors.AddError(*err)
    }
    
    return errors
}
```

### 3. ハンドラーでの使用

```go
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    
    // リクエストバインド
    if err := c.ShouldBindJSON(&req); err != nil {
        h.SendBindingError(c, err)
        return
    }
    
    // バリデーション実行
    if !h.ValidateRequest(c, func() models.ValidationErrors {
        return ValidateUserRequest(&req)
    }) {
        return
    }
    
    // ビジネスロジック処理
    // ...
}
```

## カスタムバリデーションルール

### 1. ルールベースのバリデーション

```go
// バリデーターにルールを追加
validator := NewValidator()
validator.AddRule("username", NewRequiredRule())
validator.AddRule("username", NewMinLengthRule(3))
validator.AddRule("username", NewMaxLengthRule(50))
validator.AddRule("username", NewAlphanumericRule())

// 構造体全体のバリデーション
errors := validator.ValidateStruct(userRequest)
```

### 2. 事前定義されたルール

```go
// 基本的なルール
requiredRule := NewRequiredRule()
minLengthRule := NewMinLengthRule(8)
maxLengthRule := NewMaxLengthRule(100)
emailRule := NewEmailRule()
alphanumericRule := NewAlphanumericRule()

// 数値ルール
minValueRule := NewMinValueRule(0)
maxValueRule := NewMaxValueRule(1000)

// 列挙型ルール
sportTypeRule := NewSportTypeRule()
tournamentStatusRule := NewTournamentStatusRule()
matchStatusRule := NewMatchStatusRule()

// 日時ルール
futureDateTimeRule := NewFutureDateTimeRule()

// パスワードルール
passwordRule := NewPasswordRule()
passwordRule.RequireSpecial = true
passwordRule.RequireMixedCase = true
```

### 3. カスタムルールの作成

```go
// カスタムルールの実装
type CustomRule struct {
    // ルール固有のフィールド
}

func (r CustomRule) Validate(value interface{}, fieldName string) *ValidationError {
    // バリデーションロジック
    if /* 条件 */ {
        return &ValidationError{
            Field:   fieldName,
            Message: "カスタムエラーメッセージ",
            Value:   fmt.Sprintf("%v", value),
            Code:    "CUSTOM_ERROR_CODE",
            Rule:    "custom_rule",
        }
    }
    return nil
}

func (r CustomRule) GetRuleName() string {
    return "custom_rule"
}
```

## ビジネスロジックバリデーション

### 1. 試合スコアの検証

```go
validator := NewValidator()
errors := validator.ValidateMatchScore(score1, score2, winner, team1, team2)

if errors.HasErrors() {
    // エラー処理
}
```

### 2. ビジネスルールの検証

```go
validator := NewValidator()
errors := validator.ValidateBusinessRules(tournament, match)

if errors.HasErrors() {
    // エラー処理
}
```

### 3. スポーツ固有のバリデーション

```go
validator := NewValidator()

// スポーツに対して有効なラウンドかチェック
if err := validator.ValidateRoundForSport(sport, round, "round"); err != nil {
    // エラー処理
}
```

## 多言語対応

### 1. 言語設定

```go
// 日本語（デフォルト）
validator := NewValidator()

// 英語
ctx := NewValidationContext().SetLanguage("en")
validator := NewValidator().WithContext(ctx)
```

### 2. メッセージのカスタマイズ

バリデーションメッセージは `getLocalizedMessage` 関数で管理されています。
新しい言語やメッセージを追加する場合は、この関数を拡張してください。

## エラーレスポンスの統一

### 1. ハンドラーでのエラーレスポンス

```go
// ValidationErrorsからレスポンス送信
func (h *BaseHandler) SendValidationErrors(c *gin.Context, errors ValidationErrors) {
    if !errors.HasErrors() {
        return
    }
    
    details := errors.ToValidationErrorDetails()
    h.SendValidationError(c, "入力データが無効です", details)
}

// バリデーション実行とエラーレスポンス
func (h *BaseHandler) ValidateRequest(c *gin.Context, validator func() ValidationErrors) bool {
    errors := validator()
    if errors.HasErrors() {
        h.SendValidationErrors(c, errors)
        return false
    }
    return true
}
```

### 2. レスポンス形式

```json
{
    "success": false,
    "error": "VALIDATION_ERROR",
    "message": "入力データが無効です",
    "code": 400,
    "timestamp": "2024-01-01T12:00:00Z",
    "request_id": "req_123456",
    "details": [
        {
            "field": "username",
            "message": "ユーザー名は必須です",
            "value": ""
        },
        {
            "field": "password",
            "message": "パスワードは8文字以上である必要があります",
            "value": "pass"
        }
    ]
}
```

## ベストプラクティス

### 1. バリデーション関数の命名

```go
// 良い例
func ValidateLoginRequest(req *LoginRequest) ValidationErrors
func ValidateCreateTournamentRequest(req *CreateTournamentRequest) ValidationErrors
func ValidateSubmitMatchResultRequest(req *SubmitMatchResultRequest, team1, team2 string) ValidationErrors

// 避けるべき例
func ValidateLogin(req *LoginRequest) error
func CheckTournament(req *CreateTournamentRequest) bool
```

### 2. エラーメッセージの一貫性

- フィールド名は日本語で統一
- エラーコードは英語の定数を使用
- メッセージは具体的で分かりやすく

### 3. パフォーマンス考慮

```go
// 良い例：早期リターン
func ValidateRequest(req *Request) ValidationErrors {
    var errors ValidationErrors
    validator := NewValidator()
    
    // 必須チェックを最初に実行
    if err := validator.ValidateRequired(req.Field, "field"); err != nil {
        errors.AddError(*err)
        return errors // 必須フィールドが空なら他のチェックは不要
    }
    
    // その他のバリデーション
    // ...
    
    return errors
}
```

### 4. テストの書き方

```go
func TestValidateLoginRequest(t *testing.T) {
    tests := []struct {
        name      string
        request   *LoginRequest
        wantError bool
        errorFields []string
    }{
        {
            name: "有効なリクエスト",
            request: &LoginRequest{
                Username: "admin",
                Password: "password123",
            },
            wantError: false,
        },
        {
            name: "ユーザー名が空",
            request: &LoginRequest{
                Username: "",
                Password: "password123",
            },
            wantError: true,
            errorFields: []string{"username"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            errors := ValidateLoginRequest(tt.request)
            
            if errors.HasErrors() != tt.wantError {
                t.Errorf("ValidateLoginRequest() hasErrors = %v, wantError %v", 
                    errors.HasErrors(), tt.wantError)
            }
            
            if tt.wantError {
                for _, field := range tt.errorFields {
                    fieldErrors := errors.GetFieldErrors(field)
                    if len(fieldErrors) == 0 {
                        t.Errorf("Expected error for field %s", field)
                    }
                }
            }
        })
    }
}
```

## 移行ガイド

### 1. 既存コードからの移行

```go
// 旧方式
if err := req.Validate(); err != nil {
    h.SendErrorWithCode(c, models.ErrorValidationInvalidFormat, err.Error(), http.StatusBadRequest)
    return
}

// 新方式
if !h.ValidateRequest(c, func() models.ValidationErrors {
    return models.ValidateLoginRequest(&req)
}) {
    return
}
```

### 2. 段階的移行

1. 新しいバリデーション関数を作成
2. テストを追加
3. ハンドラーを更新
4. 旧バリデーション関数を削除

この統一バリデーションシステムにより、一貫性のあるエラーハンドリングと保守しやすいコードを実現できます。