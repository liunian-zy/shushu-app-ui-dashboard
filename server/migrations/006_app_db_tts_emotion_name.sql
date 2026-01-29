ALTER TABLE `app_db_tts_presets`
  ADD COLUMN `emotion_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER `voice_id`;
