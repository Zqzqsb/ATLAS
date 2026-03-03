-- =============================================================
-- LUCID TPC-H Enterprise Demo Database
-- 用于展示两阶段 Adaptive Schema Linking（>30 表触发 LargeScale）
-- TPC-H 核心 8 表 + 30 企业扩展表 = 38 表
-- =============================================================

CREATE DATABASE IF NOT EXISTS tpch_enterprise DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE tpch_enterprise;

-- =============================================================
-- PART I: TPC-H Core Tables (8 tables)
-- 标准 TPC-H schema，SF=0.01 级别样例数据
-- =============================================================

-- Region (5 rows, fixed)
CREATE TABLE region (
    r_regionkey  INTEGER NOT NULL PRIMARY KEY,
    r_name       CHAR(25) NOT NULL COMMENT '区域名称',
    r_comment    VARCHAR(152) COMMENT '备注'
) COMMENT='TPC-H 区域表 — 全球五大区域';

INSERT INTO region VALUES
(0, 'AFRICA',      'Africa continent'),
(1, 'AMERICA',     'North and South America'),
(2, 'ASIA',        'Asia-Pacific region'),
(3, 'EUROPE',      'European countries'),
(4, 'MIDDLE EAST', 'Middle East region');

-- Nation (25 rows, fixed)
CREATE TABLE nation (
    n_nationkey  INTEGER NOT NULL PRIMARY KEY,
    n_name       CHAR(25) NOT NULL COMMENT '国家名称',
    n_regionkey  INTEGER NOT NULL COMMENT '所属区域',
    n_comment    VARCHAR(152) COMMENT '备注',
    FOREIGN KEY (n_regionkey) REFERENCES region(r_regionkey)
) COMMENT='TPC-H 国家表 — 25个国家';

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
    s_name       CHAR(25) NOT NULL COMMENT '供应商名称',
    s_address    VARCHAR(40) NOT NULL COMMENT '供应商地址',
    s_nationkey  INTEGER NOT NULL COMMENT '所在国家',
    s_phone      CHAR(15) NOT NULL COMMENT '联系电话',
    s_acctbal    DECIMAL(15,2) NOT NULL COMMENT '账户余额',
    s_comment    VARCHAR(101) COMMENT '备注',
    FOREIGN KEY (s_nationkey) REFERENCES nation(n_nationkey)
) COMMENT='TPC-H 供应商表 — 零部件供应商信息';

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
    p_name        VARCHAR(55) NOT NULL COMMENT '零件名称',
    p_mfgr        CHAR(25) NOT NULL COMMENT '制造商',
    p_brand       CHAR(10) NOT NULL COMMENT '品牌',
    p_type        VARCHAR(25) NOT NULL COMMENT '零件类型',
    p_size        INTEGER NOT NULL COMMENT '尺寸',
    p_container   CHAR(10) NOT NULL COMMENT '包装容器类型',
    p_retailprice DECIMAL(15,2) NOT NULL COMMENT '零售价格',
    p_comment     VARCHAR(23) COMMENT '备注'
) COMMENT='TPC-H 零件表 — 供应链中的零部件目录';

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
    ps_partkey    INTEGER NOT NULL COMMENT '零件编号',
    ps_suppkey    INTEGER NOT NULL COMMENT '供应商编号',
    ps_availqty   INTEGER NOT NULL COMMENT '可用库存数量',
    ps_supplycost DECIMAL(15,2) NOT NULL COMMENT '供应成本',
    ps_comment    VARCHAR(199) COMMENT '备注',
    PRIMARY KEY (ps_partkey, ps_suppkey),
    FOREIGN KEY (ps_partkey) REFERENCES part(p_partkey),
    FOREIGN KEY (ps_suppkey) REFERENCES supplier(s_suppkey)
) COMMENT='TPC-H 供应关系表 — 零件与供应商的多对多关系及成本';

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
    c_name       VARCHAR(25) NOT NULL COMMENT '客户名称',
    c_address    VARCHAR(40) NOT NULL COMMENT '客户地址',
    c_nationkey  INTEGER NOT NULL COMMENT '所在国家',
    c_phone      CHAR(15) NOT NULL COMMENT '联系电话',
    c_acctbal    DECIMAL(15,2) NOT NULL COMMENT '账户余额',
    c_mktsegment CHAR(10) NOT NULL COMMENT '市场分类 (AUTOMOBILE/BUILDING/FURNITURE/HOUSEHOLD/MACHINERY)',
    c_comment    VARCHAR(117) COMMENT '备注',
    FOREIGN KEY (c_nationkey) REFERENCES nation(n_nationkey)
) COMMENT='TPC-H 客户表 — 下单客户信息';

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
    o_custkey       INTEGER NOT NULL COMMENT '客户编号',
    o_orderstatus   CHAR(1) NOT NULL COMMENT '订单状态 (O=open, F=filled, P=partial)',
    o_totalprice    DECIMAL(15,2) NOT NULL COMMENT '订单总金额',
    o_orderdate     DATE NOT NULL COMMENT '下单日期',
    o_orderpriority CHAR(15) NOT NULL COMMENT '订单优先级 (1-URGENT, 2-HIGH, 3-MEDIUM, 4-NOT SPECIFIED, 5-LOW)',
    o_clerk         CHAR(15) NOT NULL COMMENT '处理员工编号',
    o_shippriority  INTEGER NOT NULL COMMENT '发货优先级',
    o_comment       VARCHAR(79) COMMENT '备注',
    FOREIGN KEY (o_custkey) REFERENCES customer(c_custkey)
) COMMENT='TPC-H 订单表 — 客户采购订单';

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
    l_orderkey     INTEGER NOT NULL COMMENT '订单编号',
    l_partkey      INTEGER NOT NULL COMMENT '零件编号',
    l_suppkey      INTEGER NOT NULL COMMENT '供应商编号',
    l_linenumber   INTEGER NOT NULL COMMENT '行项目号',
    l_quantity     DECIMAL(15,2) NOT NULL COMMENT '数量',
    l_extendedprice DECIMAL(15,2) NOT NULL COMMENT '金额 = 数量 × 单价',
    l_discount     DECIMAL(15,2) NOT NULL COMMENT '折扣率 (0.00-0.10)',
    l_tax          DECIMAL(15,2) NOT NULL COMMENT '税率',
    l_returnflag   CHAR(1) NOT NULL COMMENT '退货标记 (A=accepted, R=returned, N=none)',
    l_linestatus   CHAR(1) NOT NULL COMMENT '行状态 (O=open, F=filled)',
    l_shipdate     DATE NOT NULL COMMENT '发货日期',
    l_commitdate   DATE NOT NULL COMMENT '承诺交期',
    l_receiptdate  DATE NOT NULL COMMENT '实际收货日期',
    l_shipinstruct CHAR(25) NOT NULL COMMENT '发货指示',
    l_shipmode     CHAR(10) NOT NULL COMMENT '运输方式 (REG AIR/AIR/RAIL/SHIP/TRUCK/MAIL/FOB)',
    l_comment      VARCHAR(44) COMMENT '备注',
    PRIMARY KEY (l_orderkey, l_linenumber),
    FOREIGN KEY (l_orderkey) REFERENCES orders(o_orderkey)
) COMMENT='TPC-H 订单明细表 — 每笔订单的行项目，包含价格、数量、折扣、运输信息';

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
-- 模拟真实企业场景，设计干扰表以展示精排价值
-- =============================================================

