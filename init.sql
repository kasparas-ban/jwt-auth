CREATE DATABASE IF NOT EXISTS main_DB;
CREATE DATABASE IF NOT EXISTS session_DB;

-- User table
CREATE TABLE IF NOT EXISTS `main_DB`.`users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(20) NOT NULL,
  `email` VARCHAR(40) NOT NULL,
  `password` VARCHAR(70) NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  -- UNIQUE INDEX `idx_relations` (`id`, `id`) VISIBLE,
  UNIQUE INDEX `idx_email` (`email` ASC) VISIBLE,
  UNIQUE INDEX `idx_username` (`username` ASC) VISIBLE)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_unicode_ci
;

-- Friendships table
CREATE TABLE IF NOT EXISTS `main_DB`.`friendships` (
  `requester_id` BIGINT UNSIGNED NOT NULL,
  `addressee_id` BIGINT UNSIGNED NOT NULL,
  `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE INDEX `idx_relation` (`requester_id`, `addressee_id`) VISIBLE,
  FOREIGN KEY (`requester_id`)
  REFERENCES `main_DB`.`users` (`id`)
  ON DELETE CASCADE
  ON UPDATE CASCADE,
  FOREIGN KEY (`addressee_id`)
  REFERENCES `main_DB`.`users` (`id`)
  ON DELETE CASCADE
  ON UPDATE CASCADE
);