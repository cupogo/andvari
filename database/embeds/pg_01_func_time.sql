
-- return milliseconds of unix timestamp
CREATE OR REPLACE FUNCTION unix_ts_milli() RETURNS bigint
    AS 'SELECT (extract(epoch from now())*1000)::bigint;'
    LANGUAGE SQL
    IMMUTABLE
    RETURNS NULL ON NULL INPUT;

-- return milliseconds of unix timestamp
CREATE OR REPLACE FUNCTION unix_ts_milli(ts timestamp) RETURNS bigint
    AS 'SELECT (extract(epoch from ts)*1000)::bigint;'
    LANGUAGE SQL
    IMMUTABLE
    RETURNS NULL ON NULL INPUT;

-- return milliseconds of unix timestamp
CREATE OR REPLACE FUNCTION unix_ts_milli(ts timestamptz) RETURNS bigint
    AS 'SELECT (extract(epoch from ts)*1000)::bigint;'
    LANGUAGE SQL
    IMMUTABLE
    RETURNS NULL ON NULL INPUT;


