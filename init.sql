CREATE DATABASE IF NOT EXISTS main_DB;
CREATE DATABASE IF NOT EXISTS session_DB;

-- Locations table
CREATE TABLE IF NOT EXISTS `main_DB`.`locations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `city` VARCHAR(100) NOT NULL,
  `city_ascii` VARCHAR(100) NOT NULL,
  `lat` DOUBLE,
  `lng` DOUBLE,
  `country` VARCHAR(70) NOT NULL,
  `iso2` CHAR(2),
  `iso3` CHAR(3),
  `admin_name` VARCHAR(100),
  `capital` VARCHAR(100),
  `population` BIGINT,
  `location_id` BIGINT UNSIGNED,
  PRIMARY KEY (`id`)
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_unicode_ci
;

-- User table
CREATE TABLE IF NOT EXISTS `main_DB`.`users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(20) NOT NULL,
  `email` VARCHAR(255) NOT NULL,
  `password` VARCHAR(70) NOT NULL,
  `full_name` VARCHAR(50),
  `profile_pic` VARCHAR(255),
  `location` BIGINT UNSIGNED,
  `gender` VARCHAR(20),
  `about` VARCHAR(255),
  `birthday` DATE,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_email` (`email` ASC) VISIBLE,
  UNIQUE INDEX `idx_username` (`username` ASC) VISIBLE,
  FOREIGN KEY (`location`) REFERENCES `main_DB`.`locations` (`id`)
)
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
  FOREIGN KEY (`requester_id`) REFERENCES `main_DB`.`users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (`addressee_id`) REFERENCES `main_DB`.`users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COLLATE = utf8mb4_unicode_ci
;