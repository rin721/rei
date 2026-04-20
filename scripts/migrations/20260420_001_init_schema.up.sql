-- Migration: 20260420_001_init_schema (up)
-- Generated: 2026-04-19T17:55:54Z

CREATE TABLE IF NOT EXISTS "users" (
  "id" TEXT NOT NULL,
  "created_at" DATETIME,
  "updated_at" DATETIME,
  "deleted_at" TEXT,
  "username" TEXT NOT NULL,
  "email" TEXT,
  "display_name" TEXT NOT NULL,
  "password_hash" TEXT NOT NULL,
  "status" TEXT NOT NULL DEFAULT active,
  PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_users_username ON "users" ("username");

CREATE INDEX IF NOT EXISTS idx_users_email ON "users" ("email");

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON "users" ("deleted_at");

CREATE TABLE IF NOT EXISTS "roles" (
  "id" TEXT NOT NULL,
  "created_at" DATETIME,
  "updated_at" DATETIME,
  "deleted_at" TEXT,
  "name" TEXT NOT NULL,
  "description" TEXT,
  PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX IF NOT EXISTS roles_name ON "roles" ("name");

CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON "roles" ("deleted_at");

CREATE TABLE IF NOT EXISTS "user_roles" (
  "id" TEXT NOT NULL,
  "created_at" DATETIME,
  "updated_at" DATETIME,
  "deleted_at" TEXT,
  "user_id" TEXT NOT NULL,
  "role_name" TEXT NOT NULL,
  PRIMARY KEY ("id")
);

CREATE INDEX IF NOT EXISTS idx_user_role_user ON "user_roles" ("user_id");

CREATE INDEX IF NOT EXISTS idx_user_role_role ON "user_roles" ("role_name");

CREATE UNIQUE INDEX IF NOT EXISTS uidx_user_role_binding ON "user_roles" ("user_id", "role_name");

CREATE INDEX IF NOT EXISTS idx_user_roles_deleted_at ON "user_roles" ("deleted_at");

CREATE TABLE IF NOT EXISTS "policies" (
  "id" TEXT NOT NULL,
  "created_at" DATETIME,
  "updated_at" DATETIME,
  "deleted_at" TEXT,
  "subject" TEXT NOT NULL,
  "object" TEXT NOT NULL,
  "action" TEXT NOT NULL,
  PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_policy_binding ON "policies" ("subject", "object", "action");

CREATE INDEX IF NOT EXISTS idx_policies_deleted_at ON "policies" ("deleted_at");

CREATE TABLE IF NOT EXISTS "samples" (
  "id" TEXT NOT NULL,
  "created_at" DATETIME,
  "updated_at" DATETIME,
  "deleted_at" TEXT,
  "name" TEXT NOT NULL,
  "description" TEXT,
  "enabled" INTEGER NOT NULL DEFAULT true,
  PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX IF NOT EXISTS samples_name ON "samples" ("name");

CREATE INDEX IF NOT EXISTS idx_samples_deleted_at ON "samples" ("deleted_at");
