package gograte

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ddlFile represents a Data Definition Language (DDL) file
// Given the file naming convention 001-user.sql, the numbers up to
// the first dash are extracted, converted to an int and added to the
// fileNumber field to make the struct sortable using the sort package.
type ddlFile struct {
	filename   string
	fileNumber int
}

// newDDLFile initializes a DDLFile struct. File naming convention
// should be 001-user.sql where 001 represents the file number order
// to be processed
func newDDLFile(f string) (ddlFile, error) {
	i := strings.Index(f, "-")
	fileNumber := f[:i]
	fn, err := strconv.Atoi(fileNumber)
	if err != nil {
		return ddlFile{}, err
	}

	return ddlFile{filename: f, fileNumber: fn}, nil
}

func (df ddlFile) String() string {
	return fmt.Sprintf("%s: %d", df.filename, df.fileNumber)
}

// readDDLFiles reads and returns sorted DDL files from the
// up or down directory
func readDDLFiles(dir string) (ddlFiles []ddlFile, err error) {

	var files []os.DirEntry
	files, err = os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		var df ddlFile
		df, err = newDDLFile(file.Name())
		if err != nil {
			return nil, err
		}
		ddlFiles = append(ddlFiles, df)
	}

	sort.Sort(byFileNumber(ddlFiles))

	return ddlFiles, nil
}

// byFileNumber implements sort.Interface for []ddlFile based on
// the fileNumber field.
type byFileNumber []ddlFile

// Len returns the length of elements in the ByFileNumber slice for sorting
func (bfn byFileNumber) Len() int { return len(bfn) }

// Swap sets up the elements to be swapped for the ByFileNumber slice for sorting
func (bfn byFileNumber) Swap(i, j int) { bfn[i], bfn[j] = bfn[j], bfn[i] }

// Less is the sorting logic for the ByFileNumber slice
func (bfn byFileNumber) Less(i, j int) bool { return bfn[i].fileNumber < bfn[j].fileNumber }

// PSQLArgs takes a slice of DDL files to be executed and builds a
// sequence of command line arguments using the appropriate flags
// psql needs to execute files. The arguments returned for psql are as follows:
//
// -w flag is set to never prompt for a password as we are running this as a script
//
// -d flag sets the database connection using a Connection URI string.
//
// -f flag is sent before each file to tell it to process the file
func PSQLArgs(up bool, profile string) ([]string, error) {

	var (
		f   ConfigFile
		err error
	)

	// regular config path - relative to project root
	configFilePath := "./config/" + profile + ".json"

	// read JSON config file
	f, err = NewConfigFile(configFilePath)
	if err != nil {
		return nil, err
	}

	// determine directory from config file
	dir := f.Config.MigrationScriptsDir
	if up {
		dir += "/up"
	} else {
		dir += "/down"
	}

	// readDDLFiles reads and returns sorted DDL files from the up or down directory
	var ddlFiles []ddlFile
	ddlFiles, err = readDDLFiles(dir)
	if err != nil {
		return nil, err
	}

	if len(ddlFiles) == 0 {
		return nil, fmt.Errorf("there are no DDL files to process in %s", dir)
	}

	// command line args for psql are constructed
	args := []string{"-w", "-d", newPostgreSQLDSN(f).ConnectionURI(), "-c", "select current_database(), current_user, version()"}

	for _, file := range ddlFiles {
		args = append(args, "-f")
		args = append(args, dir+"/"+file.filename)
	}

	return args, nil
}

// newPostgreSQLDSN initializes a datastore.PostgreSQLDSN given a Flags struct
func newPostgreSQLDSN(f ConfigFile) PostgreSQLDSN {
	return PostgreSQLDSN{
		Host:       f.Config.Database.Host,
		Port:       f.Config.Database.Port,
		DBName:     f.Config.Database.Name,
		SearchPath: f.Config.Database.SearchPath,
		User:       f.Config.Database.User,
		Password:   f.Config.Database.Password,
	}
}

// PostgreSQLDSN is a PostgreSQL datasource name
type PostgreSQLDSN struct {
	Host       string
	Port       int
	DBName     string
	SearchPath string
	User       string
	Password   string
}

