package queries

import (
	"database/sql"
	"fmt"
	"os"
	"path"

	"github.com/dhconnelly/rtreego"
	"github.com/mattn/go-sqlite3"
	"github.com/zmaillard/whereami/config"
)

type entrypoint struct {
	lib  string
	proc string
}

var libNames = []entrypoint{
	{"mod_spatialite", "sqlite3_modspatialite_init"},
	{"mod_spatialite.dylib", "sqlite3_modspatialite_init"},
	{"libspatialite.so", "sqlite3_modspatialite_init"},
	{"libspatialite.so.5", "spatialite_init_ex"},
	{"libspatialite.so", "spatialite_init_ex"},
}

var basePath = "/opt/homebrew/Cellar/libspatialite/5.1.0_4/lib"

func RegisterExtensions() {
	libPath, exists := os.LookupEnv("SPATIALITE_PATH")
	if !exists {
		libPath = basePath
	}

	sql.Register("spatialite", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			for _, entry := range libNames {
				if err := conn.LoadExtension(path.Join(libPath, entry.lib), entry.proc); err == nil {
					return nil
				}
			}
			return fmt.Errorf("failed to load spatialite extension with any of the following entries: %v", libNames)
		},
	})
}

type Database struct {
	db *sql.DB
	rt *rtreego.Rtree
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	db, err := sql.Open("spatialite", cfg.DbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}
