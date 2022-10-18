


-- 移动删除 数字主键
CREATE OR REPLACE FUNCTION op_affect_delete(_sc_orig text, _sc_trash text, _table text, _id bigint)
RETURNS int AS
$BODY$
DECLARE
	tid bigint;
BEGIN

	IF NOT EXISTS(
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = _sc_orig AND table_name = _table) THEN
		RETURN -2;
	END IF;

	EXECUTE format('SELECT id FROM %I.%I WHERE id = $1 LIMIT 1', _sc_orig, _table)
	INTO tid
	USING _id;

	IF tid IS NULL THEN
		RAISE NOTICE 'ERR: record % NOT FOUND', _id;
		RETURN -1;
	END IF;

	EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', _sc_trash);

	IF NOT EXISTS(
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = _sc_trash AND table_name = _table) THEN
		EXECUTE format('CREATE TABLE IF NOT EXISTS %I.%I (
			LIKE %I.%I INCLUDING DEFAULTS, PRIMARY KEY (id)
			)', _sc_trash, _table, _sc_orig, _table)
		;
	ELSE
		EXECUTE format('SELECT id FROM %I.%I WHERE id = $1 LIMIT 1', _sc_trash, _table)
		INTO tid
		USING _id;
		IF tid IS NOT NULL THEN
			RAISE NOTICE 'WARN: deleted record % FOUND', _id;
			EXECUTE format('DELETE FROM %I.%I WHERE id = $1', _sc_trash, _table)
			USING _id;
			-- RETURN -3;
		END IF;

	END IF;

	EXECUTE format('INSERT INTO %I.%I SELECT * FROM %I.%I WHERE id = $1', _sc_trash, _table, _sc_orig, _table)
	USING _id;

	EXECUTE format('DELETE FROM %I.%I WHERE id = $1', _sc_orig, _table)
	USING _id;

	RETURN 1;

END;
$BODY$
LANGUAGE 'plpgsql' VOLATILE;

-- 移动删除 字串主键
CREATE OR REPLACE FUNCTION op_affect_delete(_sc_orig text, _sc_trash text, _table text, _id text)
RETURNS int AS
$BODY$
DECLARE
	tid text;
BEGIN

	IF NOT EXISTS(
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = _sc_orig AND table_name = _table) THEN
		RETURN -2;
	END IF;

	EXECUTE format('SELECT id FROM %I.%I WHERE id = $1 LIMIT 1', _sc_orig, _table)
	INTO tid
	USING _id;

	IF tid IS NULL THEN
		RAISE NOTICE 'ERR: record % NOT FOUND', _id;
		RETURN -1;
	END IF;

	EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', _sc_trash);

	IF NOT EXISTS(
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = _sc_trash AND table_name = _table) THEN
		EXECUTE format('CREATE TABLE IF NOT EXISTS %I.%I (
			LIKE %I.%I INCLUDING DEFAULTS, PRIMARY KEY (id)
			)', _sc_trash, _table, _sc_orig, _table)
		;
	ELSE
		EXECUTE format('SELECT id FROM %I.%I WHERE id = $1 LIMIT 1', _sc_trash, _table)
		INTO tid
		USING _id;
		IF tid IS NOT NULL THEN
			RAISE NOTICE 'WARN: deleted record % FOUND', _id;
			EXECUTE format('DELETE FROM %I.%I WHERE id = $1', _sc_trash, _table)
			USING _id;
			-- RETURN -3;
		END IF;

	END IF;

	EXECUTE format('INSERT INTO %I.%I SELECT * FROM %I.%I WHERE id = $1', _sc_trash, _table, _sc_orig, _table)
	USING _id;

	EXECUTE format('DELETE FROM %I.%I WHERE id = $1', _sc_orig, _table)
	USING _id;

	RETURN 1;

END;
$BODY$
LANGUAGE 'plpgsql' VOLATILE;

-- SELECT delete_with_move('aurora', 'aurora_deleted', 'terms', 'at-355nl8p4nnr6')


-- 移动删除恢复
CREATE OR REPLACE FUNCTION op_affect_undelete(_sc_orig text, _sc_trash text, _table text, _id bigint)
RETURNS int AS
$BODY$
DECLARE
	tid bigint;
BEGIN

	IF NOT EXISTS(
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = _sc_orig AND table_name = _table) THEN
		RETURN -2;
	END IF;

	IF NOT EXISTS(
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = _sc_trash AND table_name = _table) THEN
		RETURN -3;
	END IF;

	EXECUTE format('SELECT id FROM %I.%I WHERE id = $1 LIMIT 1', _sc_trash, _table)
	INTO tid
	USING _id;

	IF tid IS NULL THEN
		RAISE NOTICE 'ERR  record % NOT FOUND', _id;
		RETURN -1;
	END IF;

	EXECUTE format('SELECT id FROM %I.%I WHERE id = $1 LIMIT 1', _sc_orig, _table)
	INTO tid
	USING _id;

	IF tid IS NOT NULL THEN
		RAISE NOTICE 'ERR  record % EXISTS', _id;
		RETURN -1;
	END IF;

	EXECUTE format('INSERT INTO %I.%I SELECT * FROM %I.%I WHERE id = $1', _sc_orig, _table, _sc_trash, _table)
	USING _id;

	EXECUTE format('DELETE FROM %I.%I WHERE id = $1', _sc_trash, _table)
	USING _id;

	-- TODO:

	RETURN 1;

END;
$BODY$
LANGUAGE 'plpgsql' VOLATILE;
