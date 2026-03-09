-- Spider flight_2 database import script (MariaDB)
-- Demonstrates whitespace issues and orphan records

DROP DATABASE IF EXISTS spider_flight;
CREATE DATABASE spider_flight CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE spider_flight;

-- airlines table
CREATE TABLE airlines (
    uid INT PRIMARY KEY,
    Airline VARCHAR(100),
    Abbreviation VARCHAR(50),
    Country VARCHAR(100)
) ENGINE=InnoDB;

-- airports table
CREATE TABLE airports (
    City VARCHAR(100),
    AirportCode VARCHAR(10) PRIMARY KEY,
    AirportName VARCHAR(200),
    Country VARCHAR(100),
    CountryAbbrev VARCHAR(10)
) ENGINE=InnoDB;

-- flights table
CREATE TABLE flights (
    Airline INT,
    FlightNo INT,
    SourceAirport VARCHAR(10),
    DestAirport VARCHAR(10),
    PRIMARY KEY(Airline, FlightNo),
    FOREIGN KEY (Airline) REFERENCES airlines(uid),
    -- Note: intentionally no FK constraint to airports, to preserve orphan record issues
    INDEX idx_source (SourceAirport),
    INDEX idx_dest (DestAirport)
) ENGINE=InnoDB;
