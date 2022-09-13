package main

import (
	"github.com/gilcrest/gograte"
	"github.com/magefile/mage/sh"
)

// CueGenConfig generates a configuration file using CUE,
// example: mage -v cueGenerateConfig.
// The files are run through cue vet to ensure they are acceptable given
// the schema and are then run through cue "fmt" to format the files
func CueGenConfig() (err error) {

	var paths gograte.ConfigCueFilePaths
	paths, err = gograte.CUEPaths()
	if err != nil {
		return err
	}

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

// Up uses the psql cli to execute DDL scripts found in the up directory, example: mage -v up.
// All files will be executed, regardless of errors within an individual
// file. Check output to determine if any errors occurred. Eventually,
// I will write this to stop on errors, but for now it is what it is.
func Up() (err error) {
	var args []string

	args, err = gograte.PSQLArgs(true)
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
// found in the down directory, example: mage -v down.
// All files will be executed, regardless of errors within
// an individual file. Check output to determine if any errors occurred.
// Eventually, I will write this to stop on errors, but for now it is
// what it is.
func Down() (err error) {
	var args []string

	args, err = gograte.PSQLArgs(false)
	if err != nil {
		return err
	}

	err = sh.Run("psql", args...)
	if err != nil {
		return err
	}

	return nil
}
