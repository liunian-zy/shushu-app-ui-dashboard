CREATE TABLE IF NOT EXISTS `app_db_sync_module_jobs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `module_key` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `trigger_by` int unsigned DEFAULT NULL,
  `status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `error_message` text COLLATE utf8mb4_unicode_ci,
  `started_at` datetime DEFAULT NULL,
  `finished_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_module_key` (`module_key`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
