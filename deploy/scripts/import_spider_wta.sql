-- Spider wta_1 database import script (MariaDB)
-- Demonstrates null value abuse and whitespace issues

DROP DATABASE IF EXISTS spider_wta;
CREATE DATABASE spider_wta CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE spider_wta;

-- players table
CREATE TABLE players (
    player_id INT PRIMARY KEY,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    hand VARCHAR(10),
    birth_date DATE,
    country_code VARCHAR(10)
) ENGINE=InnoDB;

-- matches table
CREATE TABLE matches (
    best_of INT,
    draw_size INT,
    loser_age FLOAT,
    loser_entry VARCHAR(50),
    loser_hand VARCHAR(20),
    loser_ht INT,
    loser_id INT,
    loser_ioc VARCHAR(20),
    loser_name VARCHAR(200),
    loser_rank INT,
    loser_rank_points INT,
    loser_seed INT,
    match_num INT,
    minutes INT,
    round VARCHAR(50),
    score VARCHAR(100),
    surface VARCHAR(50),
    tourney_date DATE,
    tourney_id VARCHAR(50),
    tourney_level VARCHAR(20),
    tourney_name VARCHAR(200),
    winner_age FLOAT,
    winner_entry VARCHAR(50),
    winner_hand VARCHAR(20),
    winner_ht INT,
    winner_id INT,
    winner_ioc VARCHAR(20),
    winner_name VARCHAR(200),
    winner_rank INT,
    winner_rank_points INT,
    winner_seed INT,
    year INT,
    INDEX idx_loser (loser_id),
    INDEX idx_winner (winner_id)
) ENGINE=InnoDB;

-- rankings table
CREATE TABLE rankings (
    ranking_date DATE,
    ranking INT,
    player_id INT,
    ranking_points INT,
    tours INT,
    INDEX idx_player (player_id),
    INDEX idx_date (ranking_date)
) ENGINE=InnoDB;
