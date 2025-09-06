# 共通UIコンポーネント

このディレクトリには、トーナメントフロントエンドアプリケーションで使用される再利用可能なUIコンポーネントが含まれています。

## 実装済みコンポーネント

### 基本UIコンポーネント

#### Button.svelte
汎用的なボタンコンポーネント

**Props:**
- `variant`: ボタンのスタイル ('primary', 'secondary', 'success', 'danger', 'warning', 'info', 'light', 'dark')
- `size`: ボタンのサイズ ('small', 'medium', 'large')
- `disabled`: 無効状態
- `loading`: ローディング状態
- `type`: ボタンタイプ ('button', 'submit', 'reset')
- `href`: リンクとして使用する場合のURL
- `target`: リンクのターゲット
- `fullWidth`: 全幅表示
- `outline`: アウトラインスタイル

**使用例:**
```svelte
<Button variant="primary" size="medium" on:click={handleClick}>
  クリック
</Button>

<Button variant="danger" outline loading>
  処理中...
</Button>

<Button href="/admin" variant="secondary">
  管理画面へ
</Button>
```

#### Input.svelte
汎用的な入力フィールドコンポーネント

**Props:**
- `type`: 入力タイプ ('text', 'email', 'password', 'number', 'tel', 'url', 'search')
- `value`: 入力値
- `placeholder`: プレースホルダー
- `disabled`: 無効状態
- `readonly`: 読み取り専用
- `required`: 必須フィールド
- `size`: サイズ ('small', 'medium', 'large')
- `variant`: バリデーション状態 ('default', 'success', 'error', 'warning')
- `label`: ラベル
- `helperText`: ヘルパーテキスト
- `errorMessage`: エラーメッセージ
- `fullWidth`: 全幅表示

**使用例:**
```svelte
<Input
  label="ユーザー名"
  bind:value={username}
  required
  placeholder="ユーザー名を入力"
  helperText="半角英数字で入力してください"
/>

<Input
  type="email"
  label="メールアドレス"
  bind:value={email}
  variant="error"
  errorMessage="有効なメールアドレスを入力してください"
/>
```

#### Select.svelte
汎用的なセレクトボックスコンポーネント

**Props:**
- `value`: 選択値
- `options`: 選択肢の配列 (`{ value, label, disabled? }`)
- `placeholder`: プレースホルダー
- `disabled`: 無効状態
- `required`: 必須フィールド
- `size`: サイズ ('small', 'medium', 'large')
- `variant`: バリデーション状態 ('default', 'success', 'error', 'warning')
- `label`: ラベル
- `helperText`: ヘルパーテキスト
- `errorMessage`: エラーメッセージ
- `multiple`: 複数選択
- `fullWidth`: 全幅表示

**使用例:**
```svelte
<Select
  label="スポーツ"
  bind:value={selectedSport}
  options={[
    { value: 'volleyball', label: 'バレーボール' },
    { value: 'table_tennis', label: '卓球' },
    { value: 'soccer', label: 'サッカー' }
  ]}
  placeholder="スポーツを選択"
/>
```

### 通知・フィードバックコンポーネント

#### LoadingSpinner.svelte
ローディング表示コンポーネント

**Props:**
- `size`: スピナーのサイズ ('small', 'medium', 'large')
- `color`: スピナーの色

**使用例:**
```svelte
<LoadingSpinner size="large" color="#007bff" />
```

#### NotificationToast.svelte
通知トーストコンポーネント

**Props:**
- `message`: 通知メッセージ
- `type`: 通知タイプ ('success', 'error', 'warning', 'info')
- `duration`: 自動消去時間（ミリ秒）
- `dismissible`: 手動で閉じることができるか

**イベント:**
- `close`: 通知が閉じられたときに発火

**使用例:**
```svelte
<NotificationToast
  message="保存が完了しました"
  type="success"
  duration={3000}
  on:close={handleClose}
/>
```

### モーダル・ダイアログコンポーネント

#### Modal.svelte
モーダルダイアログコンポーネント

