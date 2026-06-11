-- ============================================================
-- ATLAS Demo Database 2: Spider Flight
-- Purpose: Demonstrate airline/airport/flight queries
-- Source: Spider benchmark flight_2 database
-- ============================================================

-- Create database
CREATE DATABASE IF NOT EXISTS spider_flight CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE spider_flight;

-- ============================================================
-- Table: airlines
-- ============================================================
CREATE TABLE IF NOT EXISTS airlines (
    uid INT PRIMARY KEY,
    Airline VARCHAR(100),
    Abbreviation VARCHAR(50),
    Country VARCHAR(100)
) ENGINE=InnoDB COMMENT='Airline information';

-- ============================================================
-- Table: airports
-- ============================================================
CREATE TABLE IF NOT EXISTS airports (
    City VARCHAR(100),
    AirportCode VARCHAR(10) PRIMARY KEY,
    AirportName VARCHAR(200),
    Country VARCHAR(100),
    CountryAbbrev VARCHAR(10)
) ENGINE=InnoDB COMMENT='Airport information';

-- ============================================================
-- Table: flights
-- ============================================================
CREATE TABLE IF NOT EXISTS flights (
    Airline INT,
    FlightNo INT,
    SourceAirport VARCHAR(10),
    DestAirport VARCHAR(10),
    PRIMARY KEY(Airline, FlightNo),
    FOREIGN KEY (Airline) REFERENCES airlines(uid),
    INDEX idx_source (SourceAirport),
    INDEX idx_dest (DestAirport)
) ENGINE=InnoDB COMMENT='Flight route information';

-- ============================================================
-- Insert data: airlines (12 rows)
-- ============================================================
INSERT INTO airlines (uid, Airline, Abbreviation, Country) VALUES
(1, 'United Airlines', 'UAL', 'USA'),
(2, 'US Airways', 'USAir', 'USA'),
(3, 'Delta Airlines', 'Delta', 'USA'),
(4, 'Southwest Airlines', 'Southwest', 'USA'),
(5, 'American Airlines', 'American', 'USA'),
(6, 'Northwest Airlines', 'Northwest', 'USA'),
(7, 'Continental Airlines', 'Continental', 'USA'),
(8, 'JetBlue Airways', 'JetBlue', 'USA'),
(9, 'Frontier Airlines', 'Frontier', 'USA'),
(10, 'AirTran Airways', 'AirTran', 'USA'),
(11, 'Allegiant Air', 'Allegiant', 'USA'),
(12, 'Virgin America', 'Virgin', 'USA');

