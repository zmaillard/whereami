package db

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/mattn/go-sqlite3"
	"github.com/zmaillard/whereami/config"
	"github.com/zmaillard/whereami/queries"
	"github.com/zmaillard/whereami/templates/components"
	"github.com/zmaillard/whereami/util"

	"github.com/dhconnelly/rtreego"
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

func (d *Database) GetCounty(coords queries.Coordinates) (*County, error) {
	query := fmt.Sprintf("SELECT namelsad,statefp FROM county WHERE ST_CONTAINS(CastAutomagic(geom),%s)", GeomStringFromCoordinate(coords))

	row := d.db.QueryRow(query)
	var name, stfips string
	if err := row.Scan(&name, &stfips); err != nil {
		return nil, err
	}
	return &County{Name: name, StateName: util.StateFromStateFips(stfips)}, nil
}

type watersheds struct {
	CurrentHuc  string
	Tributaries []string
}

func (w *watersheds) SetResults(dto *components.ResultDto) {
	dto.Tributaries = w.Tributaries
	dto.CurrentHuc = w.CurrentHuc
}

func (d *Database) GetStream(coords queries.Coordinates) (queries.Result, error) {
	query := fmt.Sprintf("SELECT huc12 FROM wbdhu12 WHERE ST_CONTAINS(shape,%s)", GeomStringFromCoordinate(coords))

	row := d.db.QueryRow(query)
	var huc12 string
	if err := row.Scan(&huc12); err != nil {
		return nil, err
	}

	baseQuery := `WITH RECURSIVE next_huc(h,r) AS (
					SELECT '%s',1
					UNION
					SELECT WBDHU12.tohuc,r+1
					FROM WBDHU12, next_huc
					WHERE next_huc.h = WBDHU12.huc12)
					SELECT b.name FROM next_huc
					inner join WBDHU12 b on h = b.huc12
					order by r`
	hucQuery := fmt.Sprintf(baseQuery, huc12)
	rows, err := d.db.Query(hucQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tributaries []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}

		var cleanName string
		if strings.Contains(name, "-") {
			cleanName = strings.Split(name, "-")[1]
		} else {
			cleanName = name
		}

		if len(tributaries) == 0 || tributaries[len(tributaries)-1] != cleanName {
			tributaries = append(tributaries, cleanName)

		}
	}

	return &watersheds{Tributaries: tributaries, CurrentHuc: huc12}, nil

}

type Summit struct {
	Name      string
	Elevation float64
	Distance  float64
}

func (s *Summit) SetResults(dto *components.ResultDto) {
	dto.NearestSummit = s.Name
	dto.NearestSummitElevation = s.Elevation
	dto.NearestSummitDistance = s.Distance
}

func (d *Database) GetNearestSummit(coords queries.Coordinates) (queries.Result, error) {

	results := d.rt.NearestNeighbors(1, rtreego.Point{coords.Longitude(), coords.Latitude()}, elevationFilter(12000))
	if len(results) == 0 {
		return nil, nil
	}
	si := results[0].(*summitIndex)

	fmt.Println(si.Feature_id)

	//	query := fmt.Sprintf("SELECT b.feature_name, b.elevation, a.distance_m / 1000.0 AS dist_km FROM knn2 as a join gnis as b on (b.fid = a.fid) WHERE f_table_name = 'summit' AND ref_geometry = MakePoint(%v,%v)  AND  radius = 1.0", coords.Longitude(), coords.Latitude())
	query := fmt.Sprintf("SELECT b.feature_name, b.elevation, a.distance_m / 1000.0 AS dist_km FROM knn2 as a join gnis as b on (b.fid = a.fid) WHERE f_table_name = 'summit' AND ref_geometry = MakePoint(%v,%v)  AND  radius = 1.0", coords.Longitude(), coords.Latitude())

	row := d.db.QueryRow(query)
	var name string
	var elevation, distance float64
	if err := row.Scan(&name, &elevation, &distance); err != nil {
		return nil, err
	}
	return &Summit{Name: name, Elevation: elevation, Distance: distance}, nil
}

func (d *Database) LoadSummitTree() error {
	query := fmt.Sprintf("SELECT ST_X(geom), ST_Y(geom), feature_id, elevation FROM summit")

	d.rt = rtreego.NewTree(2, 25, 50)

	rows, err := d.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var x, y, elevation float64
		var feature_id int
		if err := rows.Scan(&x, &y, &feature_id, &elevation); err != nil {
			return err
		}

		d.rt.Insert(&summitIndex{rtreego.Point{x, y}, elevation, feature_id})
	}

	return nil
}

var tol = 0.01

type summitIndex struct {
	Location   rtreego.Point
	Elevation  float64
	Feature_id int
}

func (s *summitIndex) Bounds() rtreego.Rect {
	// define the bounds of s to be a rectangle centered at s.location
	// with side lengths 2 * tol:
	return s.Location.ToRect(tol)
}
func elevationFilter(elevation float64) rtreego.Filter {
	return func(results []rtreego.Spatial, object rtreego.Spatial) (refuse, abort bool) {
		if obj, ok := object.(*summitIndex); ok {
			return obj.Elevation < elevation, false
		}
		return false, false
	}
}
