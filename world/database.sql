-- MySQL database for mmgorogue.

CREATE DATABASE IF NOT EXISTS mmgorogue;
USE mmgorogue;

CREATE USER 'server'@'localhost' IDENTIFIED BY 'mysql';
GRANT ALL PRIVILEGES ON mmgorogue.* TO 'server'@'localhost';

CREATE TABLE IF NOT EXISTS users (
       id INT NOT NULL AUTO_INCREMENT,
       user_name CHAR(16) NOT NULL,
       password CHAR(128) NOT NULL,
       email CHAR(254),
       creation_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
       last_login TIMESTAMP NULL,
       PRIMARY KEY (id),
       UNIQUE KEY (user_name)
);

CREATE TABLE IF NOT EXISTS items (
       id INT NOT NULL AUTO_INCREMENT,
       item_name CHAR(64) NOT NULL,
       PRIMARY KEY (id),
       UNIQUE KEY (item_name)
);

CREATE TABLE IF NOT EXISTS inventory (
       user_id INT NOT NULL,
       item_id INT NOT NULL,
       FOREIGN KEY (user_id) REFERENCES users(id)
       	       ON DELETE CASCADE,
       FOREIGN KEY (item_id) REFERENCES items(id)
       	       ON DELETE CASCADE
);
