-- トーナメント試合データの挿入

-- バレーボール試合データ
INSERT INTO matches (tournament_id, round, team1, team2, scheduled_at, status) VALUES
-- 1回戦 (9:30-11:00)
(1, '1st_round', '専・教', 'IE4', '2025-09-08 09:30:00', 'pending'),
(1, '1st_round', 'IS5', 'IT4', '2025-09-08 09:30:00', 'pending'),
(1, '1st_round', 'IT3', 'IT2', '2025-09-08 10:00:00', 'pending'),
(1, '1st_round', '1-1', 'IE2', '2025-09-08 10:00:00', 'pending'),
(1, '1st_round', 'IS3', 'IS2', '2025-09-08 10:30:00', 'pending'),
(1, '1st_round', 'IS4', 'IE5', '2025-09-08 10:30:00', 'pending'),
(1, '1st_round', '1-2', '1-3', '2025-09-08 11:00:00', 'pending'),
(1, '1st_round', 'IE3', 'IT5', '2025-09-08 11:00:00', 'pending'),
-- 準々決勝 (11:30, 13:00)
(1, 'quarterfinal', 'TBD', 'TBD', '2025-09-08 11:30:00', 'pending'),
(1, 'quarterfinal', 'TBD', 'TBD', '2025-09-08 11:30:00', 'pending'),
(1, 'quarterfinal', 'TBD', 'TBD', '2025-09-08 13:00:00', 'pending'),
(1, 'quarterfinal', 'TBD', 'TBD', '2025-09-08 13:00:00', 'pending'),
-- 準決勝 (14:00, 14:30)
(1, 'semifinal', 'TBD', 'TBD', '2025-09-08 14:00:00', 'pending'),
(1, 'semifinal', 'TBD', 'TBD', '2025-09-08 14:30:00', 'pending'),
-- 3位決定戦 (15:00)
(1, '3rd_place', 'TBD', 'TBD', '2025-09-08 15:00:00', 'pending'),
-- 決勝 (15:30)
(1, 'final', 'TBD', 'TBD', '2025-09-08 15:30:00', 'pending');

-- 卓球試合データ（晴天時）
INSERT INTO matches (tournament_id, round, team1, team2, scheduled_at, status) VALUES
-- 1回戦 (9:30-11:50)
(2, '1st_round', '1-2', 'IE5', '2025-09-08 09:30:00', 'pending'),
(2, '1st_round', 'IE3', 'IT2', '2025-09-08 09:50:00', 'pending'),
(2, '1st_round', '1-3', 'IE4', '2025-09-08 10:10:00', 'pending'),
(2, '1st_round', 'IT3', 'IT4', '2025-09-08 10:30:00', 'pending'),
(2, '1st_round', 'IS5', 'IE2', '2025-09-08 10:50:00', 'pending'),
(2, '1st_round', '1-1', 'IS3', '2025-09-08 11:10:00', 'pending'),
(2, '1st_round', 'IS2', 'IS4', '2025-09-08 11:30:00', 'pending'),
(2, '1st_round', 'IT5', '専・教', '2025-09-08 11:50:00', 'pending'),
-- 準々決勝 (13:00-14:00)
(2, 'quarterfinal', 'TBD', 'TBD', '2025-09-08 13:00:00', 'pending'),
(2, 'quarterfinal', 'TBD', 'TBD', '2025-09-08 13:20:00', 'pending'),
(2, 'quarterfinal', 'TBD', 'TBD', '2025-09-08 13:40:00', 'pending'),
(2, 'quarterfinal', 'TBD', 'TBD', '2025-09-08 14:00:00', 'pending'),
-- 準決勝 (14:30, 14:50)
(2, 'semifinal', 'TBD', 'TBD', '2025-09-08 14:30:00', 'pending'),
(2, 'semifinal', 'TBD', 'TBD', '2025-09-08 14:50:00', 'pending'),
-- 3位決定戦 (15:10)
(2, '3rd_place', 'TBD', 'TBD', '2025-09-08 15:10:00', 'pending'),
-- 決勝 (15:30)
(2, 'final', 'TBD', 'TBD', '2025-09-08 15:30:00', 'pending');

-- サッカー試合データ
INSERT INTO matches (tournament_id, round, team1, team2, scheduled_at, status) VALUES
-- 1回戦 (9:30-11:15)
(3, '1st_round', 'IS3', 'IE2', '2025-09-08 09:30:00', 'pending'),
(3, '1st_round', '1-1', 'IS2', '2025-09-08 09:45:00', 'pending'),
(3, '1st_round', 'IS4', 'IT5', '2025-09-08 10:00:00', 'pending'),
(3, '1st_round', 'IS5', '専・教', '2025-09-08 10:15:00', 'pending'),
(3, '1st_round', '1-2', '1-3', '2025-09-08 10:30:00', 'pending'),
(3, '1st_round', 'IE3', 'IT4', '2025-09-08 10:45:00', 'pending'),
(3, '1st_round', 'IT3', 'IE4', '2025-09-08 11:00:00', 'pending'),
(3, '1st_round', 'IT2', 'IE5', '2025-09-08 11:15:00', 'pending'),
-- 準々決勝 (11:30-13:15)
(3, 'quarterfinal', 'TBD', 'TBD', '2025-09-08 11:30:00', 'pending'),
(3, 'quarterfinal', 'TBD', 'TBD', '2025-09-08 11:45:00', 'pending'),
(3, 'quarterfinal', 'TBD', 'TBD', '2025-09-08 13:00:00', 'pending'),
(3, 'quarterfinal', 'TBD', 'TBD', '2025-09-08 13:15:00', 'pending'),
-- 準決勝 (13:30, 14:00)
(3, 'semifinal', 'TBD', 'TBD', '2025-09-08 13:30:00', 'pending'),
(3, 'semifinal', 'TBD', 'TBD', '2025-09-08 14:00:00', 'pending'),
-- 3位決定戦 (14:30)
(3, '3rd_place', 'TBD', 'TBD', '2025-09-08 14:30:00', 'pending'),
-- 決勝 (15:00)
(3, 'final', 'TBD', 'TBD', '2025-09-08 15:00:00', 'pending');