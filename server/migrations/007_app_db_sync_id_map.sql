CREATE TABLE IF NOT EXISTS `app_db_sync_id_map` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned NOT NULL,
  `module_key` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `draft_row_id` bigint unsigned NOT NULL,
  `target_row_id` bigint unsigned NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_draft_module_row` (`draft_version_id`, `module_key`, `draft_row_id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_module_key` (`module_key`),
  KEY `idx_target_row_id` (`target_row_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
