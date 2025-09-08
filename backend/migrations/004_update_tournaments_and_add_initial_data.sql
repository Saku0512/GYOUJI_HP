-- トーナメントテーブルのステータス更新と初期データ追加

-- ステータスENUMを更新（registrationを追加）
ALTER TABLE tournaments MODIFY COLUMN status ENUM('registration', 'active', 'completed') DEFAULT 'registration' COMMENT 'トーナメントステータス';

-- 初期トーナメントデータを挿入
INSERT INTO tournaments (sport, format, status, created_at, updated_at) VALUES
('volleyball', 'standard', 'registration', NOW(), NOW()),
('table_tennis', 'standard', 'registration', NOW(), NOW()),
('soccer', 'standard', 'registration', NOW(), NOW())
ON DUPLICATE KEY UPDATE
    format = VALUES(format),
    status = VALUES(status),
    updated_at = NOW();