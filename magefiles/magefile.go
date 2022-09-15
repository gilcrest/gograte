package main

import (
	"github.com/gilcrest/gograte"
	"github.com/magefile/mage/sh"
)

// CueGenConfig generates a configuration file using CUE,
// example: mage -v cueGenConfig default.
//
// The program expects that a .cue file exists which matches the profile name
// given as the profile parameter and will create a .json file with the same
// base name as well. For example, if you wanted to create a config profile for
// your test project, you would create a test.cue file, populate appropriately,
// and then run mage -v cueGenConfig test which would create a test.json file.
// These json files can then be used in the up or down migration to easily switch
// between projects.
//
// The files are run through cue vet to ensure they are acceptable given
// the schema found in schema.cue and are then run through cue "fmt" to
// format the files.
func CueGenConfig(profile string) (err error) {

	paths := gograte.CUEPaths(profile)

	// Vet input files
	vetArgs := []string{"vet"}
	vetArgs = append(vetArgs, paths.Input...)
	err = sh.Run("cue", vetArgs...)
	if err != nil {
		return err
	}

	// format input files
	fmtArgs := []string{"fmt"}
	fmtArgs = append(fmtArgs, paths.Input...)
	err = sh.Run("cue", fmtArgs...)
	if err != nil {
		return err
	}

	// Export output files
	exportArgs := []string{"export"}
	exportArgs = append(exportArgs, paths.Input...)
	exportArgs = append(exportArgs, "--force", "--out", "json", "--outfile", paths.Output)

	err = sh.Run("cue", exportArgs...)
	if err != nil {
		return err
	}

	return nil
}

// Up uses the psql cli to execute DDL scripts found in the up directory, example: mage -v up default.
//
// A json file matching the profile name is expected in the ./config directory.
// A default.json file is provided, but others may be generated easily (or just copy/paste).
//
// All files will be executed, regardless of errors within an individual file.
// Check output to determine if any errors occurred. Eventually, I will write
// this to stop on errors, but for now it is what it is.
func Up(profile string) (err error) {
	var args []string

	args, err = gograte.PSQLArgs(true, profile)
	if err != nil {
		return err
	}

	err = sh.Run("psql", args...)
	if err != nil {
		return err
	}

	return nil
}

// Down uses the psql cli to execute drop statement DDL scripts
// found in the down directory, example: mage -v down default.
//
// A json file matching the profile name is expected in the ./config directory.
// A default.json file is provided, but others may be generated easily (or just copy/paste).
//
// All files will be executed, regardless of errors within an individual file.
// Check output to determine if any errors occurred. Eventually, I will write
// this to stop on errors, but for now it is what it is.
func Down(profile string) (err error) {
	var args []string

	args, err = gograte.PSQLArgs(false, profile)
	if err != nil {
		return err
	}

	err = sh.Run("psql", args...)
	if err != nil {
		return err
	}

	return nil
}
