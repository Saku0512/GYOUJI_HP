-- 開発用データベースリセットスクリプト
-- 注意: このスクリプトは全てのデータを削除します

-- データベースの削除と再作成
DROP DATABASE IF EXISTS tournament_db;
CREATE DATABASE tournament_db 
    DEFAULT CHARACTER SET utf8mb4 
    DEFAULT COLLATE utf8mb4_unicode_ci;

USE tournament_db;

-- テーブルの作成（マイグレーションファイルの内容を実行）
-- ユーザーテーブル
CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL COMMENT 'bcryptハッシュ化されたパスワード',
    role VARCHAR(20) DEFAULT 'admin',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_username (username),
    INDEX idx_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理者ユーザーテーブル';

-- トーナメントテーブル
CREATE TABLE IF NOT EXISTS tournaments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    sport ENUM('volleyball', 'table_tennis', 'soccer') NOT NULL COMMENT 'スポーツタイプ',
    format VARCHAR(20) DEFAULT 'standard' COMMENT 'トーナメント形式（卓球の場合standard/rainy）',
    status ENUM('active', 'completed') DEFAULT 'active' COMMENT 'トーナメントステータス',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_sport (sport),
    INDEX idx_status (status),
    INDEX idx_sport_status (sport, status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='トーナメントテーブル';

-- 試合テーブル
CREATE TABLE IF NOT EXISTS matches (
    id INT PRIMARY KEY AUTO_INCREMENT,
    tournament_id INT NOT NULL COMMENT 'トーナメントID',
    round VARCHAR(50) NOT NULL COMMENT 'ラウンド名（1st_round, quarterfinal等）',
    team1 VARCHAR(100) NOT NULL COMMENT 'チーム1名',
    team2 VARCHAR(100) NOT NULL COMMENT 'チーム2名',
    score1 INT NULL COMMENT 'チーム1のスコア（試合前はNULL）',
    score2 INT NULL COMMENT 'チーム2のスコア（試合前はNULL）',
    winner VARCHAR(100) NULL COMMENT '勝者チーム名（試合前はNULL）',
    status ENUM('pending', 'completed') DEFAULT 'pending' COMMENT '試合ステータス',
    scheduled_at TIMESTAMP NOT NULL COMMENT '試合予定時刻',
    completed_at TIMESTAMP NULL COMMENT '試合完了時刻',
    
    FOREIGN KEY (tournament_id) REFERENCES tournaments(id) ON DELETE CASCADE,
    
    INDEX idx_tournament_id (tournament_id),
    INDEX idx_status (status),
    INDEX idx_round (round),
    INDEX idx_tournament_round (tournament_id, round),
    INDEX idx_tournament_status (tournament_id, status),
    INDEX idx_scheduled_at (scheduled_at),
    INDEX idx_completed_at (completed_at),
    
    CONSTRAINT chk_scores_non_negative CHECK (score1 IS NULL OR score1 >= 0),
    CONSTRAINT chk_scores_non_negative2 CHECK (score2 IS NULL OR score2 >= 0),
    CONSTRAINT chk_completed_match_has_scores CHECK (
        (status = 'completed' AND score1 IS NOT NULL AND score2 IS NOT NULL AND winner IS NOT NULL) OR
        (status = 'pending')
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='試合テーブル';

-- 初期データの挿入
-- デフォルト管理者ユーザーの作成（パスワード: admin123）
INSERT INTO users (username, password, role) VALUES 
('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin');

-- 各スポーツのトーナメント作成
INSERT INTO tournaments (sport, format, status) VALUES 
('volleyball', 'standard', 'active'),
('table_tennis', 'standard', 'active'),
('soccer', 'standard', 'active');

-- インデックスの最適化
ANALYZE TABLE users;
ANALYZE TABLE tournaments;
ANALYZE TABLE matches;

SELECT 'データベースのリセットが完了しました' AS message;