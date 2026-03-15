-- =============================================================
-- LUCID TPC-H Enterprise Demo Database
-- Demonstrates two-stage Adaptive Schema Linking (>30 tables triggers LargeScale)
-- TPC-H core 8 tables + 30 enterprise extension tables + 479 addon domain tables = 517 tables
-- =============================================================

CREATE DATABASE IF NOT EXISTS tpch_enterprise DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE tpch_enterprise;

-- =============================================================
-- PART I: TPC-H Core Tables (8 tables)
-- Standard TPC-H schema, SF=0.01 sample data
-- =============================================================

-- Region (5 rows, fixed)
CREATE TABLE region (
    r_regionkey  INTEGER NOT NULL PRIMARY KEY,
    r_name       CHAR(25) NOT NULL COMMENT 'Region name',
    r_comment    VARCHAR(152) COMMENT 'Remark'
) COMMENT='TPC-H Region table — five global regions';

INSERT INTO region VALUES
(0, 'AFRICA',      'Africa continent'),
(1, 'AMERICA',     'North and South America'),
(2, 'ASIA',        'Asia-Pacific region'),
(3, 'EUROPE',      'European countries'),
(4, 'MIDDLE EAST', 'Middle East region');

-- Nation (25 rows, fixed)
CREATE TABLE nation (
    n_nationkey  INTEGER NOT NULL PRIMARY KEY,
    n_name       CHAR(25) NOT NULL COMMENT 'Nation name',
    n_regionkey  INTEGER NOT NULL COMMENT 'Region key',
    n_comment    VARCHAR(152) COMMENT 'Remark',
    FOREIGN KEY (n_regionkey) REFERENCES region(r_regionkey)
) COMMENT='TPC-H Nation table — 25 countries';

INSERT INTO nation VALUES
(0,  'ALGERIA',        0, 'North African country'),
(1,  'ARGENTINA',      1, 'South American country'),
(2,  'BRAZIL',         1, 'Largest South American country'),
(3,  'CANADA',         1, 'North American country'),
(4,  'EGYPT',          4, 'Northeast African country'),
(5,  'ETHIOPIA',       0, 'East African country'),
(6,  'FRANCE',         3, 'Western European country'),
(7,  'GERMANY',        3, 'Central European country'),
(8,  'INDIA',          2, 'South Asian country'),
(9,  'INDONESIA',      2, 'Southeast Asian country'),
(10, 'IRAN',           4, 'Middle Eastern country'),
(11, 'IRAQ',           4, 'Middle Eastern country'),
(12, 'JAPAN',          2, 'East Asian country'),
(13, 'JORDAN',         4, 'Middle Eastern country'),
(14, 'KENYA',          0, 'East African country'),
(15, 'MOROCCO',        0, 'North African country'),
(16, 'MOZAMBIQUE',     0, 'Southeast African country'),
(17, 'PERU',           1, 'South American country'),
(18, 'CHINA',          2, 'East Asian country — largest population'),
(19, 'ROMANIA',        3, 'Eastern European country'),
(20, 'SAUDI ARABIA',   4, 'Middle Eastern country'),
(21, 'VIETNAM',        2, 'Southeast Asian country'),
(22, 'RUSSIA',         3, 'Largest country by area'),
(23, 'UNITED KINGDOM', 3, 'Western European country'),
(24, 'UNITED STATES',  1, 'North American country');

-- Supplier (100 rows @ SF=0.01)
CREATE TABLE supplier (
    s_suppkey    INTEGER NOT NULL PRIMARY KEY,
    s_name       CHAR(25) NOT NULL COMMENT 'Supplier name',
    s_address    VARCHAR(40) NOT NULL COMMENT 'Supplier address',
    s_nationkey  INTEGER NOT NULL COMMENT 'Nation key',
    s_phone      CHAR(15) NOT NULL COMMENT 'Phone number',
    s_acctbal    DECIMAL(15,2) NOT NULL COMMENT 'Account balance',
    s_comment    VARCHAR(101) COMMENT 'Remark',
    FOREIGN KEY (s_nationkey) REFERENCES nation(n_nationkey)
) COMMENT='TPC-H Supplier table — parts supplier information';

INSERT INTO supplier VALUES
(1,  'Supplier#000000001', 'N kD4on9OM Ipw3,gf0JBoQDd7tgrzrddZ',  17, '27-918-335-1736',  5755.94, 'each slyly above the careful'),
(2,  'Supplier#000000002', '89eJ5ksX3ImxJQBvxObC,',                5,  '15-679-861-2259',  4032.68, 'furiously stealthy frays thrash'),
(3,  'Supplier#000000003', 'q1,G3Pj6OjIuUYfUoH18BFTKP5aU9bEV3',   1,  '11-383-516-1199',  4192.40, 'blithely silent requests after'),
(4,  'Supplier#000000004', 'Bk7ah4CK8SYQTepEmvMkkgMwg',            15, '25-843-787-7479',  4641.08, 'furiously final courts wake'),
(5,  'Supplier#000000005', 'Gcdm2rJRzl5qlTVzc',                    11, '21-151-690-3663', -531.44,  'quiet, ironic deposits kindle'),
(6,  'Supplier#000000006', 'tQxuVm7s7CnK',                         14, '24-696-997-4969',  1365.79, 'final accounts. regular'),
(7,  'Supplier#000000007', 's,4TicNGB4uO6PaSqNBUq',                23, '33-990-965-2201',  6820.35, 'furiously regular instructions impress'),
(8,  'Supplier#000000008', '9Sq4bBH2FQEmaFOocY45sRTxo6yuoG',       17, '27-498-742-3860',  7627.85, 'final, pending deposits are'),
(9,  'Supplier#000000009', '1KhUgZegwM3ua7dsYmekYBsGx',            10, '20-403-398-8662',  5765.10, 'bold foxes sleep final'),
(10, 'Supplier#000000010', 'Saez8ksT',                              24, '34-924-489-1940',  3956.64, 'carefully ironic requests shall');

-- Part (200 rows @ SF=0.01, showing 20)
CREATE TABLE part (
    p_partkey     INTEGER NOT NULL PRIMARY KEY,
    p_name        VARCHAR(55) NOT NULL COMMENT 'Part name',
    p_mfgr        CHAR(25) NOT NULL COMMENT 'Manufacturer',
    p_brand       CHAR(10) NOT NULL COMMENT 'Brand',
    p_type        VARCHAR(25) NOT NULL COMMENT 'Part type',
    p_size        INTEGER NOT NULL COMMENT 'Size',
    p_container   CHAR(10) NOT NULL COMMENT 'Container type',
    p_retailprice DECIMAL(15,2) NOT NULL COMMENT 'Retail price',
    p_comment     VARCHAR(23) COMMENT 'Remark'
) COMMENT='TPC-H Part table — supply chain parts catalog';

INSERT INTO part VALUES
(1,  'goldenrod lavender spring chocolate lace', 'Manufacturer#1', 'Brand#13', 'PROMO BURNISHED COPPER', 7,  'JUMBO PKG',  901.00, 'ly. slyly ironi'),
(2,  'blush thistle blue yellow saddle',         'Manufacturer#1', 'Brand#13', 'LARGE BRUSHED BRASS',    1,  'LG CASE',    902.00, 'lar accounts amo'),
(3,  'spring green yellow purple cornsilk',      'Manufacturer#4', 'Brand#42', 'STANDARD POLISHED BRASS',21, 'WRAP CASE',  903.00, 'egular deposits'),
(4,  'cornflower chocolate smoke green pink',    'Manufacturer#3', 'Brand#34', 'SMALL PLATED BRASS',     14, 'MED DRUM',   904.00, 'p]ironic foxes'),
(5,  'forest brown coral puff cream',            'Manufacturer#3', 'Brand#32', 'STANDARD POLISHED TIN',  15, 'SM PKG',     905.00, 'wake carefully'),
(6,  'bisque cornflower lawn forest magenta',    'Manufacturer#2', 'Brand#24', 'PROMO PLATED STEEL',     4,  'MED BAG',    906.00, 'sly regular acc'),
(7,  'moccasin green thistle khaki floral',      'Manufacturer#1', 'Brand#11', 'SMALL PLATED COPPER',    45, 'SM BAG',     907.00, 'lyly regular re'),
(8,  'misty lace thistle blanched saddle',       'Manufacturer#4', 'Brand#44', 'PROMO BURNISHED TIN',    41, 'LG DRUM',    908.00, 'furiously final'),
(9,  'thistle dim navajo dark gainsboro',        'Manufacturer#4', 'Brand#43', 'SMALL BURNISHED STEEL',  12, 'WRAP CASE',  909.00, 'ironic deposits'),
(10, 'linen pink saddle puff powder',            'Manufacturer#5', 'Brand#54', 'LARGE BURNISHED STEEL',  44, 'LG CAN',     910.00, 'ithely final de'),
(11, 'spring maroon seashell almond orchid',     'Manufacturer#2', 'Brand#25', 'STANDARD BURNISHED NICKEL',43,'WRAP BOX',  911.00, 'ng the silently'),
(12, 'rose deep ivory midnight navy',            'Manufacturer#3', 'Brand#33', 'MEDIUM ANODIZED STEEL',  25, 'JUMBO CASE', 912.00, 'special pinto b'),
(13, 'khaki cream sandy dodger mint',            'Manufacturer#5', 'Brand#55', 'MEDIUM BURNISHED NICKEL', 1, 'JUMBO PKG',  913.00, 'uickly special'),
(14, 'slate grey violet white midnight',         'Manufacturer#1', 'Brand#11', 'SMALL POLISHED TIN',     21, 'SM BOX',     914.00, 'e carefully reg'),
(15, 'honeydew lemon chiffon sky indian',        'Manufacturer#1', 'Brand#15', 'LARGE ANODIZED BRASS',    2, 'LG CASE',    915.00, 'ets. bravely re'),
(16, 'olive magenta chocolate firebrick orchid', 'Manufacturer#3', 'Brand#32', 'PROMO PLATED TIN',       12, 'MED PACK',   916.00, 'lithely regular'),
(17, 'white smoke salmon orchid wheat',          'Manufacturer#1', 'Brand#14', 'ECONOMY BURNISHED COPPER',8, 'MED BOX',    917.00, 'against the acc'),
(18, 'brown coral indian medium spring',         'Manufacturer#2', 'Brand#21', 'PROMO BURNISHED COPPER', 29, 'SM PKG',     918.00, 'above the quick'),
(19, 'rosy plum orange chocolate lime',          'Manufacturer#3', 'Brand#33', 'SMALL ANODIZED NICKEL',  20, 'MED PKG',    919.00, 'unusual package'),
(20, 'tan olive lavender cyan wheat',            'Manufacturer#1', 'Brand#12', 'LARGE BRUSHED NICKEL',   34, 'SM BAG',     920.00, 'nding pinto bea');

-- PartSupp (composite PK)
CREATE TABLE partsupp (
    ps_partkey    INTEGER NOT NULL COMMENT 'Part key',
    ps_suppkey    INTEGER NOT NULL COMMENT 'Supplier key',
    ps_availqty   INTEGER NOT NULL COMMENT 'Available quantity',
    ps_supplycost DECIMAL(15,2) NOT NULL COMMENT 'Supply cost',
    ps_comment    VARCHAR(199) COMMENT 'Remark',
    PRIMARY KEY (ps_partkey, ps_suppkey),
    FOREIGN KEY (ps_partkey) REFERENCES part(p_partkey),
    FOREIGN KEY (ps_suppkey) REFERENCES supplier(s_suppkey)
) COMMENT='TPC-H PartSupp table — part-supplier many-to-many relationship with costs';

INSERT INTO partsupp VALUES
(1, 2, 3325, 771.64, 'requests after the carefully ironic ideas'),
(1, 4, 8076, 993.49, 'furiously regular instructions about the carefully'),
(2, 3, 8895, 378.49, 'nal foxes wake. quickly special'),
(2, 5, 4969, 915.27, 'ar deposits cajole slyly'),
(3, 6, 4651, 920.92, 'eans boost fluffily ironic ideas'),
(3, 8, 3877, 438.86, 'nts about the slyly bold deposits'),
(4, 7, 2694, 823.16, 'ly enticing accounts'),
(4, 9, 3478, 877.03, 'busily unusual instructions'),
(5, 1, 4018, 299.58, 'gifts cajole ironic deposits'),
(5, 3, 2694, 516.19, 'pinto beans are furiously'),
(6, 2, 5765, 662.99, 'special accounts sleep among the'),
(6, 4, 3000, 116.84, 'closely even pinto beans haggle'),
(7, 5, 4102, 348.16, 'packages nag blithely'),
(7, 7, 2344, 999.83, 'pending instructions across'),
(8, 6, 5765, 227.14, 'carefully unusual accounts'),
(8, 8, 7454, 364.84, 'slyly final foxes nag'),
(9, 9, 1004, 569.53, 'furiously regular deposits haggle'),
(9, 1, 2102, 418.83, 'pending pinto beans impress'),
(10,10, 8831, 194.52, 'quickly ironic deposits integrate'),
(10, 2, 5352, 878.88, 'even requests cajole furiously');

-- Customer (150 rows @ SF=0.01, showing 20)
CREATE TABLE customer (
    c_custkey    INTEGER NOT NULL PRIMARY KEY,
    c_name       VARCHAR(25) NOT NULL COMMENT 'Customer name',
    c_address    VARCHAR(40) NOT NULL COMMENT 'Customer address',
    c_nationkey  INTEGER NOT NULL COMMENT 'Nation key',
    c_phone      CHAR(15) NOT NULL COMMENT 'Phone number',
    c_acctbal    DECIMAL(15,2) NOT NULL COMMENT 'Account balance',
    c_mktsegment CHAR(10) NOT NULL COMMENT 'Market segment (AUTOMOBILE/BUILDING/FURNITURE/HOUSEHOLD/MACHINERY)',
    c_comment    VARCHAR(117) COMMENT 'Remark',
    FOREIGN KEY (c_nationkey) REFERENCES nation(n_nationkey)
) COMMENT='TPC-H Customer table — ordering customer information';

INSERT INTO customer VALUES
(1,  'Customer#000000001', 'IVhzIApeRb ot,c,E',            15, '25-989-741-2988', 711.56,   'BUILDING',   'to the even, regular platelets'),
(2,  'Customer#000000002', 'XSTf4,NCwDVaWNe6tEgvwfmRchLXak', 13, '23-768-687-3665', 121.65,  'AUTOMOBILE', 'l accounts. blithely ironic'),
(3,  'Customer#000000003', 'MG9kdTD2WBHm',                  1,  '11-719-748-3364', 7498.12,  'AUTOMOBILE', 'deposits eat slyly ironic'),
(4,  'Customer#000000004', 'XxVSJsLAGtn',                   4,  '14-128-190-5944', 2866.83,  'MACHINERY',  'requests. final, regular'),
(5,  'Customer#000000005', 'hwBtxkoBF qSW4KrIk5U 2B1AU7H',  3,  '13-750-942-6364', 794.47,  'HOUSEHOLD',  'n accounts was. unusual,'),
(6,  'Customer#000000006', 'sKZz0CsnMD7mp4Xd0YrBvx,LREYKUWAh yVn', 20, '30-114-968-4951', 7638.57, 'AUTOMOBILE', 'tions. even deposits boost'),
(7,  'Customer#000000007', 'TcGe5gaZNgVePxU5kRrvXBfkasDTea',  18, '28-190-982-9759', 9561.95, 'AUTOMOBILE', 'ainst the ironic, express'),
(8,  'Customer#000000008', 'I0B10bB0AymmC, 0PrRYBCP1yGJ8xcBPmWhl5', 17, '27-147-574-9335', 6819.74, 'BUILDING',  'among the slyly regular'),
(9,  'Customer#000000009', 'xKiAFTjUsCuxfeleNqefumTrjS',     8,  '18-338-906-3675', 8324.07, 'FURNITURE',  'r theodolites according to'),
(10, 'Customer#000000010', '6LrEaV6KR6PLVcgl2ArL Q3rqzLzcT1 v2', 5, '15-741-346-9870', 2753.54, 'HOUSEHOLD', 'es regular deposits haggle'),
(11, 'Customer#000000011', 'PkWS 3HlXqwTuzrKg633BEi',        23, '33-464-151-3439', -272.60, 'BUILDING',   'ckages. requests sleep slyly'),
(12, 'Customer#000000012', '9PWKSkkqEHkj',                    13, '23-791-276-1263', 3396.49, 'HOUSEHOLD',  'to the carefully final braids'),
(13, 'Customer#000000013', 'nsXQu0oVjD7PM659uC3SRSp',         3, '13-761-547-5974', 3857.34, 'BUILDING',   'ounts sleep carefully after'),
(14, 'Customer#000000014', 'KXkletMlL2JQEA ',                  1, '11-845-129-3851', 5266.30, 'FURNITURE',  'are fluffily. requests'),
(15, 'Customer#000000015', 'YtWggXoOLdwdo7b0y,BZaGUQMLJMX1Y', 23, '33-687-542-7601', 2788.52, 'HOUSEHOLD', 'platelets. regular deposits'),
(16, 'Customer#000000016', 'cYiaeMLZSMAOQ2 d0W,',              10,'20-781-609-3107', 4681.03, 'FURNITURE',  'kly silent courts. thinly'),
(17, 'Customer#000000017', 'izrh 6jdqtp2eqdtbkswDD8SG4SzXruMfIXyR7', 2, '12-970-682-3487', 6.34, 'AUTOMOBILE', 'packages wake! blithely'),
(18, 'Customer#000000018', 'ZJQS x1HPjbUr5T68dci3Ak',          6, '16-155-215-1315', 5494.43, 'BUILDING',  'special foxes affix'),
(19, 'Customer#000000019', 'uc,3bHIx84H,wdrmLOjVsiqXCq2tr',    18,'28-396-526-5053', 8914.71, 'HOUSEHOLD', ' nag. furiously careful packages'),
(20, 'Customer#000000020', 'JrPk8Pqplj4Ne',                    22,'32-957-234-8742', 7603.40, 'FURNITURE',  'g alongside of the special excuses');

-- Orders (showing 30 rows)
CREATE TABLE orders (
    o_orderkey      INTEGER NOT NULL PRIMARY KEY,
    o_custkey       INTEGER NOT NULL COMMENT 'Customer key',
    o_orderstatus   CHAR(1) NOT NULL COMMENT 'Order status (O=open, F=filled, P=partial)',
    o_totalprice    DECIMAL(15,2) NOT NULL COMMENT 'Total order price',
    o_orderdate     DATE NOT NULL COMMENT 'Order date',
    o_orderpriority CHAR(15) NOT NULL COMMENT 'Order priority (1-URGENT, 2-HIGH, 3-MEDIUM, 4-NOT SPECIFIED, 5-LOW)',
    o_clerk         CHAR(15) NOT NULL COMMENT 'Clerk ID',
    o_shippriority  INTEGER NOT NULL COMMENT 'Shipping priority',
    o_comment       VARCHAR(79) COMMENT 'Remark',
    FOREIGN KEY (o_custkey) REFERENCES customer(c_custkey)
) COMMENT='TPC-H Orders table — customer purchase orders';

INSERT INTO orders VALUES
(1,   4,  'O', 172799.49, '1996-01-02', '5-LOW',          'Clerk#000000951', 0, 'nstructions sleep furiously among'),
(2,   8,  'O',  46929.18, '1996-12-01', '1-URGENT',       'Clerk#000000880', 0, 'foxes. pending accounts at the'),
(3,   12, 'F', 193846.25, '1993-10-14', '5-LOW',          'Clerk#000000955', 0, 'sly final accounts boost'),
(4,   14, 'O',  32151.78, '1995-10-11', '5-LOW',          'Clerk#000000124', 0, 'sits. slyly regular warthogs'),
(5,   5,  'F',  77471.44, '1994-07-30', '5-LOW',          'Clerk#000000925', 0, 'quickly. bold deposits sleep'),
(6,   7,  'F',  36468.55, '1992-02-21', '4-NOT SPECIFIED', 'Clerk#000000058', 0, 'ggle. special, final requests'),
(7,   4,  'O', 252004.18, '1996-01-10', '2-HIGH',         'Clerk#000000470', 0, 'ly special requests'),
(32,  13, 'O', 208660.75, '1995-07-16', '2-HIGH',         'Clerk#000000616', 0, 'ise blithely bold, regular'),
(33,  7,  'F',  74150.58, '1993-10-27', '3-MEDIUM',       'Clerk#000000409', 0, 'uriously. furiously final request'),
(34,  8,  'O',  58949.67, '1998-07-21', '3-MEDIUM',       'Clerk#000000223', 0, 'ly final packages. fluffily'),
(35,  13, 'O',  73426.50, '1995-10-23', '4-NOT SPECIFIED', 'Clerk#000000259', 0, 'zzle. carefully enticing deposits'),
(36,  12, 'O', 173665.47, '1995-11-03', '1-URGENT',       'Clerk#000000358', 0, 'quick packages are blithely'),
(37,  20, 'F',  12494.03, '1992-06-03', '3-MEDIUM',       'Clerk#000000456', 0, 'kly regular pinto beans'),
(38,  13, 'O',  46366.56, '1996-08-21', '4-NOT SPECIFIED', 'Clerk#000000604', 0, 'haggle blithely. furiously express'),
(39,  1,  'O', 219707.84, '1996-09-20', '3-MEDIUM',       'Clerk#000000659', 0, 'havens boost slyly among'),
(64,  4,  'F',  20613.67, '1994-07-16', '3-MEDIUM',       'Clerk#000000661', 0, 'wake blithely. quickly bold'),
(65,  2,  'P', 110643.60, '1995-03-18', '1-URGENT',       'Clerk#000000632', 0, 'ular requests are blithely'),
(66,  13, 'F',  79258.24, '1994-01-20', '5-LOW',          'Clerk#000000743', 0, 'y alongside of the pending'),
(67,  7,  'O', 116227.05, '1996-12-19', '4-NOT SPECIFIED', 'Clerk#000000547', 0, 'ithely ironic deposits haggle'),
(68,  3,  'O', 186543.02, '1998-04-18', '3-MEDIUM',       'Clerk#000000440', 0, 'pinto beans sleep carefully');

-- Lineitem (showing 40 rows)
CREATE TABLE lineitem (
    l_orderkey     INTEGER NOT NULL COMMENT 'Order key',
    l_partkey      INTEGER NOT NULL COMMENT 'Part key',
    l_suppkey      INTEGER NOT NULL COMMENT 'Supplier key',
    l_linenumber   INTEGER NOT NULL COMMENT 'Line item number',
    l_quantity     DECIMAL(15,2) NOT NULL COMMENT 'Quantity',
    l_extendedprice DECIMAL(15,2) NOT NULL COMMENT 'Extended price = quantity x unit price',
    l_discount     DECIMAL(15,2) NOT NULL COMMENT 'Discount rate (0.00-0.10)',
    l_tax          DECIMAL(15,2) NOT NULL COMMENT 'Tax rate',
    l_returnflag   CHAR(1) NOT NULL COMMENT 'Return flag (A=accepted, R=returned, N=none)',
    l_linestatus   CHAR(1) NOT NULL COMMENT 'Line status (O=open, F=filled)',
    l_shipdate     DATE NOT NULL COMMENT 'Ship date',
    l_commitdate   DATE NOT NULL COMMENT 'Commit date',
    l_receiptdate  DATE NOT NULL COMMENT 'Receipt date',
    l_shipinstruct CHAR(25) NOT NULL COMMENT 'Shipping instructions',
    l_shipmode     CHAR(10) NOT NULL COMMENT 'Ship mode (REG AIR/AIR/RAIL/SHIP/TRUCK/MAIL/FOB)',
    l_comment      VARCHAR(44) COMMENT 'Remark',
    PRIMARY KEY (l_orderkey, l_linenumber),
    FOREIGN KEY (l_orderkey) REFERENCES orders(o_orderkey)
) COMMENT='TPC-H Lineitem table — order line items with price, quantity, discount, and shipping info';

INSERT INTO lineitem VALUES
(1, 2, 3, 1, 17.00, 21168.23, 0.04, 0.02, 'N', 'O', '1996-03-13', '1996-02-12', '1996-03-22', 'DELIVER IN PERSON', 'TRUCK',   'egular courts above the'),
(1, 3, 6, 2, 36.00, 34850.16, 0.09, 0.06, 'N', 'O', '1996-04-12', '1996-02-28', '1996-04-20', 'TAKE BACK RETURN',  'MAIL',    'ly final dependencies'),
(1, 6, 4, 3, 8.00,  13309.60, 0.10, 0.02, 'N', 'O', '1996-01-29', '1996-03-05', '1996-01-31', 'TAKE BACK RETURN',  'REG AIR', 'riously. regular, express dep'),
(1, 7, 5, 4, 28.00, 25004.40, 0.09, 0.06, 'N', 'O', '1996-04-21', '1996-03-30', '1996-05-16', 'NONE',              'AIR',     'lites. fluffily even de'),
(1, 5, 1, 5, 24.00, 22824.48, 0.10, 0.04, 'N', 'O', '1996-03-30', '1996-03-14', '1996-04-01', 'NONE',              'FOB',     'pending foxes. slyly re'),
(1, 2, 4, 6, 32.00, 28955.64, 0.07, 0.02, 'N', 'O', '1996-01-30', '1996-02-07', '1996-02-03', 'DELIVER IN PERSON', 'MAIL',    'arefully slyly ex'),
(2, 6, 2, 1, 38.00, 44694.46, 0.00, 0.05, 'N', 'O', '1997-01-28', '1997-01-14', '1997-02-02', 'TAKE BACK RETURN',  'RAIL',    'ven requests. deposits breach'),
(3, 1, 4, 1, 45.00, 54058.05, 0.06, 0.00, 'R', 'F', '1994-02-02', '1994-01-04', '1994-02-23', 'NONE',              'AIR',     'ongside of the furiously brave'),
(3, 8, 8, 2, 49.00, 46796.47, 0.10, 0.00, 'R', 'F', '1993-11-09', '1993-12-20', '1993-11-24', 'TAKE BACK RETURN',  'RAIL',    'unusual accounts. eve'),
(3, 13,6, 3, 27.00, 26159.97, 0.06, 0.07, 'A', 'F', '1994-01-16', '1993-11-22', '1994-01-23', 'DELIVER IN PERSON', 'SHIP',    'nal foxes wake.'),
(3, 4, 7, 4, 2.00,   1696.00, 0.01, 0.06, 'A', 'F', '1993-12-04', '1994-01-07', '1994-01-01', 'NONE',              'TRUCK',   'y. fluffily pending dep'),
(4, 9, 1, 1, 30.00, 26670.00, 0.03, 0.08, 'N', 'O', '1996-01-10', '1995-12-14', '1996-01-18', 'DELIVER IN PERSON', 'REG AIR', 'ironic accounts'),
(5, 11,8, 1, 15.00, 15045.00, 0.02, 0.04, 'R', 'F', '1994-10-31', '1994-08-31', '1994-11-20', 'NONE',              'AIR',     'old foxes. bold, special'),
(5, 14,5, 2, 26.00, 26324.00, 0.07, 0.08, 'R', 'F', '1994-10-16', '1994-09-25', '1994-10-19', 'NONE',              'FOB',     'are among the slyly expres'),
(5, 3, 6, 3, 50.00, 47850.00, 0.08, 0.03, 'A', 'F', '1994-08-08', '1994-10-13', '1994-08-26', 'DELIVER IN PERSON', 'AIR',     'eodolites. fluffily unusual'),
(6, 10,2, 1, 37.00, 33673.70, 0.08, 0.03, 'A', 'F', '1992-04-27', '1992-05-15', '1992-05-02', 'TAKE BACK RETURN',  'TRUCK',   'p furiously special foxes'),
(7, 16,5, 1, 12.00, 10880.40, 0.07, 0.03, 'N', 'O', '1996-05-07', '1996-03-13', '1996-06-03', 'TAKE BACK RETURN',  'FOB',     'ss pinto beans wake against th'),
(7, 18,8, 2, 9.00,   8275.71, 0.08, 0.08, 'N', 'O', '1996-02-01', '1996-03-02', '1996-02-19', 'TAKE BACK RETURN',  'SHIP',    'es. ruthlessly unusual'),
(7, 8, 6, 3, 46.00, 43529.50, 0.10, 0.07, 'N', 'O', '1996-01-15', '1996-03-27', '1996-02-03', 'COLLECT COD',       'MAIL',    'nt braids breach'),
(7, 1, 2, 4, 28.00, 25452.24, 0.03, 0.04, 'N', 'O', '1996-03-21', '1996-04-08', '1996-04-20', 'NONE',              'FOB',     'the slyly bold deposits');


-- =============================================================
-- PART II: Enterprise Extension Tables (30 tables)
-- Simulates real enterprise scenarios with distractor tables to demonstrate ranking value
-- =============================================================

-- --- Module: HR (5 tables) --- Pure distractor module ---

CREATE TABLE departments (
    dept_id      INT PRIMARY KEY AUTO_INCREMENT,
    dept_name    VARCHAR(100) NOT NULL COMMENT 'Department name',
    location     VARCHAR(100) COMMENT 'Office location',
    budget       DECIMAL(15,2) COMMENT 'Annual department budget',
    manager_name VARCHAR(100) COMMENT 'Department manager'
) COMMENT='HR — Department information';

INSERT INTO departments (dept_name, location, budget, manager_name) VALUES
('Sales',        'Shanghai', 5000000.00, 'Zhang Wei'),
('Engineering',  'Beijing',  8000000.00, 'Li Ming'),
('Marketing',    'Guangzhou', 3000000.00, 'Wang Fang'),
('Finance',      'Shanghai', 2000000.00, 'Chen Jie'),
('Logistics',    'Shenzhen', 4500000.00, 'Liu Yang'),
('Human Resources', 'Beijing', 1500000.00, 'Zhao Min'),
('Customer Service', 'Chengdu', 2500000.00, 'Sun Li'),
('R&D',          'Hangzhou',  9000000.00, 'Wu Hao');

CREATE TABLE employees (
    emp_id       INT PRIMARY KEY AUTO_INCREMENT,
    emp_name     VARCHAR(100) NOT NULL COMMENT 'Employee name',
    dept_id      INT COMMENT 'Department ID',
    title        VARCHAR(100) COMMENT 'Job title',
    hire_date    DATE COMMENT 'Hire date',
    salary       DECIMAL(12,2) COMMENT 'Monthly salary',
    email        VARCHAR(255) COMMENT 'Work email',
    phone        VARCHAR(20) COMMENT 'Phone number',
    status       VARCHAR(20) DEFAULT 'active' COMMENT 'Employment status (active/inactive/leave)',
    FOREIGN KEY (dept_id) REFERENCES departments(dept_id)
) COMMENT='HR — Employee information';

INSERT INTO employees (emp_name, dept_id, title, hire_date, salary, email, phone, status) VALUES
('Zhang Wei',   1, 'Sales Director',       '2018-03-15', 35000.00, 'zhangwei@corp.com',  '13800001001', 'active'),
('Li Ming',     2, 'VP Engineering',        '2016-06-01', 55000.00, 'liming@corp.com',    '13800001002', 'active'),
('Wang Fang',   3, 'Marketing Manager',     '2019-01-10', 28000.00, 'wangfang@corp.com',  '13800001003', 'active'),
('Chen Jie',    4, 'Finance Director',      '2017-09-01', 38000.00, 'chenjie@corp.com',   '13800001004', 'active'),
('Liu Yang',    5, 'Logistics Manager',     '2020-04-15', 25000.00, 'liuyang@corp.com',   '13800001005', 'active'),
('Zhao Min',    6, 'HR Director',           '2017-11-20', 32000.00, 'zhaomin@corp.com',   '13800001006', 'active'),
('Sun Li',      7, 'CS Team Lead',          '2021-02-01', 22000.00, 'sunli@corp.com',     '13800001007', 'active'),
('Wu Hao',      8, 'CTO',                   '2015-01-15', 65000.00, 'wuhao@corp.com',     '13800001008', 'active'),
('Huang Ying',  1, 'Sales Representative',  '2022-07-01', 15000.00, 'huangying@corp.com', '13800001009', 'active'),
('Zhou Peng',   2, 'Senior Engineer',       '2019-08-15', 42000.00, 'zhoupeng@corp.com',  '13800001010', 'active'),
('Xu Dan',      2, 'Software Engineer',     '2023-01-10', 25000.00, 'xudan@corp.com',     '13800001011', 'active'),
('Ma Yun',      3, 'Content Specialist',    '2022-03-20', 18000.00, 'mayun@corp.com',     '13800001012', 'active'),
('Gao Shan',    5, 'Warehouse Supervisor',  '2020-11-01', 20000.00, 'gaoshan@corp.com',   '13800001013', 'active'),
('Lin Mei',     4, 'Accountant',            '2021-06-15', 20000.00, 'linmei@corp.com',    '13800001014', 'active'),
('Tang Hao',    7, 'Support Agent',         '2023-04-01', 12000.00, 'tanghao@corp.com',   '13800001015', 'active');

CREATE TABLE salaries (
    salary_id  INT PRIMARY KEY AUTO_INCREMENT,
    emp_id     INT NOT NULL COMMENT 'Employee ID',
    base_pay   DECIMAL(12,2) COMMENT 'Base salary',
    bonus      DECIMAL(12,2) COMMENT 'Bonus',
    effective_date DATE COMMENT 'Effective date',
    FOREIGN KEY (emp_id) REFERENCES employees(emp_id)
) COMMENT='HR — Salary records';

INSERT INTO salaries (emp_id, base_pay, bonus, effective_date) VALUES
(1, 35000.00, 15000.00, '2024-01-01'),
(2, 55000.00, 30000.00, '2024-01-01'),
(3, 28000.00, 10000.00, '2024-01-01'),
(4, 38000.00, 12000.00, '2024-01-01'),
(5, 25000.00, 8000.00,  '2024-01-01'),
(8, 65000.00, 50000.00, '2024-01-01');

CREATE TABLE job_history (
    history_id   INT PRIMARY KEY AUTO_INCREMENT,
    emp_id       INT NOT NULL COMMENT 'Employee ID',
    old_title    VARCHAR(100) COMMENT 'Previous title',
    new_title    VARCHAR(100) COMMENT 'New title',
    change_date  DATE COMMENT 'Change date',
    reason       VARCHAR(200) COMMENT 'Reason for change',
    FOREIGN KEY (emp_id) REFERENCES employees(emp_id)
) COMMENT='HR — Job title change history';

INSERT INTO job_history (emp_id, old_title, new_title, change_date, reason) VALUES
(1, 'Sales Manager',      'Sales Director',  '2022-06-01', 'Promotion'),
(2, 'Engineering Manager', 'VP Engineering',  '2021-01-01', 'Promotion'),
(8, 'VP Engineering',      'CTO',             '2020-01-01', 'Promotion');

CREATE TABLE performance_reviews (
    review_id   INT PRIMARY KEY AUTO_INCREMENT,
    emp_id      INT NOT NULL COMMENT 'Employee ID',
    review_date DATE COMMENT 'Review date',
    rating      INT COMMENT 'Rating (1-5)',
    reviewer    VARCHAR(100) COMMENT 'Reviewer',
    comments    TEXT COMMENT 'Comments',
    FOREIGN KEY (emp_id) REFERENCES employees(emp_id)
) COMMENT='HR — Employee performance review records';

INSERT INTO performance_reviews (emp_id, review_date, rating, reviewer, comments) VALUES
(1,  '2024-06-01', 4, 'Wu Hao',    'Strong sales leadership, exceeded Q1 targets'),
(2,  '2024-06-01', 5, 'Wu Hao',    'Exceptional technical vision and team management'),
(3,  '2024-06-01', 3, 'Wu Hao',    'Good campaign results, needs improvement in ROI tracking'),
(9,  '2024-06-01', 4, 'Zhang Wei', 'Promising new hire, quick learner'),
(10, '2024-06-01', 5, 'Li Ming',   'Outstanding code quality and mentorship');


-- --- Module: CRM (5 tables) --- Semi-related (customer_feedback links to customer) ---

CREATE TABLE contacts (
    contact_id   INT PRIMARY KEY AUTO_INCREMENT,
    contact_name VARCHAR(100) NOT NULL COMMENT 'Contact name',
    company      VARCHAR(200) COMMENT 'Company name',
    email        VARCHAR(255) COMMENT 'Email',
    phone        VARCHAR(20) COMMENT 'Phone',
    vip_flag     TINYINT(1) DEFAULT 0 COMMENT 'VIP flag',
    source       VARCHAR(50) COMMENT 'Lead source (web/referral/trade_show/cold_call)',
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
) COMMENT='CRM — Prospect contact information';

INSERT INTO contacts (contact_name, company, email, phone, vip_flag, source) VALUES
('Alice Johnson',  'TechCorp Inc.',     'alice@techcorp.com',     '+1-555-0101', 1, 'referral'),
('Bob Smith',      'Global Logistics',  'bob@globallog.com',      '+1-555-0102', 0, 'trade_show'),
('Catherine Lee',  'AutoParts Ltd.',     'clee@autoparts.com',     '+86-21-55550103', 1, 'web'),
('David Wang',     'SmartBuild Co.',     'dwang@smartbuild.com',   '+86-10-55550104', 0, 'cold_call'),
('Emma Brown',     'FurniMax',          'emma@furnimax.com',       '+1-555-0105', 1, 'referral'),
('Frank Miller',   'MachineWorks',      'frank@machineworks.com',  '+49-555-0106', 0, 'trade_show'),
('Grace Chen',     'Pacific Trading',   'grace@pacifictrade.com',  '+86-755-55550107', 1, 'web'),
('Henry Zhang',    'Euro Supplies',     'henry@eurosupplies.com',  '+44-555-0108', 0, 'referral');

