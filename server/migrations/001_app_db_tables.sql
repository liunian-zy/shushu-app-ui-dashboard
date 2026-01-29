-- Draft tables for app UI plan system (prefix app_db_)

CREATE TABLE IF NOT EXISTS `app_db_version_names` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `app_version_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `location_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `feishu_field_names` text COLLATE utf8mb4_unicode_ci,
  `ai_modal` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint unsigned DEFAULT NULL,
  `draft_status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT 'draft',
  `submit_version` int unsigned DEFAULT 0,
  `last_submit_by` int unsigned DEFAULT NULL,
  `last_submit_at` datetime DEFAULT NULL,
  `confirmed_by` int unsigned DEFAULT NULL,
  `confirmed_at` datetime DEFAULT NULL,
  `sync_status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sync_message` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `synced_at` datetime DEFAULT NULL,
  `target_app_version_name_id` int unsigned DEFAULT NULL,
  `created_by` int unsigned DEFAULT NULL,
  `updated_by` int unsigned DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_app_version_name` (`app_version_name`),
  KEY `idx_location_name` (`location_name`),
  KEY `idx_draft_status` (`draft_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_app_ui_fields` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `app_version_name_id` int unsigned DEFAULT NULL,
  `home_title_left` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `home_title_right` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `home_subtitle` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `start_experience` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `step1_music` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `step1_music_text` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `step1_title` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `step2_music` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `step2_music_text` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `step2_title` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint unsigned DEFAULT NULL,
  `print_wait` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `created_by` int unsigned DEFAULT NULL,
  `updated_by` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_draft_version_id` (`draft_version_id`),
  KEY `idx_app_version_name_id` (`app_version_name_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_banners` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `title` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `image` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sort` int DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `type` tinyint unsigned DEFAULT NULL,
  `app_version_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_by` int unsigned DEFAULT NULL,
  `updated_by` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_app_version_name` (`app_version_name`),
  KEY `idx_banner_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_identities` (
  `id` int NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `image` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sort` int DEFAULT NULL,
  `status` tinyint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `app_version_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_by` int unsigned DEFAULT NULL,
  `updated_by` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_app_version_name` (`app_version_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_scenes` (
  `id` int NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `image` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `desc` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `music` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `watermark_path` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `need_watermark` tinyint(1) DEFAULT NULL,
  `sort` int DEFAULT NULL,
  `status` tinyint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `app_version_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `oss_style` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_by` int unsigned DEFAULT NULL,
  `updated_by` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_app_version_name` (`app_version_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_clothes_categories` (
  `id` int NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `image` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sort` int DEFAULT NULL,
  `status` tinyint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `music` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `desc` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `music_text` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `app_version_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_by` int unsigned DEFAULT NULL,
  `updated_by` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_app_version_name` (`app_version_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_photo_hobbies` (
  `id` int NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `image` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sort` int DEFAULT NULL,
  `status` tinyint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `music` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `music_text` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `desc` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `app_version_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_by` int unsigned DEFAULT NULL,
  `updated_by` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_app_version_name` (`app_version_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_config_extra_steps` (
  `id` int NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `app_version_name_id` int unsigned DEFAULT NULL,
  `step_index` int DEFAULT NULL,
  `field_name` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `label` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `music` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `music_text` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint unsigned DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `created_by` int unsigned DEFAULT NULL,
  `updated_by` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_app_version_name_id` (`app_version_name_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_users` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `display_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `role` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'user',
  `status` tinyint unsigned DEFAULT 1,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_role` (`role`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_tasks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `module_key` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `title` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT 'open',
  `assigned_to` int unsigned DEFAULT NULL,
  `allow_assist` tinyint(1) DEFAULT 1,
  `priority` tinyint unsigned DEFAULT 0,
  `created_by` int unsigned DEFAULT NULL,
  `updated_by` int unsigned DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_assigned_to` (`assigned_to`),
  KEY `idx_module_key` (`module_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_task_actions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `task_id` bigint unsigned DEFAULT NULL,
  `action` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `actor_id` int unsigned DEFAULT NULL,
  `detail_json` json DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_actor_id` (`actor_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_submissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `module_key` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `entity_table` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `entity_id` bigint unsigned DEFAULT NULL,
  `submit_version` int unsigned DEFAULT NULL,
  `submit_by` int unsigned DEFAULT NULL,
  `payload_json` json DEFAULT NULL,
  `diff_json` json DEFAULT NULL,
  `need_confirm` tinyint(1) DEFAULT NULL,
  `status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `prev_submission_id` bigint unsigned DEFAULT NULL,
  `confirmed_by` int unsigned DEFAULT NULL,
  `confirmed_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_module_key` (`module_key`),
  KEY `idx_entity` (`entity_table`, `entity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_field_history` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `entity_table` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `entity_id` bigint unsigned DEFAULT NULL,
  `field_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `old_value` text COLLATE utf8mb4_unicode_ci,
  `new_value` text COLLATE utf8mb4_unicode_ci,
  `submit_id` bigint unsigned DEFAULT NULL,
  `changed_by` int unsigned DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_entity` (`entity_table`, `entity_id`),
  KEY `idx_submit_id` (`submit_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_media_assets` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `module_key` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `media_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `file_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `file_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `file_size` int unsigned DEFAULT NULL,
  `width` int unsigned DEFAULT NULL,
  `height` int unsigned DEFAULT NULL,
  `duration_ms` int unsigned DEFAULT NULL,
  `format` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `hash` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `origin_asset_id` bigint unsigned DEFAULT NULL,
  `status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_by` int unsigned DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_module_key` (`module_key`),
  KEY `idx_origin_asset_id` (`origin_asset_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_media_versions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `asset_id` bigint unsigned DEFAULT NULL,
  `version_no` int unsigned DEFAULT NULL,
  `file_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `file_size` int unsigned DEFAULT NULL,
  `width` int unsigned DEFAULT NULL,
  `height` int unsigned DEFAULT NULL,
  `duration_ms` int unsigned DEFAULT NULL,
  `format` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `compress_profile` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_by` int unsigned DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_asset_id` (`asset_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_media_rules` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `module_key` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `media_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `max_size_kb` int unsigned DEFAULT NULL,
  `min_width` int unsigned DEFAULT NULL,
  `max_width` int unsigned DEFAULT NULL,
  `min_height` int unsigned DEFAULT NULL,
  `max_height` int unsigned DEFAULT NULL,
  `min_duration_ms` int unsigned DEFAULT NULL,
  `max_duration_ms` int unsigned DEFAULT NULL,
  `allow_formats` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `resize_mode` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `target_format` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `compress_quality` int unsigned DEFAULT NULL,
  `status` tinyint unsigned DEFAULT NULL,
  `created_by` int unsigned DEFAULT NULL,
  `updated_by` int unsigned DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_module_key` (`module_key`),
  KEY `idx_media_type` (`media_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_audit_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `entity_table` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `entity_id` bigint unsigned DEFAULT NULL,
  `action` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `actor_id` int unsigned DEFAULT NULL,
  `detail_json` json DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_entity` (`entity_table`, `entity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `app_db_sync_jobs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `draft_version_id` int unsigned DEFAULT NULL,
  `trigger_by` int unsigned DEFAULT NULL,
  `status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `error_message` text COLLATE utf8mb4_unicode_ci,
  `started_at` datetime DEFAULT NULL,
  `finished_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_draft_version_id` (`draft_version_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
