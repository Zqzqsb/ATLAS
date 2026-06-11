#!/usr/bin/env python3
"""
Spider database migration script: SQLite -> MariaDB
"""
import sqlite3
import pymysql
import sys
import os
from pathlib import Path

# MariaDB connection config
MARIA_CONFIG = {
    'host': os.environ.get('MARIADB_HOST', '127.0.0.1'),
    'port': int(os.environ.get('MARIADB_PORT', '19010')),
    'user': os.environ.get('MARIADB_USER', 'root'),
    'password': os.environ.get('MARIADB_ROOT_PASSWORD', ''),
    'charset': 'utf8mb4'
}

# Spider data path
SPIDER_DATA = Path('/root/workspace/ReActSql/data/spider_data/database')


def migrate_flight2():
    """Migrate flight_2 database"""
    print("=== Migrating flight_2 ===")
    
    sqlite_path = SPIDER_DATA / 'flight_2' / 'flight_2.sqlite'
    sqlite_conn = sqlite3.connect(str(sqlite_path))
    sqlite_conn.row_factory = sqlite3.Row
    
    maria_conn = pymysql.connect(**MARIA_CONFIG)
    cursor = maria_conn.cursor()
    
    # Create database and tables
    with open('/root/workspace/atlas/deploy/scripts/import_spider_flight.sql', 'r') as f:
        sql_script = f.read()
    
    for statement in sql_script.split(';'):
        statement = statement.strip()
        if statement:
            try:
                cursor.execute(statement)
            except Exception as e:
                print(f"  Warning: {e}")
    
    maria_conn.commit()
    cursor.execute("USE spider_flight")
    
    # Migrate airlines
    print("  Migrating airlines...")
    rows = sqlite_conn.execute("SELECT * FROM airlines").fetchall()
    for row in rows:
        cursor.execute(
            "INSERT INTO airlines (uid, Airline, Abbreviation, Country) VALUES (%s, %s, %s, %s)",
            (row['uid'], row['Airline'], row['Abbreviation'], row['Country'])
        )
    print(f"    Inserted {len(rows)} rows")
    
    # Migrate airports
    print("  Migrating airports...")
    rows = sqlite_conn.execute("SELECT * FROM airports").fetchall()
    for row in rows:
        cursor.execute(
            "INSERT INTO airports (City, AirportCode, AirportName, Country, CountryAbbrev) VALUES (%s, %s, %s, %s, %s)",
            (row['City'], row['AirportCode'], row['AirportName'], row['Country'], row['CountryAbbrev'])
        )
    print(f"    Inserted {len(rows)} rows")
    
    # Migrate flights
    print("  Migrating flights...")
    rows = sqlite_conn.execute("SELECT * FROM flights").fetchall()
    for row in rows:
        cursor.execute(
            "INSERT INTO flights (Airline, FlightNo, SourceAirport, DestAirport) VALUES (%s, %s, %s, %s)",
            (row['Airline'], row['FlightNo'], row['SourceAirport'], row['DestAirport'])
        )
    print(f"    Inserted {len(rows)} rows")
    
    maria_conn.commit()
    sqlite_conn.close()
    maria_conn.close()
    print("  Done!\n")


