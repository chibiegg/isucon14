DROP TABLE IF EXISTS chair_locations_latest;
CREATE TABLE chair_locations_latest
(
  chair_id   VARCHAR(26) NOT NULL COMMENT '椅子ID',
  latitude   INTEGER     NOT NULL COMMENT '経度',
  longitude  INTEGER     NOT NULL COMMENT '緯度',
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '更新日時',
  total_distance INTEGER NOT NULL DEFAULT 0 COMMENT '合計移動距離',
  PRIMARY KEY (chair_id)
);

INSERT INTO chair_locations_latest
SELECT
  chairs.id AS chair_id,
  (SELECT latitude FROM chair_locations WHERE chair_id = chairs.id ORDER BY created_at DESC LIMIT 1) AS latitude,
  (SELECT longitude FROM chair_locations WHERE chair_id = chairs.id ORDER BY created_at DESC LIMIT 1) AS longitude,
  total_distance_updated_at AS updated_at,
  IFNULL(total_distance, 0) AS total_distance
FROM chairs

LEFT JOIN (SELECT chair_id,
                  SUM(IFNULL(distance, 0)) AS total_distance,
                  MAX(created_at)          AS total_distance_updated_at
            FROM (SELECT chair_id,
                        created_at,
                        ABS(latitude - LAG(latitude) OVER (PARTITION BY chair_id ORDER BY created_at)) +
                        ABS(longitude - LAG(longitude) OVER (PARTITION BY chair_id ORDER BY created_at)) AS distance
                  FROM chair_locations) tmp
            GROUP BY chair_id) distance_table ON distance_table.chair_id = chairs.id

WHERE total_distance_updated_at IS NOT NULL;

ALTER TABLE chairs ADD COLUMN is_free BOOLEAN NOT NULL DEFAULT 1 COMMENT '乗れるかどうか' AFTER is_active,