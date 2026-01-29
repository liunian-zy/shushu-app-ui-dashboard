SET @add_password_hash := (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE app_db_users ADD COLUMN password_hash varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER display_name',
    'SELECT 1'
  )
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'app_db_users'
    AND column_name = 'password_hash'
);
PREPARE stmt FROM @add_password_hash;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @add_last_login_at := (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE app_db_users ADD COLUMN last_login_at datetime DEFAULT NULL AFTER status',
    'SELECT 1'
  )
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'app_db_users'
    AND column_name = 'last_login_at'
);
PREPARE stmt FROM @add_last_login_at;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
