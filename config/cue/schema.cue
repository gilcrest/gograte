package config

#Base: {
	migrationScriptsDir: !="" // must be specified and non-empty
}

#Database: {
	host:       !="" // must be specified and non-empty
	port:       !=0  // must be specified and non-empty
	name:       !="" // must be specified and non-empty
	user:       !="" // must be specified and non-empty
	password:   !="" // must be specified and non-empty
	searchPath: !="" // must be specified and non-empty
}

#LocalConfig: {
	#Base
	database: #Database
}
