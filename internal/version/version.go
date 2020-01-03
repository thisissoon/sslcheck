package version

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

var (
	// Version info, set at build time
	Version string
	// Timestamp set at build time
	Timestamp string
	// GitCommit set at build time
	GitCommit string
	// GitTreeState set at build time
	GitTreeState string
)

// BuildTime returns the build timestamp
func BuildTime() time.Time {
	ts, _ := strconv.ParseInt(Timestamp, 10, 64)
	return time.Unix(ts, 0).UTC()
}

// Write writes version info to an io.Writer
func Write(w io.Writer) {
	fmt.Fprintln(w, "Version:", Version)
	fmt.Fprintln(w, "Build Time:", BuildTime().Format(time.RFC1123))
	fmt.Fprintln(w, "Git Commit:", GitCommit)
	fmt.Fprintln(w, "Git Tree State:", GitTreeState)
}
