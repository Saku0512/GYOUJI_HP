# Tournament Management System - Frontend

SvelteKitを使用したトーナメント管理システムのフロントエンドアプリケーションです。

## 概要

このアプリケーションは、バレーボール、卓球、サッカーのトーナメント管理を行うWebアプリケーションです。管理者は試合結果を入力でき、一般ユーザーはリアルタイムでトーナメントの進行を閲覧できます。

## 技術スタック

- **フレームワーク**: SvelteKit
- **言語**: JavaScript/TypeScript
- **ビルドツール**: Vite
- **スタイリング**: CSS
- **リンター**: ESLint
- **フォーマッター**: Prettier

## プロジェクト構造

```
src/
├── routes/                 # SvelteKitページルーティング
│   ├── +layout.svelte     # 共通レイアウト
│   ├── +page.svelte       # ホームページ（トーナメント表示）
│   ├── login/             # ログインページ
│   └── admin/             # 管理者ダッシュボード
├── lib/
│   ├── components/        # 再利用可能コンポーネント
│   ├── stores/           # Svelteストア（状態管理）
│   ├── api/              # APIクライアント
│   └── utils/            # ユーティリティ関数
└── app.html              # HTMLテンプレート
```

## 開発環境のセットアップ

### 前提条件

- Node.js (v18以上)
- npm

### インストール

```bash
npm install
```

### 開発サーバーの起動

```bash
npm run dev
```

ブラウザで http://localhost:5173 にアクセスしてください。

### ビルド

```bash
npm run build
```

### プレビュー

```bash
npm run preview
```

## 開発コマンド

- `npm run dev` - 開発サーバーの起動
- `npm run build` - 本番用ビルド
- `npm run preview` - ビルド結果のプレビュー
- `npm run check` - TypeScript型チェック
- `npm run lint` - ESLintによるコードチェック
- `npm run format` - Prettierによるコードフォーマット

## 環境変数

`.env`ファイルで以下の環境変数を設定できます：

```env
VITE_API_BASE_URL=http://localhost:8080/api
VITE_APP_TITLE=Tournament Management System
VITE_ENABLE_POLLING=true
VITE_POLLING_INTERVAL=30000
```

## 主要機能

### 実装済み

- SvelteKitプロジェクト構造のセットアップ
- 基本的なページルーティング（ホーム、ログイン、管理者）
- 再利用可能コンポーネントの骨格
- 状態管理ストアの骨格
- APIクライアントの骨格
- ユーティリティ関数の骨格

### 今後実装予定

- 認証システム
- トーナメントブラケット表示
- 試合結果入力機能
- リアルタイム更新
- レスポンシブデザイン

## コーディング規約

- ESLintとPrettierの設定に従う
- コンポーネント名はPascalCase
- ファイル名はkebab-case
- 日本語コメントを使用

## ライセンス

このプロジェクトは内部使用のためのものです。