-- ─── Module: HR (5 tables) — 纯干扰模块 ───

CREATE TABLE departments (
    dept_id      INT PRIMARY KEY AUTO_INCREMENT,
    dept_name    VARCHAR(100) NOT NULL COMMENT '部门名称',
    location     VARCHAR(100) COMMENT '办公地点',
    budget       DECIMAL(15,2) COMMENT '部门年度预算',
    manager_name VARCHAR(100) COMMENT '部门经理'
) COMMENT='人力资源 — 部门信息表';

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
    emp_name     VARCHAR(100) NOT NULL COMMENT '员工姓名',
    dept_id      INT COMMENT '所属部门',
    title        VARCHAR(100) COMMENT '职位',
    hire_date    DATE COMMENT '入职日期',
    salary       DECIMAL(12,2) COMMENT '月薪',
    email        VARCHAR(255) COMMENT '工作邮箱',
    phone        VARCHAR(20) COMMENT '联系电话',
    status       VARCHAR(20) DEFAULT 'active' COMMENT '在职状态 (active/inactive/leave)',
    FOREIGN KEY (dept_id) REFERENCES departments(dept_id)
) COMMENT='人力资源 — 员工信息表';

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
    emp_id     INT NOT NULL COMMENT '员工编号',
    base_pay   DECIMAL(12,2) COMMENT '基本工资',
    bonus      DECIMAL(12,2) COMMENT '奖金',
    effective_date DATE COMMENT '生效日期',
    FOREIGN KEY (emp_id) REFERENCES employees(emp_id)
) COMMENT='人力资源 — 薪资记录表';

