package main

// Build-time variables for version information
var (
	buildVersion string = "1.0.0"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// main is the entry point that executes the root command
func main() {
	Execute()
}
