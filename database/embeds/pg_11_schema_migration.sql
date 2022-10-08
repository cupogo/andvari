
-- go-pg/migrations 用来维护数据库升级的版本历史 -- deprecated
CREATE TABLE IF NOT EXISTS gopg_migrations (
			id serial,
			version int,
			created_at timestamptz,
			PRIMARY KEY(id)
		);


-- bun/migrate
CREATE TABLE IF NOT EXISTS bun_migration_locks (
	id serial,
	table_name text NOT NULL UNIQUE,
	PRIMARY KEY(id)
);

-- bun/migrate
CREATE TABLE IF NOT EXISTS bun_migrations (
	id serial,
	name text,
	group_id int,
	migrated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY(id)
);
