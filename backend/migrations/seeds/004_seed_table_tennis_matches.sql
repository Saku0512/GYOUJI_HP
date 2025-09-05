-- 卓球試合データのシーディング（晴天時フォーマット）
-- READMEに基づく実際の試合データ

-- 卓球トーナメントIDを取得
SET @table_tennis_tournament_id = (SELECT id FROM tournaments WHERE sport = 'table_tennis' LIMIT 1);

-- 1回戦の試合データ（晴天時フォーマット）
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@table_tennis_tournament_id, '1st_round', '1-2', 'IE5', 'pending', DATE_ADD(CURDATE(), INTERVAL 9.5 HOUR)),
(@table_tennis_tournament_id, '1st_round', 'IE3', 'IT2', 'pending', DATE_ADD(CURDATE(), INTERVAL 9 + 50/60.0 HOUR)),
(@table_tennis_tournament_id, '1st_round', '1-3', 'IE4', 'pending', DATE_ADD(CURDATE(), INTERVAL 10 + 10/60.0 HOUR)),
(@table_tennis_tournament_id, '1st_round', 'IT3', 'IT4', 'pending', DATE_ADD(CURDATE(), INTERVAL 10.5 HOUR)),
(@table_tennis_tournament_id, '1st_round', 'IS5', 'IE2', 'pending', DATE_ADD(CURDATE(), INTERVAL 10 + 50/60.0 HOUR)),
(@table_tennis_tournament_id, '1st_round', '1-1', 'IS3', 'pending', DATE_ADD(CURDATE(), INTERVAL 11 + 10/60.0 HOUR)),
(@table_tennis_tournament_id, '1st_round', 'IS2', 'IS4', 'pending', DATE_ADD(CURDATE(), INTERVAL 11.5 HOUR)),
(@table_tennis_tournament_id, '1st_round', 'IT5', '専・教', 'pending', DATE_ADD(CURDATE(), INTERVAL 11 + 50/60.0 HOUR));

-- 準々決勝のプレースホルダー試合
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@table_tennis_tournament_id, 'quarterfinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 13 HOUR)),
(@table_tennis_tournament_id, 'quarterfinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 13 + 20/60.0 HOUR)),
(@table_tennis_tournament_id, 'quarterfinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 13 + 40/60.0 HOUR)),
(@table_tennis_tournament_id, 'quarterfinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 14 HOUR));

-- 準決勝のプレースホルダー試合
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@table_tennis_tournament_id, 'semifinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 14.5 HOUR)),
(@table_tennis_tournament_id, 'semifinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 14 + 50/60.0 HOUR));

-- 3位決定戦のプレースホルダー試合
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@table_tennis_tournament_id, 'third_place', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 15 + 10/60.0 HOUR));

-- 決勝のプレースホルダー試合
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@table_tennis_tournament_id, 'final', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 15.5 HOUR));