-- トーナメントデータのシーディング
-- 3つのスポーツ（バレーボール、卓球、サッカー）のトーナメントを作成

-- 既存のトーナメントデータをクリア（開発環境用）
DELETE FROM matches;
DELETE FROM tournaments;

-- バレーボールトーナメント
INSERT INTO tournaments (sport, format, status) VALUES 
('volleyball', 'standard', 'active');

-- 卓球トーナメント（標準フォーマット）
INSERT INTO tournaments (sport, format, status) VALUES 
('table_tennis', 'standard', 'active');

-- サッカートーナメント
INSERT INTO tournaments (sport, format, status) VALUES 
('soccer', 'standard', 'active');