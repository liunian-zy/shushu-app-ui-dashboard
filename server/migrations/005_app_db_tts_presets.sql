CREATE TABLE IF NOT EXISTS `app_db_tts_presets` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `voice_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `volume` int DEFAULT NULL,
  `speed` decimal(4,2) DEFAULT NULL,
  `pitch` int DEFAULT NULL,
  `stability` int DEFAULT NULL,
  `similarity` int DEFAULT NULL,
  `exaggeration` int DEFAULT NULL,
  `status` tinyint unsigned DEFAULT 1,
  `is_default` tinyint unsigned DEFAULT 0,
  `created_by` int unsigned DEFAULT NULL,
  `updated_by` int unsigned DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_tts_status` (`status`),
  KEY `idx_tts_default` (`is_default`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO `app_db_tts_presets`
  (`name`, `voice_id`, `volume`, `speed`, `pitch`, `stability`, `similarity`, `exaggeration`, `status`, `is_default`, `created_at`, `updated_at`)
SELECT
  '默认预设',
  '70eb6772-4cd1-11f0-9276-00163e0fe4f9',
  58,
  1.00,
  56,
  50,
  95,
  0,
  1,
  1,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
WHERE NOT EXISTS (
  SELECT 1 FROM `app_db_tts_presets` WHERE `is_default` = 1 LIMIT 1
);