**Props:**
- `open`: モーダルの表示状態
- `title`: モーダルのタイトル
- `size`: モーダルのサイズ ('small', 'medium', 'large', 'full')
- `closable`: 閉じるボタンを表示するか
- `closeOnBackdrop`: 背景クリックで閉じるか
- `closeOnEscape`: Escapeキーで閉じるか

**イベント:**
- `close`: モーダルが閉じられたときに発火

**スロット:**
- `default`: モーダルの本文
- `footer`: モーダルのフッター

**使用例:**
```svelte
<Modal
  bind:open={showModal}
  title="確認"
  size="medium"
  on:close={handleModalClose}
>
  <p>この操作を実行しますか？</p>
  
  <div slot="footer">
    <Button variant="secondary" on:click={() => showModal = false}>
      キャンセル
    </Button>
    <Button variant="primary" on:click={handleConfirm}>
      実行
    </Button>
  </div>
</Modal>
```

## アクセシビリティ対応

すべてのコンポーネントは以下のアクセシビリティ機能を実装しています：

### キーボードナビゲーション
- **Tab/Shift+Tab**: フォーカス移動
- **Enter/Space**: ボタンの実行
- **Escape**: モーダルや通知の閉じる

### ARIA属性
- `role`: 適切な役割の設定
- `aria-label`: アクセシブルな名前
- `aria-describedby`: 説明テキストとの関連付け
- `aria-invalid`: バリデーション状態
- `aria-disabled`: 無効状態
- `aria-live`: 動的コンテンツの通知

### フォーカス管理
- モーダル開閉時の適切なフォーカス移動
- フォーカストラップ（モーダル内でのフォーカス循環）
- 視覚的なフォーカスインジケーター

### スクリーンリーダー対応
- 適切な見出し構造
- フォーム要素とラベルの関連付け
- エラーメッセージの適切な通知

## レスポンシブデザイン

すべてのコンポーネントは以下のブレークポイントに対応しています：

- **Mobile**: < 768px
- **Tablet**: 768px - 1024px  
- **Desktop**: > 1024px

### モバイル最適化
- タッチフレンドリーなボタンサイズ（最小44px）
- 適切なフォントサイズ（ズーム防止のため16px以上）
- スワイプ操作対応
- 縦向き表示最適化

## テスト

各コンポーネントには包括的なテストスイートが含まれています：

### 単体テスト
- プロパティの正しい適用
- イベントハンドリング
- 状態変更の処理
- エラー状態の処理

### アクセシビリティテスト
- ARIA属性の正しい設定
- キーボードナビゲーション
- フォーカス管理
- スクリーンリーダー対応

### 統合テスト
- コンポーネント間の相互作用
- フォーム送信フロー
- モーダル操作フロー

## 使用方法

コンポーネントは以下の方法でインポートできます：

```javascript
// 個別インポート
import Button from '$lib/components/Button.svelte';
import Input from '$lib/components/Input.svelte';

// 一括インポート
import { Button, Input, Select, Modal } from '$lib/components';
```

## スタイリング

すべてのコンポーネントは一貫したデザインシステムに基づいています：

### カラーパレット
- Primary: #007bff
- Secondary: #6c757d
- Success: #28a745
- Danger: #dc3545
- Warning: #ffc107
- Info: #17a2b8

### タイポグラフィ
- フォントファミリー: システムフォント
- フォントサイズ: 0.875rem - 1.125rem
- 行間: 1.25 - 1.75

### 間隔
- 小: 0.25rem - 0.5rem
- 中: 0.75rem - 1rem
- 大: 1.5rem - 2rem

## パフォーマンス

### 最適化機能
- CSS-in-JS による最小限のスタイル読み込み
- 遅延読み込み対応
- 軽量なアニメーション
- 効率的なイベントハンドリング

### バンドルサイズ
各コンポーネントは軽量で、必要な機能のみを含んでいます：
- Button: ~2KB (gzipped)
- Input: ~3KB (gzipped)
- Select: ~3KB (gzipped)
- Modal: ~4KB (gzipped)
- NotificationToast: ~3KB (gzipped)
- LoadingSpinner: ~1KB (gzipped)