INSERT INTO salaries (emp_id, base_pay, bonus, effective_date) VALUES
(1, 35000.00, 15000.00, '2024-01-01'),
(2, 55000.00, 30000.00, '2024-01-01'),
(3, 28000.00, 10000.00, '2024-01-01'),
(4, 38000.00, 12000.00, '2024-01-01'),
(5, 25000.00, 8000.00,  '2024-01-01'),
(8, 65000.00, 50000.00, '2024-01-01');

CREATE TABLE job_history (
    history_id   INT PRIMARY KEY AUTO_INCREMENT,
    emp_id       INT NOT NULL COMMENT '员工编号',
    old_title    VARCHAR(100) COMMENT '原职位',
    new_title    VARCHAR(100) COMMENT '新职位',
    change_date  DATE COMMENT '变更日期',
    reason       VARCHAR(200) COMMENT '变更原因',
    FOREIGN KEY (emp_id) REFERENCES employees(emp_id)
) COMMENT='人力资源 — 职位变更历史';

INSERT INTO job_history (emp_id, old_title, new_title, change_date, reason) VALUES
(1, 'Sales Manager',      'Sales Director',  '2022-06-01', 'Promotion'),
(2, 'Engineering Manager', 'VP Engineering',  '2021-01-01', 'Promotion'),
(8, 'VP Engineering',      'CTO',             '2020-01-01', 'Promotion');

CREATE TABLE performance_reviews (
    review_id   INT PRIMARY KEY AUTO_INCREMENT,
    emp_id      INT NOT NULL COMMENT '员工编号',
    review_date DATE COMMENT '评审日期',
    rating      INT COMMENT '评分 (1-5分)',
    reviewer    VARCHAR(100) COMMENT '评审人',
    comments    TEXT COMMENT '评语',
    FOREIGN KEY (emp_id) REFERENCES employees(emp_id)
) COMMENT='人力资源 — 员工绩效评审记录';

INSERT INTO performance_reviews (emp_id, review_date, rating, reviewer, comments) VALUES
(1,  '2024-06-01', 4, 'Wu Hao',    'Strong sales leadership, exceeded Q1 targets'),
(2,  '2024-06-01', 5, 'Wu Hao',    'Exceptional technical vision and team management'),
(3,  '2024-06-01', 3, 'Wu Hao',    'Good campaign results, needs improvement in ROI tracking'),
(9,  '2024-06-01', 4, 'Zhang Wei', 'Promising new hire, quick learner'),
(10, '2024-06-01', 5, 'Li Ming',   'Outstanding code quality and mentorship');


-- ─── Module: CRM (5 tables) — 半相关（customer_feedback 关联 customer） ───

