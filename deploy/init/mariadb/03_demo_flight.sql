-- Spider flight_2 数据库导入脚本 (MariaDB)
-- 用于演示空格问题和孤儿记录

DROP DATABASE IF EXISTS spider_flight;
CREATE DATABASE spider_flight CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE spider_flight;

-- airlines 表
CREATE TABLE airlines (
    uid INT PRIMARY KEY,
    Airline VARCHAR(100),
    Abbreviation VARCHAR(50),
    Country VARCHAR(100)
) ENGINE=InnoDB;

-- airports 表
CREATE TABLE airports (
    City VARCHAR(100),
    AirportCode VARCHAR(10) PRIMARY KEY,
    AirportName VARCHAR(200),
    Country VARCHAR(100),
    CountryAbbrev VARCHAR(10)
) ENGINE=InnoDB;

-- flights 表
CREATE TABLE flights (
    Airline INT,
    FlightNo INT,
    SourceAirport VARCHAR(10),
    DestAirport VARCHAR(10),
    PRIMARY KEY(Airline, FlightNo),
    FOREIGN KEY (Airline) REFERENCES airlines(uid),
    -- 注意：故意不加外键约束到 airports，以保留孤儿记录问题
    INDEX idx_source (SourceAirport),
    INDEX idx_dest (DestAirport)
) ENGINE=InnoDB;
