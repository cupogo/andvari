

CREATE OR REPLACE FUNCTION ensure_ts_config(IN name varchar, IN parser varchar)
  RETURNS int AS $$
BEGIN
	IF EXISTS(SELECT oid FROM pg_ts_config WHERE cfgname = name) THEN
	    RETURN 0;
	END IF;

	IF NOT EXISTS(SELECT oid FROM pg_ts_parser WHERE prsname = parser) THEN
	    RETURN -1;
	END IF;

	EXECUTE 'CREATE TEXT SEARCH CONFIGURATION '|| quote_ident(name) ||' (PARSER = '|| quote_ident(parser) ||')';
	EXECUTE 'ALTER TEXT SEARCH CONFIGURATION '|| quote_ident(name) ||' ADD MAPPING FOR n,v,a,i,e,l WITH simple)';

	RETURN 1;

END;
$$ LANGUAGE 'plpgsql' VOLATILE;
