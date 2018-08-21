/* *****************************************************************************
// Setup the preferences
// ****************************************************************************/

SET NAMES utf8 COLLATE 'utf8_unicode_ci';
SET foreign_key_checks = 1;
SET time_zone = '+00:00';
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';
SET default_storage_engine = InnoDB;
SET CHARACTER SET utf8;

/* *****************************************************************************
// Remove old database
// ****************************************************************************/
DROP DATABASE IF EXISTS webappdb;

/* *****************************************************************************
// Create new database
// ****************************************************************************/
CREATE DATABASE webappdb DEFAULT CHARSET = utf8 COLLATE = utf8_unicode_ci;
USE webappdb;

/* *****************************************************************************
// Create the tables
// ****************************************************************************/
CREATE TABLE search_cache (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    zipcode VARCHAR(10) NOT NULL,
    times_used INT(10) UNSIGNED NOT NULL DEFAULT 1,
    currtemp VARCHAR(10) NOT NULL,
    hightemp VARCHAR(10) NOT NULL,
    lowtemp VARCHAR(10) NOT NULL,
    phrase VARCHAR(100) NOT NULL,
    icon VARCHAR(200) NOT NULL,
    iframesrc VARCHAR(200) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted TINYINT(1) UNSIGNED NOT NULL DEFAULT 0,
    
    PRIMARY KEY (id)
);

/*
CREATE TABLE user (
    id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password CHAR(60) NOT NULL,
    
    status_id TINYINT(1) UNSIGNED NOT NULL DEFAULT 1,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted TINYINT(1) UNSIGNED NOT NULL DEFAULT 0,
    
    UNIQUE KEY (email),
    CONSTRAINT `f_user_status` FOREIGN KEY (`status_id`) REFERENCES `user_status` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    
    PRIMARY KEY (id)
);


INSERT INTO `user_status` (`id`, `status`, `created_at`, `updated_at`, `deleted`) VALUES
(1, 'active',   CURRENT_TIMESTAMP,  CURRENT_TIMESTAMP,  0),
(2, 'inactive', CURRENT_TIMESTAMP,  CURRENT_TIMESTAMP,  0);
*/