def migrate_wta1():
    """Migrate wta_1 database"""
    print("=== Migrating wta_1 ===")
    
    sqlite_path = SPIDER_DATA / 'wta_1' / 'wta_1.sqlite'
    sqlite_conn = sqlite3.connect(str(sqlite_path))
    sqlite_conn.text_factory = lambda x: x.decode('utf-8', errors='replace')
    sqlite_conn.row_factory = sqlite3.Row
    
    maria_conn = pymysql.connect(**MARIA_CONFIG)
    cursor = maria_conn.cursor()
    
    # Create database and tables
    with open('/root/workspace/atlas/deploy/scripts/import_spider_wta.sql', 'r') as f:
        sql_script = f.read()
    
    for statement in sql_script.split(';'):
        statement = statement.strip()
        if statement:
            try:
                cursor.execute(statement)
            except Exception as e:
                print(f"  Warning: {e}")
    
    maria_conn.commit()
    cursor.execute("USE spider_wta")
    
    # Migrate players
    print("  Migrating players...")
    rows = sqlite_conn.execute("SELECT * FROM players").fetchall()
    batch_size = 1000
    for i in range(0, len(rows), batch_size):
        batch = rows[i:i+batch_size]
        values = []
        for row in batch:
            birth_date = row['birth_date']
            if birth_date == '' or birth_date is None:
                birth_date = None
            values.append((row['player_id'], row['first_name'], row['last_name'], 
                          row['hand'], birth_date, row['country_code']))
        cursor.executemany(
            "INSERT INTO players (player_id, first_name, last_name, hand, birth_date, country_code) VALUES (%s, %s, %s, %s, %s, %s)",
            values
        )
        maria_conn.commit()
    print(f"    Inserted {len(rows)} rows")
    
    # Migrate matches
    print("  Migrating matches...")
    rows = sqlite_conn.execute("SELECT * FROM matches").fetchall()
    for row in rows:
        cursor.execute("""
            INSERT INTO matches (best_of, draw_size, loser_age, loser_entry, loser_hand, loser_ht,
                loser_id, loser_ioc, loser_name, loser_rank, loser_rank_points, loser_seed,
                match_num, minutes, round, score, surface, tourney_date, tourney_id, tourney_level,
                tourney_name, winner_age, winner_entry, winner_hand, winner_ht, winner_id, winner_ioc,
                winner_name, winner_rank, winner_rank_points, winner_seed, year)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
        """, (row['best_of'], row['draw_size'], row['loser_age'], row['loser_entry'], row['loser_hand'],
              row['loser_ht'], row['loser_id'], row['loser_ioc'], row['loser_name'], row['loser_rank'],
              row['loser_rank_points'], row['loser_seed'], row['match_num'], row['minutes'], row['round'],
              row['score'], row['surface'], row['tourney_date'], row['tourney_id'], row['tourney_level'],
              row['tourney_name'], row['winner_age'], row['winner_entry'], row['winner_hand'],
              row['winner_ht'], row['winner_id'], row['winner_ioc'], row['winner_name'], row['winner_rank'],
              row['winner_rank_points'], row['winner_seed'], row['year']))
    print(f"    Inserted {len(rows)} rows")
    maria_conn.commit()
    
    # Migrate rankings (large table, batch insert)
    print("  Migrating rankings (large table)...")
    cursor_sqlite = sqlite_conn.cursor()
    cursor_sqlite.execute("SELECT * FROM rankings")
    
    def clean_value(v):
        """Convert empty strings to None"""
        if v == '' or v is None:
            return None
        return v
    
    batch_size = 10000
    total = 0
    while True:
        rows = cursor_sqlite.fetchmany(batch_size)
        if not rows:
            break
        values = [(clean_value(row[0]), clean_value(row[1]), clean_value(row[2]), 
                   clean_value(row[3]), clean_value(row[4])) for row in rows]
        cursor.executemany(
            "INSERT INTO rankings (ranking_date, ranking, player_id, ranking_points, tours) VALUES (%s, %s, %s, %s, %s)",
            values
        )
        maria_conn.commit()
        total += len(rows)
        print(f"    Inserted {total} rows so far...")
    
    print(f"    Total inserted: {total} rows")
    
    sqlite_conn.close()
    maria_conn.close()
    print("  Done!\n")


def verify():
    """Verify import results"""
    print("=== Verifying import results ===")
    
    maria_conn = pymysql.connect(**MARIA_CONFIG)
    cursor = maria_conn.cursor()
    
    # Verify flight_2
    cursor.execute("USE spider_flight")
    cursor.execute("SELECT COUNT(*) FROM airlines")
    print(f"  spider_flight.airlines: {cursor.fetchone()[0]} rows")
    cursor.execute("SELECT COUNT(*) FROM airports")
    print(f"  spider_flight.airports: {cursor.fetchone()[0]} rows")
    cursor.execute("SELECT COUNT(*) FROM flights")
    print(f"  spider_flight.flights: {cursor.fetchone()[0]} rows")
    
    # Verify orphan records
    cursor.execute("""
        SELECT COUNT(*) FROM flights f 
        LEFT JOIN airports a ON f.SourceAirport = a.AirportCode 
        WHERE a.AirportCode IS NULL
    """)
    orphan_source = cursor.fetchone()[0]
    cursor.execute("""
        SELECT COUNT(*) FROM flights f 
        LEFT JOIN airports a ON f.DestAirport = a.AirportCode 
        WHERE a.AirportCode IS NULL
    """)
    orphan_dest = cursor.fetchone()[0]
    print(f"  Orphan records: SourceAirport={orphan_source}, DestAirport={orphan_dest}")
    
    # Verify whitespace issues
    cursor.execute("SELECT COUNT(*) FROM airports WHERE CountryAbbrev != TRIM(CountryAbbrev)")
    whitespace = cursor.fetchone()[0]
    print(f"  Whitespace issues (CountryAbbrev): {whitespace} rows")
    
    print()
    
    # Verify wta_1
    cursor.execute("USE spider_wta")
    cursor.execute("SELECT COUNT(*) FROM players")
    print(f"  spider_wta.players: {cursor.fetchone()[0]} rows")
    cursor.execute("SELECT COUNT(*) FROM matches")
    print(f"  spider_wta.matches: {cursor.fetchone()[0]} rows")
    cursor.execute("SELECT COUNT(*) FROM rankings")
    print(f"  spider_wta.rankings: {cursor.fetchone()[0]} rows")
    
    maria_conn.close()
    print("\nVerification complete!")


if __name__ == '__main__':
    if len(sys.argv) > 1:
        if sys.argv[1] == 'flight':
            migrate_flight2()
        elif sys.argv[1] == 'wta':
            migrate_wta1()
        elif sys.argv[1] == 'verify':
            verify()
        elif sys.argv[1] == 'all':
            migrate_flight2()
            migrate_wta1()
            verify()
    else:
        print("Usage: python migrate_spider.py [flight|wta|verify|all]")
