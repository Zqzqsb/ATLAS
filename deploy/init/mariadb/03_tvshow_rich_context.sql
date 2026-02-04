-- ============================================================
-- LUCID Rich Context: Spider TV Show
-- Pre-generated semantic context for demo
-- ============================================================

USE lucid;

-- ============================================================
-- Register datasource
-- ============================================================
INSERT INTO rc_datasources (name, db_type, host, port, db_name, description, status)
VALUES ('spider_tvshow', 'mysql', 'localhost', 3306, 'spider_tvshow', 
        'Spider benchmark tv_1 database for ReAct reasoning demo', 'active')
ON DUPLICATE KEY UPDATE description = VALUES(description);

SET @ds_id = (SELECT id FROM rc_datasources WHERE name = 'spider_tvshow' LIMIT 1);

-- ============================================================
-- Table-level Rich Context
-- ============================================================
INSERT INTO rc_tables (datasource_id, table_name, description, row_count, is_expired, created_at)
VALUES 
(@ds_id, 'TV_Channel', 'TV channel information including series name, country, language, and package options. Each channel has unique features like HD support and pay-per-view availability.', 10, 0, NOW()),
(@ds_id, 'TV_series', 'TV series episode information with ratings, viewership data, and weekly rankings. Links to TV_Channel via Channel foreign key.', 10, 0, NOW()),
(@ds_id, 'Cartoon', 'Cartoon show information including title, director, writer, and air date. Links to TV_Channel via Channel foreign key.', 10, 0, NOW())
ON DUPLICATE KEY UPDATE description = VALUES(description), updated_at = NOW();

-- ============================================================
-- Column-level Rich Context
-- ============================================================

-- TV_Channel columns
INSERT INTO rc_columns (datasource_id, table_name, column_name, data_type, description, sample_values, is_nullable, created_at)
VALUES
(@ds_id, 'TV_Channel', 'id', 'INT', 'Unique channel identifier (primary key)', '1, 2, 3, 4, 5', 0, NOW()),
(@ds_id, 'TV_Channel', 'series_name', 'VARCHAR(200)', 'TV channel/series name', 'Sky News, BBC One, CNN, HBO', 1, NOW()),
(@ds_id, 'TV_Channel', 'Country', 'VARCHAR(100)', 'Country of origin', 'United Kingdom, United States, Japan, China', 1, NOW()),
(@ds_id, 'TV_Channel', 'Language', 'VARCHAR(50)', 'Primary broadcast language', 'English, Japanese, Chinese, German, French', 1, NOW()),
(@ds_id, 'TV_Channel', 'Content', 'VARCHAR(500)', 'Content description/genre', 'News, Entertainment, Documentary, Animation', 1, NOW()),
(@ds_id, 'TV_Channel', 'Hight_definition_TV', 'VARCHAR(10)', 'HD support flag: yes/no', 'yes, no', 1, NOW()),
(@ds_id, 'TV_Channel', 'Pay_per_view_PPV', 'VARCHAR(10)', 'Pay-per-view availability: yes/no', 'yes, no', 1, NOW()),
(@ds_id, 'TV_Channel', 'Package_Option', 'VARCHAR(100)', 'Subscription package tier', 'Basic, Standard, Premium', 1, NOW())
ON DUPLICATE KEY UPDATE description = VALUES(description), sample_values = VALUES(sample_values), updated_at = NOW();

-- TV_series columns
INSERT INTO rc_columns (datasource_id, table_name, column_name, data_type, description, sample_values, is_nullable, created_at)
VALUES
(@ds_id, 'TV_series', 'id', 'INT', 'Unique episode identifier (primary key)', '1, 2, 3', 0, NOW()),
(@ds_id, 'TV_series', 'Episode', 'VARCHAR(100)', 'Episode identifier in SxxExx format', 'S01E01, S01E02, S02E01, Special, Pilot', 1, NOW()),
(@ds_id, 'TV_series', 'Air_Date', 'VARCHAR(50)', 'Original air date in YYYY-MM-DD format', '2023-09-15, 2024-01-10', 1, NOW()),
(@ds_id, 'TV_series', 'Rating', 'DECIMAL(3,1)', 'Episode rating on 0-10 scale', '8.5, 9.1, 7.5', 1, NOW()),
(@ds_id, 'TV_series', 'Share', 'DECIMAL(4,1)', 'Market share percentage', '12.3, 15.0, 18.0', 1, NOW()),
(@ds_id, 'TV_series', 'Viewers_m', 'DECIMAL(5,2)', 'Number of viewers in millions', '10.25, 14.50, 16.00', 1, NOW()),
(@ds_id, 'TV_series', 'Weekly_Rank', 'INT', 'Weekly ranking position (1 = top)', '1, 2, 5, 15', 1, NOW()),
(@ds_id, 'TV_series', 'Channel', 'INT', 'Foreign key to TV_Channel.id', '2, 4, 5', 1, NOW())
ON DUPLICATE KEY UPDATE description = VALUES(description), sample_values = VALUES(sample_values), updated_at = NOW();