CREATE TABLE contacts (
    contact_id   INT PRIMARY KEY AUTO_INCREMENT,
    contact_name VARCHAR(100) NOT NULL COMMENT '联系人姓名',
    company      VARCHAR(200) COMMENT '公司名称',
    email        VARCHAR(255) COMMENT '邮箱',
    phone        VARCHAR(20) COMMENT '电话',
    vip_flag     TINYINT(1) DEFAULT 0 COMMENT 'VIP标记',
    source       VARCHAR(50) COMMENT '来源渠道 (web/referral/trade_show/cold_call)',
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
) COMMENT='CRM — 潜在客户联系人信息';

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
    contact_id  INT COMMENT '联系人编号',
    priority    VARCHAR(20) COMMENT '优先级 (hot/warm/cold)',
    deal_value  DECIMAL(15,2) COMMENT '预估成交额',
    status      VARCHAR(30) COMMENT '状态 (new/contacted/qualified/proposal/won/lost)',
    assigned_to INT COMMENT '负责销售员工',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (contact_id) REFERENCES contacts(contact_id),
    FOREIGN KEY (assigned_to) REFERENCES employees(emp_id)
) COMMENT='CRM — 销售线索跟踪';

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
    lead_id      INT COMMENT '来源线索',
    opp_name     VARCHAR(200) COMMENT '机会名称',
    stage        VARCHAR(30) COMMENT '阶段 (prospecting/analysis/proposal/negotiation/closed_won/closed_lost)',
    amount       DECIMAL(15,2) COMMENT '金额',
    close_date   DATE COMMENT '预计关闭日期',
    probability  INT COMMENT '成交概率 (%)',
    FOREIGN KEY (lead_id) REFERENCES leads(lead_id)
) COMMENT='CRM — 商机管理';

INSERT INTO opportunities (lead_id, opp_name, stage, amount, close_date, probability) VALUES
(1, 'TechCorp Server Upgrade',       'negotiation', 500000.00, '2025-03-15', 80),
(3, 'AutoParts Bulk Order',          'proposal',    800000.00, '2025-04-01', 60),
(5, 'FurniMax Annual Contract',      'negotiation', 350000.00, '2025-02-28', 75),
(7, 'Pacific Trading Partnership',   'closed_won', 1000000.00, '2025-01-15', 100);

CREATE TABLE campaigns (
    campaign_id   INT PRIMARY KEY AUTO_INCREMENT,
    campaign_name VARCHAR(200) NOT NULL COMMENT '活动名称',
    campaign_type VARCHAR(50) COMMENT '类型 (email/social/event/webinar)',
    start_date    DATE COMMENT '开始日期',
    end_date      DATE COMMENT '结束日期',
    budget        DECIMAL(12,2) COMMENT '活动预算',
    actual_cost   DECIMAL(12,2) COMMENT '实际花费',
    leads_generated INT DEFAULT 0 COMMENT '产生的线索数'
) COMMENT='CRM — 营销活动';

INSERT INTO campaigns (campaign_name, campaign_type, start_date, end_date, budget, actual_cost, leads_generated) VALUES
('2024 Spring Promotion',     'email',   '2024-03-01', '2024-03-31', 50000.00, 42000.00, 120),
('Industry Trade Show Q2',    'event',   '2024-06-15', '2024-06-18', 200000.00, 185000.00, 85),
('Social Media Campaign',     'social',  '2024-04-01', '2024-06-30', 30000.00, 28500.00, 200),
('Product Launch Webinar',    'webinar', '2024-09-10', '2024-09-10', 15000.00, 12000.00, 65),
('Year-end Customer Event',   'event',   '2024-12-10', '2024-12-12', 150000.00, 145000.00, 50);

CREATE TABLE customer_feedback (
    feedback_id   INT PRIMARY KEY AUTO_INCREMENT,
    customer_id   INT COMMENT '客户编号（关联TPC-H customer表）',
    feedback_date DATE COMMENT '反馈日期',
    rating        INT COMMENT '评分 (1-5)',
    category      VARCHAR(50) COMMENT '反馈类别 (product_quality/delivery/service/pricing)',
    comment       TEXT COMMENT '详细反馈内容',
    resolved      TINYINT(1) DEFAULT 0 COMMENT '是否已解决'
) COMMENT='CRM — 客户反馈与评价（关联 TPC-H customer 表）';

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


-- ─── Module: Logistics (5 tables) — 真相关（与 orders/lineitem 联动） ───

