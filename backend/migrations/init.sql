-- データベース初期化スクリプト
-- このスクリプトは全てのマイグレーションを順番に実行する

-- データベースの作成（存在しない場合）
CREATE DATABASE IF NOT EXISTS tournament_db 
    DEFAULT CHARACTER SET utf8mb4 
    DEFAULT COLLATE utf8mb4_unicode_ci;

USE tournament_db;

-- マイグレーション実行順序
-- 1. ユーザーテーブル
SOURCE /docker-entrypoint-initdb.d/001_create_users_table.sql;

-- 2. トーナメントテーブル
SOURCE /docker-entrypoint-initdb.d/002_create_tournaments_table.sql;

-- 3. 試合テーブル（外部キー制約があるため最後）
SOURCE /docker-entrypoint-initdb.d/003_create_matches_table.sql;

-- 初期データの挿入
-- デフォルト管理者ユーザーの作成（パスワード: admin123）
INSERT IGNORE INTO users (username, password, role) VALUES 
('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin');

-- 各スポーツのトーナメント作成
INSERT IGNORE INTO tournaments (sport, format, status) VALUES 
('volleyball', 'standard', 'active'),
('table_tennis', 'standard', 'active'),
('soccer', 'standard', 'active');

-- インデックスの最適化
ANALYZE TABLE users;
ANALYZE TABLE tournaments;
ANALYZE TABLE matches;