-- MySQL database for mmgorogue.

CREATE DATABASE IF NOT EXISTS mmgorogue;
USE mmgorogue;

CREATE USER 'server'@'localhost' IDENTIFIED BY 'mysql';
GRANT ALL PRIVILEGES ON mmgorogue.* TO 'server'@'localhost';

CREATE TABLE IF NOT EXISTS users (
       user_name CHAR(32)
);
