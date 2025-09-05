-- サッカー試合データのシーディング
-- READMEに基づく実際の試合データ

-- サッカートーナメントIDを取得
SET @soccer_tournament_id = (SELECT id FROM tournaments WHERE sport = 'soccer' LIMIT 1);

-- 1回戦の試合データ
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@soccer_tournament_id, '1st_round', 'IS3', 'IE2', 'pending', DATE_ADD(CURDATE(), INTERVAL 9.5 HOUR)),
(@soccer_tournament_id, '1st_round', '1-1', 'IS2', 'pending', DATE_ADD(CURDATE(), INTERVAL 9 + 45/60.0 HOUR)),
(@soccer_tournament_id, '1st_round', 'IS4', 'IT5', 'pending', DATE_ADD(CURDATE(), INTERVAL 10 HOUR)),
(@soccer_tournament_id, '1st_round', 'IS5', '専・教', 'pending', DATE_ADD(CURDATE(), INTERVAL 10 + 15/60.0 HOUR)),
(@soccer_tournament_id, '1st_round', '1-2', '1-3', 'pending', DATE_ADD(CURDATE(), INTERVAL 10.5 HOUR)),
(@soccer_tournament_id, '1st_round', 'IE3', 'IT4', 'pending', DATE_ADD(CURDATE(), INTERVAL 10 + 45/60.0 HOUR)),
(@soccer_tournament_id, '1st_round', 'IT3', 'IE4', 'pending', DATE_ADD(CURDATE(), INTERVAL 11 HOUR)),
(@soccer_tournament_id, '1st_round', 'IT2', 'IE5', 'pending', DATE_ADD(CURDATE(), INTERVAL 11 + 15/60.0 HOUR));

-- 準々決勝のプレースホルダー試合
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@soccer_tournament_id, 'quarterfinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 11.5 HOUR)),
(@soccer_tournament_id, 'quarterfinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 11 + 45/60.0 HOUR)),
(@soccer_tournament_id, 'quarterfinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 13 HOUR)),
(@soccer_tournament_id, 'quarterfinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 13 + 15/60.0 HOUR));

-- 準決勝のプレースホルダー試合
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@soccer_tournament_id, 'semifinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 13.5 HOUR)),
(@soccer_tournament_id, 'semifinal', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 14 HOUR));

-- 3位決定戦のプレースホルダー試合
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@soccer_tournament_id, 'third_place', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 14.5 HOUR));

-- 決勝のプレースホルダー試合
INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) VALUES
(@soccer_tournament_id, 'final', 'TBD', 'TBD', 'pending', DATE_ADD(CURDATE(), INTERVAL 15 HOUR));