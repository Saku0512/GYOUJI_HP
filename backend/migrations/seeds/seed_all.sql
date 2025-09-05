-- 全データベースシーディングスクリプト
-- 開発環境用の完全なデータセットアップ

-- 外部キー制約を一時的に無効化
SET FOREIGN_KEY_CHECKS = 0;

-- 管理者ユーザーのシーディング
SOURCE 001_seed_admin_user.sql;

-- トーナメントのシーディング
SOURCE 002_seed_tournaments.sql;

-- 各スポーツの試合データシーディング
SOURCE 003_seed_volleyball_matches.sql;
SOURCE 004_seed_table_tennis_matches.sql;
SOURCE 005_seed_soccer_matches.sql;

-- 外部キー制約を再有効化
SET FOREIGN_KEY_CHECKS = 1;

-- シーディング完了メッセージ
SELECT 'データベースシーディングが完了しました' AS message;
SELECT 
    (SELECT COUNT(*) FROM users) AS users_count,
    (SELECT COUNT(*) FROM tournaments) AS tournaments_count,
    (SELECT COUNT(*) FROM matches) AS matches_count;