CREATE TABLE leads (
    lead_id     INT PRIMARY KEY AUTO_INCREMENT,
    contact_id  INT COMMENT 'Contact ID',
    priority    VARCHAR(20) COMMENT 'Priority (hot/warm/cold)',
    deal_value  DECIMAL(15,2) COMMENT 'Estimated deal value',
    status      VARCHAR(30) COMMENT 'Status (new/contacted/qualified/proposal/won/lost)',
    assigned_to INT COMMENT 'Assigned sales employee',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (contact_id) REFERENCES contacts(contact_id),
    FOREIGN KEY (assigned_to) REFERENCES employees(emp_id)
) COMMENT='CRM — Sales lead tracking';

INSERT INTO leads (contact_id, priority, deal_value, status, assigned_to) VALUES
(1, 'hot',  500000.00, 'proposal',  1),
(2, 'warm', 200000.00, 'contacted', 9),
(3, 'hot',  800000.00, 'qualified', 1),
(4, 'cold', 50000.00,  'new',       9),
(5, 'warm', 350000.00, 'proposal',  1),
(6, 'cold', 120000.00, 'contacted', 9),
(7, 'hot',  1000000.00,'won',       1),
(8, 'warm', 180000.00, 'lost',      9);

CREATE TABLE opportunities (
    opp_id       INT PRIMARY KEY AUTO_INCREMENT,
    lead_id      INT COMMENT 'Source lead',
    opp_name     VARCHAR(200) COMMENT 'Opportunity name',
    stage        VARCHAR(30) COMMENT 'Stage (prospecting/analysis/proposal/negotiation/closed_won/closed_lost)',
    amount       DECIMAL(15,2) COMMENT 'Amount',
    close_date   DATE COMMENT 'Expected close date',
    probability  INT COMMENT 'Win probability (%)',
    FOREIGN KEY (lead_id) REFERENCES leads(lead_id)
) COMMENT='CRM — Opportunity management';

INSERT INTO opportunities (lead_id, opp_name, stage, amount, close_date, probability) VALUES
(1, 'TechCorp Server Upgrade',       'negotiation', 500000.00, '2025-03-15', 80),
(3, 'AutoParts Bulk Order',          'proposal',    800000.00, '2025-04-01', 60),
(5, 'FurniMax Annual Contract',      'negotiation', 350000.00, '2025-02-28', 75),
(7, 'Pacific Trading Partnership',   'closed_won', 1000000.00, '2025-01-15', 100);

CREATE TABLE campaigns (
    campaign_id   INT PRIMARY KEY AUTO_INCREMENT,
    campaign_name VARCHAR(200) NOT NULL COMMENT 'Campaign name',
    campaign_type VARCHAR(50) COMMENT 'Type (email/social/event/webinar)',
    start_date    DATE COMMENT 'Start date',
    end_date      DATE COMMENT 'End date',
    budget        DECIMAL(12,2) COMMENT 'Campaign budget',
    actual_cost   DECIMAL(12,2) COMMENT 'Actual cost',
    leads_generated INT DEFAULT 0 COMMENT 'Leads generated'
) COMMENT='CRM — Marketing campaigns';

INSERT INTO campaigns (campaign_name, campaign_type, start_date, end_date, budget, actual_cost, leads_generated) VALUES
('2024 Spring Promotion',     'email',   '2024-03-01', '2024-03-31', 50000.00, 42000.00, 120),
('Industry Trade Show Q2',    'event',   '2024-06-15', '2024-06-18', 200000.00, 185000.00, 85),
('Social Media Campaign',     'social',  '2024-04-01', '2024-06-30', 30000.00, 28500.00, 200),
('Product Launch Webinar',    'webinar', '2024-09-10', '2024-09-10', 15000.00, 12000.00, 65),
('Year-end Customer Event',   'event',   '2024-12-10', '2024-12-12', 150000.00, 145000.00, 50);

CREATE TABLE customer_feedback (
    feedback_id   INT PRIMARY KEY AUTO_INCREMENT,
    customer_id   INT COMMENT 'Customer ID (references TPC-H customer table)',
    feedback_date DATE COMMENT 'Feedback date',
    rating        INT COMMENT 'Rating (1-5)',
    category      VARCHAR(50) COMMENT 'Feedback category (product_quality/delivery/service/pricing)',
    comment       TEXT COMMENT 'Detailed feedback',
    resolved      TINYINT(1) DEFAULT 0 COMMENT 'Whether resolved'
) COMMENT='CRM — Customer feedback and ratings (references TPC-H customer table)';

INSERT INTO customer_feedback (customer_id, feedback_date, rating, category, comment, resolved) VALUES
(1,  '2024-08-15', 2, 'delivery',        'Shipment delayed by 10 days', 1),
(2,  '2024-09-01', 4, 'product_quality',  'Good quality parts', 1),
(3,  '2024-07-20', 1, 'delivery',        'Wrong items delivered twice', 0),
(5,  '2024-10-05', 5, 'service',         'Excellent customer support', 1),
(7,  '2024-08-22', 3, 'pricing',         'Prices higher than competitors', 1),
(8,  '2024-11-10', 2, 'product_quality',  'Parts did not match specifications', 0),
(10, '2024-09-15', 4, 'service',         'Quick response to inquiries', 1),
(12, '2024-10-20', 1, 'delivery',        'Package arrived damaged', 0),
(14, '2024-11-01', 5, 'product_quality',  'Outstanding quality, will order again', 1),
(15, '2024-07-10', 3, 'pricing',         'Volume discount could be better', 1);


-- --- Module: Logistics (5 tables) --- Directly related (links to orders/lineitem) ---

CREATE TABLE warehouses (
    warehouse_id   INT PRIMARY KEY AUTO_INCREMENT,
    warehouse_name VARCHAR(100) NOT NULL COMMENT 'Warehouse name',
    region_name    VARCHAR(50) COMMENT 'Region (East/South/North/Southwest/Overseas)',
    city           VARCHAR(50) COMMENT 'City',
    capacity       INT COMMENT 'Maximum capacity (storage slots)',
    manager_id     INT COMMENT 'Warehouse manager',
    FOREIGN KEY (manager_id) REFERENCES employees(emp_id)
) COMMENT='Logistics — Warehouse information';

INSERT INTO warehouses (warehouse_name, region_name, city, capacity, manager_id) VALUES
('Shanghai Main Warehouse',  'East',      'Shanghai',  50000, 5),
('Beijing Distribution',     'North',     'Beijing',   30000, 13),
('Shenzhen Export Hub',      'South',     'Shenzhen',  40000, 5),
('Chengdu Regional',         'Southwest', 'Chengdu',   20000, 13),
('Hong Kong Transit',        'Overseas',  'Hong Kong', 15000, NULL),
('Frankfurt EU Warehouse',   'Overseas',  'Frankfurt', 25000, NULL);

CREATE TABLE inventory (
    inventory_id   INT PRIMARY KEY AUTO_INCREMENT,
    warehouse_id   INT NOT NULL COMMENT 'Warehouse ID',
    part_id        INT NOT NULL COMMENT 'Part ID (references TPC-H part table)',
    quantity       INT NOT NULL DEFAULT 0 COMMENT 'Current stock quantity',
    min_stock      INT DEFAULT 10 COMMENT 'Minimum safety stock',
    max_stock      INT DEFAULT 1000 COMMENT 'Maximum stock',
    last_restock   DATE COMMENT 'Last restock date',
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(warehouse_id)
) COMMENT='Logistics — Warehouse inventory (references TPC-H part table, tracks stock levels per warehouse)';

INSERT INTO inventory (warehouse_id, part_id, quantity, min_stock, max_stock, last_restock) VALUES
(1, 1,  500, 50, 2000, '2025-01-15'),
(1, 2,  30,  50, 1500, '2024-12-01'),  -- below min_stock!
(1, 3,  800, 50, 2000, '2025-01-20'),
(1, 5,  15,  50, 1000, '2024-11-10'),  -- below min_stock!
(2, 1,  200, 30, 1000, '2025-01-10'),
(2, 4,  450, 30, 1000, '2025-01-18'),
(2, 7,  5,   30, 800,  '2024-10-05'),  -- below min_stock!
(3, 2,  600, 40, 1500, '2025-01-12'),
(3, 6,  350, 40, 1200, '2025-01-05'),
(3, 10, 900, 40, 2000, '2025-01-22'),
(4, 3,  180, 20, 800,  '2025-01-08'),
(4, 8,  8,   20, 600,  '2024-09-20'),  -- below min_stock!
(5, 5,  250, 20, 1000, '2025-01-16'),
(6, 1,  400, 30, 1500, '2025-01-20'),
(6, 9,  120, 30, 800,  '2024-12-28');

CREATE TABLE shipping_methods (
    method_id   INT PRIMARY KEY AUTO_INCREMENT,
    method_name VARCHAR(50) NOT NULL COMMENT 'Shipping method name',
    carrier     VARCHAR(100) COMMENT 'Carrier',
    avg_days    INT COMMENT 'Average transit days',
    cost_per_kg DECIMAL(8,2) COMMENT 'Cost per kilogram'
) COMMENT='Logistics — Shipping methods and carriers';

INSERT INTO shipping_methods (method_name, carrier, avg_days, cost_per_kg) VALUES
('Express Air',     'FedEx',          2,  25.00),
('Standard Air',    'DHL',            4,  15.00),
('Ocean Freight',   'Maersk',        21,   3.50),
('Rail Freight',    'China Railway',  7,   6.00),
('Truck Domestic',  'SF Express',     3,   8.00),
('Economy Sea',     'COSCO',         30,   2.00);

CREATE TABLE shipments (
    shipment_id    INT PRIMARY KEY AUTO_INCREMENT,
    order_id       INT COMMENT 'Order ID (references TPC-H orders table)',
    warehouse_id   INT COMMENT 'Source warehouse',
    method_id      INT COMMENT 'Shipping method',
    ship_date      DATE COMMENT 'Ship date',
    estimated_arrival DATE COMMENT 'Estimated arrival date',
    actual_arrival    DATE COMMENT 'Actual arrival date',
    tracking_no    VARCHAR(50) COMMENT 'Tracking number',
    status         VARCHAR(30) COMMENT 'Status (preparing/shipped/in_transit/delivered/returned)',
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(warehouse_id),
    FOREIGN KEY (method_id) REFERENCES shipping_methods(method_id)
) COMMENT='Logistics — Shipment records (references TPC-H orders table)';

INSERT INTO shipments (order_id, warehouse_id, method_id, ship_date, estimated_arrival, actual_arrival, tracking_no, status) VALUES
(1,  1, 5, '1996-01-05', '1996-01-08', '1996-01-09', 'SF1996010500001', 'delivered'),
(2,  1, 1, '1996-12-05', '1996-12-07', '1996-12-08', 'FX1996120500001', 'delivered'),
(3,  2, 3, '1993-10-20', '1993-11-10', '1993-11-12', 'MK1993102000001', 'delivered'),
(4,  1, 5, '1995-10-15', '1995-10-18', '1995-10-17', 'SF1995101500001', 'delivered'),
(5,  3, 2, '1994-08-05', '1994-08-09', '1994-08-10', 'DH1994080500001', 'delivered'),
(6,  1, 4, '1992-02-25', '1992-03-03', '1992-03-05', 'CR1992022500001', 'delivered'),
(7,  1, 1, '1996-01-15', '1996-01-17', '1996-01-18', 'FX1996011500001', 'delivered'),
(32, 2, 5, '1995-07-20', '1995-07-23', '1995-07-25', 'SF1995072000001', 'delivered'),
(33, 1, 4, '1993-11-01', '1993-11-08', '1993-11-07', 'CR1993110100001', 'delivered'),
(34, 3, 2, '1998-07-25', '1998-07-29', NULL,          'DH1998072500001', 'in_transit'),
(35, 2, 3, '1995-10-28', '1995-11-18', '1995-11-25', 'MK1995102800001', 'delivered'),
(36, 1, 1, '1995-11-08', '1995-11-10', '1995-11-10', 'FX1995110800001', 'delivered'),
(37, 4, 6, '1992-06-10', '1992-07-10', '1992-07-15', 'CC1992061000001', 'delivered'),
(38, 2, 5, '1996-08-25', '1996-08-28', '1996-09-05', 'SF1996082500001', 'delivered'),
(39, 1, 4, '1996-09-25', '1996-10-02', NULL,          'CR1996092500001', 'in_transit');

CREATE TABLE delivery_tracking (
    tracking_id   INT PRIMARY KEY AUTO_INCREMENT,
    shipment_id   INT NOT NULL COMMENT 'Shipment ID',
    event_time    DATETIME COMMENT 'Event time',
    location      VARCHAR(200) COMMENT 'Current location',
    event_type    VARCHAR(50) COMMENT 'Event type (pickup/departure/arrival/customs/delivered)',
    notes         VARCHAR(300) COMMENT 'Notes',
    FOREIGN KEY (shipment_id) REFERENCES shipments(shipment_id)
) COMMENT='Logistics — Delivery tracking events';

INSERT INTO delivery_tracking (shipment_id, event_time, location, event_type, notes) VALUES
(1,  '1996-01-05 09:00:00', 'Shanghai Warehouse',   'pickup',    'Package picked up'),
(1,  '1996-01-06 14:00:00', 'Shanghai Hub',         'departure', 'In transit to destination'),
(1,  '1996-01-09 10:00:00', 'Customer Address',     'delivered', 'Signed by recipient'),
(10, '1998-07-25 08:00:00', 'Shenzhen Export Hub',  'pickup',    'Package picked up'),
(10, '1998-07-26 16:00:00', 'Shenzhen Airport',     'departure', 'Dispatched via DHL'),
(10, '1998-07-28 11:00:00', 'Tokyo Customs',        'customs',   'Customs clearance in progress'),
(14, '1992-06-10 09:00:00', 'Chengdu Regional',     'pickup',    'Package picked up'),
(14, '1992-06-15 12:00:00', 'Shanghai Port',        'departure', 'Loaded on vessel'),
(14, '1992-07-08 08:00:00', 'Rotterdam Port',       'arrival',   'Arrived at destination port'),
(14, '1992-07-15 14:00:00', 'Customer Warehouse',   'delivered', 'Delivered');


-- --- Module: Finance (5 tables) --- Semi-related (invoices links to orders) ---

CREATE TABLE accounts (
    account_id    INT PRIMARY KEY AUTO_INCREMENT,
    account_name  VARCHAR(100) NOT NULL COMMENT 'Account name',
    account_type  VARCHAR(30) COMMENT 'Account type (asset/liability/equity/revenue/expense)',
    balance       DECIMAL(15,2) DEFAULT 0 COMMENT 'Current balance',
    currency      CHAR(3) DEFAULT 'CNY' COMMENT 'Currency'
) COMMENT='Finance — Chart of accounts';

INSERT INTO accounts (account_name, account_type, balance, currency) VALUES
('Cash',                'asset',     15000000.00, 'CNY'),
('Accounts Receivable', 'asset',     8500000.00,  'CNY'),
('Inventory Asset',     'asset',     12000000.00, 'CNY'),
('Accounts Payable',    'liability', 6000000.00,  'CNY'),
('Sales Revenue',       'revenue',   45000000.00, 'CNY'),
('COGS',                'expense',   28000000.00, 'CNY'),
('Operating Expenses',  'expense',   9000000.00,  'CNY'),
('USD Cash',            'asset',     2000000.00,  'USD'),
('EUR Cash',            'asset',     1500000.00,  'EUR'),
('Tax Payable',         'liability', 3200000.00,  'CNY');

CREATE TABLE invoices (
    invoice_id    INT PRIMARY KEY AUTO_INCREMENT,
    order_id      INT COMMENT 'Associated order (TPC-H orders table)',
    invoice_no    VARCHAR(30) NOT NULL COMMENT 'Invoice number',
    invoice_date  DATE COMMENT 'Invoice date',
    due_date      DATE COMMENT 'Due date',
    amount        DECIMAL(15,2) COMMENT 'Invoice amount',
    tax_amount    DECIMAL(15,2) COMMENT 'Tax amount',
    status        VARCHAR(20) COMMENT 'Status (draft/sent/paid/overdue/cancelled)',
    payment_date  DATE COMMENT 'Payment date'
) COMMENT='Finance — Invoice management (references TPC-H orders table)';

INSERT INTO invoices (order_id, invoice_no, invoice_date, due_date, amount, tax_amount, status, payment_date) VALUES
(1,  'INV-2024-0001', '1996-01-05', '1996-02-05', 172799.49, 22263.93, 'paid',    '1996-01-28'),
(2,  'INV-2024-0002', '1996-12-05', '1997-01-05', 46929.18,  6042.46,  'paid',    '1996-12-30'),
(3,  'INV-2024-0003', '1993-10-18', '1993-11-18', 193846.25, 24952.45, 'paid',    '1993-11-10'),
(5,  'INV-2024-0005', '1994-08-02', '1994-09-02', 77471.44,  9976.47,  'paid',    '1994-08-25'),
(7,  'INV-2024-0007', '1996-01-15', '1996-02-15', 252004.18, 32453.03, 'paid',    '1996-02-10'),
(34, 'INV-2024-0034', '1998-07-25', '1998-08-25', 58949.67,  7592.32,  'overdue', NULL),
(36, 'INV-2024-0036', '1995-11-08', '1995-12-08', 173665.47, 22366.83, 'paid',    '1995-12-01'),
(39, 'INV-2024-0039', '1996-09-25', '1996-10-25', 219707.84, 28288.07, 'sent',    NULL);

CREATE TABLE payments (
    payment_id     INT PRIMARY KEY AUTO_INCREMENT,
    invoice_id     INT COMMENT 'Associated invoice',
    payment_date   DATE COMMENT 'Payment date',
    amount         DECIMAL(15,2) COMMENT 'Payment amount',
    payment_method VARCHAR(30) COMMENT 'Payment method (wire_transfer/check/credit_card/cash)',
    reference_no   VARCHAR(50) COMMENT 'Payment reference number',
    FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id)
) COMMENT='Finance — Payment records';

INSERT INTO payments (invoice_id, payment_date, amount, payment_method, reference_no) VALUES
(1, '1996-01-28', 172799.49, 'wire_transfer', 'PAY-1996-0001'),
(2, '1996-12-30', 46929.18,  'wire_transfer', 'PAY-1996-0002'),
(3, '1993-11-10', 193846.25, 'check',         'PAY-1993-0003'),
(4, '1994-08-25', 77471.44,  'wire_transfer', 'PAY-1994-0005'),
(5, '1996-02-10', 252004.18, 'wire_transfer', 'PAY-1996-0007'),
(7, '1995-12-01', 173665.47, 'wire_transfer', 'PAY-1995-0036');

CREATE TABLE budget_items (
    budget_id     INT PRIMARY KEY AUTO_INCREMENT,
    dept_id       INT COMMENT 'Department ID',
    fiscal_year   INT COMMENT 'Fiscal year',
    category      VARCHAR(50) COMMENT 'Budget category (personnel/operations/equipment/travel/marketing)',
    planned       DECIMAL(15,2) COMMENT 'Planned amount',
    actual        DECIMAL(15,2) COMMENT 'Actual amount',
    variance      DECIMAL(15,2) COMMENT 'Variance (actual - planned)',
    FOREIGN KEY (dept_id) REFERENCES departments(dept_id)
) COMMENT='Finance — Department budget execution';

INSERT INTO budget_items (dept_id, fiscal_year, category, planned, actual, variance) VALUES
(1, 2024, 'personnel',  3000000.00, 3100000.00,  100000.00),
(1, 2024, 'travel',      500000.00,  480000.00,  -20000.00),
(2, 2024, 'personnel',  5000000.00, 4800000.00, -200000.00),
(2, 2024, 'equipment',  2000000.00, 2300000.00,  300000.00),
(3, 2024, 'marketing',  2000000.00, 1950000.00,  -50000.00),
(4, 2024, 'operations',  800000.00,  780000.00,  -20000.00),
(5, 2024, 'operations', 2500000.00, 2600000.00,  100000.00),
(5, 2024, 'equipment',  1000000.00,  950000.00,  -50000.00);

CREATE TABLE tax_rates (
    tax_id      INT PRIMARY KEY AUTO_INCREMENT,
    tax_name    VARCHAR(50) NOT NULL COMMENT 'Tax name',
    rate        DECIMAL(6,4) COMMENT 'Tax rate',
    region      VARCHAR(50) COMMENT 'Applicable region',
    effective_from DATE COMMENT 'Effective from date',
    effective_to   DATE COMMENT 'Effective to date'
) COMMENT='Finance — Tax rate configuration';

INSERT INTO tax_rates (tax_name, rate, region, effective_from, effective_to) VALUES
('VAT Standard',       0.1300, 'China',         '2019-04-01', NULL),
('VAT Reduced',        0.0900, 'China',         '2019-04-01', NULL),
('Export Tax Rebate',  0.1300, 'China',         '2020-01-01', NULL),
('US Sales Tax',       0.0875, 'United States', '2020-01-01', NULL),
('EU VAT Standard',    0.2000, 'Europe',        '2020-01-01', NULL),
('Import Duty',        0.0500, 'China',         '2020-01-01', NULL);


-- --- Module: Product (4 tables) --- Semi-related (links to part) ---

CREATE TABLE product_categories (
    category_id   INT PRIMARY KEY AUTO_INCREMENT,
    category_name VARCHAR(100) NOT NULL COMMENT 'Product category name',
    parent_id     INT COMMENT 'Parent category',
    description   TEXT COMMENT 'Category description',
    FOREIGN KEY (parent_id) REFERENCES product_categories(category_id)
) COMMENT='Product — Product category catalog (tree structure)';

INSERT INTO product_categories (category_name, parent_id, description) VALUES
('Industrial Parts',   NULL, 'All industrial components and parts'),
('Copper Parts',       1,    'Copper-based industrial components'),
('Brass Parts',        1,    'Brass alloy components'),
('Steel Parts',        1,    'Steel-based industrial parts'),
('Nickel Parts',       1,    'Nickel alloy components'),
('Tin Parts',          1,    'Tin-based components'),
('Consumer Goods',     NULL, 'Consumer-facing products'),
('Packaging Materials', NULL, 'Boxes, bags, drums, and containers');

CREATE TABLE product_reviews (
    review_id     INT PRIMARY KEY AUTO_INCREMENT,
    part_id       INT COMMENT 'Part ID (references TPC-H part table)',
    reviewer_name VARCHAR(100) COMMENT 'Reviewer name',
    rating        INT COMMENT 'Rating (1-5)',
    review_date   DATE COMMENT 'Review date',
    title         VARCHAR(200) COMMENT 'Review title',
    content       TEXT COMMENT 'Review content',
    verified      TINYINT(1) DEFAULT 0 COMMENT 'Verified purchase'
) COMMENT='Product — Part review records (references TPC-H part table)';

INSERT INTO product_reviews (part_id, reviewer_name, rating, review_date, title, content, verified) VALUES
(1,  'Customer#000000001', 5, '2024-06-15', 'Excellent copper component',    'Perfect fit, great quality finish. Exactly as described.', 1),
(1,  'Customer#000000007', 4, '2024-08-20', 'Good but pricey',              'Quality is there but retail price is above market average.', 1),
(2,  'Customer#000000003', 3, '2024-07-10', 'Average brass quality',        'Decent product but surface finish could be better.', 1),
(3,  'Customer#000000005', 5, '2024-09-01', 'Top quality polished brass',   'Outstanding quality, perfect tolerance.', 1),
(4,  'Customer#000000002', 2, '2024-08-05', 'Below expectations',           'Plating was uneven, returned for replacement.', 1),
(5,  'Customer#000000010', 4, '2024-10-12', 'Reliable tin product',         'Consistent quality across batches.', 1),
(6,  'Customer#000000008', 1, '2024-07-25', 'Defective steel',             'Multiple scratches found, quality control issue.', 1),
(7,  'Customer#000000012', 5, '2024-11-01', 'Superb copper work',          'Best copper component we have sourced.', 1),
(8,  'Customer#000000006', 3, '2024-09-20', 'Acceptable',                  'Meets minimum requirements but nothing special.', 0),
(9,  'Customer#000000014', 4, '2024-10-30', 'Good steel quality',          'Burn test passed, corrosion resistance is good.', 1),
(10, 'Customer#000000015', 2, '2024-11-15', 'Size variance',               'Actual size deviated from specifications by 2mm.', 1),
(12, 'Customer#000000009', 5, '2024-08-08', 'Excellent anodized steel',    'Color uniformity and hardness exceeded expectations.', 1),
(15, 'Customer#000000011', 4, '2024-07-30', 'Solid brass product',         'Good workmanship, timely delivery.', 1),
(18, 'Customer#000000004', 1, '2024-09-10', 'Very poor quality',           'Copper plating peeled off within a week of installation.', 1),
(20, 'Customer#000000017', 3, '2024-10-05', 'Average nickel finish',       'Acceptable for non-critical applications.', 0);

CREATE TABLE price_history (
    price_id    INT PRIMARY KEY AUTO_INCREMENT,
    part_id     INT COMMENT 'Part ID (references TPC-H part table)',
    old_price   DECIMAL(15,2) COMMENT 'Old price',
    new_price   DECIMAL(15,2) COMMENT 'New price',
    change_date DATE COMMENT 'Price change date',
    reason      VARCHAR(200) COMMENT 'Reason for change'
) COMMENT='Product — Part price change history';

INSERT INTO price_history (part_id, old_price, new_price, change_date, reason) VALUES
(1,  800.00,  920.00,  '2024-01-15', 'Annual price adjustment - copper cost increase'),
(2,  890.00,  902.00,  '2024-01-15', 'Annual price adjustment'),
(3,  750.00,  850.00,  '2024-03-01', 'Raw material cost increase'),
(5,  900.00,  905.00,  '2024-06-01', 'Tin market price surge'),
(10, 680.00,  780.00,  '2024-06-01', 'Steel tariff impact'),
(15, 920.00,  915.00,  '2024-09-01', 'Promotional pricing'),
(18, 925.00,  918.00,  '2024-09-01', 'Competitive price matching');

CREATE TABLE promotions (
    promo_id      INT PRIMARY KEY AUTO_INCREMENT,
    promo_name    VARCHAR(200) NOT NULL COMMENT 'Promotion name',
    discount_pct  DECIMAL(5,2) COMMENT 'Discount percentage (%)',
    start_date    DATE COMMENT 'Start date',
    end_date      DATE COMMENT 'End date',
    min_quantity  INT COMMENT 'Minimum order quantity',
    applicable_brands VARCHAR(200) COMMENT 'Applicable brands (comma-separated)',
    status        VARCHAR(20) DEFAULT 'active' COMMENT 'Status (active/expired/scheduled)'
) COMMENT='Product — Promotion configuration';

INSERT INTO promotions (promo_name, discount_pct, start_date, end_date, min_quantity, applicable_brands, status) VALUES
('New Year Bulk Discount',   10.00, '2025-01-01', '2025-01-31', 100,  'Brand#13,Brand#42',     'active'),
('Spring Clearance',          15.00, '2025-03-01', '2025-03-31', 50,   'Brand#11,Brand#24',     'scheduled'),
('Loyalty Program 5%',        5.00,  '2024-01-01', '2025-12-31', 10,   NULL,                    'active'),
('Copper Parts Special',     12.00, '2025-02-01', '2025-02-28', 200,  'Brand#13,Brand#15',     'active'),
('Quarter-end Flash Sale',   20.00, '2025-03-25', '2025-03-31', 500,  NULL,                    'scheduled');


-- --- Module: System (6 tables) --- Pure distractor module ---

CREATE TABLE user_accounts (
    user_id       INT PRIMARY KEY AUTO_INCREMENT,
    username      VARCHAR(50) NOT NULL UNIQUE COMMENT 'Username',
    password_hash VARCHAR(255) COMMENT 'Password hash',
    role          VARCHAR(30) COMMENT 'Role (admin/manager/analyst/viewer)',
    emp_id        INT COMMENT 'Associated employee',
    last_login    DATETIME COMMENT 'Last login time',
    is_active     TINYINT(1) DEFAULT 1 COMMENT 'Whether active',
    FOREIGN KEY (emp_id) REFERENCES employees(emp_id)
) COMMENT='System — User accounts';

INSERT INTO user_accounts (username, password_hash, role, emp_id, last_login, is_active) VALUES
('admin',     'pbkdf2:sha256:...', 'admin',   8,  '2025-01-20 09:30:00', 1),
('zhangwei',  'pbkdf2:sha256:...', 'manager', 1,  '2025-01-20 08:15:00', 1),
('liming',    'pbkdf2:sha256:...', 'manager', 2,  '2025-01-19 17:45:00', 1),
('chenjie',   'pbkdf2:sha256:...', 'manager', 4,  '2025-01-20 10:00:00', 1),
('liuyang',   'pbkdf2:sha256:...', 'manager', 5,  '2025-01-18 14:20:00', 1),
('analyst01', 'pbkdf2:sha256:...', 'analyst', 14, '2025-01-20 11:00:00', 1),
('viewer01',  'pbkdf2:sha256:...', 'viewer',  NULL, '2025-01-15 09:00:00', 1);

CREATE TABLE audit_log (
    log_id      INT PRIMARY KEY AUTO_INCREMENT,
    user_id     INT COMMENT 'User ID',
    action      VARCHAR(50) COMMENT 'Action type (login/logout/query/export/update/delete)',
    target      VARCHAR(200) COMMENT 'Target object',
    details     TEXT COMMENT 'Details',
    ip_address  VARCHAR(45) COMMENT 'IP address',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user_accounts(user_id)
) COMMENT='System — Audit log';

INSERT INTO audit_log (user_id, action, target, details, ip_address, created_at) VALUES
(1, 'login',  'system',             'Admin login',                    '192.168.1.100', '2025-01-20 09:30:00'),
(2, 'query',  'orders',             'SELECT * FROM orders WHERE ...',  '192.168.1.101', '2025-01-20 08:20:00'),
(2, 'export', 'sales_report_Q4',    'Exported Q4 sales data to CSV',   '192.168.1.101', '2025-01-20 08:45:00'),
(3, 'login',  'system',             'Engineering login',               '192.168.1.102', '2025-01-19 17:45:00'),
(4, 'update', 'budget_items',       'Updated Q1 2025 budget forecast', '192.168.1.103', '2025-01-20 10:15:00'),
(6, 'query',  'inventory',          'Inventory status check',          '192.168.1.104', '2025-01-20 11:05:00');

CREATE TABLE system_config (
    config_id    INT PRIMARY KEY AUTO_INCREMENT,
    config_key   VARCHAR(100) NOT NULL UNIQUE COMMENT 'Configuration key',
    config_value TEXT COMMENT 'Configuration value',
    description  VARCHAR(300) COMMENT 'Description',
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
) COMMENT='System — Global configuration';

INSERT INTO system_config (config_key, config_value, description) VALUES
('company_name',       'TPC-H Enterprise Corp.',  'Company display name'),
('default_currency',   'CNY',                     'Default currency for transactions'),
('fiscal_year_start',  '01-01',                   'Fiscal year start date (MM-DD)'),
('max_export_rows',    '100000',                  'Maximum rows for data export'),
('session_timeout',    '3600',                    'Session timeout in seconds'),
('enable_audit_log',   'true',                    'Whether to log user actions'),
('backup_retention',   '30',                      'Backup retention days'),
('email_smtp_host',    'smtp.corp.internal',      'SMTP server for notifications');

CREATE TABLE notifications (
    notification_id INT PRIMARY KEY AUTO_INCREMENT,
    user_id         INT COMMENT 'Recipient user',
    title           VARCHAR(200) COMMENT 'Notification title',
    message         TEXT COMMENT 'Notification message',
    is_read         TINYINT(1) DEFAULT 0 COMMENT 'Whether read',
    priority        VARCHAR(20) DEFAULT 'normal' COMMENT 'Priority (low/normal/high/urgent)',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user_accounts(user_id)
) COMMENT='System — Notifications';

INSERT INTO notifications (user_id, title, message, is_read, priority, created_at) VALUES
(2, 'Q4 Sales Target Achieved',    'Congratulations! Q4 sales exceeded target by 12%.', 1, 'normal',  '2025-01-15 09:00:00'),
(5, 'Low Inventory Alert',         'Parts #2, #5, #7, #8 below minimum stock levels.',  0, 'urgent',  '2025-01-18 07:00:00'),
(4, 'Budget Review Due',           'FY2025 Q1 budget review deadline: Jan 25.',          0, 'high',    '2025-01-20 08:00:00'),
(1, 'New System Update Available', 'ERP v3.2 update scheduled for Jan 22.',              1, 'low',     '2025-01-19 12:00:00'),
(3, 'Engineering Headcount Approved', 'Two new senior engineer positions approved.',     1, 'normal',  '2025-01-17 14:00:00');

CREATE TABLE user_sessions (
    session_id   VARCHAR(64) PRIMARY KEY COMMENT 'Session ID',
    user_id      INT COMMENT 'User ID',
    login_time   DATETIME COMMENT 'Login time',
    logout_time  DATETIME COMMENT 'Logout time',
    ip_address   VARCHAR(45) COMMENT 'IP address',
    user_agent   VARCHAR(300) COMMENT 'Browser user agent',
    FOREIGN KEY (user_id) REFERENCES user_accounts(user_id)
) COMMENT='System — User session records';

INSERT INTO user_sessions VALUES
('sess_abc123', 1, '2025-01-20 09:30:00', NULL, '192.168.1.100', 'Mozilla/5.0 Chrome/120'),
('sess_def456', 2, '2025-01-20 08:15:00', '2025-01-20 12:30:00', '192.168.1.101', 'Mozilla/5.0 Firefox/121'),
('sess_ghi789', 5, '2025-01-18 14:20:00', '2025-01-18 18:00:00', '192.168.1.104', 'Mozilla/5.0 Chrome/120');

CREATE TABLE data_exports (
    export_id    INT PRIMARY KEY AUTO_INCREMENT,
    user_id      INT COMMENT 'Export user',
    export_type  VARCHAR(30) COMMENT 'Export format (csv/excel/pdf)',
    table_name   VARCHAR(100) COMMENT 'Export data source',
    row_count    INT COMMENT 'Exported row count',
    file_size_kb INT COMMENT 'File size (KB)',
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user_accounts(user_id)
) COMMENT='System — Data export records';

INSERT INTO data_exports (user_id, export_type, table_name, row_count, file_size_kb, created_at) VALUES
(2, 'csv',   'orders',    15000, 2048, '2025-01-20 08:45:00'),
(4, 'excel', 'budget_items', 48, 128,  '2025-01-19 16:30:00'),
(6, 'csv',   'inventory', 500,  64,    '2025-01-20 11:10:00');


-- =============================================================
-- Grant permissions
-- =============================================================
GRANT ALL PRIVILEGES ON tpch_enterprise.* TO 'lucid'@'%';
FLUSH PRIVILEGES;


-- =============================================================
-- Register in LUCID Lake-Base (rc_datasources)
-- =============================================================
USE lucid;

INSERT INTO rc_datasources (name, db_type, host, port, db_name, username, description, status)
VALUES ('tpch_enterprise', 'mariadb', 'lucid-mariadb', 3306, 'tpch_enterprise', 'lucid',
        'TPC-H Enterprise — 517-table enterprise database for Large-Scale Adaptive Schema Linking demo', 'active')
ON DUPLICATE KEY UPDATE status = 'active', description = VALUES(description);


-- === ADDON: 479 enterprise extension tables (with intra-domain FK relations) ===
USE tpch_enterprise;

-- === HR (35 tables) ===

