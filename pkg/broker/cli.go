package broker

import (
	"flag"
)

// Options holds the options specified by the broker's code on the command
// line. Users should add their own options here and add flags for them in
// AddFlags.
type Options struct {
	CatalogPath string
	Async       bool
	dbConnStr   string
	dbHost      string
	dbPort      int
}

// AddFlags is a hook called to initialize the CLI flags for broker options.
// It is called after the flags are added for the skeleton and before flag
// parse is called.
func AddFlags(o *Options) {
	flag.StringVar(&o.CatalogPath, "catalogPath", "", "The path to the catalog")
	flag.BoolVar(&o.Async, "async", false, "Indicates whether the broker is handling the requests asynchronously.")
	flag.StringVar(&o.dbConnStr, "dbConnStr", "", "DB Privileged Connection URI")
	flag.StringVar(&o.dbHost, "dbHost", "", "DB host")
	flag.IntVar(&o.dbPort, "dbPort", 1521, "DB Port")
}