CREATE TABLE warehouses (
    warehouse_id   INT PRIMARY KEY AUTO_INCREMENT,
    warehouse_name VARCHAR(100) NOT NULL COMMENT '仓库名称',
    region_name    VARCHAR(50) COMMENT '所在区域 (华东/华南/华北/西南/海外)',
    city           VARCHAR(50) COMMENT '城市',
    capacity       INT COMMENT '最大库容（货位数）',
    manager_id     INT COMMENT '仓库主管',
    FOREIGN KEY (manager_id) REFERENCES employees(emp_id)
) COMMENT='物流 — 仓库信息表';

INSERT INTO warehouses (warehouse_name, region_name, city, capacity, manager_id) VALUES
('Shanghai Main Warehouse',  '华东', 'Shanghai',  50000, 5),
('Beijing Distribution',     '华北', 'Beijing',   30000, 13),
('Shenzhen Export Hub',      '华南', 'Shenzhen',  40000, 5),
('Chengdu Regional',         '西南', 'Chengdu',   20000, 13),
('Hong Kong Transit',        '海外', 'Hong Kong', 15000, NULL),
('Frankfurt EU Warehouse',   '海外', 'Frankfurt', 25000, NULL);

CREATE TABLE inventory (
    inventory_id   INT PRIMARY KEY AUTO_INCREMENT,
    warehouse_id   INT NOT NULL COMMENT '仓库编号',
    part_id        INT NOT NULL COMMENT '零件编号（关联 TPC-H part 表）',
    quantity       INT NOT NULL DEFAULT 0 COMMENT '当前库存数量',
    min_stock      INT DEFAULT 10 COMMENT '最低安全库存',
    max_stock      INT DEFAULT 1000 COMMENT '最大库存',
    last_restock   DATE COMMENT '最近补货日期',
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(warehouse_id)
) COMMENT='物流 — 仓库库存（关联 TPC-H part 表，跟踪各零件在各仓库的库存水位）';

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
    method_name VARCHAR(50) NOT NULL COMMENT '运输方式名称',
    carrier     VARCHAR(100) COMMENT '承运商',
    avg_days    INT COMMENT '平均运输天数',
    cost_per_kg DECIMAL(8,2) COMMENT '每公斤运费'
) COMMENT='物流 — 运输方式与承运商';

INSERT INTO shipping_methods (method_name, carrier, avg_days, cost_per_kg) VALUES
('Express Air',     'FedEx',          2,  25.00),
('Standard Air',    'DHL',            4,  15.00),
('Ocean Freight',   'Maersk',        21,   3.50),
('Rail Freight',    'China Railway',  7,   6.00),
('Truck Domestic',  'SF Express',     3,   8.00),
('Economy Sea',     'COSCO',         30,   2.00);

CREATE TABLE shipments (
    shipment_id    INT PRIMARY KEY AUTO_INCREMENT,
    order_id       INT COMMENT '订单编号（关联 TPC-H orders 表）',
    warehouse_id   INT COMMENT '出库仓库',
    method_id      INT COMMENT '运输方式',
    ship_date      DATE COMMENT '发货日期',
    estimated_arrival DATE COMMENT '预计到达日期',
    actual_arrival    DATE COMMENT '实际到达日期',
    tracking_no    VARCHAR(50) COMMENT '物流跟踪号',
    status         VARCHAR(30) COMMENT '状态 (preparing/shipped/in_transit/delivered/returned)',
    FOREIGN KEY (warehouse_id) REFERENCES warehouses(warehouse_id),
    FOREIGN KEY (method_id) REFERENCES shipping_methods(method_id)
) COMMENT='物流 — 发货记录（关联 TPC-H orders 表）';

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
    shipment_id   INT NOT NULL COMMENT '发货记录编号',
    event_time    DATETIME COMMENT '事件时间',
    location      VARCHAR(200) COMMENT '当前位置',
    event_type    VARCHAR(50) COMMENT '事件类型 (pickup/departure/arrival/customs/delivered)',
    notes         VARCHAR(300) COMMENT '备注',
    FOREIGN KEY (shipment_id) REFERENCES shipments(shipment_id)
) COMMENT='物流 — 物流轨迹追踪';

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


-- ─── Module: Finance (5 tables) — 半相关（invoices 关联 orders） ───

