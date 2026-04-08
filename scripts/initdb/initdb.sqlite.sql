CREATE TABLE IF NOT EXISTS `users` (
  `id` TEXT PRIMARY KEY NOT NULL,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `username` TEXT UNIQUE NOT NULL,
  `email` TEXT,
  `display_name` TEXT NOT NULL,
  `password_hash` TEXT NOT NULL,
  `status` TEXT NOT NULL DEFAULT 'active'
);

CREATE TABLE IF NOT EXISTS `roles` (
  `id` TEXT PRIMARY KEY NOT NULL,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `name` TEXT UNIQUE NOT NULL,
  `description` TEXT
);

CREATE TABLE IF NOT EXISTS `user_roles` (
  `id` TEXT PRIMARY KEY NOT NULL,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `user_id` TEXT NOT NULL,
  `role_name` TEXT NOT NULL,
  UNIQUE (`user_id`, `role_name`)
);

CREATE TABLE IF NOT EXISTS `policies` (
  `id` TEXT PRIMARY KEY NOT NULL,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `subject` TEXT NOT NULL,
  `object` TEXT NOT NULL,
  `action` TEXT NOT NULL,
  UNIQUE (`subject`, `object`, `action`)
);

CREATE TABLE IF NOT EXISTS `samples` (
  `id` TEXT PRIMARY KEY NOT NULL,
  `created_at` DATETIME NOT NULL,
  `updated_at` DATETIME NOT NULL,
  `deleted_at` DATETIME,
  `name` TEXT UNIQUE NOT NULL,
  `description` TEXT,
  `enabled` BOOLEAN NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS `casbin_rule` (
  `id` INTEGER PRIMARY KEY NOT NULL,
  `ptype` TEXT NOT NULL,
  `v0` TEXT,
  `v1` TEXT,
  `v2` TEXT,
  `v3` TEXT,
  `v4` TEXT,
  `v5` TEXT
);