-- ============================================================
-- Insert data: airports (100 rows)
-- ============================================================
INSERT INTO airports (City, AirportCode, AirportName, Country, CountryAbbrev) VALUES
('Aberdeen', 'APG', 'Phillips AAF', 'United States', 'US'),
('Aberdeen', 'ABR', 'Municipal', 'United States', 'US'),
('Abilene', 'DYS', 'Dyess AFB', 'United States', 'US'),
('Abilene', 'ABI', 'Municipal', 'United States', 'US'),
('Abingdon', 'VJI', 'Virginia Highlands', 'United States', 'US'),
('Ada', 'ADT', 'Ada', 'United States', 'US'),
('Adak Island', 'ADK', 'Adak Island Ns', 'United States', 'US'),
('Adrian', 'ADG', 'Lenawee County', 'United States', 'US'),
('Afton', 'AFO', 'Municipal', 'United States', 'US'),
('Aiken', 'AIK', 'Municipal', 'United States', 'US'),
('Ainsworth', 'ANW', 'Ainsworth', 'United States', 'US'),
('Akhiok', 'AKK', 'Akhiok SPB', 'United States', 'US'),
('Akiachak', 'KKI', 'Spb', 'United States', 'US'),
('Akiak', 'AKI', 'Akiak', 'United States', 'US'),
('Akron CO', 'AKO', 'Colorado Plains Regional Airport', 'United States', 'US'),
('Akron/Canton OH', 'CAK', 'Akron/canton Regional', 'United States', 'US'),
('Akron/Canton', 'AKC', 'Fulton International', 'United States', 'US'),
('Akutan', 'KQA', 'Akutan', 'United States', 'US'),
('Alakanuk', 'AUK', 'Alakanuk', 'United States', 'US'),
('Alameda', 'NGZ', 'NAS', 'United States', 'US'),
('Alamogordo', 'HMN', 'Holloman AFB', 'United States', 'US'),
('Alamogordo', 'ALM', 'Municipal', 'United States', 'US'),
('Alamosa', 'ALS', 'Municipal', 'United States', 'US'),
('Albany', 'NAB', 'Albany NAS', 'United States', 'US'),
('Albany', 'ABY', 'Dougherty County', 'United States', 'US'),
('Albany', 'ALB', 'Albany International', 'United States', 'US'),
('Albany', 'CVO', 'Albany', 'United States', 'US'),
('Albert Lea', 'AEL', 'Albert Lea', 'United States', 'US'),
('Albuquerque', 'ABQ', 'Albuquerque International', 'United States', 'US'),
('Aleknagik', 'WKK', 'Aleknagik', 'United States', 'US'),
('Aleneva', 'AED', 'Aleneva', 'United States', 'US'),
('Alexander City AL', 'ALX', 'Thomas C Russell Fld', 'United States', 'US'),
('Alexandria LA', 'AEX', 'Alexandria International', 'United States', 'US'),
('Alexandria', 'ESF', 'Esler Field', 'United States', 'US'),
('Alexandria', 'AXN', 'Alexandria', 'United States', 'US'),
('Alexandria Bay', 'AXB', 'Alexandria Bay', 'United States', 'US'),
('Algona', 'AXG', 'Algona', 'United States', 'US'),
('Alice', 'ALI', 'International', 'United States', 'US'),
('Aliceville AL', 'AIV', 'George Downer', 'United States', 'US'),
('Alitak', 'ALZ', 'Alitak SPB', 'United States', 'US'),
('Allakaket', 'AET', 'Allakaket', 'United States', 'US'),
('Alliance', 'AIA', 'Alliance', 'United States', 'US'),
('Alma', 'AMN', 'Gratiot Community', 'United States', 'US'),
('Alpena', 'APN', 'Alpena County Regional', 'United States', 'US'),
('Alpine', 'ALE', 'Alpine', 'United States', 'US'),
('Alton', 'ALN', 'Alton', 'United States', 'US'),
('Altus', 'LTS', 'Altus AFB', 'United States', 'US'),
('Altus', 'AXS', 'Municipal', 'United States', 'US'),
('Alyeska', 'AQY', 'Alyeska', 'United States', 'US'),
('Amarillo', 'AMA', 'Rick Husband Amarillo International', 'United States', 'US'),
('Amarillo', 'TDW', 'Tradewind', 'United States', 'US'),
('Ambler', 'ABL', 'Ambler', 'United States', 'US'),
('Amchitka', 'AHT', 'Amchitka', 'United States', 'US'),
('Amery', 'AHH', 'Municipal', 'United States', 'US'),
('Ames', 'AMW', 'Ames', 'United States', 'US'),
('Amityville', 'AYZ', 'Zahns', 'United States', 'US'),
('Amook', 'AOS', 'Amook', 'United States', 'US'),
('Anacortes', 'OTS', 'Anacortes', 'United States', 'US'),
('Anacostia', 'NDV', 'USN Heliport', 'United States', 'US'),
('Anaheim', 'ANA', 'Orange County Steel Salvage Heliport', 'United States', 'US'),
('Anaktuvuk', 'AKP', 'Anaktuvuk', 'United States', 'US'),
('Anchorage', 'EDF', 'Elmendorf Afb', 'United States', 'US'),
('Anchorage', 'ANC', 'Ted Stevens Anchorage International Airport', 'United States', 'US'),
('Anchorage', 'MRI', 'Merrill Field', 'United States', 'US'),
('Anderson', 'AID', 'Municipal', 'United States', 'US'),
('Anderson', 'AND', 'Anderson', 'United States', 'US'),
('Andrews', 'ADR', 'Andrews', 'United States', 'US'),
('Angel Fire', 'AXX', 'Angel Fire', 'United States', 'US'),
('Angola', 'ANQ', 'Tri-State Steuben Cty', 'United States', 'US'),
('Angoon', 'AGN', 'Angoon', 'United States', 'US'),
('Anguilla', 'RFK', 'Rollang Field', 'United States', 'US'),
('Aniak', 'ANI', 'Aniak', 'United States', 'US'),
('Anita Bay', 'AIB', 'Anita Bay', 'United States', 'US'),
('Ann Arbor MI', 'ARB', 'Municipal', 'United States', 'US'),
('Annapolis', 'ANP', 'Lee', 'United States', 'US'),
('Annette Island', 'ANN', 'Annette Island', 'United States', 'US'),
('Anniston AL', 'ANB', 'Anniston Metropolitan', 'United States', 'US'),
('Anniston', 'QAW', 'Ft Mcclellan Bus Trml', 'United States', 'US'),
('Anniston', 'RLI', 'Reilly AHP', 'United States', 'US'),
('Anthony', 'ANY', 'Anthony', 'United States', 'US'),
('Antlers', 'ATE', 'Antlers', 'United States', 'US'),
('Anvik', 'ANV', 'Anvik', 'United States', 'US'),
('Apalachicola', 'AAF', 'Municipal', 'United States', 'US'),
('Apple Valley', 'APV', 'Apple Valley', 'United States', 'US'),
('Appleton', 'ATW', 'Outagamie County', 'United States', 'US'),
('Arapahoe', 'AHF', 'Municipal', 'United States', 'US'),
('Arcata', 'ACV', 'Arcata', 'United States', 'US'),
('Arctic Village', 'ARC', 'Arctic Village', 'United States', 'US'),
('Ardmore', 'AHD', 'Downtown', 'United States', 'US'),
('Ardmore', 'ADM', 'Ardmore Municipal Arpt', 'United States', 'US'),
('Arlington Heights', 'JLH', 'US Army Heliport', 'United States', 'US'),
('Artesia', 'ATS', 'Artesia', 'United States', 'US'),
('Neptune', 'ARX', 'Asbury Park', 'United States', 'US'),
('Ashland', 'ASX', 'Ashland', 'United States', 'US'),
('Ashley', 'ASY', 'Ashley', 'United States', 'US'),
('Aspen', 'ASE', 'Aspen', 'United States', 'US'),
('Astoria', 'AST', 'Astoria', 'United States', 'US'),
('Athens', 'AHN', 'Athens', 'United States', 'US'),
('Athens', 'ATO', 'Ohio University', 'United States', 'US'),
('Athens', 'MMI', 'McMinn County', 'United States', 'US');

