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

-- 4. トーナメント初期化
SOURCE /docker-entrypoint-initdb.d/004_update_tournaments_and_add_initial_data.sql

-- 5. トーナメント情報初期化
SOURCE /docker-entrypoint-initdb.d/005_insert_tournament_matches.sql

-- インデックスの最適化
ANALYZE TABLE users;
ANALYZE TABLE tournaments;
ANALYZE TABLE matches;