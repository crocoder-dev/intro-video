DROP TABLE IF EXISTS `videos`;
DROP TABLE IF EXISTS `instances`;
DROP TABLE IF EXISTS `configurations`;

CREATE TABLE `configurations` (
	`id` INTEGER PRIMARY KEY NOT NULL,
  `theme` TEXT DEFAULT 'default',
  `bubble_enabled` BOOLEAN DEFAULT false,
  `bubble_text_content` TEXT,
  `cta_enabled` BOOLEAN DEFAULT false,
  `cta_text_content` TEXT
);
CREATE TABLE `videos` (
	`id` INTEGER PRIMARY KEY NOT NULL,
  `url` TEXT NOT NULL,
  `weight` INTEGER DEFAULT 1,
  `configuration_id` INTEGER NOT NULL,
  `instance_id` INTEGER NOT NULL,
  FOREIGN KEY(`configuration_id`) REFERENCES `configurations`(`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  FOREIGN KEY(`instance_id`) REFERENCES `instances`(`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE TABLE `instances` (
	`id` INTEGER PRIMARY KEY NOT NULL,
  `external_id` BLOB UNIQUE NOT NULL
);

CREATE INDEX instances_external_id_idx ON instances(external_id);
