
-- 01, create database
CREATE DATABASE xally CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

-- 02, create table xally.auth_users
-- drop table xally.auth_users
CREATE TABLE IF NOT EXISTS xally.auth_users (
    id              bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    created_at      datetime DEFAULT CURRENT_TIMESTAMP,
    updated_at      datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      datetime DEFAULT NULL,

    username        varchar(64) NOT NULL,
    hostname        varchar(64) NOT NULL,
    email           varchar(128) NOT NULL,
    device_info     varchar(255) NOT NULL,
    password        varchar(255) NOT NULL,
    is_actived      bigint(20) unsigned NOT NULL DEFAULT 0,
    is_verified     bigint(20) unsigned NOT NULL DEFAULT 0,

    register_at     datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expired_at      datetime DEFAULT NULL,
    activate_at     datetime DEFAULT NULL,
    deactivate_at   datetime DEFAULT NULL,

    PRIMARY KEY (id),
    UNIQUE KEY uniq_user (email),
    KEY idx_auth_users_deleted_at (deleted_at),
    KEY idx_auth_users_expired_at (expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- 03, create table xally.user_tokens
-- drop table xally.user_tokens
CREATE TABLE IF NOT EXISTS xally.user_tokens (
    id              bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    created_at      datetime DEFAULT CURRENT_TIMESTAMP,
    updated_at      datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      datetime DEFAULT NULL,

    token_type      varchar(64) NOT NULL,
    token           char(36) NOT NULL,
    expired_at      datetime DEFAULT NULL,
    user_id         bigint(20) unsigned DEFAULT NULL,
    consume_counter bigint(20) unsigned NOT NULL DEFAULT 0,

    PRIMARY KEY (id),
    KEY idx_user_tokens_expired_at (expired_at),
    KEY idx_user_tokens_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- 04, create table xally.proxy_logs
-- drop table xally.proxy_logs
CREATE TABLE IF NOT EXISTS xally.proxy_logs (
    id                      bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    created_at              datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,

    user_id                 bigint(20) unsigned NOT NULL,
    remote_addr             varchar(255) DEFAULT NULL,
    request_time            datetime NOT NULL,
    request_method          varchar(256) NOT NULL,
    request_url             varchar(256) NOT NULL,
    request_headers         text DEFAULT NULL,
    request_body            longtext DEFAULT NULL,
    response_status_code    bigint(20) NOT NULL,
    response_headers        text DEFAULT NULL,
    response_body           longtext DEFAULT NULL,

    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
