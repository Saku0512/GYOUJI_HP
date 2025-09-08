-- 試合テーブルの作成
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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日時',
    
    -- 外部キー制約
    FOREIGN KEY (tournament_id) REFERENCES tournaments(id) ON DELETE CASCADE,
    
    -- インデックス
    INDEX idx_tournament_id (tournament_id),
    INDEX idx_status (status),
    INDEX idx_round (round),
    INDEX idx_tournament_round (tournament_id, round),
    INDEX idx_tournament_status (tournament_id, status),
    INDEX idx_scheduled_at (scheduled_at),
    INDEX idx_completed_at (completed_at),
    INDEX idx_created_at (created_at),
    INDEX idx_updated_at (updated_at),
    
    -- 制約
    CONSTRAINT chk_scores_non_negative CHECK (score1 IS NULL OR score1 >= 0),
    CONSTRAINT chk_scores_non_negative2 CHECK (score2 IS NULL OR score2 >= 0),
    CONSTRAINT chk_completed_match_has_scores CHECK (
        (status = 'completed' AND score1 IS NOT NULL AND score2 IS NOT NULL AND winner IS NOT NULL) OR
        (status = 'pending')
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='試合テーブル';