CREATE TABLE accounts (
    account_id    INT PRIMARY KEY AUTO_INCREMENT,
    account_name  VARCHAR(100) NOT NULL COMMENT '科目名称',
    account_type  VARCHAR(30) COMMENT '科目类型 (asset/liability/equity/revenue/expense)',
    balance       DECIMAL(15,2) DEFAULT 0 COMMENT '当前余额',
    currency      CHAR(3) DEFAULT 'CNY' COMMENT '币种'
) COMMENT='财务 — 会计科目表';

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
    order_id      INT COMMENT '关联订单（TPC-H orders 表）',
    invoice_no    VARCHAR(30) NOT NULL COMMENT '发票编号',
    invoice_date  DATE COMMENT '开票日期',
    due_date      DATE COMMENT '到期日期',
    amount        DECIMAL(15,2) COMMENT '发票金额',
    tax_amount    DECIMAL(15,2) COMMENT '税额',
    status        VARCHAR(20) COMMENT '状态 (draft/sent/paid/overdue/cancelled)',
    payment_date  DATE COMMENT '付款日期'
) COMMENT='财务 — 发票管理（关联 TPC-H orders 表）';

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
    invoice_id     INT COMMENT '关联发票',
    payment_date   DATE COMMENT '付款日期',
    amount         DECIMAL(15,2) COMMENT '付款金额',
    payment_method VARCHAR(30) COMMENT '支付方式 (wire_transfer/check/credit_card/cash)',
    reference_no   VARCHAR(50) COMMENT '支付参考号',
    FOREIGN KEY (invoice_id) REFERENCES invoices(invoice_id)
) COMMENT='财务 — 收款记录';

INSERT INTO payments (invoice_id, payment_date, amount, payment_method, reference_no) VALUES
(1, '1996-01-28', 172799.49, 'wire_transfer', 'PAY-1996-0001'),
(2, '1996-12-30', 46929.18,  'wire_transfer', 'PAY-1996-0002'),
(3, '1993-11-10', 193846.25, 'check',         'PAY-1993-0003'),
(4, '1994-08-25', 77471.44,  'wire_transfer', 'PAY-1994-0005'),
(5, '1996-02-10', 252004.18, 'wire_transfer', 'PAY-1996-0007'),
(7, '1995-12-01', 173665.47, 'wire_transfer', 'PAY-1995-0036');

CREATE TABLE budget_items (
    budget_id     INT PRIMARY KEY AUTO_INCREMENT,
    dept_id       INT COMMENT '部门编号',
    fiscal_year   INT COMMENT '财年',
    category      VARCHAR(50) COMMENT '预算类别 (personnel/operations/equipment/travel/marketing)',
    planned       DECIMAL(15,2) COMMENT '计划金额',
    actual        DECIMAL(15,2) COMMENT '实际金额',
    variance      DECIMAL(15,2) COMMENT '偏差（实际-计划）',
    FOREIGN KEY (dept_id) REFERENCES departments(dept_id)
) COMMENT='财务 — 部门预算执行表';

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
    tax_name    VARCHAR(50) NOT NULL COMMENT '税种名称',
    rate        DECIMAL(6,4) COMMENT '税率',
    region      VARCHAR(50) COMMENT '适用地区',
    effective_from DATE COMMENT '生效日期',
    effective_to   DATE COMMENT '失效日期'
) COMMENT='财务 — 税率配置表';

INSERT INTO tax_rates (tax_name, rate, region, effective_from, effective_to) VALUES
('VAT Standard',       0.1300, 'China',         '2019-04-01', NULL),
('VAT Reduced',        0.0900, 'China',         '2019-04-01', NULL),
('Export Tax Rebate',  0.1300, 'China',         '2020-01-01', NULL),
('US Sales Tax',       0.0875, 'United States', '2020-01-01', NULL),
('EU VAT Standard',    0.2000, 'Europe',        '2020-01-01', NULL),
('Import Duty',        0.0500, 'China',         '2020-01-01', NULL);


-- ─── Module: Product (4 tables) — 半相关（关联 part） ───

