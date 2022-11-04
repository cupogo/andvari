
CREATE OR REPLACE FUNCTION b36_encode(IN digits bigint, IN min_width int = 0)
  RETURNS varchar AS $$
        DECLARE
			chars char[];
			ret varchar;
			val bigint;
		BEGIN
		chars := ARRAY['0','1','2','3','4','5','6','7','8','9'
			,'a','b','c','d','e','f','g','h','i','j','k','l','m'
			,'n','o','p','q','r','s','t','u','v','w','x','y','z'];
		val := digits;
		ret := '';
		IF val < 0 THEN
			val := val * -1;
		END IF;
		WHILE val != 0 LOOP
			ret := chars[(val % 36)+1] || ret;
			val := val / 36;
		END LOOP;

		IF min_width > 0 AND char_length(ret) < min_width THEN
			ret := lpad(ret, min_width, '0');
		END IF;

		RETURN ret;

END;
$$ LANGUAGE 'plpgsql' IMMUTABLE;


CREATE OR REPLACE FUNCTION b36_decode(IN b36 varchar)
  RETURNS bigint AS $$
        DECLARE
			a char[];
			ret bigint;
			i bigint;
			val bigint;
			chars varchar;
		BEGIN
		chars := '0123456789abcdefghijklmnopqrstuvwxyz';

		FOR i IN REVERSE char_length(b36)..1 LOOP
			a := a || substring(lower(b36) FROM i FOR 1)::char;
		END LOOP;
		i := 0;
		ret := 0;
		WHILE i < (array_length(a,1)) LOOP
			val := position(a[i+1] IN chars)-1;
			ret := ret + (val * (36 ^ i)::bigint);
			i := i + 1;
		END LOOP;

		RETURN ret;

END;
$$ LANGUAGE 'plpgsql' IMMUTABLE;


CREATE OR REPLACE FUNCTION iid_encode(IN digits bigint)
  RETURNS varchar AS $$
BEGIN

		RETURN b36_encode(digits, 10);

END;
$$ LANGUAGE 'plpgsql' IMMUTABLE;

CREATE OR REPLACE FUNCTION iid_decode(IN eiid varchar)
  RETURNS bigint AS $$
BEGIN

		IF eiid LIKE '__-____%' THEN
			RETURN b36_decode(substring(eiid from 4));
		ELSE
			RETURN b36_decode(eiid);
		END IF;

END;
$$ LANGUAGE 'plpgsql' IMMUTABLE;
