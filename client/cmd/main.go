package main

// Build-time variables (populated via -ldflags in release builds)
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// main now delegates to Cobra root command which launches the TUI by default.
func main() {
	Execute()
}