CREATE TABLE product_categories (
    category_id   INT PRIMARY KEY AUTO_INCREMENT,
    category_name VARCHAR(100) NOT NULL COMMENT '产品类别名称',
    parent_id     INT COMMENT '上级类别',
    description   TEXT COMMENT '类别描述',
    FOREIGN KEY (parent_id) REFERENCES product_categories(category_id)
) COMMENT='产品 — 产品分类目录（树形结构）';

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
    part_id       INT COMMENT '零件编号（关联 TPC-H part 表）',
    reviewer_name VARCHAR(100) COMMENT '评价人',
    rating        INT COMMENT '评分 (1-5)',
    review_date   DATE COMMENT '评价日期',
    title         VARCHAR(200) COMMENT '评价标题',
    content       TEXT COMMENT '评价内容',
    verified      TINYINT(1) DEFAULT 0 COMMENT '是否已验证购买'
) COMMENT='产品 — 零件评价记录（关联 TPC-H part 表）';

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
    part_id     INT COMMENT '零件编号（关联 TPC-H part 表）',
    old_price   DECIMAL(15,2) COMMENT '原价',
    new_price   DECIMAL(15,2) COMMENT '新价',
    change_date DATE COMMENT '调价日期',
    reason      VARCHAR(200) COMMENT '调价原因'
) COMMENT='产品 — 零件价格变更历史';

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
    promo_name    VARCHAR(200) NOT NULL COMMENT '促销活动名称',
    discount_pct  DECIMAL(5,2) COMMENT '折扣率 (%)',
    start_date    DATE COMMENT '开始日期',
    end_date      DATE COMMENT '结束日期',
    min_quantity  INT COMMENT '最低起订量',
    applicable_brands VARCHAR(200) COMMENT '适用品牌 (逗号分隔)',
    status        VARCHAR(20) DEFAULT 'active' COMMENT '状态 (active/expired/scheduled)'
) COMMENT='产品 — 促销活动配置';

INSERT INTO promotions (promo_name, discount_pct, start_date, end_date, min_quantity, applicable_brands, status) VALUES
('New Year Bulk Discount',   10.00, '2025-01-01', '2025-01-31', 100,  'Brand#13,Brand#42',     'active'),
('Spring Clearance',          15.00, '2025-03-01', '2025-03-31', 50,   'Brand#11,Brand#24',     'scheduled'),
('Loyalty Program 5%',        5.00,  '2024-01-01', '2025-12-31', 10,   NULL,                    'active'),
('Copper Parts Special',     12.00, '2025-02-01', '2025-02-28', 200,  'Brand#13,Brand#15',     'active'),
('Quarter-end Flash Sale',   20.00, '2025-03-25', '2025-03-31', 500,  NULL,                    'scheduled');


-- ─── Module: System (6 tables) — 纯干扰模块 ───

CREATE TABLE user_accounts (
    user_id       INT PRIMARY KEY AUTO_INCREMENT,
    username      VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    password_hash VARCHAR(255) COMMENT '密码哈希',
    role          VARCHAR(30) COMMENT '角色 (admin/manager/analyst/viewer)',
    emp_id        INT COMMENT '关联员工',
    last_login    DATETIME COMMENT '最后登录时间',
    is_active     TINYINT(1) DEFAULT 1 COMMENT '是否启用',
    FOREIGN KEY (emp_id) REFERENCES employees(emp_id)
) COMMENT='系统 — 用户账户';

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
    user_id     INT COMMENT '操作用户',
    action      VARCHAR(50) COMMENT '操作类型 (login/logout/query/export/update/delete)',
    target      VARCHAR(200) COMMENT '操作对象',
    details     TEXT COMMENT '详细信息',
    ip_address  VARCHAR(45) COMMENT 'IP地址',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user_accounts(user_id)
) COMMENT='系统 — 操作审计日志';

