-- ============================================================
-- ATLAS Demo Database 1: Spider TV Show
-- Purpose: Demonstrate complete ReAct reasoning pipeline
-- Source: Spider benchmark tv_1 database
-- ============================================================

-- Create database
CREATE DATABASE IF NOT EXISTS spider_tvshow CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE spider_tvshow;

-- ============================================================
-- Table: TV_Channel
-- ============================================================
CREATE TABLE TV_Channel (
    id INT PRIMARY KEY,
    series_name VARCHAR(200),
    Country VARCHAR(100),
    Language VARCHAR(50),
    Content VARCHAR(500),
    Pixel_aspect_ratio_PAR VARCHAR(20),
    Hight_definition_TV VARCHAR(10),
    Pay_per_view_PPV VARCHAR(10),
    Package_Option VARCHAR(100)
);

-- ============================================================
-- Table: TV_series
-- ============================================================
CREATE TABLE TV_series (
    id INT PRIMARY KEY,
    Episode VARCHAR(100),
    Air_Date VARCHAR(50),
    Rating DECIMAL(3,1),
    Share DECIMAL(4,1),
    18_49_Rating_Share VARCHAR(20),
    Viewers_m DECIMAL(5,2),
    Weekly_Rank INT,
    Channel INT,
    FOREIGN KEY (Channel) REFERENCES TV_Channel(id)
);

-- ============================================================
-- Table: Cartoon
-- ============================================================
CREATE TABLE Cartoon (
    id INT PRIMARY KEY,
    Title VARCHAR(200),
    Directed_by VARCHAR(100),
    Written_by VARCHAR(200),
    Original_air_date VARCHAR(50),
    Production_code VARCHAR(20),
    Channel INT,
    FOREIGN KEY (Channel) REFERENCES TV_Channel(id)
);

-- ============================================================
-- Insert sample data: TV_Channel
-- ============================================================
INSERT INTO TV_Channel VALUES
(1, 'Sky News', 'United Kingdom', 'English', 'News and current affairs', '16:9', 'yes', 'no', 'Basic'),
(2, 'BBC One', 'United Kingdom', 'English', 'Entertainment, drama, news', '16:9', 'yes', 'no', 'Basic'),
(3, 'CNN', 'United States', 'English', '24-hour news coverage', '16:9', 'yes', 'no', 'Basic'),
(4, 'HBO', 'United States', 'English', 'Premium movies and series', '16:9', 'yes', 'yes', 'Premium'),
(5, 'Discovery Channel', 'United States', 'English', 'Documentary and educational', '16:9', 'yes', 'no', 'Standard'),
(6, 'Cartoon Network', 'United States', 'English', 'Animation and cartoons', '16:9', 'yes', 'no', 'Standard'),
(7, 'NHK', 'Japan', 'Japanese', 'Public broadcasting', '16:9', 'yes', 'no', 'Basic'),
(8, 'CCTV-1', 'China', 'Chinese', 'General entertainment', '16:9', 'yes', 'no', 'Basic'),
(9, 'ZDF', 'Germany', 'German', 'Public television', '16:9', 'yes', 'no', 'Basic'),
(10, 'France 2', 'France', 'French', 'General interest', '16:9', 'yes', 'no', 'Basic');

-- ============================================================
-- Insert sample data: TV_series
-- ============================================================
INSERT INTO TV_series VALUES
(1, 'S01E01', '2023-09-15', 8.5, 12.3, '4.2/12', 10.25, 5, 2),
(2, 'S01E02', '2023-09-22', 8.7, 13.1, '4.5/13', 11.02, 4, 2),
(3, 'S01E03', '2023-09-29', 8.9, 14.2, '4.8/14', 12.15, 3, 2),
(4, 'S01E04', '2023-10-06', 9.1, 15.0, '5.1/15', 13.20, 2, 2),
(5, 'S01E05', '2023-10-13', 9.3, 16.5, '5.5/16', 14.50, 1, 2),
(6, 'S02E01', '2024-01-10', 8.2, 11.5, '3.9/11', 9.80, 8, 4),
(7, 'S02E02', '2024-01-17', 8.4, 12.0, '4.1/12', 10.30, 6, 4),
(8, 'S02E03', '2024-01-24', 8.8, 13.5, '4.6/13', 11.75, 4, 4),
(9, 'Special', '2023-12-25', 9.5, 18.0, '6.0/18', 16.00, 1, 2),
(10, 'Pilot', '2023-06-01', 7.5, 8.5, '2.8/8', 7.20, 15, 5);

-- ============================================================
-- Insert sample data: Cartoon
-- ============================================================
INSERT INTO Cartoon VALUES
(1, 'Adventure Time', 'Larry Leichliter', 'Pendleton Ward', '2010-04-05', 'AT101', 6),
(2, 'Regular Show', 'John Paul Bong', 'J.G. Quintel', '2010-09-06', 'RS101', 6),
(3, 'Steven Universe', 'Rebecca Sugar', 'Rebecca Sugar', '2013-11-04', 'SU101', 6),
(4, 'The Amazing World of Gumball', 'Mic Graves', 'Ben Bocquelet', '2011-05-03', 'TAWOG101', 6),
(5, 'Teen Titans Go!', 'Peter Rida Michail', 'Michael Jelenic', '2013-04-23', 'TTG101', 6),
(6, 'We Bare Bears', 'Manny Hernandez', 'Daniel Chong', '2015-07-27', 'WBB101', 6),
(7, 'Samurai Jack', 'Genndy Tartakovsky', 'Genndy Tartakovsky', '2001-08-10', 'SJ101', 6),
(8, 'Dexters Laboratory', 'Genndy Tartakovsky', 'Genndy Tartakovsky', '1996-04-28', 'DL101', 6),
(9, 'The Powerpuff Girls', 'Craig McCracken', 'Craig McCracken', '1998-11-18', 'PPG101', 6),
(10, 'Courage the Cowardly Dog', 'John R. Dilworth', 'John R. Dilworth', '1999-11-12', 'CCD101', 6);

-- ============================================================
-- Grant permissions to atlas user
-- ============================================================
GRANT ALL PRIVILEGES ON spider_tvshow.* TO 'atlas'@'%';
FLUSH PRIVILEGES;

SELECT 'Spider tvshow database initialized successfully!' AS status;
