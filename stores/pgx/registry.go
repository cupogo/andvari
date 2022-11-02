package pgx

import "io/fs"

var (
	allmodels []any
	alldbfs   []fs.FS
	alterfs   []fs.FS
)

// RegisterModel all tables will be created by InitSchemas()
func RegisterModel(m ...any) {
	allmodels = append(allmodels, m...)
}

// Deprecated by RegisterInitFs()
func RegisterDbFs(dbfs ...fs.FS) { RegisterInitFs(dbfs...) }

// RegisterInitFs special sql files will be executed by InitSchemas()
func RegisterInitFs(dbfs ...fs.FS) {
	alldbfs = append(alldbfs, dbfs...)
}

// RegisterMigrateFs special sql files in FS will be executed by RunMigrations()
func RegisterMigrationFs(dbfs ...fs.FS) {
	alterfs = append(alterfs, dbfs...)
}