INSERT INTO audit_log (user_id, action, target, details, ip_address, created_at) VALUES
(1, 'login',  'system',             'Admin login',                    '192.168.1.100', '2025-01-20 09:30:00'),
(2, 'query',  'orders',             'SELECT * FROM orders WHERE ...',  '192.168.1.101', '2025-01-20 08:20:00'),
(2, 'export', 'sales_report_Q4',    'Exported Q4 sales data to CSV',   '192.168.1.101', '2025-01-20 08:45:00'),
(3, 'login',  'system',             'Engineering login',               '192.168.1.102', '2025-01-19 17:45:00'),
(4, 'update', 'budget_items',       'Updated Q1 2025 budget forecast', '192.168.1.103', '2025-01-20 10:15:00'),
(6, 'query',  'inventory',          'Inventory status check',          '192.168.1.104', '2025-01-20 11:05:00');

CREATE TABLE system_config (
    config_id    INT PRIMARY KEY AUTO_INCREMENT,
    config_key   VARCHAR(100) NOT NULL UNIQUE COMMENT '配置项',
    config_value TEXT COMMENT '配置值',
    description  VARCHAR(300) COMMENT '说明',
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
) COMMENT='系统 — 全局配置表';

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
    user_id         INT COMMENT '接收用户',
    title           VARCHAR(200) COMMENT '通知标题',
    message         TEXT COMMENT '通知内容',
    is_read         TINYINT(1) DEFAULT 0 COMMENT '是否已读',
    priority        VARCHAR(20) DEFAULT 'normal' COMMENT '优先级 (low/normal/high/urgent)',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user_accounts(user_id)
) COMMENT='系统 — 消息通知';

INSERT INTO notifications (user_id, title, message, is_read, priority, created_at) VALUES
(2, 'Q4 Sales Target Achieved',    'Congratulations! Q4 sales exceeded target by 12%.', 1, 'normal',  '2025-01-15 09:00:00'),
(5, 'Low Inventory Alert',         'Parts #2, #5, #7, #8 below minimum stock levels.',  0, 'urgent',  '2025-01-18 07:00:00'),
(4, 'Budget Review Due',           'FY2025 Q1 budget review deadline: Jan 25.',          0, 'high',    '2025-01-20 08:00:00'),
(1, 'New System Update Available', 'ERP v3.2 update scheduled for Jan 22.',              1, 'low',     '2025-01-19 12:00:00'),
(3, 'Engineering Headcount Approved', 'Two new senior engineer positions approved.',     1, 'normal',  '2025-01-17 14:00:00');

CREATE TABLE user_sessions (
    session_id   VARCHAR(64) PRIMARY KEY COMMENT '会话ID',
    user_id      INT COMMENT '用户编号',
    login_time   DATETIME COMMENT '登录时间',
    logout_time  DATETIME COMMENT '登出时间',
    ip_address   VARCHAR(45) COMMENT 'IP地址',
    user_agent   VARCHAR(300) COMMENT '浏览器标识',
    FOREIGN KEY (user_id) REFERENCES user_accounts(user_id)
) COMMENT='系统 — 用户会话记录';

INSERT INTO user_sessions VALUES
('sess_abc123', 1, '2025-01-20 09:30:00', NULL, '192.168.1.100', 'Mozilla/5.0 Chrome/120'),
('sess_def456', 2, '2025-01-20 08:15:00', '2025-01-20 12:30:00', '192.168.1.101', 'Mozilla/5.0 Firefox/121'),
('sess_ghi789', 5, '2025-01-18 14:20:00', '2025-01-18 18:00:00', '192.168.1.104', 'Mozilla/5.0 Chrome/120');

CREATE TABLE data_exports (
    export_id    INT PRIMARY KEY AUTO_INCREMENT,
    user_id      INT COMMENT '导出用户',
    export_type  VARCHAR(30) COMMENT '导出格式 (csv/excel/pdf)',
    table_name   VARCHAR(100) COMMENT '导出数据源',
    row_count    INT COMMENT '导出行数',
    file_size_kb INT COMMENT '文件大小 (KB)',
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user_accounts(user_id)
) COMMENT='系统 — 数据导出记录';

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
        'TPC-H Enterprise — 38-table supply chain database for Large-Scale Adaptive Schema Linking demo', 'active')
ON DUPLICATE KEY UPDATE status = 'active', description = VALUES(description);
