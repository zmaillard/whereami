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
	//{"mod_spatialite.so.8.1.0", "sqlite3_modspatialite_init"},
	{"mod_spatialite.dylib", "sqlite3_modspatialite_init"},
	//{"libspatialite.so", "spatialite_init_ex"},
	//{"libspatialite.so.7.1.2", "spatialite_init_ex"},
	//{"libspatialite.so.8", "spatialite_init_ex"},
	{"mod_spatialite", "sqlite3_modspatialite_init"},
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
				//return conn.LoadExtension(path.Join(libPath, entry.lib), entry.proc)
			}
			return fmt.Errorf("failed to load spatialite extension with any of the following entries: %v in %s", libNames, libPath)
		},
	})
}

type Database struct {
	db *sql.DB
	rt *rtreego.Rtree
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	db, err := sql.Open("spatialite", fmt.Sprintf("%s?cache=shared&_stmt_cache_size=20&mode=ro", cfg.DbPath))
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}