-- Cartoon columns
INSERT INTO rc_columns (datasource_id, table_name, column_name, data_type, description, sample_values, is_nullable, created_at)
VALUES
(@ds_id, 'Cartoon', 'id', 'INT', 'Unique cartoon identifier (primary key)', '1, 2, 3', 0, NOW()),
(@ds_id, 'Cartoon', 'Title', 'VARCHAR(200)', 'Cartoon show title', 'Adventure Time, Steven Universe, Samurai Jack', 1, NOW()),
(@ds_id, 'Cartoon', 'Directed_by', 'VARCHAR(100)', 'Director name', 'Genndy Tartakovsky, Rebecca Sugar', 1, NOW()),
(@ds_id, 'Cartoon', 'Written_by', 'VARCHAR(200)', 'Writer name(s)', 'Pendleton Ward, Craig McCracken', 1, NOW()),
(@ds_id, 'Cartoon', 'Original_air_date', 'VARCHAR(50)', 'First air date', '2010-04-05, 2001-08-10', 1, NOW()),
(@ds_id, 'Cartoon', 'Production_code', 'VARCHAR(20)', 'Production episode code', 'AT101, SJ101, PPG101', 1, NOW()),
(@ds_id, 'Cartoon', 'Channel', 'INT', 'Foreign key to TV_Channel.id (most cartoons on channel 6 = Cartoon Network)', '6', 1, NOW())
ON DUPLICATE KEY UPDATE description = VALUES(description), sample_values = VALUES(sample_values), updated_at = NOW();

-- ============================================================
-- Table Relations
-- ============================================================
INSERT INTO rc_relations (datasource_id, from_table, from_column, to_table, to_column, relation_type, description, created_at)
VALUES
(@ds_id, 'TV_series', 'Channel', 'TV_Channel', 'id', 'many_to_one', 'Each TV series episode belongs to one channel', NOW()),
(@ds_id, 'Cartoon', 'Channel', 'TV_Channel', 'id', 'many_to_one', 'Each cartoon belongs to one channel (mostly Cartoon Network)', NOW())
ON DUPLICATE KEY UPDATE description = VALUES(description), updated_at = NOW();

-- ============================================================
-- Business Terms
-- ============================================================
INSERT INTO rc_terms (datasource_id, term, definition, synonyms, examples, created_at)
VALUES
(@ds_id, 'rating', 'Numerical score representing episode quality on a 0-10 scale', 'score, grade', 'A rating of 9.0 means excellent quality', NOW()),
(@ds_id, 'viewers', 'Number of people watching, typically measured in millions (Viewers_m)', 'audience, viewership', '10.25 means 10.25 million viewers', NOW()),
(@ds_id, 'weekly_rank', 'Position in weekly rankings where 1 is the highest/best', 'ranking, position', 'Weekly_Rank=1 means top-rated show of the week', NOW()),
(@ds_id, 'HD', 'High Definition television support', 'high definition, Hight_definition_TV', 'Hight_definition_TV=yes means channel supports HD', NOW()),
(@ds_id, 'PPV', 'Pay-Per-View, premium content requiring additional payment', 'pay per view, Pay_per_view_PPV', 'Pay_per_view_PPV=yes means extra payment required', NOW())
ON DUPLICATE KEY UPDATE definition = VALUES(definition), updated_at = NOW();

-- ============================================================
-- Sample Queries (for demo reference)
-- ============================================================
INSERT INTO rc_change_log (datasource_id, change_type, entity_type, entity_name, old_value, new_value, changed_by, created_at)
VALUES
(@ds_id, 'init', 'system', 'sample_queries', NULL, 
 'Q1: What are all the channels from the United Kingdom?\nQ2: Which cartoon has the earliest air date?\nQ3: What is the average rating of TV series on BBC One?\nQ4: List all cartoons directed by Genndy Tartakovsky.\nQ5: Which channel has the most cartoons?',
 'system', NOW());

SELECT 'Rich Context for spider_tvshow initialized successfully!' AS status;
