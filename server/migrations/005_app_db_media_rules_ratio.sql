SET @exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'app_db_media_rules'
    AND COLUMN_NAME = 'ratio_width'
);
SET @sql := IF(@exists = 0,
  'ALTER TABLE `app_db_media_rules` ADD COLUMN `ratio_width` int unsigned DEFAULT NULL AFTER `max_height`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'app_db_media_rules'
    AND COLUMN_NAME = 'ratio_height'
);
SET @sql := IF(@exists = 0,
  'ALTER TABLE `app_db_media_rules` ADD COLUMN `ratio_height` int unsigned DEFAULT NULL AFTER `ratio_width`',
  'SELECT 1'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
