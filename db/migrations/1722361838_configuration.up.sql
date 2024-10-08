DROP TABLE IF EXISTS `videos`;
DROP TABLE IF EXISTS `instances`;
DROP TABLE IF EXISTS `configurations`;

CREATE TABLE `configurations` (
	`id` BLOB UNIQUE NOT NULL,
  `theme` TEXT DEFAULT 'default',
  `bubble_enabled` BOOLEAN DEFAULT false,
  `bubble_text_content` TEXT,
  `cta_enabled` BOOLEAN DEFAULT false,
  `cta_text_content` TEXT,
  `video_url` TEXT
);
