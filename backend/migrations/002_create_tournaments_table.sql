-- トーナメントテーブルの作成
CREATE TABLE IF NOT EXISTS tournaments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    sport ENUM('volleyball', 'table_tennis', 'soccer') NOT NULL COMMENT 'スポーツタイプ',
    format VARCHAR(20) DEFAULT 'standard' COMMENT 'トーナメント形式（卓球の場合standard/rainy）',
    status ENUM('registration', 'active', 'completed') DEFAULT 'registration' COMMENT 'トーナメントステータス',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- インデックス
    INDEX idx_sport (sport),
    INDEX idx_status (status),
    INDEX idx_sport_status (sport, status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='トーナメントテーブル';