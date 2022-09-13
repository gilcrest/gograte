package config

config: #LocalConfig

config: migrationScriptsDir: "./scripts/db/migrations"

config: database: host:       "localhost"
config: database: port:       5432
config: database: name:       "dga_local"
config: database: user:       "demo_user"
config: database: password:   "REPLACE_ME"
config: database: searchPath: "demo"
