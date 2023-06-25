
CREATE OR REPLACE FUNCTION op_reset_table_id(_table text)
RETURNS int AS $$
DECLARE
	max_id int;
	next_id int;
BEGIN

	EXECUTE format('SELECT max(id) FROM %I', _table)
	INTO max_id;

	IF max_id IS NULL THEN
		next_id := 1;
	ELSE
		next_id := max_id+1;
	END IF;

	EXECUTE format('SELECT setval(%L, $1, false)', _table || '_id_seq')
	USING next_id;

	RAISE NOTICE 'next %.id: %', _table, next_id;

	RETURN next_id;
END $$

LANGUAGE 'plpgsql' VOLATILE;