// ConnectionURI returns a formatted PostgreSQL datasource "Keyword/Value Connection String"
// The general form for a connection URI is:
// postgresql://[userspec@][hostspec][/dbname][?paramspec]
// where userspec is
//
//	user[:password]
//
// and hostspec is:
//
//	[host][:port][,...]
//
// and paramspec is:
//
//	name=value[&...]
//
// The URI scheme designator can be either postgresql:// or postgres://.
// Each of the remaining URI parts is optional.
// The following examples illustrate valid URI syntax:
//
//	postgresql://
//	postgresql://localhost
//	postgresql://localhost:5433
//	postgresql://localhost/mydb
//	postgresql://user@localhost
//	postgresql://user:secret@localhost
//	postgresql://other@localhost/otherdb?connect_timeout=10&application_name=myapp
//	postgresql://host1:123,host2:456/somedb?target_session_attrs=any&application_name=myapp
func (dsn PostgreSQLDSN) ConnectionURI() string {

	const uriSchemeDesignator string = "postgresql"

	var h string
	h = dsn.Host
	if dsn.Port != 0 {
		h += ":" + strconv.Itoa(dsn.Port)
	}

	u := url.URL{
		Scheme: uriSchemeDesignator,
		User:   url.User(dsn.User),
		Host:   h,
		Path:   dsn.DBName,
	}

	if dsn.SearchPath != "" {
		q := u.Query()
		q.Set("options", fmt.Sprintf("-csearch_path=%s", dsn.SearchPath))
		u.RawQuery = q.Encode()
	}

	return u.String()
}

// KeywordValueConnectionString returns a formatted PostgreSQL datasource "Keyword/Value Connection String"
func (dsn PostgreSQLDSN) KeywordValueConnectionString() string {

	var s string

	// if db connection does not have a password (should only be for local testing and preferably never),
	// the password parameter must be removed from the string, otherwise the connection will fail.
	switch dsn.Password {
	case "":
		s = fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable", dsn.Host, dsn.Port, dsn.DBName, dsn.User)
	default:
		s = fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable", dsn.Host, dsn.Port, dsn.DBName, dsn.User, dsn.Password)
	}

	// if search path needs to be explicitly set, will be added to the end of the datasource string
	switch dsn.SearchPath {
	case "":
		return s
	default:
		return s + " " + fmt.Sprintf("search_path=%s", dsn.SearchPath)
	}
}

// ConfigFile defines the configuration file.
type ConfigFile struct {
	Config struct {
		Database struct {
			Host       string `json:"host"`
			Port       int    `json:"port"`
			Name       string `json:"name"`
			User       string `json:"user"`
			Password   string `json:"password"`
			SearchPath string `json:"searchPath"`
		} `json:"database"`
		MigrationScriptsDir string `json:"migrationScriptsDir"`
	} `json:"config"`
}

// NewConfigFile initializes a Config struct from a JSON file at a
// predetermined file path (path is relative to project root)
//
// Local:      ./config/local.json
func NewConfigFile(configFilePath string) (ConfigFile, error) {
	var (
		b   []byte
		err error
	)
	b, err = os.ReadFile(configFilePath)
	if err != nil {
		return ConfigFile{}, err
	}

	f := ConfigFile{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		return ConfigFile{}, err
	}

	return f, nil
}

// ConfigCueFilePaths defines the paths for config files processed through CUE.
type ConfigCueFilePaths struct {
	// Input defines the list of paths for files to be taken as input for CUE
	Input []string
	// Output defines the path for the JSON output of CUE
	Output string
}

// CUEPaths returns the ConfigCueFilePaths.
// Paths are relative to the project root.
func CUEPaths(profile string) ConfigCueFilePaths {
	const schemaInput = "./config/cue/schema.cue"

	// cue config path - relative to project root
	profileInput := "./config/cue/" + profile + ".cue"
	// regular config path - relative to project root
	profileOutput := "./config/" + profile + ".json"

	return ConfigCueFilePaths{
		Input:  []string{schemaInput, profileInput},
		Output: profileOutput,
	}
}
