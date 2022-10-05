package pgx

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/cupogo/andvari/utils/zlog"
)

func TestMain(m *testing.M) {
	lgr, _ := zap.NewDevelopment()
	defer func() {
		_ = lgr.Sync() // flushes buffer, if any
	}()
	sugar := lgr.Sugar()
	zlog.Set(sugar)

	ret := m.Run()

	os.Exit(ret)
}

// CREATE ROLE testing WITH LOGIN PASSWORD 'testing1';
// CREATE DATABASE testing WITH OWNER = testing ENCODING = 'UTF8';
// GRANT ALL PRIVILEGES ON DATABASE testing to testing;
func TestOpen(t *testing.T) {
	dsn := "postgres://testing@localhost"
	ftscfg := ""
	debug := true
	db, err := Open(dsn, ftscfg, debug)
	assert.Error(t, err)
	assert.Nil(t, db)

	dsn = "postgres://testing:testing1@localhost/postgres?sslmode=disable"
	db, err = Open(dsn, ftscfg, debug)
	assert.Error(t, err)
	assert.Nil(t, db)

	dsn = "postgres://testing:testing1@localhost/testing?sslmode=disable"
	db, err = Open(dsn, ftscfg, debug)
	assert.NoError(t, err)
	assert.NotNil(t, db)
}