-- ============================================================
-- Insert data: flights (100 rows, sampled from 1200)
-- ============================================================
INSERT INTO flights (Airline, FlightNo, SourceAirport, DestAirport) VALUES
(1, 28, 'APG', 'ASY'),
(1, 148, 'HMN', 'ABL'),
(1, 370, 'AKC', 'WKK'),
(1, 560, 'AHF', 'ATO'),
(1, 744, 'AED', 'OTS'),
(1, 888, 'ALZ', 'ANW'),
(1, 1026, 'AHN', 'DYS'),
(1, 1168, 'ADK', 'NAB'),
(1, 1284, 'ADK', 'ADG'),
(2, 124, 'OTS', 'ALZ'),
(2, 330, 'AYZ', 'AND'),
(2, 484, 'ARX', 'AQY'),
(2, 670, 'KKI', 'ATS'),
(2, 830, 'ANQ', 'AMA'),
(2, 988, 'ALS', 'ADG'),
(2, 1158, 'ABI', 'QAW'),
(2, 1330, 'AXX', 'ANP'),
(3, 36, 'CAK', 'OTS'),
(3, 162, 'AIK', 'ADT'),
(3, 284, 'AXB', 'AXN'),
(3, 438, 'AFO', 'AUK'),
(3, 626, 'KQA', 'ANV'),
(3, 782, 'AXG', 'JLH'),
(3, 970, 'ANP', 'AIB'),
(3, 1140, 'ATW', 'ALB'),
(4, 10, 'ASY', 'ATS'),
(4, 136, 'LTS', 'JLH'),
(4, 264, 'ARB', 'ATW'),
(4, 430, 'CVO', 'LTS'),
(4, 558, 'AGN', 'AHN'),
(4, 666, 'AAF', 'AMA'),
(4, 838, 'APV', 'ALM'),
(4, 1064, 'ANV', 'AET'),
(4, 1206, 'AXN', 'NDV'),
(5, 104, 'ALB', 'ANY'),
(5, 242, 'ADR', 'LTS'),
(5, 360, 'ARX', 'ANW'),
(5, 504, 'ASX', 'AND'),
(5, 688, 'ANI', 'DYS'),
(5, 866, 'AND', 'ALE'),
(5, 1010, 'RLI', 'LTS'),
(5, 1248, 'ABR', 'ALE'),
(6, 50, 'AIB', 'ALZ'),
(6, 178, 'ABQ', 'ALI'),
(6, 368, 'JLH', 'NAB'),
(6, 478, 'APG', 'AEL'),
(6, 658, 'ALM', 'CVO'),
(6, 754, 'AAF', 'ARX'),
(6, 906, 'ARC', 'RFK'),
(6, 1070, 'ADK', 'ABR'),
(7, 44, 'AKI', 'ABR'),
(7, 120, 'AKI', 'ALX'),
(7, 294, 'AET', 'AKP'),
(7, 408, 'AMN', 'ADG'),
(7, 570, 'OTS', 'ADG'),
(7, 722, 'AST', 'CAK'),
(7, 844, 'MMI', 'AOS'),
(7, 980, 'AUK', 'AIA'),
(7, 1126, 'AKI', 'ESF'),
(8, 174, 'DYS', 'ABQ'),
(8, 420, 'AKP', 'ADT'),
(8, 588, 'AEX', 'AXN'),
(8, 726, 'ACV', 'ASY'),
(8, 872, 'ABY', 'CAK'),
(8, 1086, 'EDF', 'AFO'),
(8, 1280, 'ATO', 'AET'),
(8, 1420, 'ALM', 'ANV'),
(9, 78, 'ASE', 'JLH'),
(9, 230, 'ANY', 'ABQ'),
(9, 390, 'DYS', 'AIV'),
(9, 616, 'HMN', 'AND'),
(9, 680, 'ABY', 'AGN'),
(9, 828, 'ADG', 'HMN'),
(9, 968, 'CAK', 'ALN'),
(9, 1114, 'LTS', 'ABI'),
(10, 6, 'TDW', 'AXN'),
(10, 218, 'AKO', 'ANN'),
(10, 382, 'APG', 'ALE'),
(10, 600, 'ABQ', 'AKI'),
(10, 784, 'NGZ', 'NAB'),
(10, 934, 'ABI', 'AKP'),
(10, 1128, 'ABI', 'AMN'),
(10, 1256, 'AED', 'ADM'),
(10, 1426, 'JLH', 'ABL'),
(11, 110, 'AXS', 'AMW'),
(11, 240, 'ARC', 'AXG'),
(11, 400, 'ANA', 'AEL'),
(11, 566, 'LTS', 'AET'),
(11, 640, 'AND', 'AIA'),
(11, 774, 'ALN', 'AKP'),
(11, 908, 'ANQ', 'ATO'),
(11, 1078, 'ALZ', 'ADG'),
(12, 120, 'ASY', 'ADK'),
(12, 284, 'ADM', 'RLI'),
(12, 450, 'ALB', 'HMN'),
(12, 690, 'AEX', 'ANI'),
(12, 834, 'ABI', 'AIV'),
(12, 970, 'AST', 'ADM'),
(12, 1126, 'RLI', 'ABY'),
(12, 1254, 'AIB', 'ANV');

-- ============================================================
-- Grant permissions to lucid user
-- ============================================================
GRANT ALL PRIVILEGES ON spider_flight.* TO 'lucid'@'%';
FLUSH PRIVILEGES;

SELECT 'Spider flight database initialized successfully!' AS status;
