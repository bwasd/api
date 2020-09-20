DROP SCHEMA IF EXISTS test CASCADE;
CREATE SCHEMA test;

CREATE TABLE test.product (
    id      SERIAL PRIMARY KEY,
    sku     text UNIQUE NOT NULL,
    attrs   JSONB
);

INSERT INTO test.product (sku,attrs)
SELECT 'sku_' || i, json_build_object(
    'size',     'small',
    'grams',    (i % 20) || '',
    'width',    '3.141592653589793',
    'foo',      to_hex(1&i),
    'bar',      md5(random()::text),
    'baz',      md5(random()::text),
    'sec',      i || ''
) FROM
    generate_series(1,(1 << 20)) AS i;

INSERT INTO
    test.product (sku, attrs)
SELECT
    'sku_' || i, json_build_object(
    md5(random()::text) || '', md5(random()::text) || '',
    md5(random()::text) || '', md5(random()::text) || '',
    md5(random()::text) || '', md5(random()::text) || '',
    md5(random()::text) || '', md5(random()::text) || '',
) FROM
    generate_series(1,(1 << 20) AS i;

INSERT INTO
    test.product (sku, attrs)
SELECT 'sku_' || i, json_build_object(
    md5(random()::text) || '', md5(random()::text) || '',
    md5(random()::text) || '', md5(random()::text) || '',
    md5(random()::text) || '', md5(random()::text) || '',
    md5(random()::text) || '', md5(random()::text) || '',
) FROM
    generate_series(1,(1 << 24) AS i;

-- Update attribute field
UPDATE
    test.product
SET
    attrs = jsonb_set(attrs, '{"grams"}', '"100"')
WHERE
    sku = sku_md5(120);

-- Lookup SKU by attribute key-value
SELECT
    sku
FROM
    test.product
WHERE
    attrs ->> 'grams' = 100 % 20 || '';

-- These JSONB Unicode regression tests were copied from postgres's upstream
-- mirror on Github:
-- github.com/postgres/postgres/blob/bbd93667bd/src/test/regress/sql/jsonb.sql

-- Basic Unicode input

-- ERROR, incomplete escape
SELECT '"\u"'::jsonb;
-- ERROR, incomplete escape
SELECT '"\u00"'::jsonb;
-- ERROR, g is not a hex digit
SELECT '"\u000g"'::jsonb;
-- OK, legal escape
SELECT '"\u0045"'::jsonb;
-- ERROR, we don't support U+0000
SELECT '"\u0000"'::jsonb;

-- Use octet_length here so we don't get an odd unicode char in the output
-- OK, uppercase and lower case both OK
SELECT
    octet_length('"\uaBcD"'::jsonb::text);

-- Handling of Unicode surrogate pairs
SELECT
    octet_length((jsonb '{ "a":  "\ud83d\ude04\ud83d\udc36" }' -> 'a')::text)
AS correct_in_utf8;

-- 2 high surrogates in a row
SELECT
    jsonb '{ "a":  "\ud83d\ud83d" }' -> 'a';

-- Surrogates in wrong order
SELECT
    jsonb '{ "a":  "\ude04\ud83d" }' -> 'a';

-- Orphan high surrogate
SELECT
    jsonb '{ "a":  "\ud83dX" }' -> 'a';

-- Orphan low surrogate
SELECT
    jsonb '{ "a":  "\ude04X" }' -> 'a';

-- Simple Unicode escape handling cases
SELECT
    jsonb'{ "a":  "the Copyright \u00a9 sign" }'
AS correct_in_utf8;

SELECT
    jsonb '{ "a":  "dollar \u0024 character" }'
AS correct_everywhere;

SELECT
    jsonb '{ "a":  "dollar \\u0024 character" }'
AS not_an_escape;

SELECT
    jsonb '{ "a":  "null \u0000 escape" }'
AS fails;

SELECT
    jsonb '{ "a":  "null \\u0000 escape" }'
AS not_an_escape;

SELECT
    jsonb '{ "a":  "the Copyright \u00a9 sign" }' ->> 'a'
AS correct_in_utf8;

SELECT
    jsonb '{ "a":  "dollar \u0024 character" }' ->> 'a'
AS correct_everywhere;

SELECT
    jsonb '{ "a":  "dollar \\u0024 character" }' ->> 'a'
AS not_an_escape;

SELECT
    jsonb '{ "a":  "null \u0000 escape" }' ->> 'a'
AS fails;

SELECT
    jsonb '{ "a":  "null \\u0000 escape" }' ->> 'a'
AS not_an_escape;

-- Theory
-- Generalized Inverted Indexes (GIN) are
-- https://www.postgresql.org/docs/9.1/textsearch-indexes.html
-- Inverted Indexes
-- https://en.wikipedia.org/wiki/Inverted_index