CREATE TABLE IF NOT EXISTS hr_departments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='HR — Departments';
INSERT IGNORE INTO hr_departments (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-02-02 00:00:00', 'done'),
('Beta-2', 'descri_2', 'type_B', '2025-05-01 07:00:00', 'active'),
('Gamma-3', 'descri_3', 'type_C', '2025-03-16 23:00:00', 'active'),
('Delta-4', 'descri_4', 'type_D', '2025-12-11 23:00:00', 'done');

CREATE TABLE IF NOT EXISTS hr_employees (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Employees';
INSERT IGNORE INTO hr_employees (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 76, 'descri_1', 'type_A', '2025-08-21 01:00:00', 'active'),
('Beta-2', 12, 'descri_2', 'type_B', '2025-04-28 07:00:00', 'done');

CREATE TABLE IF NOT EXISTS hr_positions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Positions';
INSERT IGNORE INTO hr_positions (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 4, 'descri_1', 'type_A', '2025-10-08 06:00:00', 'done'),
('Beta-2', 84, 'descri_2', 'type_B', '2025-12-24 17:00:00', 'pending'),
('Gamma-3', 29, 'descri_3', 'type_C', '2025-08-06 18:00:00', 'pending'),
('Delta-4', 1, 'descri_4', 'type_D', '2025-03-26 22:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_recruitment_requisitions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Recruitment Requisitions';
INSERT IGNORE INTO hr_recruitment_requisitions (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 36, 'descri_1', 'type_A', '2025-03-24 06:00:00', 'pending'),
('Beta-2', 14, 'descri_2', 'type_B', '2025-02-20 12:00:00', 'active'),
('Gamma-3', 46, 'descri_3', 'type_C', '2025-06-09 19:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_candidates (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Candidates';
INSERT IGNORE INTO hr_candidates (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 94, 'descri_1', 'type_A', '2025-08-12 17:00:00', 'active'),
('Beta-2', 49, 'descri_2', 'type_B', '2025-02-13 17:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_interviews (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Interviews';
INSERT IGNORE INTO hr_interviews (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 80, 'descri_1', 'type_A', '2025-07-18 18:00:00', 'active'),
('Beta-2', 91, 'descri_2', 'type_B', '2025-02-08 01:00:00', 'done'),
('Gamma-3', 30, 'descri_3', 'type_C', '2025-05-09 02:00:00', 'active'),
('Delta-4', 13, 'descri_4', 'type_D', '2025-07-27 08:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_offer_letters (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Offer Letters';
INSERT IGNORE INTO hr_offer_letters (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 47, 'descri_1', 'type_A', '2025-03-28 11:00:00', 'pending'),
('Beta-2', 27, 'descri_2', 'type_B', '2025-12-08 08:00:00', 'done'),
('Gamma-3', 88, 'descri_3', 'type_C', '2025-12-24 02:00:00', 'done'),
('Delta-4', 82, 'descri_4', 'type_D', '2025-03-04 17:00:00', 'done');

CREATE TABLE IF NOT EXISTS hr_training_courses (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Training Courses';
INSERT IGNORE INTO hr_training_courses (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 21, 'descri_1', 'type_A', '2025-08-13 12:00:00', 'pending'),
('Beta-2', 82, 'descri_2', 'type_B', '2025-12-17 17:00:00', 'active');

CREATE TABLE IF NOT EXISTS hr_training_enrollments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Training Enrollments';
INSERT IGNORE INTO hr_training_enrollments (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 42, 'descri_1', 'type_A', '2025-01-01 07:00:00', 'active'),
('Beta-2', 41, 'descri_2', 'type_B', '2025-07-10 08:00:00', 'active'),
('Gamma-3', 28, 'descri_3', 'type_C', '2025-10-11 22:00:00', 'pending'),
('Delta-4', 28, 'descri_4', 'type_D', '2025-12-28 15:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_performance_goals (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Performance Goals';
INSERT IGNORE INTO hr_performance_goals (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 59, 'descri_1', 'type_A', '2025-03-18 08:00:00', 'active'),
('Beta-2', 32, 'descri_2', 'type_B', '2025-10-08 17:00:00', 'pending'),
('Gamma-3', 96, 'descri_3', 'type_C', '2025-10-20 13:00:00', 'done'),
('Delta-4', 52, 'descri_4', 'type_D', '2025-07-18 07:00:00', 'active');

CREATE TABLE IF NOT EXISTS hr_benefits_plans (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Benefits Plans';
INSERT IGNORE INTO hr_benefits_plans (name, departments_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 64, 2024, 918.09, 480.65, 'active'),
('Beta-2', 20, 2025, 6277.56, 7922.08, 'pending'),
('Gamma-3', 77, 2026, 644.58, 3822.00, 'pending'),
('Delta-4', 68, 2024, 2521.38, 5536.17, 'active');

CREATE TABLE IF NOT EXISTS hr_benefits_enrollments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Benefits Enrollments';
INSERT IGNORE INTO hr_benefits_enrollments (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 93, 'descri_1', 'type_A', '2025-02-03 21:00:00', 'done'),
('Beta-2', 97, 'descri_2', 'type_B', '2025-05-25 20:00:00', 'pending'),
('Gamma-3', 15, 'descri_3', 'type_C', '2025-06-11 13:00:00', 'active'),
('Delta-4', 59, 'descri_4', 'type_D', '2025-01-02 23:00:00', 'done');

CREATE TABLE IF NOT EXISTS hr_leave_requests (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Leave Requests';
INSERT IGNORE INTO hr_leave_requests (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 65, 'descri_1', 'type_A', '2025-04-08 16:00:00', 'active'),
('Beta-2', 81, 'descri_2', 'type_B', '2025-06-13 20:00:00', 'done'),
('Gamma-3', 78, 'descri_3', 'type_C', '2025-04-18 04:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_leave_balances (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Leave Balances';
INSERT IGNORE INTO hr_leave_balances (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 70, 'descri_1', 'type_A', '2025-10-20 00:00:00', 'done'),
('Beta-2', 42, 'descri_2', 'type_B', '2025-09-27 00:00:00', 'active');

CREATE TABLE IF NOT EXISTS hr_timesheets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Timesheets';
INSERT IGNORE INTO hr_timesheets (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 40, 'descri_1', 'type_A', '2025-05-11 01:00:00', 'active'),
('Beta-2', 73, 'descri_2', 'type_B', '2025-02-13 02:00:00', 'done'),
('Gamma-3', 63, 'descri_3', 'type_C', '2025-02-08 17:00:00', 'active');

CREATE TABLE IF NOT EXISTS hr_expense_reports (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Expense Reports';
INSERT IGNORE INTO hr_expense_reports (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 85, 'descri_1', 'type_A', '2025-09-20 17:00:00', 'active'),
('Beta-2', 34, 'descri_2', 'type_B', '2025-10-19 19:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_expense_items (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Expense Items';
INSERT IGNORE INTO hr_expense_items (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 70, 'descri_1', 'type_A', '2025-12-18 06:00:00', 'done'),
('Beta-2', 40, 'descri_2', 'type_B', '2025-07-09 21:00:00', 'done');

CREATE TABLE IF NOT EXISTS hr_employee_documents (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Employee Documents';
INSERT IGNORE INTO hr_employee_documents (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 57, 'descri_1', 'type_A', '2025-09-13 14:00:00', 'active'),
('Beta-2', 32, 'descri_2', 'type_B', '2025-04-04 02:00:00', 'pending'),
('Gamma-3', 3, 'descri_3', 'type_C', '2025-11-22 17:00:00', 'active');

CREATE TABLE IF NOT EXISTS hr_org_chart_history (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Org Chart History';
INSERT IGNORE INTO hr_org_chart_history (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 29, 'descri_1', 'type_A', '2025-01-04 02:00:00', 'done'),
('Beta-2', 81, 'descri_2', 'type_B', '2025-02-03 07:00:00', 'active'),
('Gamma-3', 5, 'descri_3', 'type_C', '2025-06-02 02:00:00', 'done'),
('Delta-4', 31, 'descri_4', 'type_D', '2025-05-03 21:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_salary_bands (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Salary Bands';
INSERT IGNORE INTO hr_salary_bands (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 70, 'descri_1', 'type_A', '2025-03-12 23:00:00', 'done'),
('Beta-2', 74, 'descri_2', 'type_B', '2025-09-19 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_compensation_history (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Compensation History';
INSERT IGNORE INTO hr_compensation_history (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 25, 'descri_1', 'type_A', '2025-02-21 03:00:00', 'done'),
('Beta-2', 56, 'descri_2', 'type_B', '2025-07-14 13:00:00', 'pending'),
('Gamma-3', 60, 'descri_3', 'type_C', '2025-01-28 21:00:00', 'done');

CREATE TABLE IF NOT EXISTS hr_skills (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Skills';
INSERT IGNORE INTO hr_skills (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 13, 'descri_1', 'type_A', '2025-02-04 12:00:00', 'done'),
('Beta-2', 44, 'descri_2', 'type_B', '2025-02-28 07:00:00', 'active'),
('Gamma-3', 25, 'descri_3', 'type_C', '2025-10-23 14:00:00', 'active'),
('Delta-4', 55, 'descri_4', 'type_D', '2025-04-10 08:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_employee_skills (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Employee Skills';
INSERT IGNORE INTO hr_employee_skills (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 10, 'descri_1', 'type_A', '2025-08-03 17:00:00', 'active'),
('Beta-2', 7, 'descri_2', 'type_B', '2025-12-26 17:00:00', 'active');

CREATE TABLE IF NOT EXISTS hr_employee_certifications (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Employee Certifications';
INSERT IGNORE INTO hr_employee_certifications (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 97, 'descri_1', 'type_A', '2025-05-10 05:00:00', 'pending'),
('Beta-2', 63, 'descri_2', 'type_B', '2025-09-23 06:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_disciplinary_actions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Disciplinary Actions';
INSERT IGNORE INTO hr_disciplinary_actions (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 22, 'descri_1', 'type_A', '2025-07-27 00:00:00', 'pending'),
('Beta-2', 34, 'descri_2', 'type_B', '2025-08-09 09:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_employee_surveys (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Employee Surveys';
INSERT IGNORE INTO hr_employee_surveys (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 94, 'descri_1', 'type_A', '2025-10-05 21:00:00', 'done'),
('Beta-2', 63, 'descri_2', 'type_B', '2025-03-24 06:00:00', 'pending'),
('Gamma-3', 28, 'descri_3', 'type_C', '2025-01-02 18:00:00', 'done'),
('Delta-4', 70, 'descri_4', 'type_D', '2025-02-04 23:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_succession_plans (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Succession Plans';
INSERT IGNORE INTO hr_succession_plans (name, departments_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 7, 2024, 5845.35, 5032.97, 'done'),
('Beta-2', 21, 2025, 578.17, 5082.70, 'active');

CREATE TABLE IF NOT EXISTS hr_onboarding_tasks (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Onboarding Tasks';
INSERT IGNORE INTO hr_onboarding_tasks (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 77, 'descri_1', 'type_A', '2025-02-07 21:00:00', 'active'),
('Beta-2', 52, 'descri_2', 'type_B', '2025-03-06 18:00:00', 'active');

CREATE TABLE IF NOT EXISTS hr_exit_interviews (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Exit Interviews';
INSERT IGNORE INTO hr_exit_interviews (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 77, 'descri_1', 'type_A', '2025-01-21 19:00:00', 'active'),
('Beta-2', 54, 'descri_2', 'type_B', '2025-12-01 18:00:00', 'done'),
('Gamma-3', 67, 'descri_3', 'type_C', '2025-06-22 08:00:00', 'active'),
('Delta-4', 86, 'descri_4', 'type_D', '2025-06-21 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS hr_payroll_runs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Payroll Runs';
INSERT IGNORE INTO hr_payroll_runs (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 17, 'descri_1', 'type_A', '2025-12-08 20:00:00', 'pending'),
('Beta-2', 59, 'descri_2', 'type_B', '2025-06-22 02:00:00', 'active'),
('Gamma-3', 59, 'descri_3', 'type_C', '2025-11-11 18:00:00', 'active');

CREATE TABLE IF NOT EXISTS hr_payroll_details (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Payroll Details';
INSERT IGNORE INTO hr_payroll_details (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 69, 'descri_1', 'type_A', '2025-04-26 16:00:00', 'pending'),
('Beta-2', 17, 'descri_2', 'type_B', '2025-06-11 02:00:00', 'active');

CREATE TABLE IF NOT EXISTS hr_workforce_forecast (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Workforce Forecast';
INSERT IGNORE INTO hr_workforce_forecast (name, departments_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 37, 2024, 1585.73, 8338.28, 'done'),
('Beta-2', 39, 2025, 6120.05, 9871.47, 'done'),
('Gamma-3', 68, 2026, 88.15, 8172.05, 'pending');

CREATE TABLE IF NOT EXISTS hr_emergency_contacts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Emergency Contacts';
INSERT IGNORE INTO hr_emergency_contacts (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 14, 'descri_1', 'type_A', '2025-03-13 08:00:00', 'active'),
('Beta-2', 14, 'descri_2', 'type_B', '2025-10-04 04:00:00', 'pending'),
('Gamma-3', 37, 'descri_3', 'type_C', '2025-11-02 06:00:00', 'done'),
('Delta-4', 44, 'descri_4', 'type_D', '2025-04-21 21:00:00', 'done');

CREATE TABLE IF NOT EXISTS hr_grievances (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Grievances';
INSERT IGNORE INTO hr_grievances (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 65, 'descri_1', 'type_A', '2025-09-27 08:00:00', 'active'),
('Beta-2', 12, 'descri_2', 'type_B', '2025-11-17 13:00:00', 'pending'),
('Gamma-3', 6, 'descri_3', 'type_C', '2025-01-02 10:00:00', 'active');

CREATE TABLE IF NOT EXISTS hr_workplace_incidents (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    departments_id               INT COMMENT 'FK to hr_departments' COMMENT 'Ref hr_departments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (departments_id) REFERENCES hr_departments(id)
) COMMENT='HR — Workplace Incidents';
INSERT IGNORE INTO hr_workplace_incidents (name, departments_id, description, category, created_at, status) VALUES
('Alpha-1', 34, 'descri_1', 'type_A', '2025-03-27 23:00:00', 'pending'),
('Beta-2', 71, 'descri_2', 'type_B', '2025-13-26 13:00:00', 'done'),
('Gamma-3', 2, 'descri_3', 'type_C', '2025-02-02 02:00:00', 'done'),
('Delta-4', 20, 'descri_4', 'type_D', '2025-10-28 01:00:00', 'pending');

-- === Finance (25 tables) ===

CREATE TABLE IF NOT EXISTS fin_chart_of_accounts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Finance — Chart Of Accounts';
INSERT IGNORE INTO fin_chart_of_accounts (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-10-03 04:00:00', 'pending'),
('Beta-2', 'descri_2', 'type_B', '2025-03-10 01:00:00', 'pending'),
('Gamma-3', 'descri_3', 'type_C', '2025-07-19 01:00:00', 'pending'),
('Delta-4', 'descri_4', 'type_D', '2025-04-24 21:00:00', 'active');

CREATE TABLE IF NOT EXISTS fin_journal_entries (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Journal Entries';
INSERT IGNORE INTO fin_journal_entries (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 27, 'descri_1', 'type_A', '2025-07-14 17:00:00', 'pending'),
('Beta-2', 159, 'descri_2', 'type_B', '2025-03-24 07:00:00', 'active'),
('Gamma-3', 46, 'descri_3', 'type_C', '2025-08-16 00:00:00', 'active'),
('Delta-4', 189, 'descri_4', 'type_D', '2025-06-03 13:00:00', 'done');

CREATE TABLE IF NOT EXISTS fin_journal_lines (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Journal Lines';
INSERT IGNORE INTO fin_journal_lines (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 64, 'descri_1', 'type_A', '2025-05-25 05:00:00', 'done'),
('Beta-2', 28, 'descri_2', 'type_B', '2025-07-28 01:00:00', 'pending'),
('Gamma-3', 57, 'descri_3', 'type_C', '2025-04-19 14:00:00', 'pending'),
('Delta-4', 79, 'descri_4', 'type_D', '2025-04-05 07:00:00', 'active');

CREATE TABLE IF NOT EXISTS fin_fiscal_periods (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Fiscal Periods';
INSERT IGNORE INTO fin_fiscal_periods (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 50, 'descri_1', 'type_A', '2025-07-09 10:00:00', 'pending'),
('Beta-2', 18, 'descri_2', 'type_B', '2025-05-03 11:00:00', 'done'),
('Gamma-3', 131, 'descri_3', 'type_C', '2025-07-09 21:00:00', 'done'),
('Delta-4', 85, 'descri_4', 'type_D', '2025-01-15 03:00:00', 'pending');

CREATE TABLE IF NOT EXISTS fin_vendor_master (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Vendor Master';
INSERT IGNORE INTO fin_vendor_master (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 149, 'descri_1', 'type_A', '2025-05-24 01:00:00', 'active'),
('Beta-2', 153, 'descri_2', 'type_B', '2025-08-27 11:00:00', 'done');

CREATE TABLE IF NOT EXISTS fin_purchase_orders (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Purchase Orders';
INSERT IGNORE INTO fin_purchase_orders (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 112, 'descri_1', 'type_A', '2025-11-03 16:00:00', 'active'),
('Beta-2', 99, 'descri_2', 'type_B', '2025-10-16 06:00:00', 'pending'),
('Gamma-3', 12, 'descri_3', 'type_C', '2025-13-27 13:00:00', 'active');

CREATE TABLE IF NOT EXISTS fin_po_line_items (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Po Line Items';
INSERT IGNORE INTO fin_po_line_items (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 138, 'descri_1', 'type_A', '2025-12-16 23:00:00', 'done'),
('Beta-2', 189, 'descri_2', 'type_B', '2025-12-08 06:00:00', 'pending'),
('Gamma-3', 111, 'descri_3', 'type_C', '2025-02-08 21:00:00', 'pending'),
('Delta-4', 160, 'descri_4', 'type_D', '2025-06-21 21:00:00', 'active');

CREATE TABLE IF NOT EXISTS fin_ap_invoices (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Ap Invoices';
INSERT IGNORE INTO fin_ap_invoices (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 77, 'descri_1', 'type_A', '2025-09-08 09:00:00', 'done'),
('Beta-2', 105, 'descri_2', 'type_B', '2025-06-28 12:00:00', 'done'),
('Gamma-3', 76, 'descri_3', 'type_C', '2025-10-04 04:00:00', 'active'),
('Delta-4', 108, 'descri_4', 'type_D', '2025-12-05 12:00:00', 'done');

CREATE TABLE IF NOT EXISTS fin_ap_payments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Ap Payments';
INSERT IGNORE INTO fin_ap_payments (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 45, 'descri_1', 'type_A', '2025-11-08 18:00:00', 'pending'),
('Beta-2', 104, 'descri_2', 'type_B', '2025-10-01 00:00:00', 'pending'),
('Gamma-3', 74, 'descri_3', 'type_C', '2025-04-24 13:00:00', 'done'),
('Delta-4', 156, 'descri_4', 'type_D', '2025-12-28 10:00:00', 'pending');

CREATE TABLE IF NOT EXISTS fin_ar_invoices (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Ar Invoices';
INSERT IGNORE INTO fin_ar_invoices (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 114, 'descri_1', 'type_A', '2025-12-10 06:00:00', 'done'),
('Beta-2', 122, 'descri_2', 'type_B', '2025-03-03 21:00:00', 'active'),
('Gamma-3', 73, 'descri_3', 'type_C', '2025-09-12 21:00:00', 'done');

CREATE TABLE IF NOT EXISTS fin_ar_receipts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Ar Receipts';
INSERT IGNORE INTO fin_ar_receipts (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 86, 'descri_1', 'type_A', '2025-02-20 07:00:00', 'done'),
('Beta-2', 80, 'descri_2', 'type_B', '2025-04-04 06:00:00', 'active'),
('Gamma-3', 7, 'descri_3', 'type_C', '2025-01-24 07:00:00', 'pending'),
('Delta-4', 157, 'descri_4', 'type_D', '2025-02-10 14:00:00', 'pending');

CREATE TABLE IF NOT EXISTS fin_bank_accounts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Bank Accounts';
INSERT IGNORE INTO fin_bank_accounts (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 148, 'descri_1', 'type_A', '2025-04-16 22:00:00', 'done'),
('Beta-2', 99, 'descri_2', 'type_B', '2025-09-02 12:00:00', 'active'),
('Gamma-3', 38, 'descri_3', 'type_C', '2025-12-28 22:00:00', 'active'),
('Delta-4', 193, 'descri_4', 'type_D', '2025-02-27 13:00:00', 'active');

CREATE TABLE IF NOT EXISTS fin_bank_transactions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Bank Transactions';
INSERT IGNORE INTO fin_bank_transactions (name, chart_of_accounts_id, event_date, description, created_by, status) VALUES
('Alpha-1', 179, '2025-09-14 14:00:00', 'Sample data row 1', 'create_1', 'active'),
('Beta-2', 143, '2025-05-16 03:00:00', 'Sample data row 2', 'create_2', 'pending');

CREATE TABLE IF NOT EXISTS fin_fixed_assets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Fixed Assets';
INSERT IGNORE INTO fin_fixed_assets (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 119, 'descri_1', 'type_A', '2025-12-06 16:00:00', 'done'),
('Beta-2', 153, 'descri_2', 'type_B', '2025-06-23 14:00:00', 'done');

CREATE TABLE IF NOT EXISTS fin_depreciation_schedules (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Depreciation Schedules';
INSERT IGNORE INTO fin_depreciation_schedules (name, chart_of_accounts_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 130, 2024, 4272.71, 9068.57, 'pending'),
('Beta-2', 41, 2025, 7438.37, 4751.52, 'pending'),
('Gamma-3', 193, 2026, 2479.68, 6379.60, 'done'),
('Delta-4', 125, 2024, 6270.59, 2752.95, 'active');

CREATE TABLE IF NOT EXISTS fin_budgets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Budgets';
INSERT IGNORE INTO fin_budgets (name, chart_of_accounts_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 74, 2024, 2352.36, 3364.78, 'done'),
('Beta-2', 21, 2025, 1392.22, 2320.07, 'done'),
('Gamma-3', 40, 2026, 7066.42, 651.58, 'pending'),
('Delta-4', 85, 2024, 5430.14, 4163.17, 'active');

CREATE TABLE IF NOT EXISTS fin_cost_centers (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Cost Centers';
INSERT IGNORE INTO fin_cost_centers (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 100, 'descri_1', 'type_A', '2025-10-20 22:00:00', 'active'),
('Beta-2', 196, 'descri_2', 'type_B', '2025-10-15 12:00:00', 'pending'),
('Gamma-3', 2, 'descri_3', 'type_C', '2025-07-13 09:00:00', 'pending');

CREATE TABLE IF NOT EXISTS fin_exchange_rates (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Exchange Rates';
INSERT IGNORE INTO fin_exchange_rates (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 138, 'descri_1', 'type_A', '2025-10-28 19:00:00', 'active'),
('Beta-2', 125, 'descri_2', 'type_B', '2025-04-01 08:00:00', 'pending'),
('Gamma-3', 125, 'descri_3', 'type_C', '2025-01-15 12:00:00', 'pending');

CREATE TABLE IF NOT EXISTS fin_tax_jurisdictions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Tax Jurisdictions';
INSERT IGNORE INTO fin_tax_jurisdictions (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 174, 'descri_1', 'type_A', '2025-07-12 23:00:00', 'active'),
('Beta-2', 120, 'descri_2', 'type_B', '2025-03-10 19:00:00', 'done'),
('Gamma-3', 7, 'descri_3', 'type_C', '2025-07-06 18:00:00', 'done'),
('Delta-4', 170, 'descri_4', 'type_D', '2025-01-14 02:00:00', 'done');

CREATE TABLE IF NOT EXISTS fin_intercompany_txns (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Intercompany Txns';
INSERT IGNORE INTO fin_intercompany_txns (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 35, 'descri_1', 'type_A', '2025-08-13 05:00:00', 'active'),
('Beta-2', 67, 'descri_2', 'type_B', '2025-07-27 10:00:00', 'active'),
('Gamma-3', 117, 'descri_3', 'type_C', '2025-06-28 10:00:00', 'pending');

CREATE TABLE IF NOT EXISTS fin_credit_memos (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Credit Memos';
INSERT IGNORE INTO fin_credit_memos (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 193, 'descri_1', 'type_A', '2025-08-20 08:00:00', 'active'),
('Beta-2', 121, 'descri_2', 'type_B', '2025-01-10 23:00:00', 'done'),
('Gamma-3', 14, 'descri_3', 'type_C', '2025-06-12 07:00:00', 'done');

CREATE TABLE IF NOT EXISTS fin_financial_reports (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Financial Reports';
INSERT IGNORE INTO fin_financial_reports (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 200, 'descri_1', 'type_A', '2025-12-26 01:00:00', 'active'),
('Beta-2', 64, 'descri_2', 'type_B', '2025-04-19 00:00:00', 'done');

CREATE TABLE IF NOT EXISTS fin_audit_findings (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Audit Findings';
INSERT IGNORE INTO fin_audit_findings (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 62, 'descri_1', 'type_A', '2025-03-09 15:00:00', 'done'),
('Beta-2', 30, 'descri_2', 'type_B', '2025-10-09 06:00:00', 'pending');

CREATE TABLE IF NOT EXISTS fin_cash_flow_forecasts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Cash Flow Forecasts';
INSERT IGNORE INTO fin_cash_flow_forecasts (name, chart_of_accounts_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 66, 2024, 7670.55, 1686.05, 'done'),
('Beta-2', 192, 2025, 7185.67, 7779.36, 'active'),
('Gamma-3', 80, 2026, 1089.80, 266.50, 'pending'),
('Delta-4', 148, 2024, 6776.02, 9581.19, 'pending');

CREATE TABLE IF NOT EXISTS fin_revenue_recognition (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    chart_of_accounts_id         INT COMMENT 'FK to fin_chart_of_accounts' COMMENT 'Ref fin_chart_of_accounts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (chart_of_accounts_id) REFERENCES fin_chart_of_accounts(id)
) COMMENT='Finance — Revenue Recognition';
INSERT IGNORE INTO fin_revenue_recognition (name, chart_of_accounts_id, description, category, created_at, status) VALUES
('Alpha-1', 51, 'descri_1', 'type_A', '2025-02-11 18:00:00', 'done'),
('Beta-2', 161, 'descri_2', 'type_B', '2025-05-13 03:00:00', 'done'),
('Gamma-3', 198, 'descri_3', 'type_C', '2025-06-15 21:00:00', 'done'),
('Delta-4', 31, 'descri_4', 'type_D', '2025-10-10 01:00:00', 'pending');

-- === Supply Chain (14 tables) ===

CREATE TABLE IF NOT EXISTS scm_procurement_requests (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Supply Chain — Procurement Requests';
INSERT IGNORE INTO scm_procurement_requests (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-08-24 21:00:00', 'pending'),
('Beta-2', 'descri_2', 'type_B', '2025-02-08 16:00:00', 'done'),
('Gamma-3', 'descri_3', 'type_C', '2025-06-07 00:00:00', 'pending'),
('Delta-4', 'descri_4', 'type_D', '2025-09-27 03:00:00', 'pending');

CREATE TABLE IF NOT EXISTS scm_supplier_evaluations (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    procurement_requests_id      INT COMMENT 'FK to scm_procurement_requests' COMMENT 'Ref scm_procurement_requests',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (procurement_requests_id) REFERENCES scm_procurement_requests(id)
) COMMENT='Supply Chain — Supplier Evaluations';
INSERT IGNORE INTO scm_supplier_evaluations (name, procurement_requests_id, description, category, created_at, status) VALUES
('Alpha-1', 82, 'descri_1', 'type_A', '2025-08-12 22:00:00', 'active'),
('Beta-2', 56, 'descri_2', 'type_B', '2025-04-07 23:00:00', 'done'),
('Gamma-3', 84, 'descri_3', 'type_C', '2025-05-27 19:00:00', 'done');

CREATE TABLE IF NOT EXISTS scm_goods_receipts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    procurement_requests_id      INT COMMENT 'FK to scm_procurement_requests' COMMENT 'Ref scm_procurement_requests',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (procurement_requests_id) REFERENCES scm_procurement_requests(id)
) COMMENT='Supply Chain — Goods Receipts';
INSERT IGNORE INTO scm_goods_receipts (name, procurement_requests_id, description, category, created_at, status) VALUES
('Alpha-1', 60, 'descri_1', 'type_A', '2025-08-28 23:00:00', 'done'),
('Beta-2', 35, 'descri_2', 'type_B', '2025-06-26 07:00:00', 'active'),
('Gamma-3', 36, 'descri_3', 'type_C', '2025-08-07 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS scm_returns_to_vendor (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    procurement_requests_id      INT COMMENT 'FK to scm_procurement_requests' COMMENT 'Ref scm_procurement_requests',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (procurement_requests_id) REFERENCES scm_procurement_requests(id)
) COMMENT='Supply Chain — Returns To Vendor';
INSERT IGNORE INTO scm_returns_to_vendor (name, procurement_requests_id, description, category, created_at, status) VALUES
('Alpha-1', 79, 'descri_1', 'type_A', '2025-12-07 12:00:00', 'pending'),
('Beta-2', 4, 'descri_2', 'type_B', '2025-09-02 10:00:00', 'active'),
('Gamma-3', 63, 'descri_3', 'type_C', '2025-04-25 11:00:00', 'pending'),
('Delta-4', 44, 'descri_4', 'type_D', '2025-05-04 19:00:00', 'done');

CREATE TABLE IF NOT EXISTS scm_shipping_carriers (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    procurement_requests_id      INT COMMENT 'FK to scm_procurement_requests' COMMENT 'Ref scm_procurement_requests',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (procurement_requests_id) REFERENCES scm_procurement_requests(id)
) COMMENT='Supply Chain — Shipping Carriers';
INSERT IGNORE INTO scm_shipping_carriers (name, procurement_requests_id, description, category, created_at, status) VALUES
('Alpha-1', 72, 'descri_1', 'type_A', '2025-01-06 16:00:00', 'active'),
('Beta-2', 11, 'descri_2', 'type_B', '2025-05-12 23:00:00', 'pending'),
('Gamma-3', 63, 'descri_3', 'type_C', '2025-10-05 07:00:00', 'done');

CREATE TABLE IF NOT EXISTS scm_shipment_tracking (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    procurement_requests_id      INT COMMENT 'FK to scm_procurement_requests' COMMENT 'Ref scm_procurement_requests',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (procurement_requests_id) REFERENCES scm_procurement_requests(id)
) COMMENT='Supply Chain — Shipment Tracking';
INSERT IGNORE INTO scm_shipment_tracking (name, procurement_requests_id, description, category, created_at, status) VALUES
('Alpha-1', 83, 'descri_1', 'type_A', '2025-13-01 15:00:00', 'pending'),
('Beta-2', 3, 'descri_2', 'type_B', '2025-02-20 09:00:00', 'active'),
('Gamma-3', 52, 'descri_3', 'type_C', '2025-12-19 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS scm_demand_forecasts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    procurement_requests_id      INT COMMENT 'FK to scm_procurement_requests' COMMENT 'Ref scm_procurement_requests',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (procurement_requests_id) REFERENCES scm_procurement_requests(id)
) COMMENT='Supply Chain — Demand Forecasts';
INSERT IGNORE INTO scm_demand_forecasts (name, procurement_requests_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 75, 2024, 3696.15, 5538.64, 'pending'),
('Beta-2', 55, 2025, 9965.35, 5507.83, 'pending'),
('Gamma-3', 90, 2026, 4542.46, 3072.86, 'active'),
('Delta-4', 16, 2024, 7214.91, 3162.01, 'done');

CREATE TABLE IF NOT EXISTS scm_safety_stock_levels (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    procurement_requests_id      INT COMMENT 'FK to scm_procurement_requests' COMMENT 'Ref scm_procurement_requests',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (procurement_requests_id) REFERENCES scm_procurement_requests(id)
) COMMENT='Supply Chain — Safety Stock Levels';
INSERT IGNORE INTO scm_safety_stock_levels (name, procurement_requests_id, description, category, created_at, status) VALUES
('Alpha-1', 98, 'descri_1', 'type_A', '2025-12-18 05:00:00', 'active'),
('Beta-2', 28, 'descri_2', 'type_B', '2025-09-24 08:00:00', 'done'),
('Gamma-3', 76, 'descri_3', 'type_C', '2025-09-17 19:00:00', 'pending'),
('Delta-4', 13, 'descri_4', 'type_D', '2025-04-16 09:00:00', 'active');

CREATE TABLE IF NOT EXISTS scm_customs_declarations (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    procurement_requests_id      INT COMMENT 'FK to scm_procurement_requests' COMMENT 'Ref scm_procurement_requests',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (procurement_requests_id) REFERENCES scm_procurement_requests(id)
) COMMENT='Supply Chain — Customs Declarations';
INSERT IGNORE INTO scm_customs_declarations (name, procurement_requests_id, description, category, created_at, status) VALUES
('Alpha-1', 23, 'descri_1', 'type_A', '2025-06-15 00:00:00', 'done'),
('Beta-2', 69, 'descri_2', 'type_B', '2025-03-09 08:00:00', 'active'),
('Gamma-3', 7, 'descri_3', 'type_C', '2025-10-04 09:00:00', 'done');

CREATE TABLE IF NOT EXISTS scm_inbound_deliveries (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    procurement_requests_id      INT COMMENT 'FK to scm_procurement_requests' COMMENT 'Ref scm_procurement_requests',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (procurement_requests_id) REFERENCES scm_procurement_requests(id)
) COMMENT='Supply Chain — Inbound Deliveries';
INSERT IGNORE INTO scm_inbound_deliveries (name, procurement_requests_id, description, category, created_at, status) VALUES
('Alpha-1', 82, 'descri_1', 'type_A', '2025-09-28 03:00:00', 'active'),
('Beta-2', 74, 'descri_2', 'type_B', '2025-05-06 15:00:00', 'pending');

CREATE TABLE IF NOT EXISTS scm_outbound_shipments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    procurement_requests_id      INT COMMENT 'FK to scm_procurement_requests' COMMENT 'Ref scm_procurement_requests',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (procurement_requests_id) REFERENCES scm_procurement_requests(id)
) COMMENT='Supply Chain — Outbound Shipments';
INSERT IGNORE INTO scm_outbound_shipments (name, procurement_requests_id, description, category, created_at, status) VALUES
('Alpha-1', 44, 'descri_1', 'type_A', '2025-04-11 01:00:00', 'pending'),
('Beta-2', 62, 'descri_2', 'type_B', '2025-02-03 02:00:00', 'pending'),
('Gamma-3', 63, 'descri_3', 'type_C', '2025-02-10 18:00:00', 'done');

CREATE TABLE IF NOT EXISTS scm_supplier_contracts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    procurement_requests_id      INT COMMENT 'FK to scm_procurement_requests' COMMENT 'Ref scm_procurement_requests',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (procurement_requests_id) REFERENCES scm_procurement_requests(id)
) COMMENT='Supply Chain — Supplier Contracts';
INSERT IGNORE INTO scm_supplier_contracts (name, procurement_requests_id, description, category, created_at, status) VALUES
('Alpha-1', 7, 'descri_1', 'type_A', '2025-03-22 04:00:00', 'done'),
('Beta-2', 39, 'descri_2', 'type_B', '2025-02-16 07:00:00', 'active'),
('Gamma-3', 72, 'descri_3', 'type_C', '2025-08-18 19:00:00', 'done'),
('Delta-4', 80, 'descri_4', 'type_D', '2025-04-04 16:00:00', 'pending');

CREATE TABLE IF NOT EXISTS scm_material_requirements (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    procurement_requests_id      INT COMMENT 'FK to scm_procurement_requests' COMMENT 'Ref scm_procurement_requests',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (procurement_requests_id) REFERENCES scm_procurement_requests(id)
) COMMENT='Supply Chain — Material Requirements';
INSERT IGNORE INTO scm_material_requirements (name, procurement_requests_id, description, category, created_at, status) VALUES
('Alpha-1', 57, 'descri_1', 'type_A', '2025-06-13 18:00:00', 'pending'),
('Beta-2', 40, 'descri_2', 'type_B', '2025-10-12 19:00:00', 'active'),
('Gamma-3', 79, 'descri_3', 'type_C', '2025-02-23 06:00:00', 'done');

CREATE TABLE IF NOT EXISTS scm_freight_invoices (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    procurement_requests_id      INT COMMENT 'FK to scm_procurement_requests' COMMENT 'Ref scm_procurement_requests',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (procurement_requests_id) REFERENCES scm_procurement_requests(id)
) COMMENT='Supply Chain — Freight Invoices';
INSERT IGNORE INTO scm_freight_invoices (name, procurement_requests_id, description, category, created_at, status) VALUES
('Alpha-1', 34, 'descri_1', 'type_A', '2025-12-03 02:00:00', 'active'),
('Beta-2', 31, 'descri_2', 'type_B', '2025-03-05 17:00:00', 'active');

-- === CRM (10 tables) ===

CREATE TABLE IF NOT EXISTS crm_customer_segments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='CRM — Customer Segments';
INSERT IGNORE INTO crm_customer_segments (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-01-02 13:00:00', 'pending'),
('Beta-2', 'descri_2', 'type_B', '2025-12-17 19:00:00', 'pending');

CREATE TABLE IF NOT EXISTS crm_contact_interactions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_segments_id         INT COMMENT 'FK to crm_customer_segments' COMMENT 'Ref crm_customer_segments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_segments_id) REFERENCES crm_customer_segments(id)
) COMMENT='CRM — Contact Interactions';
INSERT IGNORE INTO crm_contact_interactions (name, customer_segments_id, description, category, created_at, status) VALUES
('Alpha-1', 5, 'descri_1', 'type_A', '2025-04-07 09:00:00', 'done'),
('Beta-2', 37, 'descri_2', 'type_B', '2025-12-24 14:00:00', 'active'),
('Gamma-3', 88, 'descri_3', 'type_C', '2025-04-08 08:00:00', 'done');

CREATE TABLE IF NOT EXISTS crm_support_tickets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_segments_id         INT COMMENT 'FK to crm_customer_segments' COMMENT 'Ref crm_customer_segments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_segments_id) REFERENCES crm_customer_segments(id)
) COMMENT='CRM — Support Tickets';
INSERT IGNORE INTO crm_support_tickets (name, customer_segments_id, description, category, created_at, status) VALUES
('Alpha-1', 85, 'descri_1', 'type_A', '2025-04-18 13:00:00', 'active'),
('Beta-2', 70, 'descri_2', 'type_B', '2025-04-04 20:00:00', 'active'),
('Gamma-3', 35, 'descri_3', 'type_C', '2025-03-17 02:00:00', 'active'),
('Delta-4', 22, 'descri_4', 'type_D', '2025-06-18 19:00:00', 'done');

CREATE TABLE IF NOT EXISTS crm_ticket_responses (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_segments_id         INT COMMENT 'FK to crm_customer_segments' COMMENT 'Ref crm_customer_segments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_segments_id) REFERENCES crm_customer_segments(id)
) COMMENT='CRM — Ticket Responses';
INSERT IGNORE INTO crm_ticket_responses (name, customer_segments_id, description, category, created_at, status) VALUES
('Alpha-1', 37, 'descri_1', 'type_A', '2025-08-01 03:00:00', 'pending'),
('Beta-2', 89, 'descri_2', 'type_B', '2025-06-16 22:00:00', 'pending'),
('Gamma-3', 35, 'descri_3', 'type_C', '2025-09-05 17:00:00', 'pending'),
('Delta-4', 57, 'descri_4', 'type_D', '2025-02-14 19:00:00', 'active');

CREATE TABLE IF NOT EXISTS crm_loyalty_programs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_segments_id         INT COMMENT 'FK to crm_customer_segments' COMMENT 'Ref crm_customer_segments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_segments_id) REFERENCES crm_customer_segments(id)
) COMMENT='CRM — Loyalty Programs';
INSERT IGNORE INTO crm_loyalty_programs (name, customer_segments_id, description, category, created_at, status) VALUES
('Alpha-1', 95, 'descri_1', 'type_A', '2025-06-26 19:00:00', 'pending'),
('Beta-2', 4, 'descri_2', 'type_B', '2025-02-19 07:00:00', 'done'),
('Gamma-3', 74, 'descri_3', 'type_C', '2025-11-21 00:00:00', 'done');

CREATE TABLE IF NOT EXISTS crm_loyalty_transactions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_segments_id         INT COMMENT 'FK to crm_customer_segments' COMMENT 'Ref crm_customer_segments',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_segments_id) REFERENCES crm_customer_segments(id)
) COMMENT='CRM — Loyalty Transactions';
INSERT IGNORE INTO crm_loyalty_transactions (name, customer_segments_id, event_date, description, created_by, status) VALUES
('Alpha-1', 74, '2025-01-21 05:00:00', 'Sample data row 1', 'create_1', 'pending'),
('Beta-2', 67, '2025-12-26 14:00:00', 'Sample data row 2', 'create_2', 'pending'),
('Gamma-3', 24, '2025-10-20 13:00:00', 'Sample data row 3', 'create_3', 'done');

CREATE TABLE IF NOT EXISTS crm_feedback_forms (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_segments_id         INT COMMENT 'FK to crm_customer_segments' COMMENT 'Ref crm_customer_segments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_segments_id) REFERENCES crm_customer_segments(id)
) COMMENT='CRM — Feedback Forms';
INSERT IGNORE INTO crm_feedback_forms (name, customer_segments_id, description, category, created_at, status) VALUES
('Alpha-1', 12, 'descri_1', 'type_A', '2025-09-17 11:00:00', 'pending'),
('Beta-2', 43, 'descri_2', 'type_B', '2025-06-25 21:00:00', 'active'),
('Gamma-3', 21, 'descri_3', 'type_C', '2025-06-01 13:00:00', 'done');

CREATE TABLE IF NOT EXISTS crm_nps_scores (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_segments_id         INT COMMENT 'FK to crm_customer_segments' COMMENT 'Ref crm_customer_segments',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_segments_id) REFERENCES crm_customer_segments(id)
) COMMENT='CRM — Nps Scores';
INSERT IGNORE INTO crm_nps_scores (name, customer_segments_id, metric_date, value, target, status) VALUES
('Alpha-1', 37, '2025-12-04', 9451.54, 8136.71, 'done'),
('Beta-2', 5, '2025-08-09', 889.63, 2531.21, 'active'),
('Gamma-3', 99, '2025-07-11', 8649.59, 8250.20, 'active');

CREATE TABLE IF NOT EXISTS crm_customer_addresses (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_segments_id         INT COMMENT 'FK to crm_customer_segments' COMMENT 'Ref crm_customer_segments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_segments_id) REFERENCES crm_customer_segments(id)
) COMMENT='CRM — Customer Addresses';
INSERT IGNORE INTO crm_customer_addresses (name, customer_segments_id, description, category, created_at, status) VALUES
('Alpha-1', 70, 'descri_1', 'type_A', '2025-08-13 13:00:00', 'active'),
('Beta-2', 25, 'descri_2', 'type_B', '2025-09-14 11:00:00', 'done'),
('Gamma-3', 97, 'descri_3', 'type_C', '2025-09-04 20:00:00', 'pending'),
('Delta-4', 98, 'descri_4', 'type_D', '2025-01-27 06:00:00', 'pending');

CREATE TABLE IF NOT EXISTS crm_referral_programs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_segments_id         INT COMMENT 'FK to crm_customer_segments' COMMENT 'Ref crm_customer_segments',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_segments_id) REFERENCES crm_customer_segments(id)
) COMMENT='CRM — Referral Programs';
INSERT IGNORE INTO crm_referral_programs (name, customer_segments_id, description, category, created_at, status) VALUES
('Alpha-1', 17, 'descri_1', 'type_A', '2025-05-08 14:00:00', 'done'),
('Beta-2', 63, 'descri_2', 'type_B', '2025-03-07 00:00:00', 'done'),
('Gamma-3', 78, 'descri_3', 'type_C', '2025-05-11 22:00:00', 'active'),
('Delta-4', 40, 'descri_4', 'type_D', '2025-10-03 00:00:00', 'done');

-- === Manufacturing (10 tables) ===

CREATE TABLE IF NOT EXISTS mfg_bill_of_materials (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Manufacturing — Bill Of Materials';
INSERT IGNORE INTO mfg_bill_of_materials (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-02-20 07:00:00', 'active'),
('Beta-2', 'descri_2', 'type_B', '2025-08-13 03:00:00', 'done'),
('Gamma-3', 'descri_3', 'type_C', '2025-03-23 15:00:00', 'done');

CREATE TABLE IF NOT EXISTS mfg_work_orders (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    bill_of_materials_id         INT COMMENT 'FK to mfg_bill_of_materials' COMMENT 'Ref mfg_bill_of_materials',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (bill_of_materials_id) REFERENCES mfg_bill_of_materials(id)
) COMMENT='Manufacturing — Work Orders';
INSERT IGNORE INTO mfg_work_orders (name, bill_of_materials_id, description, category, created_at, status) VALUES
('Alpha-1', 66, 'descri_1', 'type_A', '2025-13-26 08:00:00', 'pending'),
('Beta-2', 62, 'descri_2', 'type_B', '2025-09-18 07:00:00', 'pending'),
('Gamma-3', 71, 'descri_3', 'type_C', '2025-03-19 12:00:00', 'active');

CREATE TABLE IF NOT EXISTS mfg_work_centers (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    bill_of_materials_id         INT COMMENT 'FK to mfg_bill_of_materials' COMMENT 'Ref mfg_bill_of_materials',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (bill_of_materials_id) REFERENCES mfg_bill_of_materials(id)
) COMMENT='Manufacturing — Work Centers';
INSERT IGNORE INTO mfg_work_centers (name, bill_of_materials_id, description, category, created_at, status) VALUES
('Alpha-1', 66, 'descri_1', 'type_A', '2025-03-14 02:00:00', 'pending'),
('Beta-2', 99, 'descri_2', 'type_B', '2025-08-17 10:00:00', 'done'),
('Gamma-3', 35, 'descri_3', 'type_C', '2025-01-02 09:00:00', 'done'),
('Delta-4', 39, 'descri_4', 'type_D', '2025-11-21 18:00:00', 'done');

CREATE TABLE IF NOT EXISTS mfg_production_schedules (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    bill_of_materials_id         INT COMMENT 'FK to mfg_bill_of_materials' COMMENT 'Ref mfg_bill_of_materials',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (bill_of_materials_id) REFERENCES mfg_bill_of_materials(id)
) COMMENT='Manufacturing — Production Schedules';
INSERT IGNORE INTO mfg_production_schedules (name, bill_of_materials_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 20, 2024, 4470.56, 4848.09, 'pending'),
('Beta-2', 71, 2025, 7630.34, 3777.95, 'pending'),
('Gamma-3', 25, 2026, 9806.96, 2395.43, 'pending');

CREATE TABLE IF NOT EXISTS mfg_quality_inspections (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    bill_of_materials_id         INT COMMENT 'FK to mfg_bill_of_materials' COMMENT 'Ref mfg_bill_of_materials',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (bill_of_materials_id) REFERENCES mfg_bill_of_materials(id)
) COMMENT='Manufacturing — Quality Inspections';
INSERT IGNORE INTO mfg_quality_inspections (name, bill_of_materials_id, description, category, created_at, status) VALUES
('Alpha-1', 100, 'descri_1', 'type_A', '2025-08-15 01:00:00', 'pending'),
('Beta-2', 96, 'descri_2', 'type_B', '2025-09-19 22:00:00', 'pending');

CREATE TABLE IF NOT EXISTS mfg_scrap_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    bill_of_materials_id         INT COMMENT 'FK to mfg_bill_of_materials' COMMENT 'Ref mfg_bill_of_materials',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (bill_of_materials_id) REFERENCES mfg_bill_of_materials(id)
) COMMENT='Manufacturing — Scrap Records';
INSERT IGNORE INTO mfg_scrap_records (name, bill_of_materials_id, event_date, description, created_by, status) VALUES
('Alpha-1', 85, '2025-12-26 04:00:00', 'Sample data row 1', 'create_1', 'pending'),
('Beta-2', 5, '2025-03-09 16:00:00', 'Sample data row 2', 'create_2', 'done'),
('Gamma-3', 43, '2025-02-24 14:00:00', 'Sample data row 3', 'create_3', 'active');

CREATE TABLE IF NOT EXISTS mfg_equipment_master (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    bill_of_materials_id         INT COMMENT 'FK to mfg_bill_of_materials' COMMENT 'Ref mfg_bill_of_materials',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (bill_of_materials_id) REFERENCES mfg_bill_of_materials(id)
) COMMENT='Manufacturing — Equipment Master';
INSERT IGNORE INTO mfg_equipment_master (name, bill_of_materials_id, description, category, created_at, status) VALUES
('Alpha-1', 59, 'descri_1', 'type_A', '2025-01-08 23:00:00', 'active'),
('Beta-2', 53, 'descri_2', 'type_B', '2025-12-28 04:00:00', 'active'),
('Gamma-3', 61, 'descri_3', 'type_C', '2025-05-24 10:00:00', 'done'),
('Delta-4', 89, 'descri_4', 'type_D', '2025-07-08 20:00:00', 'active');

CREATE TABLE IF NOT EXISTS mfg_maintenance_orders (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    bill_of_materials_id         INT COMMENT 'FK to mfg_bill_of_materials' COMMENT 'Ref mfg_bill_of_materials',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (bill_of_materials_id) REFERENCES mfg_bill_of_materials(id)
) COMMENT='Manufacturing — Maintenance Orders';
INSERT IGNORE INTO mfg_maintenance_orders (name, bill_of_materials_id, description, category, created_at, status) VALUES
('Alpha-1', 87, 'descri_1', 'type_A', '2025-10-22 12:00:00', 'pending'),
('Beta-2', 81, 'descri_2', 'type_B', '2025-09-26 17:00:00', 'active'),
('Gamma-3', 80, 'descri_3', 'type_C', '2025-02-08 07:00:00', 'done');

CREATE TABLE IF NOT EXISTS mfg_tooling_inventory (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    bill_of_materials_id         INT COMMENT 'FK to mfg_bill_of_materials' COMMENT 'Ref mfg_bill_of_materials',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (bill_of_materials_id) REFERENCES mfg_bill_of_materials(id)
) COMMENT='Manufacturing — Tooling Inventory';
INSERT IGNORE INTO mfg_tooling_inventory (name, bill_of_materials_id, description, category, created_at, status) VALUES
('Alpha-1', 37, 'descri_1', 'type_A', '2025-04-05 23:00:00', 'active'),
('Beta-2', 56, 'descri_2', 'type_B', '2025-02-23 20:00:00', 'done'),
('Gamma-3', 13, 'descri_3', 'type_C', '2025-08-04 05:00:00', 'done'),
('Delta-4', 39, 'descri_4', 'type_D', '2025-01-15 01:00:00', 'pending');

CREATE TABLE IF NOT EXISTS mfg_production_kpis (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    bill_of_materials_id         INT COMMENT 'FK to mfg_bill_of_materials' COMMENT 'Ref mfg_bill_of_materials',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (bill_of_materials_id) REFERENCES mfg_bill_of_materials(id)
) COMMENT='Manufacturing — Production Kpis';
INSERT IGNORE INTO mfg_production_kpis (name, bill_of_materials_id, metric_date, value, target, status) VALUES
('Alpha-1', 38, '2025-07-16', 3754.29, 1464.10, 'done'),
('Beta-2', 53, '2025-10-10', 6820.50, 1808.59, 'active');

-- === Sales (10 tables) ===

CREATE TABLE IF NOT EXISTS sales_price_lists (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Sales — Price Lists';
INSERT IGNORE INTO sales_price_lists (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-11-05 12:00:00', 'done'),
('Beta-2', 'descri_2', 'type_B', '2025-12-14 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS sales_price_list_items (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    price_lists_id               INT COMMENT 'FK to sales_price_lists' COMMENT 'Ref sales_price_lists',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (price_lists_id) REFERENCES sales_price_lists(id)
) COMMENT='Sales — Price List Items';
INSERT IGNORE INTO sales_price_list_items (name, price_lists_id, description, category, created_at, status) VALUES
('Alpha-1', 19, 'descri_1', 'type_A', '2025-04-07 14:00:00', 'done'),
('Beta-2', 33, 'descri_2', 'type_B', '2025-08-12 08:00:00', 'done'),
('Gamma-3', 2, 'descri_3', 'type_C', '2025-08-15 09:00:00', 'done'),
('Delta-4', 70, 'descri_4', 'type_D', '2025-03-25 02:00:00', 'pending');

CREATE TABLE IF NOT EXISTS sales_sales_quotations (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    price_lists_id               INT COMMENT 'FK to sales_price_lists' COMMENT 'Ref sales_price_lists',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (price_lists_id) REFERENCES sales_price_lists(id)
) COMMENT='Sales — Sales Quotations';
INSERT IGNORE INTO sales_sales_quotations (name, price_lists_id, description, category, created_at, status) VALUES
('Alpha-1', 76, 'descri_1', 'type_A', '2025-06-14 20:00:00', 'pending'),
('Beta-2', 89, 'descri_2', 'type_B', '2025-05-17 14:00:00', 'pending'),
('Gamma-3', 26, 'descri_3', 'type_C', '2025-07-01 15:00:00', 'active');

CREATE TABLE IF NOT EXISTS sales_quote_line_items (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    price_lists_id               INT COMMENT 'FK to sales_price_lists' COMMENT 'Ref sales_price_lists',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (price_lists_id) REFERENCES sales_price_lists(id)
) COMMENT='Sales — Quote Line Items';
INSERT IGNORE INTO sales_quote_line_items (name, price_lists_id, description, category, created_at, status) VALUES
('Alpha-1', 49, 'descri_1', 'type_A', '2025-10-13 11:00:00', 'done'),
('Beta-2', 38, 'descri_2', 'type_B', '2025-12-23 09:00:00', 'active');

CREATE TABLE IF NOT EXISTS sales_sales_orders (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    price_lists_id               INT COMMENT 'FK to sales_price_lists' COMMENT 'Ref sales_price_lists',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (price_lists_id) REFERENCES sales_price_lists(id)
) COMMENT='Sales — Sales Orders';
INSERT IGNORE INTO sales_sales_orders (name, price_lists_id, description, category, created_at, status) VALUES
('Alpha-1', 51, 'descri_1', 'type_A', '2025-05-01 00:00:00', 'done'),
('Beta-2', 88, 'descri_2', 'type_B', '2025-01-26 19:00:00', 'done'),
('Gamma-3', 64, 'descri_3', 'type_C', '2025-05-07 07:00:00', 'done'),
('Delta-4', 46, 'descri_4', 'type_D', '2025-04-01 20:00:00', 'active');

CREATE TABLE IF NOT EXISTS sales_sales_order_items (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    price_lists_id               INT COMMENT 'FK to sales_price_lists' COMMENT 'Ref sales_price_lists',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (price_lists_id) REFERENCES sales_price_lists(id)
) COMMENT='Sales — Sales Order Items';
INSERT IGNORE INTO sales_sales_order_items (name, price_lists_id, description, category, created_at, status) VALUES
('Alpha-1', 33, 'descri_1', 'type_A', '2025-12-12 23:00:00', 'done'),
('Beta-2', 88, 'descri_2', 'type_B', '2025-03-15 20:00:00', 'active'),
('Gamma-3', 81, 'descri_3', 'type_C', '2025-12-23 01:00:00', 'pending'),
('Delta-4', 57, 'descri_4', 'type_D', '2025-01-18 18:00:00', 'pending');

CREATE TABLE IF NOT EXISTS sales_sales_territories (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    price_lists_id               INT COMMENT 'FK to sales_price_lists' COMMENT 'Ref sales_price_lists',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (price_lists_id) REFERENCES sales_price_lists(id)
) COMMENT='Sales — Sales Territories';
INSERT IGNORE INTO sales_sales_territories (name, price_lists_id, description, category, created_at, status) VALUES
('Alpha-1', 17, 'descri_1', 'type_A', '2025-02-19 09:00:00', 'pending'),
('Beta-2', 96, 'descri_2', 'type_B', '2025-08-17 05:00:00', 'active'),
('Gamma-3', 17, 'descri_3', 'type_C', '2025-10-25 11:00:00', 'done'),
('Delta-4', 65, 'descri_4', 'type_D', '2025-05-28 05:00:00', 'pending');

CREATE TABLE IF NOT EXISTS sales_sales_commissions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    price_lists_id               INT COMMENT 'FK to sales_price_lists' COMMENT 'Ref sales_price_lists',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (price_lists_id) REFERENCES sales_price_lists(id)
) COMMENT='Sales — Sales Commissions';
INSERT IGNORE INTO sales_sales_commissions (name, price_lists_id, description, category, created_at, status) VALUES
('Alpha-1', 38, 'descri_1', 'type_A', '2025-06-06 03:00:00', 'pending'),
('Beta-2', 10, 'descri_2', 'type_B', '2025-03-17 07:00:00', 'done'),
('Gamma-3', 93, 'descri_3', 'type_C', '2025-12-10 12:00:00', 'done');

CREATE TABLE IF NOT EXISTS sales_sales_returns (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    price_lists_id               INT COMMENT 'FK to sales_price_lists' COMMENT 'Ref sales_price_lists',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (price_lists_id) REFERENCES sales_price_lists(id)
) COMMENT='Sales — Sales Returns';
INSERT IGNORE INTO sales_sales_returns (name, price_lists_id, description, category, created_at, status) VALUES
('Alpha-1', 12, 'descri_1', 'type_A', '2025-07-07 00:00:00', 'pending'),
('Beta-2', 69, 'descri_2', 'type_B', '2025-03-08 14:00:00', 'pending'),
('Gamma-3', 87, 'descri_3', 'type_C', '2025-12-09 08:00:00', 'done');

CREATE TABLE IF NOT EXISTS sales_sales_targets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    price_lists_id               INT COMMENT 'FK to sales_price_lists' COMMENT 'Ref sales_price_lists',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (price_lists_id) REFERENCES sales_price_lists(id)
) COMMENT='Sales — Sales Targets';
INSERT IGNORE INTO sales_sales_targets (name, price_lists_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 82, 2024, 9402.58, 1092.27, 'active'),
('Beta-2', 61, 2025, 259.97, 8842.62, 'done'),
('Gamma-3', 42, 2026, 9152.49, 2221.24, 'active');

-- === Project Mgmt (25 tables) ===

CREATE TABLE IF NOT EXISTS proj_tasks (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Project Mgmt — Tasks';
INSERT IGNORE INTO proj_tasks (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-08-14 22:00:00', 'pending'),
('Beta-2', 'descri_2', 'type_B', '2025-12-25 13:00:00', 'active'),
('Gamma-3', 'descri_3', 'type_C', '2025-03-16 01:00:00', 'active'),
('Delta-4', 'descri_4', 'type_D', '2025-06-16 15:00:00', 'active');

CREATE TABLE IF NOT EXISTS proj_milestones (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Milestones';
INSERT IGNORE INTO proj_milestones (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 31, 'descri_1', 'type_A', '2025-10-24 04:00:00', 'pending'),
('Beta-2', 59, 'descri_2', 'type_B', '2025-07-22 21:00:00', 'done');

CREATE TABLE IF NOT EXISTS proj_resource_allocations (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Resource Allocations';
INSERT IGNORE INTO proj_resource_allocations (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 70, 'descri_1', 'type_A', '2025-08-19 18:00:00', 'done'),
('Beta-2', 94, 'descri_2', 'type_B', '2025-03-24 13:00:00', 'done'),
('Gamma-3', 13, 'descri_3', 'type_C', '2025-09-27 19:00:00', 'pending'),
('Delta-4', 36, 'descri_4', 'type_D', '2025-01-17 22:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proj_risk_register (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Risk Register';
INSERT IGNORE INTO proj_risk_register (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 57, 'descri_1', 'type_A', '2025-08-04 07:00:00', 'pending'),
('Beta-2', 13, 'descri_2', 'type_B', '2025-12-16 11:00:00', 'done');

CREATE TABLE IF NOT EXISTS proj_project_budgets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Project Budgets';
INSERT IGNORE INTO proj_project_budgets (name, tasks_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 46, 2024, 614.83, 2765.78, 'active'),
('Beta-2', 59, 2025, 925.73, 2128.72, 'done'),
('Gamma-3', 77, 2026, 9710.25, 515.23, 'pending'),
('Delta-4', 32, 2024, 9887.33, 7871.12, 'active');

CREATE TABLE IF NOT EXISTS proj_status_reports (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Status Reports';
INSERT IGNORE INTO proj_status_reports (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 98, 'descri_1', 'type_A', '2025-10-04 06:00:00', 'done'),
('Beta-2', 28, 'descri_2', 'type_B', '2025-04-08 10:00:00', 'active');

CREATE TABLE IF NOT EXISTS proj_project_documents (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Project Documents';
INSERT IGNORE INTO proj_project_documents (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 1, 'descri_1', 'type_A', '2025-05-02 04:00:00', 'active'),
('Beta-2', 70, 'descri_2', 'type_B', '2025-05-17 05:00:00', 'active'),
('Gamma-3', 85, 'descri_3', 'type_C', '2025-01-14 04:00:00', 'active'),
('Delta-4', 46, 'descri_4', 'type_D', '2025-05-10 18:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proj_stakeholders (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Stakeholders';
INSERT IGNORE INTO proj_stakeholders (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 23, 'descri_1', 'type_A', '2025-05-24 01:00:00', 'active'),
('Beta-2', 95, 'descri_2', 'type_B', '2025-08-20 16:00:00', 'active');

CREATE TABLE IF NOT EXISTS proj_change_requests (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Change Requests';
INSERT IGNORE INTO proj_change_requests (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 9, 'descri_1', 'type_A', '2025-09-20 14:00:00', 'pending'),
('Beta-2', 66, 'descri_2', 'type_B', '2025-11-24 03:00:00', 'pending'),
('Gamma-3', 65, 'descri_3', 'type_C', '2025-04-02 19:00:00', 'active'),
('Delta-4', 94, 'descri_4', 'type_D', '2025-12-02 16:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proj_project_phases (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Project Phases';
INSERT IGNORE INTO proj_project_phases (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 83, 'descri_1', 'type_A', '2025-01-16 01:00:00', 'pending'),
('Beta-2', 52, 'descri_2', 'type_B', '2025-08-23 21:00:00', 'active'),
('Gamma-3', 63, 'descri_3', 'type_C', '2025-13-01 14:00:00', 'active');

CREATE TABLE IF NOT EXISTS proj_dependencies (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Dependencies';
INSERT IGNORE INTO proj_dependencies (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 42, 'descri_1', 'type_A', '2025-11-04 04:00:00', 'active'),
('Beta-2', 17, 'descri_2', 'type_B', '2025-05-01 19:00:00', 'done');

CREATE TABLE IF NOT EXISTS proj_issue_log (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Issue Log';
INSERT IGNORE INTO proj_issue_log (name, tasks_id, event_date, description, created_by, status) VALUES
('Alpha-1', 71, '2025-13-01 10:00:00', 'Sample data row 1', 'create_1', 'pending'),
('Beta-2', 77, '2025-10-20 09:00:00', 'Sample data row 2', 'create_2', 'pending'),
('Gamma-3', 65, '2025-11-02 13:00:00', 'Sample data row 3', 'create_3', 'active'),
('Delta-4', 90, '2025-02-03 20:00:00', 'Sample data row 4', 'create_4', 'done');

CREATE TABLE IF NOT EXISTS proj_lessons_learned (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Lessons Learned';
INSERT IGNORE INTO proj_lessons_learned (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 93, 'descri_1', 'type_A', '2025-04-27 13:00:00', 'pending'),
('Beta-2', 30, 'descri_2', 'type_B', '2025-08-16 10:00:00', 'pending'),
('Gamma-3', 52, 'descri_3', 'type_C', '2025-08-17 23:00:00', 'active'),
('Delta-4', 41, 'descri_4', 'type_D', '2025-08-23 10:00:00', 'done');

CREATE TABLE IF NOT EXISTS proj_meeting_minutes (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Meeting Minutes';
INSERT IGNORE INTO proj_meeting_minutes (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 48, 'descri_1', 'type_A', '2025-03-23 21:00:00', 'pending'),
('Beta-2', 9, 'descri_2', 'type_B', '2025-02-19 02:00:00', 'active'),
('Gamma-3', 56, 'descri_3', 'type_C', '2025-02-22 23:00:00', 'done');

CREATE TABLE IF NOT EXISTS proj_deliverables (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Deliverables';
INSERT IGNORE INTO proj_deliverables (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 17, 'descri_1', 'type_A', '2025-10-05 01:00:00', 'done'),
('Beta-2', 72, 'descri_2', 'type_B', '2025-10-08 10:00:00', 'done'),
('Gamma-3', 16, 'descri_3', 'type_C', '2025-08-15 11:00:00', 'done');

CREATE TABLE IF NOT EXISTS proj_work_packages (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Work Packages';
INSERT IGNORE INTO proj_work_packages (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 93, 'descri_1', 'type_A', '2025-01-27 09:00:00', 'done'),
('Beta-2', 40, 'descri_2', 'type_B', '2025-07-13 03:00:00', 'done'),
('Gamma-3', 65, 'descri_3', 'type_C', '2025-04-25 04:00:00', 'done');

CREATE TABLE IF NOT EXISTS proj_gantt_schedules (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Gantt Schedules';
INSERT IGNORE INTO proj_gantt_schedules (name, tasks_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 29, 2024, 8473.43, 3507.29, 'done'),
('Beta-2', 48, 2025, 1157.59, 2792.62, 'active'),
('Gamma-3', 55, 2026, 8452.25, 9744.75, 'done');

CREATE TABLE IF NOT EXISTS proj_project_portfolios (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Project Portfolios';
INSERT IGNORE INTO proj_project_portfolios (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 87, 'descri_1', 'type_A', '2025-11-22 17:00:00', 'active'),
('Beta-2', 78, 'descri_2', 'type_B', '2025-12-01 22:00:00', 'pending'),
('Gamma-3', 4, 'descri_3', 'type_C', '2025-04-09 08:00:00', 'done'),
('Delta-4', 98, 'descri_4', 'type_D', '2025-06-19 10:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proj_program_master (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Program Master';
INSERT IGNORE INTO proj_program_master (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 24, 'descri_1', 'type_A', '2025-03-18 18:00:00', 'done'),
('Beta-2', 52, 'descri_2', 'type_B', '2025-02-08 04:00:00', 'done');

CREATE TABLE IF NOT EXISTS proj_project_templates (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Project Templates';
INSERT IGNORE INTO proj_project_templates (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 4, 'descri_1', 'type_A', '2025-02-19 23:00:00', 'done'),
('Beta-2', 28, 'descri_2', 'type_B', '2025-07-25 13:00:00', 'pending'),
('Gamma-3', 44, 'descri_3', 'type_C', '2025-03-25 11:00:00', 'pending'),
('Delta-4', 93, 'descri_4', 'type_D', '2025-06-27 18:00:00', 'done');

CREATE TABLE IF NOT EXISTS proj_approval_workflows (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Approval Workflows';
INSERT IGNORE INTO proj_approval_workflows (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 7, 'descri_1', 'type_A', '2025-03-24 05:00:00', 'done'),
('Beta-2', 7, 'descri_2', 'type_B', '2025-12-10 02:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proj_time_entries (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Time Entries';
INSERT IGNORE INTO proj_time_entries (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 85, 'descri_1', 'type_A', '2025-08-22 15:00:00', 'done'),
('Beta-2', 57, 'descri_2', 'type_B', '2025-08-17 08:00:00', 'active'),
('Gamma-3', 97, 'descri_3', 'type_C', '2025-09-11 03:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proj_resource_skills (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Resource Skills';
INSERT IGNORE INTO proj_resource_skills (name, tasks_id, description, category, created_at, status) VALUES
('Alpha-1', 15, 'descri_1', 'type_A', '2025-05-06 21:00:00', 'done'),
('Beta-2', 76, 'descri_2', 'type_B', '2025-09-26 16:00:00', 'done'),
('Gamma-3', 40, 'descri_3', 'type_C', '2025-01-24 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proj_project_kpis (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Project Kpis';
INSERT IGNORE INTO proj_project_kpis (name, tasks_id, metric_date, value, target, status) VALUES
('Alpha-1', 8, '2025-01-04', 2051.98, 9452.84, 'active'),
('Beta-2', 98, '2025-05-19', 2901.29, 1208.52, 'pending'),
('Gamma-3', 96, '2025-08-25', 1765.02, 3807.89, 'done'),
('Delta-4', 30, '2025-09-05', 5591.09, 6683.13, 'pending');

CREATE TABLE IF NOT EXISTS proj_sprint_backlogs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tasks_id                     INT COMMENT 'FK to proj_tasks' COMMENT 'Ref proj_tasks',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tasks_id) REFERENCES proj_tasks(id)
) COMMENT='Project Mgmt — Sprint Backlogs';
INSERT IGNORE INTO proj_sprint_backlogs (name, tasks_id, event_date, description, created_by, status) VALUES
('Alpha-1', 51, '2025-01-22 13:00:00', 'Sample data row 1', 'create_1', 'active'),
('Beta-2', 59, '2025-02-12 10:00:00', 'Sample data row 2', 'create_2', 'done');

-- === Quality (25 tables) ===

CREATE TABLE IF NOT EXISTS qa_inspection_plans (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Quality — Inspection Plans';
INSERT IGNORE INTO qa_inspection_plans (name, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 2024, 5737.69, 7097.19, 'pending'),
('Beta-2', 2025, 2901.97, 4055.87, 'pending'),
('Gamma-3', 2026, 1726.84, 9475.72, 'pending');

CREATE TABLE IF NOT EXISTS qa_test_cases (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Test Cases';
INSERT IGNORE INTO qa_test_cases (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 47, 'descri_1', 'type_A', '2025-02-18 13:00:00', 'active'),
('Beta-2', 32, 'descri_2', 'type_B', '2025-08-28 18:00:00', 'pending'),
('Gamma-3', 68, 'descri_3', 'type_C', '2025-02-13 12:00:00', 'pending'),
('Delta-4', 96, 'descri_4', 'type_D', '2025-06-06 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS qa_test_runs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Test Runs';
INSERT IGNORE INTO qa_test_runs (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 10, 'descri_1', 'type_A', '2025-09-10 20:00:00', 'active'),
('Beta-2', 68, 'descri_2', 'type_B', '2025-09-10 06:00:00', 'pending');

CREATE TABLE IF NOT EXISTS qa_defect_reports (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Defect Reports';
INSERT IGNORE INTO qa_defect_reports (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 94, 'descri_1', 'type_A', '2025-12-23 04:00:00', 'active'),
('Beta-2', 14, 'descri_2', 'type_B', '2025-03-19 08:00:00', 'active'),
('Gamma-3', 23, 'descri_3', 'type_C', '2025-11-01 04:00:00', 'done');

CREATE TABLE IF NOT EXISTS qa_corrective_actions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Corrective Actions';
INSERT IGNORE INTO qa_corrective_actions (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 23, 'descri_1', 'type_A', '2025-11-14 15:00:00', 'pending'),
('Beta-2', 97, 'descri_2', 'type_B', '2025-10-09 18:00:00', 'pending');

CREATE TABLE IF NOT EXISTS qa_preventive_actions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Preventive Actions';
INSERT IGNORE INTO qa_preventive_actions (name, inspection_plans_id, event_date, description, created_by, status) VALUES
('Alpha-1', 73, '2025-11-22 20:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 42, '2025-11-14 10:00:00', 'Sample data row 2', 'create_2', 'active'),
('Gamma-3', 57, '2025-02-07 15:00:00', 'Sample data row 3', 'create_3', 'pending'),
('Delta-4', 81, '2025-06-16 08:00:00', 'Sample data row 4', 'create_4', 'done');

CREATE TABLE IF NOT EXISTS qa_audit_schedules (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Audit Schedules';
INSERT IGNORE INTO qa_audit_schedules (name, inspection_plans_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 46, 2024, 5077.79, 3110.66, 'pending'),
('Beta-2', 5, 2025, 578.27, 8317.82, 'active');

CREATE TABLE IF NOT EXISTS qa_audit_findings (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Audit Findings';
INSERT IGNORE INTO qa_audit_findings (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 12, 'descri_1', 'type_A', '2025-11-07 19:00:00', 'done'),
('Beta-2', 50, 'descri_2', 'type_B', '2025-08-13 18:00:00', 'done'),
('Gamma-3', 95, 'descri_3', 'type_C', '2025-01-22 14:00:00', 'done'),
('Delta-4', 84, 'descri_4', 'type_D', '2025-04-13 10:00:00', 'done');

CREATE TABLE IF NOT EXISTS qa_nonconformance_reports (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Nonconformance Reports';
INSERT IGNORE INTO qa_nonconformance_reports (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 65, 'descri_1', 'type_A', '2025-03-22 01:00:00', 'pending'),
('Beta-2', 14, 'descri_2', 'type_B', '2025-06-08 22:00:00', 'active'),
('Gamma-3', 65, 'descri_3', 'type_C', '2025-12-23 05:00:00', 'active');

CREATE TABLE IF NOT EXISTS qa_control_plans (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Control Plans';
INSERT IGNORE INTO qa_control_plans (name, inspection_plans_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 91, 2024, 4382.77, 4399.38, 'done'),
('Beta-2', 79, 2025, 1595.72, 3734.42, 'pending');

CREATE TABLE IF NOT EXISTS qa_measurement_systems (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Measurement Systems';
INSERT IGNORE INTO qa_measurement_systems (name, inspection_plans_id, metric_date, value, target, status) VALUES
('Alpha-1', 53, '2025-06-06', 6791.51, 533.08, 'done'),
('Beta-2', 83, '2025-06-04', 668.39, 954.02, 'done'),
('Gamma-3', 50, '2025-05-06', 2527.19, 8516.30, 'done');

CREATE TABLE IF NOT EXISTS qa_calibration_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Calibration Records';
INSERT IGNORE INTO qa_calibration_records (name, inspection_plans_id, event_date, description, created_by, status) VALUES
('Alpha-1', 43, '2025-02-14 18:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 19, '2025-06-12 09:00:00', 'Sample data row 2', 'create_2', 'done');

CREATE TABLE IF NOT EXISTS qa_spc_data (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Spc Data';
INSERT IGNORE INTO qa_spc_data (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 85, 'descri_1', 'type_A', '2025-07-05 04:00:00', 'done'),
('Beta-2', 91, 'descri_2', 'type_B', '2025-02-16 09:00:00', 'done'),
('Gamma-3', 49, 'descri_3', 'type_C', '2025-11-22 10:00:00', 'active'),
('Delta-4', 86, 'descri_4', 'type_D', '2025-12-24 23:00:00', 'done');

CREATE TABLE IF NOT EXISTS qa_process_capability (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Process Capability';
INSERT IGNORE INTO qa_process_capability (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 12, 'descri_1', 'type_A', '2025-12-23 21:00:00', 'pending'),
('Beta-2', 66, 'descri_2', 'type_B', '2025-07-18 00:00:00', 'pending'),
('Gamma-3', 40, 'descri_3', 'type_C', '2025-04-09 06:00:00', 'pending'),
('Delta-4', 99, 'descri_4', 'type_D', '2025-09-25 06:00:00', 'active');

CREATE TABLE IF NOT EXISTS qa_reliability_tests (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Reliability Tests';
INSERT IGNORE INTO qa_reliability_tests (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 20, 'descri_1', 'type_A', '2025-02-12 09:00:00', 'active'),
('Beta-2', 65, 'descri_2', 'type_B', '2025-10-25 23:00:00', 'done');

CREATE TABLE IF NOT EXISTS qa_environmental_tests (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Environmental Tests';
INSERT IGNORE INTO qa_environmental_tests (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 85, 'descri_1', 'type_A', '2025-06-05 19:00:00', 'active'),
('Beta-2', 77, 'descri_2', 'type_B', '2025-07-25 04:00:00', 'active');

CREATE TABLE IF NOT EXISTS qa_supplier_quality (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Supplier Quality';
INSERT IGNORE INTO qa_supplier_quality (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 89, 'descri_1', 'type_A', '2025-11-12 05:00:00', 'done'),
('Beta-2', 57, 'descri_2', 'type_B', '2025-01-23 13:00:00', 'pending');

CREATE TABLE IF NOT EXISTS qa_incoming_inspections (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Incoming Inspections';
INSERT IGNORE INTO qa_incoming_inspections (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 93, 'descri_1', 'type_A', '2025-05-10 14:00:00', 'done'),
('Beta-2', 37, 'descri_2', 'type_B', '2025-08-06 07:00:00', 'done'),
('Gamma-3', 31, 'descri_3', 'type_C', '2025-06-19 15:00:00', 'active'),
('Delta-4', 48, 'descri_4', 'type_D', '2025-12-12 18:00:00', 'pending');

CREATE TABLE IF NOT EXISTS qa_final_inspections (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Final Inspections';
INSERT IGNORE INTO qa_final_inspections (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 99, 'descri_1', 'type_A', '2025-05-05 12:00:00', 'done'),
('Beta-2', 68, 'descri_2', 'type_B', '2025-08-19 05:00:00', 'active'),
('Gamma-3', 78, 'descri_3', 'type_C', '2025-03-15 08:00:00', 'active');

CREATE TABLE IF NOT EXISTS qa_customer_complaints (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Customer Complaints';
INSERT IGNORE INTO qa_customer_complaints (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 62, 'descri_1', 'type_A', '2025-07-23 17:00:00', 'active'),
('Beta-2', 92, 'descri_2', 'type_B', '2025-09-13 03:00:00', 'pending'),
('Gamma-3', 11, 'descri_3', 'type_C', '2025-03-27 08:00:00', 'pending'),
('Delta-4', 66, 'descri_4', 'type_D', '2025-03-20 13:00:00', 'active');

CREATE TABLE IF NOT EXISTS qa_root_cause_analysis (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Root Cause Analysis';
INSERT IGNORE INTO qa_root_cause_analysis (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 58, 'descri_1', 'type_A', '2025-06-11 00:00:00', 'pending'),
('Beta-2', 7, 'descri_2', 'type_B', '2025-07-07 16:00:00', 'pending');

CREATE TABLE IF NOT EXISTS qa_capa_tracking (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Capa Tracking';
INSERT IGNORE INTO qa_capa_tracking (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 50, 'descri_1', 'type_A', '2025-02-14 11:00:00', 'active'),
('Beta-2', 4, 'descri_2', 'type_B', '2025-06-24 03:00:00', 'done');

CREATE TABLE IF NOT EXISTS qa_quality_costs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Quality Costs';
INSERT IGNORE INTO qa_quality_costs (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 43, 'descri_1', 'type_A', '2025-03-19 04:00:00', 'active'),
('Beta-2', 37, 'descri_2', 'type_B', '2025-09-18 22:00:00', 'active'),
('Gamma-3', 98, 'descri_3', 'type_C', '2025-13-26 15:00:00', 'pending'),
('Delta-4', 79, 'descri_4', 'type_D', '2025-01-03 02:00:00', 'active');

CREATE TABLE IF NOT EXISTS qa_quality_objectives (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Quality Objectives';
INSERT IGNORE INTO qa_quality_objectives (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 28, 'descri_1', 'type_A', '2025-03-21 17:00:00', 'done'),
('Beta-2', 78, 'descri_2', 'type_B', '2025-10-19 13:00:00', 'active'),
('Gamma-3', 100, 'descri_3', 'type_C', '2025-05-08 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS qa_document_controls (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    inspection_plans_id          INT COMMENT 'FK to qa_inspection_plans' COMMENT 'Ref qa_inspection_plans',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (inspection_plans_id) REFERENCES qa_inspection_plans(id)
) COMMENT='Quality — Document Controls';
INSERT IGNORE INTO qa_document_controls (name, inspection_plans_id, description, category, created_at, status) VALUES
('Alpha-1', 7, 'descri_1', 'type_A', '2025-05-11 13:00:00', 'done'),
('Beta-2', 80, 'descri_2', 'type_B', '2025-08-11 02:00:00', 'active');

-- === IT (25 tables) ===

CREATE TABLE IF NOT EXISTS it_it_assets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='IT — It Assets';
INSERT IGNORE INTO it_it_assets (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-11-26 17:00:00', 'active'),
('Beta-2', 'descri_2', 'type_B', '2025-11-16 16:00:00', 'done'),
('Gamma-3', 'descri_3', 'type_C', '2025-05-12 22:00:00', 'active');

CREATE TABLE IF NOT EXISTS it_software_licenses (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Software Licenses';
INSERT IGNORE INTO it_software_licenses (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 55, 'descri_1', 'type_A', '2025-01-01 19:00:00', 'pending'),
('Beta-2', 31, 'descri_2', 'type_B', '2025-10-13 13:00:00', 'active'),
('Gamma-3', 86, 'descri_3', 'type_C', '2025-12-07 02:00:00', 'done');

CREATE TABLE IF NOT EXISTS it_hardware_inventory (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Hardware Inventory';
INSERT IGNORE INTO it_hardware_inventory (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 9, 'descri_1', 'type_A', '2025-09-18 17:00:00', 'done'),
('Beta-2', 65, 'descri_2', 'type_B', '2025-10-04 00:00:00', 'pending'),
('Gamma-3', 61, 'descri_3', 'type_C', '2025-01-23 20:00:00', 'pending');

CREATE TABLE IF NOT EXISTS it_help_desk_tickets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Help Desk Tickets';
INSERT IGNORE INTO it_help_desk_tickets (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 33, 'descri_1', 'type_A', '2025-01-09 11:00:00', 'active'),
('Beta-2', 45, 'descri_2', 'type_B', '2025-05-12 23:00:00', 'done'),
('Gamma-3', 81, 'descri_3', 'type_C', '2025-02-26 18:00:00', 'done');

CREATE TABLE IF NOT EXISTS it_ticket_comments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Ticket Comments';
INSERT IGNORE INTO it_ticket_comments (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 18, 'descri_1', 'type_A', '2025-01-23 11:00:00', 'done'),
('Beta-2', 44, 'descri_2', 'type_B', '2025-11-22 05:00:00', 'done'),
('Gamma-3', 60, 'descri_3', 'type_C', '2025-12-21 15:00:00', 'done');

CREATE TABLE IF NOT EXISTS it_change_requests (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Change Requests';
INSERT IGNORE INTO it_change_requests (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 18, 'descri_1', 'type_A', '2025-02-05 22:00:00', 'pending'),
('Beta-2', 5, 'descri_2', 'type_B', '2025-06-11 06:00:00', 'active');

CREATE TABLE IF NOT EXISTS it_change_approvals (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Change Approvals';
INSERT IGNORE INTO it_change_approvals (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 6, 'descri_1', 'type_A', '2025-06-22 09:00:00', 'done'),
('Beta-2', 51, 'descri_2', 'type_B', '2025-10-27 15:00:00', 'pending');

CREATE TABLE IF NOT EXISTS it_sla_definitions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Sla Definitions';
INSERT IGNORE INTO it_sla_definitions (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 97, 'descri_1', 'type_A', '2025-12-24 06:00:00', 'pending'),
('Beta-2', 46, 'descri_2', 'type_B', '2025-01-25 20:00:00', 'pending');

CREATE TABLE IF NOT EXISTS it_sla_metrics (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Sla Metrics';
INSERT IGNORE INTO it_sla_metrics (name, it_assets_id, metric_date, value, target, status) VALUES
('Alpha-1', 16, '2025-07-21', 4374.71, 4005.82, 'pending'),
('Beta-2', 50, '2025-06-06', 9736.52, 4966.57, 'pending'),
('Gamma-3', 48, '2025-09-14', 2674.27, 835.21, 'pending');

CREATE TABLE IF NOT EXISTS it_network_devices (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Network Devices';
INSERT IGNORE INTO it_network_devices (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 56, 'descri_1', 'type_A', '2025-11-01 05:00:00', 'done'),
('Beta-2', 38, 'descri_2', 'type_B', '2025-06-25 03:00:00', 'active');

CREATE TABLE IF NOT EXISTS it_server_inventory (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Server Inventory';
INSERT IGNORE INTO it_server_inventory (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 85, 'descri_1', 'type_A', '2025-06-12 09:00:00', 'pending'),
('Beta-2', 78, 'descri_2', 'type_B', '2025-08-23 05:00:00', 'done'),
('Gamma-3', 57, 'descri_3', 'type_C', '2025-06-12 14:00:00', 'active');

CREATE TABLE IF NOT EXISTS it_backup_schedules (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Backup Schedules';
INSERT IGNORE INTO it_backup_schedules (name, it_assets_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 46, 2024, 6151.75, 4354.62, 'done'),
('Beta-2', 8, 2025, 759.59, 6378.11, 'pending'),
('Gamma-3', 66, 2026, 8012.15, 6796.63, 'active'),
('Delta-4', 19, 2024, 8487.36, 6783.43, 'pending');

CREATE TABLE IF NOT EXISTS it_backup_logs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Backup Logs';
INSERT IGNORE INTO it_backup_logs (name, it_assets_id, event_date, description, created_by, status) VALUES
('Alpha-1', 17, '2025-02-07 07:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 47, '2025-07-18 12:00:00', 'Sample data row 2', 'create_2', 'done');

CREATE TABLE IF NOT EXISTS it_security_incidents (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Security Incidents';
INSERT IGNORE INTO it_security_incidents (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 78, 'descri_1', 'type_A', '2025-03-23 21:00:00', 'pending'),
('Beta-2', 48, 'descri_2', 'type_B', '2025-07-23 14:00:00', 'active');

CREATE TABLE IF NOT EXISTS it_vulnerability_scans (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Vulnerability Scans';
INSERT IGNORE INTO it_vulnerability_scans (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 18, 'descri_1', 'type_A', '2025-10-20 11:00:00', 'pending'),
('Beta-2', 41, 'descri_2', 'type_B', '2025-12-25 08:00:00', 'active'),
('Gamma-3', 15, 'descri_3', 'type_C', '2025-01-14 23:00:00', 'active'),
('Delta-4', 64, 'descri_4', 'type_D', '2025-09-14 12:00:00', 'done');

CREATE TABLE IF NOT EXISTS it_patch_management (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Patch Management';
INSERT IGNORE INTO it_patch_management (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 34, 'descri_1', 'type_A', '2025-05-22 22:00:00', 'pending'),
('Beta-2', 28, 'descri_2', 'type_B', '2025-11-06 09:00:00', 'done');

CREATE TABLE IF NOT EXISTS it_access_requests (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Access Requests';
INSERT IGNORE INTO it_access_requests (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 26, 'descri_1', 'type_A', '2025-03-07 04:00:00', 'active'),
('Beta-2', 58, 'descri_2', 'type_B', '2025-03-05 22:00:00', 'pending'),
('Gamma-3', 12, 'descri_3', 'type_C', '2025-12-14 10:00:00', 'done');

CREATE TABLE IF NOT EXISTS it_it_projects (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — It Projects';
INSERT IGNORE INTO it_it_projects (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 91, 'descri_1', 'type_A', '2025-02-06 17:00:00', 'done'),
('Beta-2', 38, 'descri_2', 'type_B', '2025-06-14 05:00:00', 'done'),
('Gamma-3', 91, 'descri_3', 'type_C', '2025-12-23 20:00:00', 'active');

CREATE TABLE IF NOT EXISTS it_service_catalog (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Service Catalog';
INSERT IGNORE INTO it_service_catalog (name, it_assets_id, event_date, description, created_by, status) VALUES
('Alpha-1', 66, '2025-04-03 03:00:00', 'Sample data row 1', 'create_1', 'active'),
('Beta-2', 18, '2025-05-10 15:00:00', 'Sample data row 2', 'create_2', 'active'),
('Gamma-3', 47, '2025-10-04 18:00:00', 'Sample data row 3', 'create_3', 'pending');

CREATE TABLE IF NOT EXISTS it_capacity_metrics (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Capacity Metrics';
INSERT IGNORE INTO it_capacity_metrics (name, it_assets_id, metric_date, value, target, status) VALUES
('Alpha-1', 71, '2025-03-11', 6122.75, 871.21, 'pending'),
('Beta-2', 51, '2025-09-22', 5261.17, 7693.18, 'done'),
('Gamma-3', 10, '2025-03-09', 9723.10, 6426.32, 'pending');

CREATE TABLE IF NOT EXISTS it_configuration_items (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Configuration Items';
INSERT IGNORE INTO it_configuration_items (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 88, 'descri_1', 'type_A', '2025-09-13 11:00:00', 'active'),
('Beta-2', 100, 'descri_2', 'type_B', '2025-10-03 20:00:00', 'done'),
('Gamma-3', 24, 'descri_3', 'type_C', '2025-03-11 13:00:00', 'done');

CREATE TABLE IF NOT EXISTS it_release_management (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Release Management';
INSERT IGNORE INTO it_release_management (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 16, 'descri_1', 'type_A', '2025-09-14 04:00:00', 'pending'),
('Beta-2', 22, 'descri_2', 'type_B', '2025-03-27 10:00:00', 'done');

CREATE TABLE IF NOT EXISTS it_incident_escalations (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Incident Escalations';
INSERT IGNORE INTO it_incident_escalations (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 45, 'descri_1', 'type_A', '2025-09-14 09:00:00', 'active'),
('Beta-2', 33, 'descri_2', 'type_B', '2025-04-17 20:00:00', 'done');

CREATE TABLE IF NOT EXISTS it_knowledge_articles (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — Knowledge Articles';
INSERT IGNORE INTO it_knowledge_articles (name, it_assets_id, description, category, created_at, status) VALUES
('Alpha-1', 17, 'descri_1', 'type_A', '2025-11-13 09:00:00', 'done'),
('Beta-2', 69, 'descri_2', 'type_B', '2025-02-20 16:00:00', 'done'),
('Gamma-3', 22, 'descri_3', 'type_C', '2025-11-24 18:00:00', 'active');

CREATE TABLE IF NOT EXISTS it_it_budgets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    it_assets_id                 INT COMMENT 'FK to it_it_assets' COMMENT 'Ref it_it_assets',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (it_assets_id) REFERENCES it_it_assets(id)
) COMMENT='IT — It Budgets';
INSERT IGNORE INTO it_it_budgets (name, it_assets_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 85, 2024, 6247.92, 9020.09, 'pending'),
('Beta-2', 73, 2025, 421.49, 8638.31, 'active');

-- === Legal (20 tables) ===

CREATE TABLE IF NOT EXISTS legal_contracts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Legal — Contracts';
INSERT IGNORE INTO legal_contracts (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-11-21 18:00:00', 'pending'),
('Beta-2', 'descri_2', 'type_B', '2025-12-26 06:00:00', 'done');

CREATE TABLE IF NOT EXISTS legal_contract_amendments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Contract Amendments';
INSERT IGNORE INTO legal_contract_amendments (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 80, 'descri_1', 'type_A', '2025-11-20 00:00:00', 'pending'),
('Beta-2', 81, 'descri_2', 'type_B', '2025-10-28 09:00:00', 'done'),
('Gamma-3', 39, 'descri_3', 'type_C', '2025-09-24 07:00:00', 'done');

CREATE TABLE IF NOT EXISTS legal_legal_cases (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Legal Cases';
INSERT IGNORE INTO legal_legal_cases (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 39, 'descri_1', 'type_A', '2025-08-09 02:00:00', 'done'),
('Beta-2', 8, 'descri_2', 'type_B', '2025-03-25 14:00:00', 'pending'),
('Gamma-3', 62, 'descri_3', 'type_C', '2025-08-14 06:00:00', 'pending');

CREATE TABLE IF NOT EXISTS legal_case_documents (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Case Documents';
INSERT IGNORE INTO legal_case_documents (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 19, 'descri_1', 'type_A', '2025-06-21 22:00:00', 'pending'),
('Beta-2', 94, 'descri_2', 'type_B', '2025-06-09 12:00:00', 'active'),
('Gamma-3', 98, 'descri_3', 'type_C', '2025-07-22 16:00:00', 'done'),
('Delta-4', 14, 'descri_4', 'type_D', '2025-06-24 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS legal_intellectual_property (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Intellectual Property';
INSERT IGNORE INTO legal_intellectual_property (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 35, 'descri_1', 'type_A', '2025-08-07 07:00:00', 'active'),
('Beta-2', 13, 'descri_2', 'type_B', '2025-01-26 09:00:00', 'pending');

CREATE TABLE IF NOT EXISTS legal_trademark_registrations (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Trademark Registrations';
INSERT IGNORE INTO legal_trademark_registrations (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 54, 'descri_1', 'type_A', '2025-05-16 05:00:00', 'pending'),
('Beta-2', 74, 'descri_2', 'type_B', '2025-06-21 06:00:00', 'active'),
('Gamma-3', 64, 'descri_3', 'type_C', '2025-09-12 14:00:00', 'pending'),
('Delta-4', 40, 'descri_4', 'type_D', '2025-09-03 00:00:00', 'active');

CREATE TABLE IF NOT EXISTS legal_patent_filings (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Patent Filings';
INSERT IGNORE INTO legal_patent_filings (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 65, 'descri_1', 'type_A', '2025-08-11 07:00:00', 'active'),
('Beta-2', 75, 'descri_2', 'type_B', '2025-07-13 01:00:00', 'active'),
('Gamma-3', 37, 'descri_3', 'type_C', '2025-09-02 19:00:00', 'done');

CREATE TABLE IF NOT EXISTS legal_regulatory_requirements (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Regulatory Requirements';
INSERT IGNORE INTO legal_regulatory_requirements (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 61, 'descri_1', 'type_A', '2025-05-07 17:00:00', 'active'),
('Beta-2', 14, 'descri_2', 'type_B', '2025-08-25 04:00:00', 'pending'),
('Gamma-3', 94, 'descri_3', 'type_C', '2025-07-20 12:00:00', 'pending'),
('Delta-4', 6, 'descri_4', 'type_D', '2025-07-10 01:00:00', 'done');

CREATE TABLE IF NOT EXISTS legal_compliance_checklist (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Compliance Checklist';
INSERT IGNORE INTO legal_compliance_checklist (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 25, 'descri_1', 'type_A', '2025-07-18 17:00:00', 'pending'),
('Beta-2', 10, 'descri_2', 'type_B', '2025-07-02 16:00:00', 'pending'),
('Gamma-3', 98, 'descri_3', 'type_C', '2025-10-02 08:00:00', 'done'),
('Delta-4', 88, 'descri_4', 'type_D', '2025-11-05 03:00:00', 'active');

CREATE TABLE IF NOT EXISTS legal_compliance_audits (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Compliance Audits';
INSERT IGNORE INTO legal_compliance_audits (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 51, 'descri_1', 'type_A', '2025-07-24 10:00:00', 'done'),
('Beta-2', 47, 'descri_2', 'type_B', '2025-03-18 06:00:00', 'done');

CREATE TABLE IF NOT EXISTS legal_legal_holds (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Legal Holds';
INSERT IGNORE INTO legal_legal_holds (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 52, 'descri_1', 'type_A', '2025-09-05 01:00:00', 'active'),
('Beta-2', 5, 'descri_2', 'type_B', '2025-03-15 22:00:00', 'pending'),
('Gamma-3', 61, 'descri_3', 'type_C', '2025-09-14 14:00:00', 'active'),
('Delta-4', 78, 'descri_4', 'type_D', '2025-09-12 04:00:00', 'pending');

CREATE TABLE IF NOT EXISTS legal_litigation_matters (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Litigation Matters';
INSERT IGNORE INTO legal_litigation_matters (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 41, 'descri_1', 'type_A', '2025-03-28 12:00:00', 'done'),
('Beta-2', 95, 'descri_2', 'type_B', '2025-06-14 18:00:00', 'pending'),
('Gamma-3', 65, 'descri_3', 'type_C', '2025-09-09 17:00:00', 'pending'),
('Delta-4', 91, 'descri_4', 'type_D', '2025-10-09 09:00:00', 'pending');

CREATE TABLE IF NOT EXISTS legal_settlement_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Settlement Records';
INSERT IGNORE INTO legal_settlement_records (name, contracts_id, event_date, description, created_by, status) VALUES
('Alpha-1', 48, '2025-06-02 21:00:00', 'Sample data row 1', 'create_1', 'active'),
('Beta-2', 54, '2025-10-19 09:00:00', 'Sample data row 2', 'create_2', 'done');

CREATE TABLE IF NOT EXISTS legal_legal_fees (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Legal Fees';
INSERT IGNORE INTO legal_legal_fees (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 81, 'descri_1', 'type_A', '2025-01-14 19:00:00', 'pending'),
('Beta-2', 34, 'descri_2', 'type_B', '2025-12-28 18:00:00', 'done'),
('Gamma-3', 30, 'descri_3', 'type_C', '2025-01-27 18:00:00', 'pending'),
('Delta-4', 22, 'descri_4', 'type_D', '2025-09-17 20:00:00', 'done');

CREATE TABLE IF NOT EXISTS legal_attorney_assignments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Attorney Assignments';
INSERT IGNORE INTO legal_attorney_assignments (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 100, 'descri_1', 'type_A', '2025-07-27 04:00:00', 'done'),
('Beta-2', 32, 'descri_2', 'type_B', '2025-01-17 18:00:00', 'done'),
('Gamma-3', 15, 'descri_3', 'type_C', '2025-04-14 00:00:00', 'pending'),
('Delta-4', 41, 'descri_4', 'type_D', '2025-08-19 04:00:00', 'pending');

CREATE TABLE IF NOT EXISTS legal_nda_register (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Nda Register';
INSERT IGNORE INTO legal_nda_register (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 27, 'descri_1', 'type_A', '2025-08-15 16:00:00', 'done'),
('Beta-2', 61, 'descri_2', 'type_B', '2025-02-04 22:00:00', 'active'),
('Gamma-3', 67, 'descri_3', 'type_C', '2025-04-23 17:00:00', 'pending'),
('Delta-4', 85, 'descri_4', 'type_D', '2025-09-21 16:00:00', 'pending');

CREATE TABLE IF NOT EXISTS legal_corporate_filings (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Corporate Filings';
INSERT IGNORE INTO legal_corporate_filings (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 23, 'descri_1', 'type_A', '2025-08-12 17:00:00', 'pending'),
('Beta-2', 70, 'descri_2', 'type_B', '2025-07-14 21:00:00', 'done'),
('Gamma-3', 88, 'descri_3', 'type_C', '2025-11-22 22:00:00', 'pending');

CREATE TABLE IF NOT EXISTS legal_policy_documents (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Policy Documents';
INSERT IGNORE INTO legal_policy_documents (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 62, 'descri_1', 'type_A', '2025-04-15 07:00:00', 'pending'),
('Beta-2', 72, 'descri_2', 'type_B', '2025-06-13 07:00:00', 'pending'),
('Gamma-3', 99, 'descri_3', 'type_C', '2025-05-08 22:00:00', 'active'),
('Delta-4', 89, 'descri_4', 'type_D', '2025-13-25 15:00:00', 'pending');

CREATE TABLE IF NOT EXISTS legal_regulatory_fines (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Regulatory Fines';
INSERT IGNORE INTO legal_regulatory_fines (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 45, 'descri_1', 'type_A', '2025-10-07 23:00:00', 'pending'),
('Beta-2', 37, 'descri_2', 'type_B', '2025-03-07 18:00:00', 'done'),
('Gamma-3', 70, 'descri_3', 'type_C', '2025-07-27 12:00:00', 'pending');

CREATE TABLE IF NOT EXISTS legal_legal_calendar (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    contracts_id                 INT COMMENT 'FK to legal_contracts' COMMENT 'Ref legal_contracts',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (contracts_id) REFERENCES legal_contracts(id)
) COMMENT='Legal — Legal Calendar';
INSERT IGNORE INTO legal_legal_calendar (name, contracts_id, description, category, created_at, status) VALUES
('Alpha-1', 38, 'descri_1', 'type_A', '2025-01-22 09:00:00', 'done'),
('Beta-2', 11, 'descri_2', 'type_B', '2025-06-10 14:00:00', 'done');

-- === Marketing (25 tables) ===

CREATE TABLE IF NOT EXISTS mktg_marketing_campaigns (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Marketing — Marketing Campaigns';
INSERT IGNORE INTO mktg_marketing_campaigns (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-09-22 06:00:00', 'active'),
('Beta-2', 'descri_2', 'type_B', '2025-10-24 08:00:00', 'done'),
('Gamma-3', 'descri_3', 'type_C', '2025-12-21 08:00:00', 'active');

CREATE TABLE IF NOT EXISTS mktg_campaign_budgets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Campaign Budgets';
INSERT IGNORE INTO mktg_campaign_budgets (name, marketing_campaigns_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 79, 2024, 7410.76, 2400.19, 'active'),
('Beta-2', 86, 2025, 9051.88, 2264.23, 'active');

CREATE TABLE IF NOT EXISTS mktg_campaign_performance (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Campaign Performance';
INSERT IGNORE INTO mktg_campaign_performance (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 13, 'descri_1', 'type_A', '2025-08-16 10:00:00', 'done'),
('Beta-2', 61, 'descri_2', 'type_B', '2025-02-24 21:00:00', 'active');

CREATE TABLE IF NOT EXISTS mktg_email_campaigns (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Email Campaigns';
INSERT IGNORE INTO mktg_email_campaigns (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 71, 'descri_1', 'type_A', '2025-03-25 13:00:00', 'done'),
('Beta-2', 61, 'descri_2', 'type_B', '2025-09-21 20:00:00', 'active');

CREATE TABLE IF NOT EXISTS mktg_email_templates (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Email Templates';
INSERT IGNORE INTO mktg_email_templates (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 42, 'descri_1', 'type_A', '2025-05-07 20:00:00', 'active'),
('Beta-2', 99, 'descri_2', 'type_B', '2025-02-18 20:00:00', 'done'),
('Gamma-3', 30, 'descri_3', 'type_C', '2025-10-22 23:00:00', 'done');

CREATE TABLE IF NOT EXISTS mktg_social_media_posts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Social Media Posts';
INSERT IGNORE INTO mktg_social_media_posts (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 23, 'descri_1', 'type_A', '2025-08-18 05:00:00', 'active'),
('Beta-2', 51, 'descri_2', 'type_B', '2025-09-02 05:00:00', 'done');

CREATE TABLE IF NOT EXISTS mktg_social_media_metrics (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Social Media Metrics';
INSERT IGNORE INTO mktg_social_media_metrics (name, marketing_campaigns_id, metric_date, value, target, status) VALUES
('Alpha-1', 5, '2025-01-05', 2990.64, 6031.20, 'pending'),
('Beta-2', 37, '2025-08-09', 9890.11, 5437.03, 'pending'),
('Gamma-3', 18, '2025-09-07', 4685.06, 1935.99, 'active');

CREATE TABLE IF NOT EXISTS mktg_content_calendar (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Content Calendar';
INSERT IGNORE INTO mktg_content_calendar (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 21, 'descri_1', 'type_A', '2025-08-11 20:00:00', 'pending'),
('Beta-2', 92, 'descri_2', 'type_B', '2025-04-12 00:00:00', 'done'),
('Gamma-3', 44, 'descri_3', 'type_C', '2025-06-12 18:00:00', 'done');

CREATE TABLE IF NOT EXISTS mktg_content_assets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Content Assets';
INSERT IGNORE INTO mktg_content_assets (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 23, 'descri_1', 'type_A', '2025-11-05 20:00:00', 'pending'),
('Beta-2', 55, 'descri_2', 'type_B', '2025-09-12 10:00:00', 'active');

CREATE TABLE IF NOT EXISTS mktg_seo_keywords (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Seo Keywords';
INSERT IGNORE INTO mktg_seo_keywords (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 86, 'descri_1', 'type_A', '2025-02-21 05:00:00', 'active'),
('Beta-2', 62, 'descri_2', 'type_B', '2025-06-26 07:00:00', 'active'),
('Gamma-3', 34, 'descri_3', 'type_C', '2025-07-01 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS mktg_seo_rankings (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Seo Rankings';
INSERT IGNORE INTO mktg_seo_rankings (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 43, 'descri_1', 'type_A', '2025-06-15 18:00:00', 'done'),
('Beta-2', 74, 'descri_2', 'type_B', '2025-01-06 08:00:00', 'done'),
('Gamma-3', 47, 'descri_3', 'type_C', '2025-12-19 07:00:00', 'active');

CREATE TABLE IF NOT EXISTS mktg_ppc_campaigns (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Ppc Campaigns';
INSERT IGNORE INTO mktg_ppc_campaigns (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 16, 'descri_1', 'type_A', '2025-08-15 09:00:00', 'active'),
('Beta-2', 52, 'descri_2', 'type_B', '2025-12-16 16:00:00', 'done'),
('Gamma-3', 99, 'descri_3', 'type_C', '2025-06-20 22:00:00', 'active'),
('Delta-4', 82, 'descri_4', 'type_D', '2025-06-12 11:00:00', 'done');

CREATE TABLE IF NOT EXISTS mktg_ppc_ad_groups (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Ppc Ad Groups';
INSERT IGNORE INTO mktg_ppc_ad_groups (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 29, 'descri_1', 'type_A', '2025-03-13 15:00:00', 'active'),
('Beta-2', 59, 'descri_2', 'type_B', '2025-11-03 11:00:00', 'pending');

CREATE TABLE IF NOT EXISTS mktg_landing_pages (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Landing Pages';
INSERT IGNORE INTO mktg_landing_pages (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 71, 'descri_1', 'type_A', '2025-09-18 17:00:00', 'done'),
('Beta-2', 28, 'descri_2', 'type_B', '2025-05-15 21:00:00', 'done'),
('Gamma-3', 11, 'descri_3', 'type_C', '2025-09-18 14:00:00', 'done'),
('Delta-4', 91, 'descri_4', 'type_D', '2025-07-18 02:00:00', 'done');

CREATE TABLE IF NOT EXISTS mktg_conversion_tracking (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Conversion Tracking';
INSERT IGNORE INTO mktg_conversion_tracking (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 8, 'descri_1', 'type_A', '2025-10-01 16:00:00', 'active'),
('Beta-2', 74, 'descri_2', 'type_B', '2025-10-23 04:00:00', 'active');

CREATE TABLE IF NOT EXISTS mktg_webinar_events (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Webinar Events';
INSERT IGNORE INTO mktg_webinar_events (name, marketing_campaigns_id, event_date, description, created_by, status) VALUES
('Alpha-1', 67, '2025-08-03 03:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 27, '2025-10-19 15:00:00', 'Sample data row 2', 'create_2', 'active'),
('Gamma-3', 66, '2025-08-05 01:00:00', 'Sample data row 3', 'create_3', 'pending');

CREATE TABLE IF NOT EXISTS mktg_webinar_registrations (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Webinar Registrations';
INSERT IGNORE INTO mktg_webinar_registrations (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 66, 'descri_1', 'type_A', '2025-08-17 14:00:00', 'done'),
('Beta-2', 8, 'descri_2', 'type_B', '2025-10-07 14:00:00', 'done');

CREATE TABLE IF NOT EXISTS mktg_trade_shows (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Trade Shows';
INSERT IGNORE INTO mktg_trade_shows (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 93, 'descri_1', 'type_A', '2025-01-12 12:00:00', 'pending'),
('Beta-2', 1, 'descri_2', 'type_B', '2025-04-28 18:00:00', 'active'),
('Gamma-3', 6, 'descri_3', 'type_C', '2025-08-21 11:00:00', 'done');

CREATE TABLE IF NOT EXISTS mktg_brand_guidelines (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Brand Guidelines';
INSERT IGNORE INTO mktg_brand_guidelines (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 70, 'descri_1', 'type_A', '2025-02-03 02:00:00', 'pending'),
('Beta-2', 5, 'descri_2', 'type_B', '2025-05-07 13:00:00', 'active');

CREATE TABLE IF NOT EXISTS mktg_market_research (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Market Research';
INSERT IGNORE INTO mktg_market_research (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 99, 'descri_1', 'type_A', '2025-11-21 23:00:00', 'done'),
('Beta-2', 54, 'descri_2', 'type_B', '2025-07-24 12:00:00', 'pending');

CREATE TABLE IF NOT EXISTS mktg_competitor_analysis (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Competitor Analysis';
INSERT IGNORE INTO mktg_competitor_analysis (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 49, 'descri_1', 'type_A', '2025-02-14 21:00:00', 'done'),
('Beta-2', 70, 'descri_2', 'type_B', '2025-03-13 20:00:00', 'pending'),
('Gamma-3', 16, 'descri_3', 'type_C', '2025-04-08 17:00:00', 'pending');

CREATE TABLE IF NOT EXISTS mktg_press_releases (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Press Releases';
INSERT IGNORE INTO mktg_press_releases (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 17, 'descri_1', 'type_A', '2025-04-03 00:00:00', 'active'),
('Beta-2', 39, 'descri_2', 'type_B', '2025-08-14 21:00:00', 'done'),
('Gamma-3', 70, 'descri_3', 'type_C', '2025-08-22 17:00:00', 'pending'),
('Delta-4', 30, 'descri_4', 'type_D', '2025-05-15 14:00:00', 'pending');

CREATE TABLE IF NOT EXISTS mktg_influencer_partnerships (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Influencer Partnerships';
INSERT IGNORE INTO mktg_influencer_partnerships (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 36, 'descri_1', 'type_A', '2025-04-13 23:00:00', 'active'),
('Beta-2', 5, 'descri_2', 'type_B', '2025-12-03 13:00:00', 'done');

CREATE TABLE IF NOT EXISTS mktg_media_buys (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Media Buys';
INSERT IGNORE INTO mktg_media_buys (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 31, 'descri_1', 'type_A', '2025-04-22 02:00:00', 'active'),
('Beta-2', 77, 'descri_2', 'type_B', '2025-01-18 14:00:00', 'done');

CREATE TABLE IF NOT EXISTS mktg_affiliate_programs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    marketing_campaigns_id       INT COMMENT 'FK to mktg_marketing_campaigns' COMMENT 'Ref mktg_marketing_campaigns',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (marketing_campaigns_id) REFERENCES mktg_marketing_campaigns(id)
) COMMENT='Marketing — Affiliate Programs';
INSERT IGNORE INTO mktg_affiliate_programs (name, marketing_campaigns_id, description, category, created_at, status) VALUES
('Alpha-1', 91, 'descri_1', 'type_A', '2025-01-25 07:00:00', 'done'),
('Beta-2', 6, 'descri_2', 'type_B', '2025-07-10 14:00:00', 'active'),
('Gamma-3', 70, 'descri_3', 'type_C', '2025-04-28 01:00:00', 'active'),
('Delta-4', 65, 'descri_4', 'type_D', '2025-05-09 07:00:00', 'done');

-- === Warehouse (25 tables) ===

CREATE TABLE IF NOT EXISTS wms_warehouse_zones (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Warehouse — Warehouse Zones';
INSERT IGNORE INTO wms_warehouse_zones (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-06-24 18:00:00', 'done'),
('Beta-2', 'descri_2', 'type_B', '2025-12-09 10:00:00', 'active'),
('Gamma-3', 'descri_3', 'type_C', '2025-06-15 04:00:00', 'done'),
('Delta-4', 'descri_4', 'type_D', '2025-09-15 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS wms_storage_bins (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Storage Bins';
INSERT IGNORE INTO wms_storage_bins (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 36, 'descri_1', 'type_A', '2025-02-04 17:00:00', 'done'),
('Beta-2', 95, 'descri_2', 'type_B', '2025-03-06 20:00:00', 'done'),
('Gamma-3', 55, 'descri_3', 'type_C', '2025-10-05 15:00:00', 'active');

CREATE TABLE IF NOT EXISTS wms_bin_assignments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Bin Assignments';
INSERT IGNORE INTO wms_bin_assignments (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 83, 'descri_1', 'type_A', '2025-12-08 12:00:00', 'done'),
('Beta-2', 41, 'descri_2', 'type_B', '2025-12-21 13:00:00', 'pending'),
('Gamma-3', 20, 'descri_3', 'type_C', '2025-06-14 12:00:00', 'active');

CREATE TABLE IF NOT EXISTS wms_pick_lists (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Pick Lists';
INSERT IGNORE INTO wms_pick_lists (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 61, 'descri_1', 'type_A', '2025-05-12 07:00:00', 'pending'),
('Beta-2', 91, 'descri_2', 'type_B', '2025-03-18 14:00:00', 'active'),
('Gamma-3', 72, 'descri_3', 'type_C', '2025-08-16 13:00:00', 'done'),
('Delta-4', 68, 'descri_4', 'type_D', '2025-03-13 12:00:00', 'active');

CREATE TABLE IF NOT EXISTS wms_pick_tasks (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Pick Tasks';
INSERT IGNORE INTO wms_pick_tasks (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 27, 'descri_1', 'type_A', '2025-06-02 20:00:00', 'active'),
('Beta-2', 58, 'descri_2', 'type_B', '2025-07-22 02:00:00', 'done'),
('Gamma-3', 93, 'descri_3', 'type_C', '2025-04-14 01:00:00', 'pending');

CREATE TABLE IF NOT EXISTS wms_pack_stations (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Pack Stations';
INSERT IGNORE INTO wms_pack_stations (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 87, 'descri_1', 'type_A', '2025-11-02 19:00:00', 'active'),
('Beta-2', 10, 'descri_2', 'type_B', '2025-04-13 18:00:00', 'done'),
('Gamma-3', 86, 'descri_3', 'type_C', '2025-10-07 06:00:00', 'pending');

CREATE TABLE IF NOT EXISTS wms_packing_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Packing Records';
INSERT IGNORE INTO wms_packing_records (name, warehouse_zones_id, event_date, description, created_by, status) VALUES
('Alpha-1', 43, '2025-06-16 00:00:00', 'Sample data row 1', 'create_1', 'active'),
('Beta-2', 25, '2025-03-05 23:00:00', 'Sample data row 2', 'create_2', 'pending');

CREATE TABLE IF NOT EXISTS wms_receiving_docks (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Receiving Docks';
INSERT IGNORE INTO wms_receiving_docks (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 90, 'descri_1', 'type_A', '2025-11-02 22:00:00', 'active'),
('Beta-2', 51, 'descri_2', 'type_B', '2025-05-11 17:00:00', 'pending');

CREATE TABLE IF NOT EXISTS wms_putaway_tasks (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Putaway Tasks';
INSERT IGNORE INTO wms_putaway_tasks (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 49, 'descri_1', 'type_A', '2025-08-15 17:00:00', 'done'),
('Beta-2', 46, 'descri_2', 'type_B', '2025-06-18 08:00:00', 'pending'),
('Gamma-3', 66, 'descri_3', 'type_C', '2025-09-03 14:00:00', 'active');

CREATE TABLE IF NOT EXISTS wms_cycle_count_plans (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Cycle Count Plans';
INSERT IGNORE INTO wms_cycle_count_plans (name, warehouse_zones_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 61, 2024, 7617.91, 3208.81, 'active'),
('Beta-2', 48, 2025, 3135.78, 466.65, 'active'),
('Gamma-3', 95, 2026, 1467.62, 2614.87, 'done'),
('Delta-4', 75, 2024, 7202.20, 2958.68, 'active');

CREATE TABLE IF NOT EXISTS wms_cycle_count_results (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Cycle Count Results';
INSERT IGNORE INTO wms_cycle_count_results (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 30, 'descri_1', 'type_A', '2025-07-27 18:00:00', 'active'),
('Beta-2', 64, 'descri_2', 'type_B', '2025-10-02 20:00:00', 'done'),
('Gamma-3', 44, 'descri_3', 'type_C', '2025-05-20 15:00:00', 'done');

CREATE TABLE IF NOT EXISTS wms_wave_planning (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Wave Planning';
INSERT IGNORE INTO wms_wave_planning (name, warehouse_zones_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 95, 2024, 4910.56, 1690.49, 'pending'),
('Beta-2', 22, 2025, 1408.05, 5469.45, 'active'),
('Gamma-3', 70, 2026, 9902.34, 9571.84, 'done'),
('Delta-4', 5, 2024, 8391.46, 752.49, 'done');

CREATE TABLE IF NOT EXISTS wms_replenishment_tasks (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Replenishment Tasks';
INSERT IGNORE INTO wms_replenishment_tasks (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 98, 'descri_1', 'type_A', '2025-01-04 13:00:00', 'active'),
('Beta-2', 81, 'descri_2', 'type_B', '2025-04-07 02:00:00', 'done');

CREATE TABLE IF NOT EXISTS wms_cross_dock_orders (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Cross Dock Orders';
INSERT IGNORE INTO wms_cross_dock_orders (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 2, 'descri_1', 'type_A', '2025-04-28 16:00:00', 'pending'),
('Beta-2', 48, 'descri_2', 'type_B', '2025-02-03 19:00:00', 'done');

CREATE TABLE IF NOT EXISTS wms_kitting_orders (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Kitting Orders';
INSERT IGNORE INTO wms_kitting_orders (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 79, 'descri_1', 'type_A', '2025-09-24 21:00:00', 'pending'),
('Beta-2', 3, 'descri_2', 'type_B', '2025-01-04 17:00:00', 'done'),
('Gamma-3', 53, 'descri_3', 'type_C', '2025-01-07 00:00:00', 'done'),
('Delta-4', 93, 'descri_4', 'type_D', '2025-05-01 17:00:00', 'pending');

CREATE TABLE IF NOT EXISTS wms_kit_components (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Kit Components';
INSERT IGNORE INTO wms_kit_components (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 65, 'descri_1', 'type_A', '2025-12-22 21:00:00', 'pending'),
('Beta-2', 23, 'descri_2', 'type_B', '2025-02-27 03:00:00', 'done');

CREATE TABLE IF NOT EXISTS wms_returns_processing (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Returns Processing';
INSERT IGNORE INTO wms_returns_processing (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 31, 'descri_1', 'type_A', '2025-04-15 19:00:00', 'done'),
('Beta-2', 33, 'descri_2', 'type_B', '2025-07-14 08:00:00', 'pending');

CREATE TABLE IF NOT EXISTS wms_quarantine_areas (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Quarantine Areas';
INSERT IGNORE INTO wms_quarantine_areas (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 48, 'descri_1', 'type_A', '2025-07-12 14:00:00', 'done'),
('Beta-2', 32, 'descri_2', 'type_B', '2025-12-21 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS wms_temperature_logs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Temperature Logs';
INSERT IGNORE INTO wms_temperature_logs (name, warehouse_zones_id, event_date, description, created_by, status) VALUES
('Alpha-1', 11, '2025-12-27 20:00:00', 'Sample data row 1', 'create_1', 'active'),
('Beta-2', 12, '2025-07-12 12:00:00', 'Sample data row 2', 'create_2', 'pending'),
('Gamma-3', 71, '2025-09-20 01:00:00', 'Sample data row 3', 'create_3', 'done'),
('Delta-4', 2, '2025-12-24 05:00:00', 'Sample data row 4', 'create_4', 'active');

CREATE TABLE IF NOT EXISTS wms_warehouse_kpis (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Warehouse Kpis';
INSERT IGNORE INTO wms_warehouse_kpis (name, warehouse_zones_id, metric_date, value, target, status) VALUES
('Alpha-1', 56, '2025-12-23', 7845.13, 5662.34, 'active'),
('Beta-2', 68, '2025-01-22', 2305.14, 9012.82, 'done'),
('Gamma-3', 73, '2025-09-20', 2717.97, 9241.43, 'done');

CREATE TABLE IF NOT EXISTS wms_labor_tracking (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Labor Tracking';
INSERT IGNORE INTO wms_labor_tracking (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 70, 'descri_1', 'type_A', '2025-10-09 21:00:00', 'active'),
('Beta-2', 23, 'descri_2', 'type_B', '2025-06-22 00:00:00', 'active'),
('Gamma-3', 76, 'descri_3', 'type_C', '2025-03-18 22:00:00', 'pending');

CREATE TABLE IF NOT EXISTS wms_forklift_assignments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Forklift Assignments';
INSERT IGNORE INTO wms_forklift_assignments (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 39, 'descri_1', 'type_A', '2025-03-28 18:00:00', 'active'),
('Beta-2', 73, 'descri_2', 'type_B', '2025-07-04 21:00:00', 'done');

CREATE TABLE IF NOT EXISTS wms_yard_management (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Yard Management';
INSERT IGNORE INTO wms_yard_management (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 50, 'descri_1', 'type_A', '2025-03-16 22:00:00', 'done'),
('Beta-2', 11, 'descri_2', 'type_B', '2025-09-05 23:00:00', 'pending'),
('Gamma-3', 7, 'descri_3', 'type_C', '2025-02-23 13:00:00', 'active');

CREATE TABLE IF NOT EXISTS wms_dock_schedules (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Dock Schedules';
INSERT IGNORE INTO wms_dock_schedules (name, warehouse_zones_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 44, 2024, 6055.03, 6141.90, 'done'),
('Beta-2', 51, 2025, 9688.55, 3270.18, 'done');

CREATE TABLE IF NOT EXISTS wms_barcode_labels (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    warehouse_zones_id           INT COMMENT 'FK to wms_warehouse_zones' COMMENT 'Ref wms_warehouse_zones',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (warehouse_zones_id) REFERENCES wms_warehouse_zones(id)
) COMMENT='Warehouse — Barcode Labels';
INSERT IGNORE INTO wms_barcode_labels (name, warehouse_zones_id, description, category, created_at, status) VALUES
('Alpha-1', 58, 'descri_1', 'type_A', '2025-09-28 07:00:00', 'pending'),
('Beta-2', 71, 'descri_2', 'type_B', '2025-07-25 13:00:00', 'active'),
('Gamma-3', 88, 'descri_3', 'type_C', '2025-10-20 21:00:00', 'pending');

-- === Fleet (20 tables) ===

CREATE TABLE IF NOT EXISTS fleet_vehicles (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Fleet — Vehicles';
INSERT IGNORE INTO fleet_vehicles (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-11-09 09:00:00', 'active'),
('Beta-2', 'descri_2', 'type_B', '2025-02-10 02:00:00', 'pending');

CREATE TABLE IF NOT EXISTS fleet_vehicle_maintenance (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Vehicle Maintenance';
INSERT IGNORE INTO fleet_vehicle_maintenance (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 49, 'descri_1', 'type_A', '2025-13-01 20:00:00', 'active'),
('Beta-2', 95, 'descri_2', 'type_B', '2025-07-04 10:00:00', 'pending');

CREATE TABLE IF NOT EXISTS fleet_drivers (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Drivers';
INSERT IGNORE INTO fleet_drivers (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 12, 'descri_1', 'type_A', '2025-01-03 09:00:00', 'pending'),
('Beta-2', 47, 'descri_2', 'type_B', '2025-05-27 03:00:00', 'active');

CREATE TABLE IF NOT EXISTS fleet_driver_licenses (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Driver Licenses';
INSERT IGNORE INTO fleet_driver_licenses (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 24, 'descri_1', 'type_A', '2025-08-25 14:00:00', 'done'),
('Beta-2', 72, 'descri_2', 'type_B', '2025-09-11 13:00:00', 'active');

CREATE TABLE IF NOT EXISTS fleet_routes (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Routes';
INSERT IGNORE INTO fleet_routes (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 12, 'descri_1', 'type_A', '2025-07-14 17:00:00', 'active'),
('Beta-2', 77, 'descri_2', 'type_B', '2025-11-27 10:00:00', 'pending');

CREATE TABLE IF NOT EXISTS fleet_route_stops (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Route Stops';
INSERT IGNORE INTO fleet_route_stops (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 38, 'descri_1', 'type_A', '2025-08-16 12:00:00', 'active'),
('Beta-2', 93, 'descri_2', 'type_B', '2025-10-07 07:00:00', 'done');

CREATE TABLE IF NOT EXISTS fleet_fuel_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Fuel Records';
INSERT IGNORE INTO fleet_fuel_records (name, vehicles_id, event_date, description, created_by, status) VALUES
('Alpha-1', 22, '2025-12-15 12:00:00', 'Sample data row 1', 'create_1', 'active'),
('Beta-2', 18, '2025-05-26 09:00:00', 'Sample data row 2', 'create_2', 'pending'),
('Gamma-3', 64, '2025-03-19 02:00:00', 'Sample data row 3', 'create_3', 'active'),
('Delta-4', 56, '2025-05-02 13:00:00', 'Sample data row 4', 'create_4', 'pending');

CREATE TABLE IF NOT EXISTS fleet_fuel_cards (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Fuel Cards';
INSERT IGNORE INTO fleet_fuel_cards (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 10, 'descri_1', 'type_A', '2025-07-17 08:00:00', 'active'),
('Beta-2', 93, 'descri_2', 'type_B', '2025-11-13 15:00:00', 'done'),
('Gamma-3', 79, 'descri_3', 'type_C', '2025-04-17 14:00:00', 'active');

CREATE TABLE IF NOT EXISTS fleet_trip_logs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Trip Logs';
INSERT IGNORE INTO fleet_trip_logs (name, vehicles_id, event_date, description, created_by, status) VALUES
('Alpha-1', 39, '2025-01-04 12:00:00', 'Sample data row 1', 'create_1', 'pending'),
('Beta-2', 80, '2025-07-27 10:00:00', 'Sample data row 2', 'create_2', 'pending');

CREATE TABLE IF NOT EXISTS fleet_vehicle_inspections (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Vehicle Inspections';
INSERT IGNORE INTO fleet_vehicle_inspections (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 56, 'descri_1', 'type_A', '2025-12-26 19:00:00', 'active'),
('Beta-2', 39, 'descri_2', 'type_B', '2025-06-25 19:00:00', 'done'),
('Gamma-3', 26, 'descri_3', 'type_C', '2025-09-22 10:00:00', 'active');

CREATE TABLE IF NOT EXISTS fleet_accident_reports (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Accident Reports';
INSERT IGNORE INTO fleet_accident_reports (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 41, 'descri_1', 'type_A', '2025-05-10 23:00:00', 'done'),
('Beta-2', 82, 'descri_2', 'type_B', '2025-09-28 18:00:00', 'active'),
('Gamma-3', 42, 'descri_3', 'type_C', '2025-07-25 08:00:00', 'pending');

CREATE TABLE IF NOT EXISTS fleet_insurance_policies (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Insurance Policies';
INSERT IGNORE INTO fleet_insurance_policies (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 15, 'descri_1', 'type_A', '2025-10-09 06:00:00', 'done'),
('Beta-2', 70, 'descri_2', 'type_B', '2025-04-09 21:00:00', 'done'),
('Gamma-3', 4, 'descri_3', 'type_C', '2025-08-13 22:00:00', 'active');

CREATE TABLE IF NOT EXISTS fleet_vehicle_assignments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Vehicle Assignments';
INSERT IGNORE INTO fleet_vehicle_assignments (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 38, 'descri_1', 'type_A', '2025-12-19 02:00:00', 'pending'),
('Beta-2', 87, 'descri_2', 'type_B', '2025-09-04 04:00:00', 'done'),
('Gamma-3', 39, 'descri_3', 'type_C', '2025-05-12 08:00:00', 'done');

CREATE TABLE IF NOT EXISTS fleet_toll_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Toll Records';
INSERT IGNORE INTO fleet_toll_records (name, vehicles_id, event_date, description, created_by, status) VALUES
('Alpha-1', 92, '2025-08-21 12:00:00', 'Sample data row 1', 'create_1', 'active'),
('Beta-2', 58, '2025-11-27 15:00:00', 'Sample data row 2', 'create_2', 'done');

CREATE TABLE IF NOT EXISTS fleet_parking_permits (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Parking Permits';
INSERT IGNORE INTO fleet_parking_permits (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 69, 'descri_1', 'type_A', '2025-09-08 22:00:00', 'pending'),
('Beta-2', 70, 'descri_2', 'type_B', '2025-01-19 11:00:00', 'done'),
('Gamma-3', 69, 'descri_3', 'type_C', '2025-11-25 20:00:00', 'active');

CREATE TABLE IF NOT EXISTS fleet_gps_tracking (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Gps Tracking';
INSERT IGNORE INTO fleet_gps_tracking (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 99, 'descri_1', 'type_A', '2025-05-16 21:00:00', 'done'),
('Beta-2', 46, 'descri_2', 'type_B', '2025-03-02 20:00:00', 'done');

CREATE TABLE IF NOT EXISTS fleet_vehicle_leases (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Vehicle Leases';
INSERT IGNORE INTO fleet_vehicle_leases (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 73, 'descri_1', 'type_A', '2025-12-23 21:00:00', 'done'),
('Beta-2', 52, 'descri_2', 'type_B', '2025-06-02 13:00:00', 'active');

CREATE TABLE IF NOT EXISTS fleet_tire_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Tire Records';
INSERT IGNORE INTO fleet_tire_records (name, vehicles_id, event_date, description, created_by, status) VALUES
('Alpha-1', 13, '2025-05-21 07:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 96, '2025-09-14 17:00:00', 'Sample data row 2', 'create_2', 'done');

CREATE TABLE IF NOT EXISTS fleet_emissions_tests (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Emissions Tests';
INSERT IGNORE INTO fleet_emissions_tests (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 74, 'descri_1', 'type_A', '2025-04-02 14:00:00', 'pending'),
('Beta-2', 51, 'descri_2', 'type_B', '2025-08-14 21:00:00', 'done'),
('Gamma-3', 89, 'descri_3', 'type_C', '2025-09-05 04:00:00', 'pending'),
('Delta-4', 4, 'descri_4', 'type_D', '2025-09-24 03:00:00', 'pending');

CREATE TABLE IF NOT EXISTS fleet_fleet_costs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    vehicles_id                  INT COMMENT 'FK to fleet_vehicles' COMMENT 'Ref fleet_vehicles',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (vehicles_id) REFERENCES fleet_vehicles(id)
) COMMENT='Fleet — Fleet Costs';
INSERT IGNORE INTO fleet_fleet_costs (name, vehicles_id, description, category, created_at, status) VALUES
('Alpha-1', 11, 'descri_1', 'type_A', '2025-02-04 23:00:00', 'active'),
('Beta-2', 45, 'descri_2', 'type_B', '2025-06-20 10:00:00', 'pending'),
('Gamma-3', 27, 'descri_3', 'type_C', '2025-09-15 15:00:00', 'pending');

-- === ESG (20 tables) ===

CREATE TABLE IF NOT EXISTS esg_carbon_emissions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='ESG — Carbon Emissions';
INSERT IGNORE INTO esg_carbon_emissions (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-02-22 14:00:00', 'done'),
('Beta-2', 'descri_2', 'type_B', '2025-12-21 14:00:00', 'pending'),
('Gamma-3', 'descri_3', 'type_C', '2025-02-07 09:00:00', 'active');

CREATE TABLE IF NOT EXISTS esg_emission_sources (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Emission Sources';
INSERT IGNORE INTO esg_emission_sources (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 15, 'descri_1', 'type_A', '2025-01-12 10:00:00', 'done'),
('Beta-2', 14, 'descri_2', 'type_B', '2025-12-11 05:00:00', 'done'),
('Gamma-3', 32, 'descri_3', 'type_C', '2025-09-13 05:00:00', 'done'),
('Delta-4', 21, 'descri_4', 'type_D', '2025-06-02 17:00:00', 'pending');

CREATE TABLE IF NOT EXISTS esg_energy_consumption (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Energy Consumption';
INSERT IGNORE INTO esg_energy_consumption (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 30, 'descri_1', 'type_A', '2025-07-12 20:00:00', 'active'),
('Beta-2', 24, 'descri_2', 'type_B', '2025-11-20 21:00:00', 'pending'),
('Gamma-3', 51, 'descri_3', 'type_C', '2025-01-15 23:00:00', 'done');

CREATE TABLE IF NOT EXISTS esg_energy_sources (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Energy Sources';
INSERT IGNORE INTO esg_energy_sources (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 58, 'descri_1', 'type_A', '2025-11-24 13:00:00', 'pending'),
('Beta-2', 1, 'descri_2', 'type_B', '2025-13-25 06:00:00', 'active');

CREATE TABLE IF NOT EXISTS esg_water_usage (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Water Usage';
INSERT IGNORE INTO esg_water_usage (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 97, 'descri_1', 'type_A', '2025-13-25 02:00:00', 'done'),
('Beta-2', 14, 'descri_2', 'type_B', '2025-10-23 05:00:00', 'pending'),
('Gamma-3', 42, 'descri_3', 'type_C', '2025-04-17 14:00:00', 'active');

CREATE TABLE IF NOT EXISTS esg_waste_generation (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Waste Generation';
INSERT IGNORE INTO esg_waste_generation (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 86, 'descri_1', 'type_A', '2025-09-27 16:00:00', 'done'),
('Beta-2', 41, 'descri_2', 'type_B', '2025-11-26 12:00:00', 'done'),
('Gamma-3', 51, 'descri_3', 'type_C', '2025-11-21 03:00:00', 'pending');

CREATE TABLE IF NOT EXISTS esg_waste_disposal (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Waste Disposal';
INSERT IGNORE INTO esg_waste_disposal (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 59, 'descri_1', 'type_A', '2025-11-09 05:00:00', 'done'),
('Beta-2', 91, 'descri_2', 'type_B', '2025-06-13 19:00:00', 'done'),
('Gamma-3', 11, 'descri_3', 'type_C', '2025-12-09 04:00:00', 'pending');

CREATE TABLE IF NOT EXISTS esg_recycling_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Recycling Records';
INSERT IGNORE INTO esg_recycling_records (name, carbon_emissions_id, event_date, description, created_by, status) VALUES
('Alpha-1', 31, '2025-06-17 03:00:00', 'Sample data row 1', 'create_1', 'active'),
('Beta-2', 48, '2025-12-20 04:00:00', 'Sample data row 2', 'create_2', 'done');

CREATE TABLE IF NOT EXISTS esg_sustainability_goals (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Sustainability Goals';
INSERT IGNORE INTO esg_sustainability_goals (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 54, 'descri_1', 'type_A', '2025-11-26 04:00:00', 'done'),
('Beta-2', 50, 'descri_2', 'type_B', '2025-08-21 05:00:00', 'pending'),
('Gamma-3', 82, 'descri_3', 'type_C', '2025-10-24 22:00:00', 'done');

CREATE TABLE IF NOT EXISTS esg_sustainability_metrics (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Sustainability Metrics';
INSERT IGNORE INTO esg_sustainability_metrics (name, carbon_emissions_id, metric_date, value, target, status) VALUES
('Alpha-1', 72, '2025-03-02', 4901.03, 1411.42, 'pending'),
('Beta-2', 58, '2025-11-10', 545.52, 3596.44, 'active');

CREATE TABLE IF NOT EXISTS esg_social_impact_projects (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Social Impact Projects';
INSERT IGNORE INTO esg_social_impact_projects (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 18, 'descri_1', 'type_A', '2025-04-16 12:00:00', 'done'),
('Beta-2', 65, 'descri_2', 'type_B', '2025-12-27 15:00:00', 'pending'),
('Gamma-3', 88, 'descri_3', 'type_C', '2025-09-28 13:00:00', 'done');

CREATE TABLE IF NOT EXISTS esg_community_investments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Community Investments';
INSERT IGNORE INTO esg_community_investments (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 63, 'descri_1', 'type_A', '2025-03-02 02:00:00', 'done'),
('Beta-2', 4, 'descri_2', 'type_B', '2025-04-01 09:00:00', 'active'),
('Gamma-3', 36, 'descri_3', 'type_C', '2025-04-04 17:00:00', 'pending');

CREATE TABLE IF NOT EXISTS esg_diversity_metrics (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Diversity Metrics';
INSERT IGNORE INTO esg_diversity_metrics (name, carbon_emissions_id, metric_date, value, target, status) VALUES
('Alpha-1', 100, '2025-08-11', 5660.28, 7726.98, 'pending'),
('Beta-2', 71, '2025-09-10', 1140.86, 1150.81, 'done');

CREATE TABLE IF NOT EXISTS esg_safety_metrics (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Safety Metrics';
INSERT IGNORE INTO esg_safety_metrics (name, carbon_emissions_id, metric_date, value, target, status) VALUES
('Alpha-1', 70, '2025-01-22', 7601.68, 4418.81, 'active'),
('Beta-2', 55, '2025-02-25', 7361.79, 6539.53, 'active'),
('Gamma-3', 39, '2025-01-17', 4502.38, 3491.94, 'active');

CREATE TABLE IF NOT EXISTS esg_supply_chain_ethics (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Supply Chain Ethics';
INSERT IGNORE INTO esg_supply_chain_ethics (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 16, 'descri_1', 'type_A', '2025-05-10 06:00:00', 'done'),
('Beta-2', 76, 'descri_2', 'type_B', '2025-12-17 11:00:00', 'done'),
('Gamma-3', 79, 'descri_3', 'type_C', '2025-11-16 13:00:00', 'active');

CREATE TABLE IF NOT EXISTS esg_environmental_audits (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Environmental Audits';
INSERT IGNORE INTO esg_environmental_audits (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 18, 'descri_1', 'type_A', '2025-04-22 06:00:00', 'active'),
('Beta-2', 73, 'descri_2', 'type_B', '2025-06-12 17:00:00', 'pending'),
('Gamma-3', 77, 'descri_3', 'type_C', '2025-10-24 05:00:00', 'pending'),
('Delta-4', 91, 'descri_4', 'type_D', '2025-05-10 09:00:00', 'done');

CREATE TABLE IF NOT EXISTS esg_carbon_offsets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Carbon Offsets';
INSERT IGNORE INTO esg_carbon_offsets (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 66, 'descri_1', 'type_A', '2025-12-12 03:00:00', 'active'),
('Beta-2', 97, 'descri_2', 'type_B', '2025-08-15 01:00:00', 'pending'),
('Gamma-3', 84, 'descri_3', 'type_C', '2025-03-09 22:00:00', 'active');

CREATE TABLE IF NOT EXISTS esg_renewable_energy_certs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Renewable Energy Certs';
INSERT IGNORE INTO esg_renewable_energy_certs (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 19, 'descri_1', 'type_A', '2025-13-01 10:00:00', 'active'),
('Beta-2', 98, 'descri_2', 'type_B', '2025-12-11 12:00:00', 'pending');

CREATE TABLE IF NOT EXISTS esg_esg_reports (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Esg Reports';
INSERT IGNORE INTO esg_esg_reports (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 74, 'descri_1', 'type_A', '2025-11-16 08:00:00', 'done'),
('Beta-2', 54, 'descri_2', 'type_B', '2025-07-25 14:00:00', 'active');

CREATE TABLE IF NOT EXISTS esg_stakeholder_engagement (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    carbon_emissions_id          INT COMMENT 'FK to esg_carbon_emissions' COMMENT 'Ref esg_carbon_emissions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (carbon_emissions_id) REFERENCES esg_carbon_emissions(id)
) COMMENT='ESG — Stakeholder Engagement';
INSERT IGNORE INTO esg_stakeholder_engagement (name, carbon_emissions_id, description, category, created_at, status) VALUES
('Alpha-1', 99, 'descri_1', 'type_A', '2025-02-20 12:00:00', 'done'),
('Beta-2', 96, 'descri_2', 'type_B', '2025-05-03 22:00:00', 'pending'),
('Gamma-3', 59, 'descri_3', 'type_C', '2025-09-26 10:00:00', 'done'),
('Delta-4', 1, 'descri_4', 'type_D', '2025-02-20 23:00:00', 'pending');

-- === R&D (20 tables) ===

CREATE TABLE IF NOT EXISTS rnd_research_projects (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='R&D — Research Projects';
INSERT IGNORE INTO rnd_research_projects (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-12-05 22:00:00', 'pending'),
('Beta-2', 'descri_2', 'type_B', '2025-02-05 17:00:00', 'pending'),
('Gamma-3', 'descri_3', 'type_C', '2025-04-28 13:00:00', 'active'),
('Delta-4', 'descri_4', 'type_D', '2025-09-02 08:00:00', 'pending');

CREATE TABLE IF NOT EXISTS rnd_experiments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Experiments';
INSERT IGNORE INTO rnd_experiments (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 44, 'descri_1', 'type_A', '2025-10-01 18:00:00', 'active'),
('Beta-2', 73, 'descri_2', 'type_B', '2025-09-25 10:00:00', 'done'),
('Gamma-3', 98, 'descri_3', 'type_C', '2025-01-25 01:00:00', 'active');

CREATE TABLE IF NOT EXISTS rnd_experiment_results (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Experiment Results';
INSERT IGNORE INTO rnd_experiment_results (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 59, 'descri_1', 'type_A', '2025-01-09 03:00:00', 'active'),
('Beta-2', 57, 'descri_2', 'type_B', '2025-08-10 00:00:00', 'pending'),
('Gamma-3', 26, 'descri_3', 'type_C', '2025-12-17 04:00:00', 'done'),
('Delta-4', 39, 'descri_4', 'type_D', '2025-03-26 08:00:00', 'active');

CREATE TABLE IF NOT EXISTS rnd_prototypes (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Prototypes';
INSERT IGNORE INTO rnd_prototypes (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 47, 'descri_1', 'type_A', '2025-05-17 02:00:00', 'pending'),
('Beta-2', 86, 'descri_2', 'type_B', '2025-12-25 05:00:00', 'active'),
('Gamma-3', 51, 'descri_3', 'type_C', '2025-11-13 09:00:00', 'done'),
('Delta-4', 90, 'descri_4', 'type_D', '2025-04-08 13:00:00', 'done');

CREATE TABLE IF NOT EXISTS rnd_prototype_tests (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Prototype Tests';
INSERT IGNORE INTO rnd_prototype_tests (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 91, 'descri_1', 'type_A', '2025-02-21 00:00:00', 'active'),
('Beta-2', 62, 'descri_2', 'type_B', '2025-02-12 04:00:00', 'done');

CREATE TABLE IF NOT EXISTS rnd_patent_applications (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Patent Applications';
INSERT IGNORE INTO rnd_patent_applications (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 67, 'descri_1', 'type_A', '2025-12-13 14:00:00', 'active'),
('Beta-2', 2, 'descri_2', 'type_B', '2025-12-21 10:00:00', 'active');

CREATE TABLE IF NOT EXISTS rnd_patent_citations (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Patent Citations';
INSERT IGNORE INTO rnd_patent_citations (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 89, 'descri_1', 'type_A', '2025-03-12 15:00:00', 'active'),
('Beta-2', 30, 'descri_2', 'type_B', '2025-07-01 02:00:00', 'done'),
('Gamma-3', 14, 'descri_3', 'type_C', '2025-02-25 10:00:00', 'pending');

CREATE TABLE IF NOT EXISTS rnd_publications (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Publications';
INSERT IGNORE INTO rnd_publications (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 18, 'descri_1', 'type_A', '2025-07-28 04:00:00', 'done'),
('Beta-2', 87, 'descri_2', 'type_B', '2025-03-18 02:00:00', 'done'),
('Gamma-3', 73, 'descri_3', 'type_C', '2025-01-05 19:00:00', 'done');

CREATE TABLE IF NOT EXISTS rnd_research_grants (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Research Grants';
INSERT IGNORE INTO rnd_research_grants (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 57, 'descri_1', 'type_A', '2025-06-12 23:00:00', 'active'),
('Beta-2', 81, 'descri_2', 'type_B', '2025-03-22 13:00:00', 'done');

CREATE TABLE IF NOT EXISTS rnd_grant_milestones (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Grant Milestones';
INSERT IGNORE INTO rnd_grant_milestones (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 57, 'descri_1', 'type_A', '2025-04-28 02:00:00', 'active'),
('Beta-2', 18, 'descri_2', 'type_B', '2025-03-08 18:00:00', 'done'),
('Gamma-3', 50, 'descri_3', 'type_C', '2025-06-12 13:00:00', 'pending'),
('Delta-4', 18, 'descri_4', 'type_D', '2025-05-15 08:00:00', 'done');

CREATE TABLE IF NOT EXISTS rnd_lab_equipment (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Lab Equipment';
INSERT IGNORE INTO rnd_lab_equipment (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 32, 'descri_1', 'type_A', '2025-10-04 19:00:00', 'done'),
('Beta-2', 93, 'descri_2', 'type_B', '2025-11-04 09:00:00', 'done');

CREATE TABLE IF NOT EXISTS rnd_lab_bookings (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Lab Bookings';
INSERT IGNORE INTO rnd_lab_bookings (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 85, 'descri_1', 'type_A', '2025-06-15 06:00:00', 'done'),
('Beta-2', 78, 'descri_2', 'type_B', '2025-09-09 06:00:00', 'done');

CREATE TABLE IF NOT EXISTS rnd_reagent_inventory (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Reagent Inventory';
INSERT IGNORE INTO rnd_reagent_inventory (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 38, 'descri_1', 'type_A', '2025-12-23 01:00:00', 'active'),
('Beta-2', 64, 'descri_2', 'type_B', '2025-07-02 03:00:00', 'active'),
('Gamma-3', 64, 'descri_3', 'type_C', '2025-11-19 19:00:00', 'active');

CREATE TABLE IF NOT EXISTS rnd_clinical_trials (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Clinical Trials';
INSERT IGNORE INTO rnd_clinical_trials (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 2, 'descri_1', 'type_A', '2025-07-17 10:00:00', 'active'),
('Beta-2', 50, 'descri_2', 'type_B', '2025-10-12 13:00:00', 'pending'),
('Gamma-3', 70, 'descri_3', 'type_C', '2025-12-14 05:00:00', 'pending'),
('Delta-4', 99, 'descri_4', 'type_D', '2025-02-10 00:00:00', 'done');

CREATE TABLE IF NOT EXISTS rnd_trial_participants (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Trial Participants';
INSERT IGNORE INTO rnd_trial_participants (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 2, 'descri_1', 'type_A', '2025-05-23 06:00:00', 'active'),
('Beta-2', 8, 'descri_2', 'type_B', '2025-07-08 16:00:00', 'pending');

CREATE TABLE IF NOT EXISTS rnd_innovation_ideas (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Innovation Ideas';
INSERT IGNORE INTO rnd_innovation_ideas (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 92, 'descri_1', 'type_A', '2025-09-05 13:00:00', 'pending'),
('Beta-2', 90, 'descri_2', 'type_B', '2025-07-10 02:00:00', 'done'),
('Gamma-3', 69, 'descri_3', 'type_C', '2025-10-24 19:00:00', 'active'),
('Delta-4', 36, 'descri_4', 'type_D', '2025-02-15 09:00:00', 'active');

CREATE TABLE IF NOT EXISTS rnd_idea_evaluations (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Idea Evaluations';
INSERT IGNORE INTO rnd_idea_evaluations (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 27, 'descri_1', 'type_A', '2025-03-24 17:00:00', 'pending'),
('Beta-2', 51, 'descri_2', 'type_B', '2025-11-21 20:00:00', 'done'),
('Gamma-3', 88, 'descri_3', 'type_C', '2025-11-22 02:00:00', 'pending'),
('Delta-4', 90, 'descri_4', 'type_D', '2025-08-28 23:00:00', 'active');

CREATE TABLE IF NOT EXISTS rnd_technology_roadmaps (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Technology Roadmaps';
INSERT IGNORE INTO rnd_technology_roadmaps (name, research_projects_id, event_date, description, created_by, status) VALUES
('Alpha-1', 32, '2025-02-17 13:00:00', 'Sample data row 1', 'create_1', 'active'),
('Beta-2', 59, '2025-11-06 19:00:00', 'Sample data row 2', 'create_2', 'active');

CREATE TABLE IF NOT EXISTS rnd_research_partnerships (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Research Partnerships';
INSERT IGNORE INTO rnd_research_partnerships (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 86, 'descri_1', 'type_A', '2025-12-01 23:00:00', 'active'),
('Beta-2', 16, 'descri_2', 'type_B', '2025-01-07 22:00:00', 'active'),
('Gamma-3', 90, 'descri_3', 'type_C', '2025-01-06 05:00:00', 'pending');

CREATE TABLE IF NOT EXISTS rnd_technical_reviews (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    research_projects_id         INT COMMENT 'FK to rnd_research_projects' COMMENT 'Ref rnd_research_projects',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (research_projects_id) REFERENCES rnd_research_projects(id)
) COMMENT='R&D — Technical Reviews';
INSERT IGNORE INTO rnd_technical_reviews (name, research_projects_id, description, category, created_at, status) VALUES
('Alpha-1', 67, 'descri_1', 'type_A', '2025-09-13 23:00:00', 'pending'),
('Beta-2', 22, 'descri_2', 'type_B', '2025-07-23 04:00:00', 'done'),
('Gamma-3', 100, 'descri_3', 'type_C', '2025-05-26 23:00:00', 'active');

-- === BI (20 tables) ===

CREATE TABLE IF NOT EXISTS bi_kpi_definitions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='BI — Kpi Definitions';
INSERT IGNORE INTO bi_kpi_definitions (name, metric_date, value, target, status) VALUES
('Alpha-1', '2025-06-04', 8072.20, 2726.81, 'active'),
('Beta-2', '2025-05-23', 7125.40, 5770.17, 'active');

CREATE TABLE IF NOT EXISTS bi_kpi_measurements (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Kpi Measurements';
INSERT IGNORE INTO bi_kpi_measurements (name, kpi_definitions_id, metric_date, value, target, status) VALUES
('Alpha-1', 59, '2025-09-10', 3607.28, 5000.60, 'done'),
('Beta-2', 22, '2025-07-21', 1581.37, 7734.78, 'done'),
('Gamma-3', 87, '2025-02-04', 2291.53, 5092.65, 'active');

CREATE TABLE IF NOT EXISTS bi_dashboard_definitions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Dashboard Definitions';
INSERT IGNORE INTO bi_dashboard_definitions (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 32, 'descri_1', 'type_A', '2025-01-24 15:00:00', 'pending'),
('Beta-2', 50, 'descri_2', 'type_B', '2025-03-21 05:00:00', 'active');

CREATE TABLE IF NOT EXISTS bi_dashboard_widgets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Dashboard Widgets';
INSERT IGNORE INTO bi_dashboard_widgets (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 92, 'descri_1', 'type_A', '2025-12-28 13:00:00', 'active'),
('Beta-2', 42, 'descri_2', 'type_B', '2025-05-16 13:00:00', 'done'),
('Gamma-3', 41, 'descri_3', 'type_C', '2025-05-27 02:00:00', 'done'),
('Delta-4', 48, 'descri_4', 'type_D', '2025-03-06 16:00:00', 'done');

CREATE TABLE IF NOT EXISTS bi_data_sources (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Data Sources';
INSERT IGNORE INTO bi_data_sources (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 23, 'descri_1', 'type_A', '2025-04-03 16:00:00', 'active'),
('Beta-2', 52, 'descri_2', 'type_B', '2025-02-08 14:00:00', 'pending');

CREATE TABLE IF NOT EXISTS bi_etl_jobs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Etl Jobs';
INSERT IGNORE INTO bi_etl_jobs (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 42, 'descri_1', 'type_A', '2025-02-17 17:00:00', 'pending'),
('Beta-2', 2, 'descri_2', 'type_B', '2025-07-21 06:00:00', 'pending'),
('Gamma-3', 73, 'descri_3', 'type_C', '2025-06-16 23:00:00', 'done');

CREATE TABLE IF NOT EXISTS bi_etl_job_runs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Etl Job Runs';
INSERT IGNORE INTO bi_etl_job_runs (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 60, 'descri_1', 'type_A', '2025-07-21 18:00:00', 'pending'),
('Beta-2', 99, 'descri_2', 'type_B', '2025-04-17 23:00:00', 'done');

CREATE TABLE IF NOT EXISTS bi_data_quality_rules (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Data Quality Rules';
INSERT IGNORE INTO bi_data_quality_rules (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 20, 'descri_1', 'type_A', '2025-01-04 13:00:00', 'active'),
('Beta-2', 30, 'descri_2', 'type_B', '2025-10-26 11:00:00', 'done');

CREATE TABLE IF NOT EXISTS bi_data_quality_scores (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Data Quality Scores';
INSERT IGNORE INTO bi_data_quality_scores (name, kpi_definitions_id, metric_date, value, target, status) VALUES
('Alpha-1', 2, '2025-06-04', 24.60, 6575.81, 'done'),
('Beta-2', 94, '2025-06-18', 1048.35, 5302.32, 'active'),
('Gamma-3', 54, '2025-09-28', 9611.28, 9724.89, 'active'),
('Delta-4', 92, '2025-05-04', 939.24, 9212.61, 'done');

CREATE TABLE IF NOT EXISTS bi_report_definitions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Report Definitions';
INSERT IGNORE INTO bi_report_definitions (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 90, 'descri_1', 'type_A', '2025-07-23 14:00:00', 'done'),
('Beta-2', 11, 'descri_2', 'type_B', '2025-10-18 03:00:00', 'done'),
('Gamma-3', 18, 'descri_3', 'type_C', '2025-11-19 12:00:00', 'active');

CREATE TABLE IF NOT EXISTS bi_report_schedules (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Report Schedules';
INSERT IGNORE INTO bi_report_schedules (name, kpi_definitions_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 73, 2024, 632.14, 6627.15, 'active'),
('Beta-2', 31, 2025, 2923.55, 8509.78, 'pending'),
('Gamma-3', 90, 2026, 8319.85, 3268.93, 'pending'),
('Delta-4', 35, 2024, 2343.54, 9435.17, 'active');

CREATE TABLE IF NOT EXISTS bi_report_distributions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Report Distributions';
INSERT IGNORE INTO bi_report_distributions (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 14, 'descri_1', 'type_A', '2025-03-24 03:00:00', 'active'),
('Beta-2', 58, 'descri_2', 'type_B', '2025-08-15 10:00:00', 'pending'),
('Gamma-3', 16, 'descri_3', 'type_C', '2025-10-23 11:00:00', 'active'),
('Delta-4', 58, 'descri_4', 'type_D', '2025-06-18 14:00:00', 'pending');

CREATE TABLE IF NOT EXISTS bi_data_catalogs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Data Catalogs';
INSERT IGNORE INTO bi_data_catalogs (name, kpi_definitions_id, event_date, description, created_by, status) VALUES
('Alpha-1', 12, '2025-03-25 21:00:00', 'Sample data row 1', 'create_1', 'pending'),
('Beta-2', 91, '2025-12-21 19:00:00', 'Sample data row 2', 'create_2', 'active');

CREATE TABLE IF NOT EXISTS bi_data_lineage (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Data Lineage';
INSERT IGNORE INTO bi_data_lineage (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 42, 'descri_1', 'type_A', '2025-03-20 02:00:00', 'done'),
('Beta-2', 32, 'descri_2', 'type_B', '2025-07-15 12:00:00', 'done');

CREATE TABLE IF NOT EXISTS bi_metric_alerts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Metric Alerts';
INSERT IGNORE INTO bi_metric_alerts (name, kpi_definitions_id, metric_date, value, target, status) VALUES
('Alpha-1', 87, '2025-06-14', 2608.81, 9905.87, 'active'),
('Beta-2', 52, '2025-08-08', 5563.42, 5505.16, 'active');

CREATE TABLE IF NOT EXISTS bi_alert_notifications (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Alert Notifications';
INSERT IGNORE INTO bi_alert_notifications (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 14, 'descri_1', 'type_A', '2025-03-16 03:00:00', 'active'),
('Beta-2', 8, 'descri_2', 'type_B', '2025-04-02 04:00:00', 'active'),
('Gamma-3', 52, 'descri_3', 'type_C', '2025-07-23 21:00:00', 'done');

CREATE TABLE IF NOT EXISTS bi_dimension_tables (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Dimension Tables';
INSERT IGNORE INTO bi_dimension_tables (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 11, 'descri_1', 'type_A', '2025-10-19 18:00:00', 'pending'),
('Beta-2', 10, 'descri_2', 'type_B', '2025-01-12 02:00:00', 'active'),
('Gamma-3', 83, 'descri_3', 'type_C', '2025-08-04 04:00:00', 'active'),
('Delta-4', 43, 'descri_4', 'type_D', '2025-03-07 01:00:00', 'pending');

CREATE TABLE IF NOT EXISTS bi_fact_tables (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Fact Tables';
INSERT IGNORE INTO bi_fact_tables (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 22, 'descri_1', 'type_A', '2025-10-14 13:00:00', 'done'),
('Beta-2', 51, 'descri_2', 'type_B', '2025-09-03 00:00:00', 'pending');

CREATE TABLE IF NOT EXISTS bi_cube_definitions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Cube Definitions';
INSERT IGNORE INTO bi_cube_definitions (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 55, 'descri_1', 'type_A', '2025-03-05 11:00:00', 'active'),
('Beta-2', 24, 'descri_2', 'type_B', '2025-05-01 08:00:00', 'pending'),
('Gamma-3', 20, 'descri_3', 'type_C', '2025-01-18 19:00:00', 'done'),
('Delta-4', 79, 'descri_4', 'type_D', '2025-05-14 20:00:00', 'pending');

CREATE TABLE IF NOT EXISTS bi_data_governance_policies (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    kpi_definitions_id           INT COMMENT 'FK to bi_kpi_definitions' COMMENT 'Ref bi_kpi_definitions',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (kpi_definitions_id) REFERENCES bi_kpi_definitions(id)
) COMMENT='BI — Data Governance Policies';
INSERT IGNORE INTO bi_data_governance_policies (name, kpi_definitions_id, description, category, created_at, status) VALUES
('Alpha-1', 53, 'descri_1', 'type_A', '2025-10-02 15:00:00', 'active'),
('Beta-2', 12, 'descri_2', 'type_B', '2025-05-04 12:00:00', 'active'),
('Gamma-3', 54, 'descri_3', 'type_C', '2025-04-19 20:00:00', 'done');

-- === Accounting (20 tables) ===

CREATE TABLE IF NOT EXISTS acct_tax_filings (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Accounting — Tax Filings';
INSERT IGNORE INTO acct_tax_filings (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-11-15 17:00:00', 'active'),
('Beta-2', 'descri_2', 'type_B', '2025-07-28 22:00:00', 'pending');

CREATE TABLE IF NOT EXISTS acct_tax_schedules (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Tax Schedules';
INSERT IGNORE INTO acct_tax_schedules (name, tax_filings_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 70, 2024, 7815.76, 3446.56, 'done'),
('Beta-2', 65, 2025, 3222.51, 2711.38, 'active'),
('Gamma-3', 41, 2026, 5974.60, 302.71, 'pending');

CREATE TABLE IF NOT EXISTS acct_withholding_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Withholding Records';
INSERT IGNORE INTO acct_withholding_records (name, tax_filings_id, event_date, description, created_by, status) VALUES
('Alpha-1', 61, '2025-10-20 11:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 30, '2025-03-26 03:00:00', 'Sample data row 2', 'create_2', 'active');

CREATE TABLE IF NOT EXISTS acct_expense_categories (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Expense Categories';
INSERT IGNORE INTO acct_expense_categories (name, tax_filings_id, description, category, created_at, status) VALUES
('Alpha-1', 31, 'descri_1', 'type_A', '2025-05-25 17:00:00', 'done'),
('Beta-2', 8, 'descri_2', 'type_B', '2025-04-01 18:00:00', 'pending'),
('Gamma-3', 46, 'descri_3', 'type_C', '2025-03-05 05:00:00', 'active'),
('Delta-4', 76, 'descri_4', 'type_D', '2025-06-23 23:00:00', 'done');

CREATE TABLE IF NOT EXISTS acct_expense_policies (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Expense Policies';
INSERT IGNORE INTO acct_expense_policies (name, tax_filings_id, description, category, created_at, status) VALUES
('Alpha-1', 76, 'descri_1', 'type_A', '2025-01-15 22:00:00', 'done'),
('Beta-2', 46, 'descri_2', 'type_B', '2025-10-11 18:00:00', 'active'),
('Gamma-3', 73, 'descri_3', 'type_C', '2025-04-13 15:00:00', 'done');

CREATE TABLE IF NOT EXISTS acct_travel_requests (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Travel Requests';
INSERT IGNORE INTO acct_travel_requests (name, tax_filings_id, description, category, created_at, status) VALUES
('Alpha-1', 23, 'descri_1', 'type_A', '2025-09-28 01:00:00', 'active'),
('Beta-2', 8, 'descri_2', 'type_B', '2025-04-08 19:00:00', 'active'),
('Gamma-3', 3, 'descri_3', 'type_C', '2025-09-18 15:00:00', 'active');

CREATE TABLE IF NOT EXISTS acct_travel_itineraries (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Travel Itineraries';
INSERT IGNORE INTO acct_travel_itineraries (name, tax_filings_id, description, category, created_at, status) VALUES
('Alpha-1', 79, 'descri_1', 'type_A', '2025-04-20 04:00:00', 'pending'),
('Beta-2', 92, 'descri_2', 'type_B', '2025-04-07 10:00:00', 'active'),
('Gamma-3', 3, 'descri_3', 'type_C', '2025-03-20 18:00:00', 'done');

CREATE TABLE IF NOT EXISTS acct_mileage_claims (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Mileage Claims';
INSERT IGNORE INTO acct_mileage_claims (name, tax_filings_id, description, category, created_at, status) VALUES
('Alpha-1', 100, 'descri_1', 'type_A', '2025-02-01 16:00:00', 'pending'),
('Beta-2', 10, 'descri_2', 'type_B', '2025-07-24 22:00:00', 'done');

CREATE TABLE IF NOT EXISTS acct_petty_cash_funds (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Petty Cash Funds';
INSERT IGNORE INTO acct_petty_cash_funds (name, tax_filings_id, description, category, created_at, status) VALUES
('Alpha-1', 76, 'descri_1', 'type_A', '2025-02-24 10:00:00', 'pending'),
('Beta-2', 42, 'descri_2', 'type_B', '2025-03-14 05:00:00', 'done'),
('Gamma-3', 56, 'descri_3', 'type_C', '2025-11-20 15:00:00', 'done');

CREATE TABLE IF NOT EXISTS acct_petty_cash_transactions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Petty Cash Transactions';
INSERT IGNORE INTO acct_petty_cash_transactions (name, tax_filings_id, event_date, description, created_by, status) VALUES
('Alpha-1', 23, '2025-13-27 17:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 79, '2025-07-15 07:00:00', 'Sample data row 2', 'create_2', 'done'),
('Gamma-3', 76, '2025-03-06 12:00:00', 'Sample data row 3', 'create_3', 'pending');

CREATE TABLE IF NOT EXISTS acct_asset_disposals (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Asset Disposals';
INSERT IGNORE INTO acct_asset_disposals (name, tax_filings_id, description, category, created_at, status) VALUES
('Alpha-1', 89, 'descri_1', 'type_A', '2025-06-11 04:00:00', 'active'),
('Beta-2', 94, 'descri_2', 'type_B', '2025-01-01 22:00:00', 'done'),
('Gamma-3', 51, 'descri_3', 'type_C', '2025-10-11 01:00:00', 'active'),
('Delta-4', 78, 'descri_4', 'type_D', '2025-06-23 19:00:00', 'active');

CREATE TABLE IF NOT EXISTS acct_lease_agreements (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Lease Agreements';
INSERT IGNORE INTO acct_lease_agreements (name, tax_filings_id, description, category, created_at, status) VALUES
('Alpha-1', 73, 'descri_1', 'type_A', '2025-02-26 15:00:00', 'active'),
('Beta-2', 43, 'descri_2', 'type_B', '2025-02-12 07:00:00', 'pending'),
('Gamma-3', 41, 'descri_3', 'type_C', '2025-03-02 20:00:00', 'active'),
('Delta-4', 91, 'descri_4', 'type_D', '2025-12-08 20:00:00', 'done');

CREATE TABLE IF NOT EXISTS acct_lease_payments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Lease Payments';
INSERT IGNORE INTO acct_lease_payments (name, tax_filings_id, description, category, created_at, status) VALUES
('Alpha-1', 58, 'descri_1', 'type_A', '2025-01-06 08:00:00', 'active'),
('Beta-2', 32, 'descri_2', 'type_B', '2025-12-19 02:00:00', 'pending'),
('Gamma-3', 33, 'descri_3', 'type_C', '2025-02-28 23:00:00', 'active');

CREATE TABLE IF NOT EXISTS acct_loan_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Loan Records';
INSERT IGNORE INTO acct_loan_records (name, tax_filings_id, event_date, description, created_by, status) VALUES
('Alpha-1', 50, '2025-08-01 23:00:00', 'Sample data row 1', 'create_1', 'pending'),
('Beta-2', 22, '2025-08-16 15:00:00', 'Sample data row 2', 'create_2', 'pending');

CREATE TABLE IF NOT EXISTS acct_loan_payments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Loan Payments';
INSERT IGNORE INTO acct_loan_payments (name, tax_filings_id, description, category, created_at, status) VALUES
('Alpha-1', 70, 'descri_1', 'type_A', '2025-07-25 03:00:00', 'pending'),
('Beta-2', 74, 'descri_2', 'type_B', '2025-12-26 21:00:00', 'active'),
('Gamma-3', 21, 'descri_3', 'type_C', '2025-08-08 02:00:00', 'active');

CREATE TABLE IF NOT EXISTS acct_investment_portfolios (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Investment Portfolios';
INSERT IGNORE INTO acct_investment_portfolios (name, tax_filings_id, description, category, created_at, status) VALUES
('Alpha-1', 3, 'descri_1', 'type_A', '2025-06-24 08:00:00', 'active'),
('Beta-2', 10, 'descri_2', 'type_B', '2025-06-08 05:00:00', 'pending'),
('Gamma-3', 21, 'descri_3', 'type_C', '2025-02-11 17:00:00', 'active');

CREATE TABLE IF NOT EXISTS acct_investment_transactions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Investment Transactions';
INSERT IGNORE INTO acct_investment_transactions (name, tax_filings_id, event_date, description, created_by, status) VALUES
('Alpha-1', 77, '2025-11-07 15:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 4, '2025-08-25 20:00:00', 'Sample data row 2', 'create_2', 'active'),
('Gamma-3', 78, '2025-08-27 05:00:00', 'Sample data row 3', 'create_3', 'active');

CREATE TABLE IF NOT EXISTS acct_dividend_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Dividend Records';
INSERT IGNORE INTO acct_dividend_records (name, tax_filings_id, event_date, description, created_by, status) VALUES
('Alpha-1', 43, '2025-04-22 06:00:00', 'Sample data row 1', 'create_1', 'pending'),
('Beta-2', 90, '2025-10-04 23:00:00', 'Sample data row 2', 'create_2', 'done');

CREATE TABLE IF NOT EXISTS acct_insurance_premiums (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Insurance Premiums';
INSERT IGNORE INTO acct_insurance_premiums (name, tax_filings_id, description, category, created_at, status) VALUES
('Alpha-1', 34, 'descri_1', 'type_A', '2025-12-03 09:00:00', 'pending'),
('Beta-2', 31, 'descri_2', 'type_B', '2025-02-22 01:00:00', 'pending'),
('Gamma-3', 74, 'descri_3', 'type_C', '2025-10-02 15:00:00', 'active'),
('Delta-4', 7, 'descri_4', 'type_D', '2025-07-17 00:00:00', 'pending');

CREATE TABLE IF NOT EXISTS acct_insurance_claims (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    tax_filings_id               INT COMMENT 'FK to acct_tax_filings' COMMENT 'Ref acct_tax_filings',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (tax_filings_id) REFERENCES acct_tax_filings(id)
) COMMENT='Accounting — Insurance Claims';
INSERT IGNORE INTO acct_insurance_claims (name, tax_filings_id, description, category, created_at, status) VALUES
('Alpha-1', 38, 'descri_1', 'type_A', '2025-12-03 20:00:00', 'done'),
('Beta-2', 62, 'descri_2', 'type_B', '2025-04-19 03:00:00', 'active');

-- === Procurement (20 tables) ===

CREATE TABLE IF NOT EXISTS proc_rfq_documents (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Procurement — Rfq Documents';
INSERT IGNORE INTO proc_rfq_documents (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-03-04 20:00:00', 'pending'),
('Beta-2', 'descri_2', 'type_B', '2025-02-14 15:00:00', 'active');

CREATE TABLE IF NOT EXISTS proc_rfq_responses (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Rfq Responses';
INSERT IGNORE INTO proc_rfq_responses (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 51, 'descri_1', 'type_A', '2025-09-18 15:00:00', 'done'),
('Beta-2', 34, 'descri_2', 'type_B', '2025-02-19 20:00:00', 'done'),
('Gamma-3', 50, 'descri_3', 'type_C', '2025-04-11 11:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proc_bid_evaluations (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Bid Evaluations';
INSERT IGNORE INTO proc_bid_evaluations (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 24, 'descri_1', 'type_A', '2025-08-05 01:00:00', 'pending'),
('Beta-2', 57, 'descri_2', 'type_B', '2025-08-15 08:00:00', 'active'),
('Gamma-3', 35, 'descri_3', 'type_C', '2025-10-09 01:00:00', 'active');

CREATE TABLE IF NOT EXISTS proc_framework_agreements (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Framework Agreements';
INSERT IGNORE INTO proc_framework_agreements (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 86, 'descri_1', 'type_A', '2025-02-22 02:00:00', 'done'),
('Beta-2', 45, 'descri_2', 'type_B', '2025-10-26 13:00:00', 'done'),
('Gamma-3', 30, 'descri_3', 'type_C', '2025-10-04 01:00:00', 'pending'),
('Delta-4', 67, 'descri_4', 'type_D', '2025-08-19 17:00:00', 'done');

CREATE TABLE IF NOT EXISTS proc_blanket_orders (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Blanket Orders';
INSERT IGNORE INTO proc_blanket_orders (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 74, 'descri_1', 'type_A', '2025-05-09 15:00:00', 'pending'),
('Beta-2', 11, 'descri_2', 'type_B', '2025-07-07 22:00:00', 'active'),
('Gamma-3', 65, 'descri_3', 'type_C', '2025-10-13 16:00:00', 'done');

CREATE TABLE IF NOT EXISTS proc_blanket_releases (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Blanket Releases';
INSERT IGNORE INTO proc_blanket_releases (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 80, 'descri_1', 'type_A', '2025-03-20 03:00:00', 'pending'),
('Beta-2', 23, 'descri_2', 'type_B', '2025-03-02 06:00:00', 'active'),
('Gamma-3', 17, 'descri_3', 'type_C', '2025-01-22 13:00:00', 'active'),
('Delta-4', 88, 'descri_4', 'type_D', '2025-08-27 06:00:00', 'done');

CREATE TABLE IF NOT EXISTS proc_catalog_items (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Catalog Items';
INSERT IGNORE INTO proc_catalog_items (name, rfq_documents_id, event_date, description, created_by, status) VALUES
('Alpha-1', 77, '2025-05-21 10:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 9, '2025-02-17 12:00:00', 'Sample data row 2', 'create_2', 'done');

CREATE TABLE IF NOT EXISTS proc_catalog_prices (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Catalog Prices';
INSERT IGNORE INTO proc_catalog_prices (name, rfq_documents_id, event_date, description, created_by, status) VALUES
('Alpha-1', 71, '2025-06-27 08:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 59, '2025-01-07 22:00:00', 'Sample data row 2', 'create_2', 'done'),
('Gamma-3', 76, '2025-09-16 13:00:00', 'Sample data row 3', 'create_3', 'active');

CREATE TABLE IF NOT EXISTS proc_vendor_onboarding (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Vendor Onboarding';
INSERT IGNORE INTO proc_vendor_onboarding (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 20, 'descri_1', 'type_A', '2025-03-20 18:00:00', 'done'),
('Beta-2', 75, 'descri_2', 'type_B', '2025-02-24 03:00:00', 'done'),
('Gamma-3', 100, 'descri_3', 'type_C', '2025-02-24 09:00:00', 'done');

CREATE TABLE IF NOT EXISTS proc_vendor_documents (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Vendor Documents';
INSERT IGNORE INTO proc_vendor_documents (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 53, 'descri_1', 'type_A', '2025-05-24 12:00:00', 'done'),
('Beta-2', 63, 'descri_2', 'type_B', '2025-10-13 19:00:00', 'pending'),
('Gamma-3', 5, 'descri_3', 'type_C', '2025-03-05 08:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proc_vendor_contacts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Vendor Contacts';
INSERT IGNORE INTO proc_vendor_contacts (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 80, 'descri_1', 'type_A', '2025-11-04 21:00:00', 'done'),
('Beta-2', 52, 'descri_2', 'type_B', '2025-01-20 12:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proc_vendor_performance_kpis (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Vendor Performance Kpis';
INSERT IGNORE INTO proc_vendor_performance_kpis (name, rfq_documents_id, metric_date, value, target, status) VALUES
('Alpha-1', 31, '2025-01-26', 7459.29, 9208.22, 'pending'),
('Beta-2', 3, '2025-06-05', 8174.87, 3037.22, 'pending'),
('Gamma-3', 63, '2025-12-21', 6850.99, 9724.27, 'active'),
('Delta-4', 18, '2025-06-14', 9741.48, 4436.23, 'pending');

CREATE TABLE IF NOT EXISTS proc_sourcing_events (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Sourcing Events';
INSERT IGNORE INTO proc_sourcing_events (name, rfq_documents_id, event_date, description, created_by, status) VALUES
('Alpha-1', 94, '2025-08-18 19:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 12, '2025-04-13 14:00:00', 'Sample data row 2', 'create_2', 'active'),
('Gamma-3', 53, '2025-09-25 16:00:00', 'Sample data row 3', 'create_3', 'pending');

CREATE TABLE IF NOT EXISTS proc_sourcing_awards (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Sourcing Awards';
INSERT IGNORE INTO proc_sourcing_awards (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 66, 'descri_1', 'type_A', '2025-03-26 02:00:00', 'pending'),
('Beta-2', 91, 'descri_2', 'type_B', '2025-09-08 12:00:00', 'active');

CREATE TABLE IF NOT EXISTS proc_contract_milestones (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Contract Milestones';
INSERT IGNORE INTO proc_contract_milestones (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 73, 'descri_1', 'type_A', '2025-01-15 05:00:00', 'active'),
('Beta-2', 26, 'descri_2', 'type_B', '2025-01-02 07:00:00', 'active'),
('Gamma-3', 59, 'descri_3', 'type_C', '2025-01-26 11:00:00', 'done'),
('Delta-4', 26, 'descri_4', 'type_D', '2025-05-01 11:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proc_contract_deliverables (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Contract Deliverables';
INSERT IGNORE INTO proc_contract_deliverables (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 51, 'descri_1', 'type_A', '2025-11-18 03:00:00', 'done'),
('Beta-2', 4, 'descri_2', 'type_B', '2025-05-12 11:00:00', 'pending'),
('Gamma-3', 77, 'descri_3', 'type_C', '2025-08-05 05:00:00', 'pending'),
('Delta-4', 76, 'descri_4', 'type_D', '2025-10-03 11:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proc_spend_analysis (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Spend Analysis';
INSERT IGNORE INTO proc_spend_analysis (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 34, 'descri_1', 'type_A', '2025-12-17 02:00:00', 'pending'),
('Beta-2', 4, 'descri_2', 'type_B', '2025-07-02 01:00:00', 'active');

CREATE TABLE IF NOT EXISTS proc_commodity_codes (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Commodity Codes';
INSERT IGNORE INTO proc_commodity_codes (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 100, 'descri_1', 'type_A', '2025-04-26 07:00:00', 'done'),
('Beta-2', 29, 'descri_2', 'type_B', '2025-12-11 06:00:00', 'pending'),
('Gamma-3', 84, 'descri_3', 'type_C', '2025-07-14 16:00:00', 'active'),
('Delta-4', 100, 'descri_4', 'type_D', '2025-12-14 00:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proc_approved_vendor_list (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Approved Vendor List';
INSERT IGNORE INTO proc_approved_vendor_list (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 83, 'descri_1', 'type_A', '2025-03-05 19:00:00', 'active'),
('Beta-2', 29, 'descri_2', 'type_B', '2025-05-18 19:00:00', 'pending');

CREATE TABLE IF NOT EXISTS proc_purchase_req_lines (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    rfq_documents_id             INT COMMENT 'FK to proc_rfq_documents' COMMENT 'Ref proc_rfq_documents',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (rfq_documents_id) REFERENCES proc_rfq_documents(id)
) COMMENT='Procurement — Purchase Req Lines';
INSERT IGNORE INTO proc_purchase_req_lines (name, rfq_documents_id, description, category, created_at, status) VALUES
('Alpha-1', 91, 'descri_1', 'type_A', '2025-11-22 08:00:00', 'pending'),
('Beta-2', 49, 'descri_2', 'type_B', '2025-06-11 14:00:00', 'pending'),
('Gamma-3', 28, 'descri_3', 'type_C', '2025-08-15 09:00:00', 'done'),
('Delta-4', 67, 'descri_4', 'type_D', '2025-11-12 12:00:00', 'done');

-- === Customer Success (20 tables) ===

CREATE TABLE IF NOT EXISTS cust_customer_tiers (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Customer Success — Customer Tiers';
INSERT IGNORE INTO cust_customer_tiers (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-01-04 16:00:00', 'done'),
('Beta-2', 'descri_2', 'type_B', '2025-07-24 17:00:00', 'done');

CREATE TABLE IF NOT EXISTS cust_tier_benefits (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Tier Benefits';
INSERT IGNORE INTO cust_tier_benefits (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 78, 'descri_1', 'type_A', '2025-05-05 09:00:00', 'done'),
('Beta-2', 14, 'descri_2', 'type_B', '2025-09-23 02:00:00', 'pending'),
('Gamma-3', 36, 'descri_3', 'type_C', '2025-11-20 10:00:00', 'pending'),
('Delta-4', 34, 'descri_4', 'type_D', '2025-12-28 22:00:00', 'done');

CREATE TABLE IF NOT EXISTS cust_customer_preferences (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Customer Preferences';
INSERT IGNORE INTO cust_customer_preferences (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 25, 'descri_1', 'type_A', '2025-03-21 16:00:00', 'active'),
('Beta-2', 8, 'descri_2', 'type_B', '2025-11-28 12:00:00', 'done'),
('Gamma-3', 41, 'descri_3', 'type_C', '2025-12-13 04:00:00', 'done');

CREATE TABLE IF NOT EXISTS cust_comm_preferences (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Comm Preferences';
INSERT IGNORE INTO cust_comm_preferences (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 84, 'descri_1', 'type_A', '2025-09-03 09:00:00', 'pending'),
('Beta-2', 54, 'descri_2', 'type_B', '2025-07-12 12:00:00', 'done');

CREATE TABLE IF NOT EXISTS cust_subscription_plans (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Subscription Plans';
INSERT IGNORE INTO cust_subscription_plans (name, customer_tiers_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 75, 2024, 7082.90, 5724.13, 'active'),
('Beta-2', 44, 2025, 7267.51, 2231.19, 'done');

CREATE TABLE IF NOT EXISTS cust_subscription_billing (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Subscription Billing';
INSERT IGNORE INTO cust_subscription_billing (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 61, 'descri_1', 'type_A', '2025-12-08 11:00:00', 'done'),
('Beta-2', 39, 'descri_2', 'type_B', '2025-03-04 21:00:00', 'done'),
('Gamma-3', 24, 'descri_3', 'type_C', '2025-06-11 03:00:00', 'pending'),
('Delta-4', 17, 'descri_4', 'type_D', '2025-05-21 23:00:00', 'done');

CREATE TABLE IF NOT EXISTS cust_health_scores (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Health Scores';
INSERT IGNORE INTO cust_health_scores (name, customer_tiers_id, metric_date, value, target, status) VALUES
('Alpha-1', 88, '2025-06-04', 6547.15, 2230.09, 'active'),
('Beta-2', 96, '2025-06-18', 4210.10, 8868.68, 'pending');

CREATE TABLE IF NOT EXISTS cust_churn_predictions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Churn Predictions';
INSERT IGNORE INTO cust_churn_predictions (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 75, 'descri_1', 'type_A', '2025-06-11 14:00:00', 'active'),
('Beta-2', 61, 'descri_2', 'type_B', '2025-01-28 20:00:00', 'done'),
('Gamma-3', 75, 'descri_3', 'type_C', '2025-11-10 02:00:00', 'pending');

CREATE TABLE IF NOT EXISTS cust_win_back_campaigns (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Win Back Campaigns';
INSERT IGNORE INTO cust_win_back_campaigns (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 65, 'descri_1', 'type_A', '2025-03-06 21:00:00', 'pending'),
('Beta-2', 68, 'descri_2', 'type_B', '2025-06-13 13:00:00', 'active');

CREATE TABLE IF NOT EXISTS cust_customer_journeys (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Customer Journeys';
INSERT IGNORE INTO cust_customer_journeys (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 18, 'descri_1', 'type_A', '2025-04-19 10:00:00', 'pending'),
('Beta-2', 73, 'descri_2', 'type_B', '2025-08-11 04:00:00', 'pending');

CREATE TABLE IF NOT EXISTS cust_touchpoint_analysis (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Touchpoint Analysis';
INSERT IGNORE INTO cust_touchpoint_analysis (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 24, 'descri_1', 'type_A', '2025-02-14 15:00:00', 'pending'),
('Beta-2', 81, 'descri_2', 'type_B', '2025-04-07 10:00:00', 'done'),
('Gamma-3', 8, 'descri_3', 'type_C', '2025-01-03 14:00:00', 'pending'),
('Delta-4', 27, 'descri_4', 'type_D', '2025-03-03 18:00:00', 'active');

CREATE TABLE IF NOT EXISTS cust_customer_360_views (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Customer 360 Views';
INSERT IGNORE INTO cust_customer_360_views (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 98, 'descri_1', 'type_A', '2025-02-18 04:00:00', 'done'),
('Beta-2', 56, 'descri_2', 'type_B', '2025-11-17 13:00:00', 'pending'),
('Gamma-3', 55, 'descri_3', 'type_C', '2025-09-21 12:00:00', 'active');

CREATE TABLE IF NOT EXISTS cust_account_teams (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Account Teams';
INSERT IGNORE INTO cust_account_teams (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 69, 'descri_1', 'type_A', '2025-04-19 23:00:00', 'pending'),
('Beta-2', 2, 'descri_2', 'type_B', '2025-06-25 16:00:00', 'active');

CREATE TABLE IF NOT EXISTS cust_account_plans (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Account Plans';
INSERT IGNORE INTO cust_account_plans (name, customer_tiers_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 87, 2024, 51.19, 7337.96, 'active'),
('Beta-2', 89, 2025, 8770.20, 3127.49, 'active');

CREATE TABLE IF NOT EXISTS cust_strategic_accounts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Strategic Accounts';
INSERT IGNORE INTO cust_strategic_accounts (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 65, 'descri_1', 'type_A', '2025-11-22 09:00:00', 'active'),
('Beta-2', 9, 'descri_2', 'type_B', '2025-01-24 09:00:00', 'pending'),
('Gamma-3', 59, 'descri_3', 'type_C', '2025-09-15 19:00:00', 'done');

CREATE TABLE IF NOT EXISTS cust_customer_portals (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Customer Portals';
INSERT IGNORE INTO cust_customer_portals (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 56, 'descri_1', 'type_A', '2025-12-16 04:00:00', 'pending'),
('Beta-2', 63, 'descri_2', 'type_B', '2025-07-16 06:00:00', 'active'),
('Gamma-3', 52, 'descri_3', 'type_C', '2025-01-10 07:00:00', 'active');

CREATE TABLE IF NOT EXISTS cust_portal_activity_logs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Portal Activity Logs';
INSERT IGNORE INTO cust_portal_activity_logs (name, customer_tiers_id, event_date, description, created_by, status) VALUES
('Alpha-1', 17, '2025-04-25 23:00:00', 'Sample data row 1', 'create_1', 'active'),
('Beta-2', 76, '2025-09-07 05:00:00', 'Sample data row 2', 'create_2', 'active'),
('Gamma-3', 47, '2025-13-01 20:00:00', 'Sample data row 3', 'create_3', 'active'),
('Delta-4', 49, '2025-11-20 08:00:00', 'Sample data row 4', 'create_4', 'done');

CREATE TABLE IF NOT EXISTS cust_self_service_tickets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Self Service Tickets';
INSERT IGNORE INTO cust_self_service_tickets (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 7, 'descri_1', 'type_A', '2025-11-25 01:00:00', 'done'),
('Beta-2', 14, 'descri_2', 'type_B', '2025-12-27 00:00:00', 'active'),
('Gamma-3', 89, 'descri_3', 'type_C', '2025-02-04 13:00:00', 'pending'),
('Delta-4', 49, 'descri_4', 'type_D', '2025-03-08 17:00:00', 'pending');

CREATE TABLE IF NOT EXISTS cust_knowledge_base_articles (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Knowledge Base Articles';
INSERT IGNORE INTO cust_knowledge_base_articles (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 90, 'descri_1', 'type_A', '2025-03-23 06:00:00', 'done'),
('Beta-2', 88, 'descri_2', 'type_B', '2025-11-15 00:00:00', 'pending'),
('Gamma-3', 54, 'descri_3', 'type_C', '2025-12-03 03:00:00', 'done');

CREATE TABLE IF NOT EXISTS cust_faq_categories (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    customer_tiers_id            INT COMMENT 'FK to cust_customer_tiers' COMMENT 'Ref cust_customer_tiers',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (customer_tiers_id) REFERENCES cust_customer_tiers(id)
) COMMENT='Customer Success — Faq Categories';
INSERT IGNORE INTO cust_faq_categories (name, customer_tiers_id, description, category, created_at, status) VALUES
('Alpha-1', 35, 'descri_1', 'type_A', '2025-11-11 19:00:00', 'done'),
('Beta-2', 18, 'descri_2', 'type_B', '2025-08-19 03:00:00', 'done'),
('Gamma-3', 93, 'descri_3', 'type_C', '2025-11-09 03:00:00', 'pending'),
('Delta-4', 15, 'descri_4', 'type_D', '2025-02-27 15:00:00', 'active');

-- === Plant Ops (20 tables) ===

CREATE TABLE IF NOT EXISTS plnt_plant_master (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Plant Ops — Plant Master';
INSERT IGNORE INTO plnt_plant_master (name, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 2024, 2005.30, 2637.05, 'done'),
('Beta-2', 2025, 9730.49, 3585.28, 'pending'),
('Gamma-3', 2026, 2960.26, 403.62, 'pending'),
('Delta-4', 2024, 9103.55, 6884.47, 'pending');

CREATE TABLE IF NOT EXISTS plnt_plant_areas (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Plant Areas';
INSERT IGNORE INTO plnt_plant_areas (name, plant_master_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 31, 2024, 37.35, 6718.94, 'active'),
('Beta-2', 74, 2025, 6692.86, 1499.80, 'active'),
('Gamma-3', 100, 2026, 5121.42, 7237.68, 'active');

CREATE TABLE IF NOT EXISTS plnt_production_lines (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Production Lines';
INSERT IGNORE INTO plnt_production_lines (name, plant_master_id, description, category, created_at, status) VALUES
('Alpha-1', 83, 'descri_1', 'type_A', '2025-04-07 14:00:00', 'pending'),
('Beta-2', 34, 'descri_2', 'type_B', '2025-08-14 01:00:00', 'active'),
('Gamma-3', 23, 'descri_3', 'type_C', '2025-01-21 09:00:00', 'pending');

CREATE TABLE IF NOT EXISTS plnt_line_stations (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Line Stations';
INSERT IGNORE INTO plnt_line_stations (name, plant_master_id, description, category, created_at, status) VALUES
('Alpha-1', 41, 'descri_1', 'type_A', '2025-08-23 23:00:00', 'active'),
('Beta-2', 13, 'descri_2', 'type_B', '2025-01-24 00:00:00', 'active'),
('Gamma-3', 85, 'descri_3', 'type_C', '2025-12-24 05:00:00', 'pending'),
('Delta-4', 46, 'descri_4', 'type_D', '2025-08-01 19:00:00', 'pending');

CREATE TABLE IF NOT EXISTS plnt_station_assignments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Station Assignments';
INSERT IGNORE INTO plnt_station_assignments (name, plant_master_id, description, category, created_at, status) VALUES
('Alpha-1', 12, 'descri_1', 'type_A', '2025-07-23 10:00:00', 'active'),
('Beta-2', 15, 'descri_2', 'type_B', '2025-07-10 12:00:00', 'pending'),
('Gamma-3', 35, 'descri_3', 'type_C', '2025-07-02 22:00:00', 'pending'),
('Delta-4', 54, 'descri_4', 'type_D', '2025-12-25 05:00:00', 'active');

CREATE TABLE IF NOT EXISTS plnt_shift_schedules (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Shift Schedules';
INSERT IGNORE INTO plnt_shift_schedules (name, plant_master_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 90, 2024, 7266.12, 9814.89, 'active'),
('Beta-2', 14, 2025, 4165.70, 5911.00, 'done');

CREATE TABLE IF NOT EXISTS plnt_shift_assignments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Shift Assignments';
INSERT IGNORE INTO plnt_shift_assignments (name, plant_master_id, description, category, created_at, status) VALUES
('Alpha-1', 57, 'descri_1', 'type_A', '2025-04-09 19:00:00', 'pending'),
('Beta-2', 100, 'descri_2', 'type_B', '2025-07-16 19:00:00', 'active'),
('Gamma-3', 88, 'descri_3', 'type_C', '2025-03-11 15:00:00', 'pending'),
('Delta-4', 15, 'descri_4', 'type_D', '2025-08-15 14:00:00', 'active');

CREATE TABLE IF NOT EXISTS plnt_downtime_events (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Downtime Events';
INSERT IGNORE INTO plnt_downtime_events (name, plant_master_id, event_date, description, created_by, status) VALUES
('Alpha-1', 34, '2025-12-28 10:00:00', 'Sample data row 1', 'create_1', 'active'),
('Beta-2', 91, '2025-12-12 16:00:00', 'Sample data row 2', 'create_2', 'done');

CREATE TABLE IF NOT EXISTS plnt_downtime_reasons (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Downtime Reasons';
INSERT IGNORE INTO plnt_downtime_reasons (name, plant_master_id, description, category, created_at, status) VALUES
('Alpha-1', 93, 'descri_1', 'type_A', '2025-10-10 07:00:00', 'done'),
('Beta-2', 44, 'descri_2', 'type_B', '2025-09-14 16:00:00', 'done'),
('Gamma-3', 90, 'descri_3', 'type_C', '2025-11-09 03:00:00', 'pending'),
('Delta-4', 85, 'descri_4', 'type_D', '2025-05-16 15:00:00', 'pending');

CREATE TABLE IF NOT EXISTS plnt_oee_metrics (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    metric_date                  DATE COMMENT 'Date',
    value                        DECIMAL(14,4) COMMENT 'Value',
    target                       DECIMAL(14,4) COMMENT 'Target',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Oee Metrics';
INSERT IGNORE INTO plnt_oee_metrics (name, plant_master_id, metric_date, value, target, status) VALUES
('Alpha-1', 87, '2025-11-21', 3854.53, 8391.10, 'done'),
('Beta-2', 79, '2025-01-28', 126.32, 9343.44, 'done'),
('Gamma-3', 61, '2025-09-02', 8694.59, 757.83, 'pending'),
('Delta-4', 42, '2025-05-12', 3315.31, 2788.51, 'done');

CREATE TABLE IF NOT EXISTS plnt_takt_time_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Takt Time Records';
INSERT IGNORE INTO plnt_takt_time_records (name, plant_master_id, event_date, description, created_by, status) VALUES
('Alpha-1', 71, '2025-11-22 12:00:00', 'Sample data row 1', 'create_1', 'pending'),
('Beta-2', 31, '2025-02-13 14:00:00', 'Sample data row 2', 'create_2', 'pending');

CREATE TABLE IF NOT EXISTS plnt_andon_alerts (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Andon Alerts';
INSERT IGNORE INTO plnt_andon_alerts (name, plant_master_id, description, category, created_at, status) VALUES
('Alpha-1', 58, 'descri_1', 'type_A', '2025-02-18 15:00:00', 'pending'),
('Beta-2', 41, 'descri_2', 'type_B', '2025-03-05 15:00:00', 'done'),
('Gamma-3', 84, 'descri_3', 'type_C', '2025-01-10 03:00:00', 'pending'),
('Delta-4', 53, 'descri_4', 'type_D', '2025-01-19 17:00:00', 'done');

CREATE TABLE IF NOT EXISTS plnt_kanban_cards (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Kanban Cards';
INSERT IGNORE INTO plnt_kanban_cards (name, plant_master_id, description, category, created_at, status) VALUES
('Alpha-1', 12, 'descri_1', 'type_A', '2025-11-02 19:00:00', 'done'),
('Beta-2', 39, 'descri_2', 'type_B', '2025-09-09 17:00:00', 'active');

CREATE TABLE IF NOT EXISTS plnt_kanban_boards (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Kanban Boards';
INSERT IGNORE INTO plnt_kanban_boards (name, plant_master_id, description, category, created_at, status) VALUES
('Alpha-1', 86, 'descri_1', 'type_A', '2025-08-10 10:00:00', 'pending'),
('Beta-2', 7, 'descri_2', 'type_B', '2025-04-03 14:00:00', 'pending'),
('Gamma-3', 72, 'descri_3', 'type_C', '2025-11-10 19:00:00', 'pending'),
('Delta-4', 92, 'descri_4', 'type_D', '2025-12-01 11:00:00', 'done');

CREATE TABLE IF NOT EXISTS plnt_visual_mgmt_boards (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Visual Mgmt Boards';
INSERT IGNORE INTO plnt_visual_mgmt_boards (name, plant_master_id, description, category, created_at, status) VALUES
('Alpha-1', 15, 'descri_1', 'type_A', '2025-02-22 21:00:00', 'active'),
('Beta-2', 1, 'descri_2', 'type_B', '2025-06-07 11:00:00', 'done'),
('Gamma-3', 38, 'descri_3', 'type_C', '2025-10-01 16:00:00', 'pending');

CREATE TABLE IF NOT EXISTS plnt_gemba_walk_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Gemba Walk Records';
INSERT IGNORE INTO plnt_gemba_walk_records (name, plant_master_id, event_date, description, created_by, status) VALUES
('Alpha-1', 14, '2025-01-24 05:00:00', 'Sample data row 1', 'create_1', 'pending'),
('Beta-2', 18, '2025-07-14 03:00:00', 'Sample data row 2', 'create_2', 'pending'),
('Gamma-3', 79, '2025-04-17 06:00:00', 'Sample data row 3', 'create_3', 'active'),
('Delta-4', 51, '2025-04-27 14:00:00', 'Sample data row 4', 'create_4', 'active');

CREATE TABLE IF NOT EXISTS plnt_kaizen_suggestions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Kaizen Suggestions';
INSERT IGNORE INTO plnt_kaizen_suggestions (name, plant_master_id, description, category, created_at, status) VALUES
('Alpha-1', 44, 'descri_1', 'type_A', '2025-02-28 13:00:00', 'active'),
('Beta-2', 82, 'descri_2', 'type_B', '2025-12-09 18:00:00', 'active'),
('Gamma-3', 99, 'descri_3', 'type_C', '2025-08-07 14:00:00', 'done'),
('Delta-4', 76, 'descri_4', 'type_D', '2025-09-07 04:00:00', 'pending');

CREATE TABLE IF NOT EXISTS plnt_suggestion_reviews (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Suggestion Reviews';
INSERT IGNORE INTO plnt_suggestion_reviews (name, plant_master_id, description, category, created_at, status) VALUES
('Alpha-1', 68, 'descri_1', 'type_A', '2025-01-25 17:00:00', 'pending'),
('Beta-2', 75, 'descri_2', 'type_B', '2025-09-24 16:00:00', 'active');

CREATE TABLE IF NOT EXISTS plnt_ci_projects (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Ci Projects';
INSERT IGNORE INTO plnt_ci_projects (name, plant_master_id, description, category, created_at, status) VALUES
('Alpha-1', 75, 'descri_1', 'type_A', '2025-03-05 23:00:00', 'active'),
('Beta-2', 14, 'descri_2', 'type_B', '2025-07-03 21:00:00', 'done'),
('Gamma-3', 78, 'descri_3', 'type_C', '2025-06-25 16:00:00', 'pending'),
('Delta-4', 54, 'descri_4', 'type_D', '2025-11-04 22:00:00', 'active');

CREATE TABLE IF NOT EXISTS plnt_ci_results (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    plant_master_id              INT COMMENT 'FK to plnt_plant_master' COMMENT 'Ref plnt_plant_master',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (plant_master_id) REFERENCES plnt_plant_master(id)
) COMMENT='Plant Ops — Ci Results';
INSERT IGNORE INTO plnt_ci_results (name, plant_master_id, description, category, created_at, status) VALUES
('Alpha-1', 51, 'descri_1', 'type_A', '2025-06-08 09:00:00', 'pending'),
('Beta-2', 18, 'descri_2', 'type_B', '2025-03-15 13:00:00', 'done'),
('Gamma-3', 90, 'descri_3', 'type_C', '2025-09-11 09:00:00', 'done');

-- === Operations (70 tables) ===

CREATE TABLE IF NOT EXISTS ops_batch_jobs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status'
) COMMENT='Operations — Batch Jobs';
INSERT IGNORE INTO ops_batch_jobs (name, description, category, created_at, status) VALUES
('Alpha-1', 'descri_1', 'type_A', '2025-06-21 22:00:00', 'done'),
('Beta-2', 'descri_2', 'type_B', '2025-10-04 20:00:00', 'active'),
('Gamma-3', 'descri_3', 'type_C', '2025-04-19 06:00:00', 'done'),
('Delta-4', 'descri_4', 'type_D', '2025-05-09 21:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_batch_steps (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Batch Steps';
INSERT IGNORE INTO ops_batch_steps (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 17, 'descri_1', 'type_A', '2025-13-28 20:00:00', 'active'),
('Beta-2', 65, 'descri_2', 'type_B', '2025-10-26 22:00:00', 'done'),
('Gamma-3', 42, 'descri_3', 'type_C', '2025-13-28 03:00:00', 'pending'),
('Delta-4', 72, 'descri_4', 'type_D', '2025-09-22 18:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_mq_topics (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Mq Topics';
INSERT IGNORE INTO ops_mq_topics (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 91, 'descri_1', 'type_A', '2025-08-18 21:00:00', 'done'),
('Beta-2', 69, 'descri_2', 'type_B', '2025-05-07 13:00:00', 'active'),
('Gamma-3', 67, 'descri_3', 'type_C', '2025-02-27 00:00:00', 'pending'),
('Delta-4', 19, 'descri_4', 'type_D', '2025-01-28 21:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_mq_subscriptions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Mq Subscriptions';
INSERT IGNORE INTO ops_mq_subscriptions (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 35, 'descri_1', 'type_A', '2025-03-27 09:00:00', 'done'),
('Beta-2', 33, 'descri_2', 'type_B', '2025-03-19 01:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_api_endpoints (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Api Endpoints';
INSERT IGNORE INTO ops_api_endpoints (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 27, 'descri_1', 'type_A', '2025-10-26 01:00:00', 'pending'),
('Beta-2', 58, 'descri_2', 'type_B', '2025-02-25 19:00:00', 'done'),
('Gamma-3', 95, 'descri_3', 'type_C', '2025-10-08 07:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_api_usage_logs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Api Usage Logs';
INSERT IGNORE INTO ops_api_usage_logs (name, batch_jobs_id, event_date, description, created_by, status) VALUES
('Alpha-1', 72, '2025-07-04 07:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 90, '2025-05-08 22:00:00', 'Sample data row 2', 'create_2', 'done'),
('Gamma-3', 7, '2025-07-28 12:00:00', 'Sample data row 3', 'create_3', 'pending'),
('Delta-4', 88, '2025-06-21 17:00:00', 'Sample data row 4', 'create_4', 'active');

CREATE TABLE IF NOT EXISTS ops_webhook_configs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Webhook Configs';
INSERT IGNORE INTO ops_webhook_configs (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 96, 'descri_1', 'type_A', '2025-05-17 06:00:00', 'done'),
('Beta-2', 100, 'descri_2', 'type_B', '2025-08-03 07:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_webhook_deliveries (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Webhook Deliveries';
INSERT IGNORE INTO ops_webhook_deliveries (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 91, 'descri_1', 'type_A', '2025-12-24 11:00:00', 'done'),
('Beta-2', 70, 'descri_2', 'type_B', '2025-11-01 06:00:00', 'done'),
('Gamma-3', 25, 'descri_3', 'type_C', '2025-06-11 14:00:00', 'active'),
('Delta-4', 88, 'descri_4', 'type_D', '2025-02-11 05:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_doc_templates (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Doc Templates';
INSERT IGNORE INTO ops_doc_templates (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 66, 'descri_1', 'type_A', '2025-03-06 12:00:00', 'active'),
('Beta-2', 55, 'descri_2', 'type_B', '2025-05-03 17:00:00', 'pending'),
('Gamma-3', 17, 'descri_3', 'type_C', '2025-03-27 18:00:00', 'pending'),
('Delta-4', 1, 'descri_4', 'type_D', '2025-06-02 14:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_doc_versions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Doc Versions';
INSERT IGNORE INTO ops_doc_versions (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 6, 'descri_1', 'type_A', '2025-03-23 10:00:00', 'done'),
('Beta-2', 7, 'descri_2', 'type_B', '2025-11-13 19:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_workflow_defs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Workflow Defs';
INSERT IGNORE INTO ops_workflow_defs (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 63, 'descri_1', 'type_A', '2025-10-14 17:00:00', 'pending'),
('Beta-2', 10, 'descri_2', 'type_B', '2025-13-27 10:00:00', 'done'),
('Gamma-3', 29, 'descri_3', 'type_C', '2025-04-11 16:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_workflow_instances (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Workflow Instances';
INSERT IGNORE INTO ops_workflow_instances (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 21, 'descri_1', 'type_A', '2025-08-17 17:00:00', 'done'),
('Beta-2', 52, 'descri_2', 'type_B', '2025-02-20 11:00:00', 'active'),
('Gamma-3', 28, 'descri_3', 'type_C', '2025-12-04 10:00:00', 'pending'),
('Delta-4', 47, 'descri_4', 'type_D', '2025-06-11 06:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_workflow_steps (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Workflow Steps';
INSERT IGNORE INTO ops_workflow_steps (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 61, 'descri_1', 'type_A', '2025-10-08 20:00:00', 'active'),
('Beta-2', 15, 'descri_2', 'type_B', '2025-12-07 11:00:00', 'pending'),
('Gamma-3', 31, 'descri_3', 'type_C', '2025-11-21 19:00:00', 'done'),
('Delta-4', 32, 'descri_4', 'type_D', '2025-01-21 21:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_notif_templates (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Notif Templates';
INSERT IGNORE INTO ops_notif_templates (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 15, 'descri_1', 'type_A', '2025-07-02 08:00:00', 'done'),
('Beta-2', 93, 'descri_2', 'type_B', '2025-05-05 00:00:00', 'done'),
('Gamma-3', 99, 'descri_3', 'type_C', '2025-07-24 16:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_notif_channels (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Notif Channels';
INSERT IGNORE INTO ops_notif_channels (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 63, 'descri_1', 'type_A', '2025-01-23 09:00:00', 'done'),
('Beta-2', 25, 'descri_2', 'type_B', '2025-06-25 16:00:00', 'active'),
('Gamma-3', 13, 'descri_3', 'type_C', '2025-03-02 17:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_scheduled_tasks (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Scheduled Tasks';
INSERT IGNORE INTO ops_scheduled_tasks (name, batch_jobs_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 81, 2024, 664.00, 6705.98, 'active'),
('Beta-2', 88, 2025, 4247.43, 1048.12, 'done');

CREATE TABLE IF NOT EXISTS ops_task_exec_logs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Task Exec Logs';
INSERT IGNORE INTO ops_task_exec_logs (name, batch_jobs_id, event_date, description, created_by, status) VALUES
('Alpha-1', 95, '2025-08-27 20:00:00', 'Sample data row 1', 'create_1', 'active'),
('Beta-2', 96, '2025-12-16 04:00:00', 'Sample data row 2', 'create_2', 'active'),
('Gamma-3', 59, '2025-06-02 01:00:00', 'Sample data row 3', 'create_3', 'active'),
('Delta-4', 10, '2025-01-28 05:00:00', 'Sample data row 4', 'create_4', 'pending');

CREATE TABLE IF NOT EXISTS ops_feature_flags (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Feature Flags';
INSERT IGNORE INTO ops_feature_flags (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 10, 'descri_1', 'type_A', '2025-10-15 07:00:00', 'pending'),
('Beta-2', 53, 'descri_2', 'type_B', '2025-07-03 14:00:00', 'done'),
('Gamma-3', 52, 'descri_3', 'type_C', '2025-08-26 10:00:00', 'active'),
('Delta-4', 50, 'descri_4', 'type_D', '2025-12-09 03:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_ab_experiments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Ab Experiments';
INSERT IGNORE INTO ops_ab_experiments (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 83, 'descri_1', 'type_A', '2025-11-09 02:00:00', 'done'),
('Beta-2', 75, 'descri_2', 'type_B', '2025-01-24 22:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_ab_results (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Ab Results';
INSERT IGNORE INTO ops_ab_results (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 65, 'descri_1', 'type_A', '2025-02-27 09:00:00', 'done'),
('Beta-2', 39, 'descri_2', 'type_B', '2025-11-03 02:00:00', 'pending'),
('Gamma-3', 98, 'descri_3', 'type_C', '2025-08-01 12:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_geo_regions (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Geo Regions';
INSERT IGNORE INTO ops_geo_regions (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 4, 'descri_1', 'type_A', '2025-09-19 05:00:00', 'done'),
('Beta-2', 29, 'descri_2', 'type_B', '2025-03-14 23:00:00', 'pending'),
('Gamma-3', 70, 'descri_3', 'type_C', '2025-06-12 20:00:00', 'active'),
('Delta-4', 39, 'descri_4', 'type_D', '2025-12-11 23:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_geo_cities (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Geo Cities';
INSERT IGNORE INTO ops_geo_cities (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 2, 'descri_1', 'type_A', '2025-10-04 17:00:00', 'active'),
('Beta-2', 16, 'descri_2', 'type_B', '2025-01-22 00:00:00', 'done'),
('Gamma-3', 77, 'descri_3', 'type_C', '2025-07-06 17:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_timezone_defs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Timezone Defs';
INSERT IGNORE INTO ops_timezone_defs (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 40, 'descri_1', 'type_A', '2025-04-24 23:00:00', 'done'),
('Beta-2', 93, 'descri_2', 'type_B', '2025-06-09 06:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_currency_master (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Currency Master';
INSERT IGNORE INTO ops_currency_master (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 66, 'descri_1', 'type_A', '2025-03-17 05:00:00', 'active'),
('Beta-2', 29, 'descri_2', 'type_B', '2025-11-08 08:00:00', 'active'),
('Gamma-3', 15, 'descri_3', 'type_C', '2025-04-09 21:00:00', 'done'),
('Delta-4', 7, 'descri_4', 'type_D', '2025-10-02 14:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_language_packs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Language Packs';
INSERT IGNORE INTO ops_language_packs (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 10, 'descri_1', 'type_A', '2025-05-09 21:00:00', 'active'),
('Beta-2', 33, 'descri_2', 'type_B', '2025-02-23 06:00:00', 'done'),
('Gamma-3', 75, 'descri_3', 'type_C', '2025-09-26 10:00:00', 'pending'),
('Delta-4', 17, 'descri_4', 'type_D', '2025-12-08 20:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_translation_keys (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Translation Keys';
INSERT IGNORE INTO ops_translation_keys (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 37, 'descri_1', 'type_A', '2025-11-10 02:00:00', 'active'),
('Beta-2', 41, 'descri_2', 'type_B', '2025-09-25 14:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_user_preferences (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — User Preferences';
INSERT IGNORE INTO ops_user_preferences (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 85, 'descri_1', 'type_A', '2025-06-17 22:00:00', 'done'),
('Beta-2', 20, 'descri_2', 'type_B', '2025-10-10 18:00:00', 'pending'),
('Gamma-3', 41, 'descri_3', 'type_C', '2025-08-24 05:00:00', 'active'),
('Delta-4', 41, 'descri_4', 'type_D', '2025-05-15 07:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_system_params (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — System Params';
INSERT IGNORE INTO ops_system_params (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 83, 'descri_1', 'type_A', '2025-08-28 08:00:00', 'pending'),
('Beta-2', 18, 'descri_2', 'type_B', '2025-12-12 23:00:00', 'pending'),
('Gamma-3', 62, 'descri_3', 'type_C', '2025-08-13 14:00:00', 'pending'),
('Delta-4', 100, 'descri_4', 'type_D', '2025-06-17 15:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_health_checks (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Health Checks';
INSERT IGNORE INTO ops_health_checks (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 23, 'descri_1', 'type_A', '2025-12-23 19:00:00', 'active'),
('Beta-2', 36, 'descri_2', 'type_B', '2025-03-16 17:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_perf_baselines (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Perf Baselines';
INSERT IGNORE INTO ops_perf_baselines (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 94, 'descri_1', 'type_A', '2025-11-13 02:00:00', 'done'),
('Beta-2', 10, 'descri_2', 'type_B', '2025-01-10 16:00:00', 'done'),
('Gamma-3', 4, 'descri_3', 'type_C', '2025-10-18 18:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_capacity_plans (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Capacity Plans';
INSERT IGNORE INTO ops_capacity_plans (name, batch_jobs_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 28, 2024, 9063.07, 5519.87, 'done'),
('Beta-2', 70, 2025, 4725.85, 7753.45, 'pending'),
('Gamma-3', 51, 2026, 2240.61, 7722.95, 'done'),
('Delta-4', 38, 2024, 1347.34, 5015.72, 'active');

CREATE TABLE IF NOT EXISTS ops_dr_plans (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Dr Plans';
INSERT IGNORE INTO ops_dr_plans (name, batch_jobs_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 47, 2024, 5026.15, 6629.50, 'done'),
('Beta-2', 88, 2025, 1666.61, 5145.71, 'pending'),
('Gamma-3', 21, 2026, 1799.48, 9486.68, 'done');

CREATE TABLE IF NOT EXISTS ops_backup_policies (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Backup Policies';
INSERT IGNORE INTO ops_backup_policies (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 95, 'descri_1', 'type_A', '2025-09-24 11:00:00', 'active'),
('Beta-2', 29, 'descri_2', 'type_B', '2025-09-26 07:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_data_class_rules (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Data Class Rules';
INSERT IGNORE INTO ops_data_class_rules (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 48, 'descri_1', 'type_A', '2025-04-08 01:00:00', 'done'),
('Beta-2', 27, 'descri_2', 'type_B', '2025-09-14 11:00:00', 'pending'),
('Gamma-3', 61, 'descri_3', 'type_C', '2025-08-09 01:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_key_rotations (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Key Rotations';
INSERT IGNORE INTO ops_key_rotations (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 42, 'descri_1', 'type_A', '2025-02-25 22:00:00', 'done'),
('Beta-2', 82, 'descri_2', 'type_B', '2025-05-01 23:00:00', 'pending'),
('Gamma-3', 75, 'descri_3', 'type_C', '2025-10-27 18:00:00', 'active'),
('Delta-4', 49, 'descri_4', 'type_D', '2025-12-27 12:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_cert_mgmt (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Cert Mgmt';
INSERT IGNORE INTO ops_cert_mgmt (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 84, 'descri_1', 'type_A', '2025-02-12 16:00:00', 'pending'),
('Beta-2', 49, 'descri_2', 'type_B', '2025-04-03 16:00:00', 'pending'),
('Gamma-3', 46, 'descri_3', 'type_C', '2025-06-09 15:00:00', 'pending'),
('Delta-4', 78, 'descri_4', 'type_D', '2025-01-04 16:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_dns_records (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Dns Records';
INSERT IGNORE INTO ops_dns_records (name, batch_jobs_id, event_date, description, created_by, status) VALUES
('Alpha-1', 22, '2025-04-06 02:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 83, '2025-05-04 06:00:00', 'Sample data row 2', 'create_2', 'active'),
('Gamma-3', 24, '2025-03-25 11:00:00', 'Sample data row 3', 'create_3', 'done');

CREATE TABLE IF NOT EXISTS ops_lb_configs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Lb Configs';
INSERT IGNORE INTO ops_lb_configs (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 16, 'descri_1', 'type_A', '2025-12-03 07:00:00', 'pending'),
('Beta-2', 43, 'descri_2', 'type_B', '2025-02-03 15:00:00', 'done'),
('Gamma-3', 62, 'descri_3', 'type_C', '2025-09-22 06:00:00', 'done'),
('Delta-4', 21, 'descri_4', 'type_D', '2025-12-24 13:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_container_deploys (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Container Deploys';
INSERT IGNORE INTO ops_container_deploys (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 6, 'descri_1', 'type_A', '2025-03-10 18:00:00', 'done'),
('Beta-2', 21, 'descri_2', 'type_B', '2025-03-13 20:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_mesh_policies (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Mesh Policies';
INSERT IGNORE INTO ops_mesh_policies (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 74, 'descri_1', 'type_A', '2025-03-20 01:00:00', 'active'),
('Beta-2', 34, 'descri_2', 'type_B', '2025-04-09 16:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_rate_limits (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Rate Limits';
INSERT IGNORE INTO ops_rate_limits (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 73, 'descri_1', 'type_A', '2025-11-19 00:00:00', 'done'),
('Beta-2', 94, 'descri_2', 'type_B', '2025-12-03 09:00:00', 'active'),
('Gamma-3', 28, 'descri_3', 'type_C', '2025-08-06 19:00:00', 'pending'),
('Delta-4', 62, 'descri_4', 'type_D', '2025-03-04 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_oauth_clients (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Oauth Clients';
INSERT IGNORE INTO ops_oauth_clients (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 19, 'descri_1', 'type_A', '2025-10-19 05:00:00', 'done'),
('Beta-2', 85, 'descri_2', 'type_B', '2025-09-10 08:00:00', 'active'),
('Gamma-3', 93, 'descri_3', 'type_C', '2025-12-16 10:00:00', 'pending'),
('Delta-4', 77, 'descri_4', 'type_D', '2025-11-23 02:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_oauth_tokens (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Oauth Tokens';
INSERT IGNORE INTO ops_oauth_tokens (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 50, 'descri_1', 'type_A', '2025-03-12 03:00:00', 'active'),
('Beta-2', 26, 'descri_2', 'type_B', '2025-09-16 16:00:00', 'done'),
('Gamma-3', 54, 'descri_3', 'type_C', '2025-03-21 19:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_saml_providers (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Saml Providers';
INSERT IGNORE INTO ops_saml_providers (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 61, 'descri_1', 'type_A', '2025-11-07 19:00:00', 'active'),
('Beta-2', 69, 'descri_2', 'type_B', '2025-09-06 03:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_mfa_configs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Mfa Configs';
INSERT IGNORE INTO ops_mfa_configs (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 85, 'descri_1', 'type_A', '2025-03-11 19:00:00', 'done'),
('Beta-2', 1, 'descri_2', 'type_B', '2025-12-12 20:00:00', 'pending'),
('Gamma-3', 67, 'descri_3', 'type_C', '2025-08-20 11:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_login_audit (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Login Audit';
INSERT IGNORE INTO ops_login_audit (name, batch_jobs_id, event_date, description, created_by, status) VALUES
('Alpha-1', 68, '2025-08-24 07:00:00', 'Sample data row 1', 'create_1', 'pending'),
('Beta-2', 99, '2025-07-28 11:00:00', 'Sample data row 2', 'create_2', 'done'),
('Gamma-3', 95, '2025-12-21 18:00:00', 'Sample data row 3', 'create_3', 'active'),
('Delta-4', 55, '2025-03-13 08:00:00', 'Sample data row 4', 'create_4', 'pending');

CREATE TABLE IF NOT EXISTS ops_permission_sets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Permission Sets';
INSERT IGNORE INTO ops_permission_sets (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 11, 'descri_1', 'type_A', '2025-05-06 19:00:00', 'done'),
('Beta-2', 53, 'descri_2', 'type_B', '2025-05-20 06:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_role_assignments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Role Assignments';
INSERT IGNORE INTO ops_role_assignments (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 96, 'descri_1', 'type_A', '2025-01-04 18:00:00', 'done'),
('Beta-2', 16, 'descri_2', 'type_B', '2025-09-15 00:00:00', 'pending'),
('Gamma-3', 84, 'descri_3', 'type_C', '2025-04-22 22:00:00', 'pending'),
('Delta-4', 53, 'descri_4', 'type_D', '2025-07-18 03:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_resource_quotas (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Resource Quotas';
INSERT IGNORE INTO ops_resource_quotas (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 84, 'descri_1', 'type_A', '2025-10-20 08:00:00', 'pending'),
('Beta-2', 69, 'descri_2', 'type_B', '2025-08-21 18:00:00', 'pending'),
('Gamma-3', 79, 'descri_3', 'type_C', '2025-05-02 21:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_cost_tags (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Cost Tags';
INSERT IGNORE INTO ops_cost_tags (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 53, 'descri_1', 'type_A', '2025-10-03 17:00:00', 'pending'),
('Beta-2', 23, 'descri_2', 'type_B', '2025-07-22 17:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_tag_defs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Tag Defs';
INSERT IGNORE INTO ops_tag_defs (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 33, 'descri_1', 'type_A', '2025-06-18 18:00:00', 'pending'),
('Beta-2', 100, 'descri_2', 'type_B', '2025-13-26 12:00:00', 'done'),
('Gamma-3', 18, 'descri_3', 'type_C', '2025-12-02 03:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_tag_assignments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Tag Assignments';
INSERT IGNORE INTO ops_tag_assignments (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 58, 'descri_1', 'type_A', '2025-12-23 05:00:00', 'active'),
('Beta-2', 28, 'descri_2', 'type_B', '2025-09-09 12:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_event_subs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Event Subs';
INSERT IGNORE INTO ops_event_subs (name, batch_jobs_id, event_date, description, created_by, status) VALUES
('Alpha-1', 79, '2025-12-21 21:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 7, '2025-03-27 12:00:00', 'Sample data row 2', 'create_2', 'done'),
('Gamma-3', 1, '2025-02-19 04:00:00', 'Sample data row 3', 'create_3', 'pending'),
('Delta-4', 77, '2025-08-27 02:00:00', 'Sample data row 4', 'create_4', 'active');

CREATE TABLE IF NOT EXISTS ops_event_handlers (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Event Handlers';
INSERT IGNORE INTO ops_event_handlers (name, batch_jobs_id, event_date, description, created_by, status) VALUES
('Alpha-1', 29, '2025-07-22 02:00:00', 'Sample data row 1', 'create_1', 'done'),
('Beta-2', 5, '2025-07-01 12:00:00', 'Sample data row 2', 'create_2', 'done'),
('Gamma-3', 84, '2025-09-15 09:00:00', 'Sample data row 3', 'create_3', 'done');

CREATE TABLE IF NOT EXISTS ops_cron_schedules (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    fiscal_year                  INT COMMENT 'Year',
    planned_value                DECIMAL(14,2) COMMENT 'Planned',
    actual_value                 DECIMAL(14,2) DEFAULT 0 COMMENT 'Actual',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Cron Schedules';
INSERT IGNORE INTO ops_cron_schedules (name, batch_jobs_id, fiscal_year, planned_value, actual_value, status) VALUES
('Alpha-1', 8, 2024, 9456.76, 8203.26, 'done'),
('Beta-2', 64, 2025, 4394.89, 2024.85, 'done'),
('Gamma-3', 67, 2026, 2866.92, 345.72, 'done'),
('Delta-4', 66, 2024, 5162.46, 2324.16, 'active');

CREATE TABLE IF NOT EXISTS ops_cron_history (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Cron History';
INSERT IGNORE INTO ops_cron_history (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 5, 'descri_1', 'type_A', '2025-07-17 07:00:00', 'done'),
('Beta-2', 1, 'descri_2', 'type_B', '2025-03-25 02:00:00', 'pending'),
('Gamma-3', 100, 'descri_3', 'type_C', '2025-07-09 11:00:00', 'done'),
('Delta-4', 51, 'descri_4', 'type_D', '2025-07-12 18:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_file_buckets (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — File Buckets';
INSERT IGNORE INTO ops_file_buckets (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 2, 'descri_1', 'type_A', '2025-13-25 02:00:00', 'active'),
('Beta-2', 71, 'descri_2', 'type_B', '2025-05-06 23:00:00', 'done'),
('Gamma-3', 75, 'descri_3', 'type_C', '2025-06-11 11:00:00', 'pending'),
('Delta-4', 29, 'descri_4', 'type_D', '2025-11-03 06:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_file_metadata (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — File Metadata';
INSERT IGNORE INTO ops_file_metadata (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 74, 'descri_1', 'type_A', '2025-01-03 08:00:00', 'active'),
('Beta-2', 53, 'descri_2', 'type_B', '2025-09-08 03:00:00', 'done'),
('Gamma-3', 10, 'descri_3', 'type_C', '2025-09-06 09:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_img_proc_jobs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Img Proc Jobs';
INSERT IGNORE INTO ops_img_proc_jobs (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 66, 'descri_1', 'type_A', '2025-05-20 17:00:00', 'done'),
('Beta-2', 33, 'descri_2', 'type_B', '2025-12-20 14:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_pdf_gen_jobs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Pdf Gen Jobs';
INSERT IGNORE INTO ops_pdf_gen_jobs (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 62, 'descri_1', 'type_A', '2025-05-13 10:00:00', 'pending'),
('Beta-2', 88, 'descri_2', 'type_B', '2025-08-12 18:00:00', 'active'),
('Gamma-3', 62, 'descri_3', 'type_C', '2025-02-04 08:00:00', 'active'),
('Delta-4', 49, 'descri_4', 'type_D', '2025-11-17 07:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_email_logs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Email Logs';
INSERT IGNORE INTO ops_email_logs (name, batch_jobs_id, event_date, description, created_by, status) VALUES
('Alpha-1', 30, '2025-06-22 13:00:00', 'Sample data row 1', 'create_1', 'pending'),
('Beta-2', 67, '2025-10-23 00:00:00', 'Sample data row 2', 'create_2', 'done'),
('Gamma-3', 34, '2025-05-16 09:00:00', 'Sample data row 3', 'create_3', 'active'),
('Delta-4', 52, '2025-12-09 23:00:00', 'Sample data row 4', 'create_4', 'pending');

CREATE TABLE IF NOT EXISTS ops_sms_logs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Sms Logs';
INSERT IGNORE INTO ops_sms_logs (name, batch_jobs_id, event_date, description, created_by, status) VALUES
('Alpha-1', 71, '2025-04-08 15:00:00', 'Sample data row 1', 'create_1', 'pending'),
('Beta-2', 46, '2025-04-28 18:00:00', 'Sample data row 2', 'create_2', 'pending');

CREATE TABLE IF NOT EXISTS ops_push_notif_logs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    event_date                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Date',
    description                  TEXT COMMENT 'Description',
    created_by                   VARCHAR(80) COMMENT 'Creator',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Push Notif Logs';
INSERT IGNORE INTO ops_push_notif_logs (name, batch_jobs_id, event_date, description, created_by, status) VALUES
('Alpha-1', 96, '2025-12-14 06:00:00', 'Sample data row 1', 'create_1', 'pending'),
('Beta-2', 95, '2025-04-19 07:00:00', 'Sample data row 2', 'create_2', 'done');

CREATE TABLE IF NOT EXISTS ops_in_app_msgs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — In App Msgs';
INSERT IGNORE INTO ops_in_app_msgs (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 34, 'descri_1', 'type_A', '2025-04-22 07:00:00', 'active'),
('Beta-2', 32, 'descri_2', 'type_B', '2025-11-09 11:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_activity_feed (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Activity Feed';
INSERT IGNORE INTO ops_activity_feed (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 16, 'descri_1', 'type_A', '2025-02-13 21:00:00', 'done'),
('Beta-2', 64, 'descri_2', 'type_B', '2025-11-07 00:00:00', 'done');

CREATE TABLE IF NOT EXISTS ops_feed_comments (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Feed Comments';
INSERT IGNORE INTO ops_feed_comments (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 62, 'descri_1', 'type_A', '2025-07-23 19:00:00', 'pending'),
('Beta-2', 64, 'descri_2', 'type_B', '2025-10-15 08:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_global_settings (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Global Settings';
INSERT IGNORE INTO ops_global_settings (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 98, 'descri_1', 'type_A', '2025-09-10 08:00:00', 'pending'),
('Beta-2', 29, 'descri_2', 'type_B', '2025-10-16 04:00:00', 'active');

CREATE TABLE IF NOT EXISTS ops_tenant_configs (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Tenant Configs';
INSERT IGNORE INTO ops_tenant_configs (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 55, 'descri_1', 'type_A', '2025-09-21 17:00:00', 'active'),
('Beta-2', 65, 'descri_2', 'type_B', '2025-06-10 12:00:00', 'done'),
('Gamma-3', 100, 'descri_3', 'type_C', '2025-03-06 14:00:00', 'active'),
('Delta-4', 90, 'descri_4', 'type_D', '2025-09-14 14:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_feature_entitlements (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Feature Entitlements';
INSERT IGNORE INTO ops_feature_entitlements (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 2, 'descri_1', 'type_A', '2025-01-07 21:00:00', 'active'),
('Beta-2', 49, 'descri_2', 'type_B', '2025-10-18 03:00:00', 'pending'),
('Gamma-3', 47, 'descri_3', 'type_C', '2025-11-06 08:00:00', 'active'),
('Delta-4', 71, 'descri_4', 'type_D', '2025-05-23 07:00:00', 'pending');

CREATE TABLE IF NOT EXISTS ops_usage_metering (
    id                           INT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    name                         VARCHAR(120) COMMENT 'Name',
    batch_jobs_id                INT COMMENT 'FK to ops_batch_jobs' COMMENT 'Ref ops_batch_jobs',
    description                  VARCHAR(300) COMMENT 'Description',
    category                     VARCHAR(50) COMMENT 'Category',
    created_at                   DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Created',
    status                       VARCHAR(20) DEFAULT 'active' COMMENT 'Status',
    FOREIGN KEY (batch_jobs_id) REFERENCES ops_batch_jobs(id)
) COMMENT='Operations — Usage Metering';
INSERT IGNORE INTO ops_usage_metering (name, batch_jobs_id, description, category, created_at, status) VALUES
('Alpha-1', 79, 'descri_1', 'type_A', '2025-06-08 11:00:00', 'active'),
('Beta-2', 64, 'descri_2', 'type_B', '2025-08-22 06:00:00', 'pending'),
('Gamma-3', 90, 'descri_3', 'type_C', '2025-07-23 11:00:00', 'pending');



-- Update registration with correct table count
USE lucid;
UPDATE rc_datasources SET description = 'TPC-H Enterprise — 517-table enterprise database for Large-Scale Adaptive Schema Linking demo'
WHERE name = 'tpch_enterprise';
