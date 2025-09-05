-- バレーボール試合データのシーディング
-- READMEに基づく実際の試合データ

-- バレーボールトーナメントIDを取得（@volleyball_tournament_id変数に設定）
SET @volleyball_tournament_id = (SELECT id FROM tournaments WHERE sport = 'volleyball' LIMIT 1);

-- 1回戦の試合データ（READMEに基づく）
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@volleyball_tournament_id, '1st_round', '専・教', 'IE4', 'pending', DATE_ADD(CURDATE(), INTERVAL 9.5 HOUR)),
(@volleyball_tournament_id, '1st_round', 'IS5', 'IT4', 'pending', DATE_ADD(CURDATE(), INTERVAL 9.5 HOUR)),
(@volleyball_tournament_id, '1st_round', 'IT3', 'IT2', 'pending', DATE_ADD(CURDATE(), INTERVAL 10 HOUR)),
(@volleyball_tournament_id, '1st_round', '1-1', 'IE2', 'pending', DATE_ADD(CURDATE(), INTERVAL 10 HOUR)),
(@volleyball_tournament_id, '1st_round', 'IS3', 'IS2', 'pending', DATE_ADD(CURDATE(), INTERVAL 10.5 HOUR)),
(@volleyball_tournament_id, '1st_round', 'IS4', 'IE5', 'pending', DATE_ADD(CURDATE(), INTERVAL 10.5 HOUR)),
(@volleyball_tournament_id, '1st_round', '1-2', '1-3', 'pending', DATE_ADD(CURDATE(), INTERVAL 11 HOUR)),
(@volleyball_tournament_id, '1st_round', 'IE3', 'IT5', 'pending', DATE_ADD(CURDATE(), INTERVAL 11 HOUR));

-- 準々決勝のプレースホルダー試合
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@volleyball_tournament_id, 'quarterfinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 11.5 HOUR)),
(@volleyball_tournament_id, 'quarterfinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 13 HOUR));

-- 準決勝のプレースホルダー試合
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@volleyball_tournament_id, 'semifinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 14 HOUR)),
(@volleyball_tournament_id, 'semifinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 14.5 HOUR));

-- 3位決定戦のプレースホルダー試合
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@volleyball_tournament_id, 'third_place', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 15 HOUR));

-- 決勝のプレースホルダー試合
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@volleyball_tournament_id, 'final', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 15.5 HOUR));