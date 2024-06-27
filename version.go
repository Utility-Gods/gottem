package version

// Version holds the current version of the Gottem application.
// This is typically updated manually or by a build script before a new release.
const Version = "1.4.0"

// Additional version-related information can be added here.
// For example:
// const (
//     BuildDate = "2023-06-27"
//     GitCommit = "abc123" // This could be filled in by your build process
// )

// VersionInfo returns a string with the current version information.
func VersionInfo() string {
	return "Gottem version " + Version
	// You could expand this to include more information:
	// return fmt.Sprintf("Gottem version %s (Built on %s, Commit %s)", Version, BuildDate, GitCommit)
}
