-- 管理者ユーザーのシーディング
-- パスワード: admin123 (bcryptハッシュ化済み)

INSERT INTO users (username, password, role) VALUES 
('admin', '$2a$10$rh2lI/npcfCnXsskVbKQGO5cbb4i2Vxh7iZfg5PqgjFz/NWnCWLIO', 'admin')
ON DUPLICATE KEY UPDATE 
    password = VALUES(password),
    role = VALUES(role);

-- 注意: 本番環境では必ず強力なパスワードに変更してください
-- このハッシュは 'admin123' に対応しています