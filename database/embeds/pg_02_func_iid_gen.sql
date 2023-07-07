CREATE SEQUENCE IF NOT EXISTS global_iid_seq;

-- fixed from https://instagram-engineering.com/sharding-ids-at-instagram-1cf5a71e5a5c
-- 43 bits for time in milliseconds.
-- 10 bits that represent the shard id.
-- 11 bits that represent an auto-incrementing sequence.
CREATE OR REPLACE FUNCTION iid_generator(shard_id int)
    RETURNS bigint
    LANGUAGE 'plpgsql'
AS $$
DECLARE
    our_epoch bigint := 1451606400000; -- 2016-01-01 00:00:00 +0000 UTC
    seq_id bigint;
    now_millis bigint;
    result bigint := 0;
BEGIN
    SELECT nextval('global_iid_seq') % 2048 INTO seq_id;
    shard_id := shard_id % 1024;

    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;
    result := (now_millis - our_epoch) << 21;
    result := result | (shard_id << 11);
    result := result | (seq_id);
    return result;
END;
$$;

-- SELECT iid_encode(iid_generator(0)), iid_encode(iid_generator(1)), iid_encode(iid_generator(6));
