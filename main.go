package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

const usage = `dbredraw - filter a GatherDB by search term, pipe, or both (AND)

Usage:
  dbredraw -db <file> -search <term> [-o <file>]   search DB, redraw matches
  dbredraw -db <file> [-o <file>]                   pipe lines in, redraw those devices
  dbredraw -db <file> -search <term> [-o <file>]    pipe + search = AND (chain searches)

Flags:
  -db <file>      path to the GatherDB file to redraw from (required)
  -search <term>  text to search for (case-insensitive)
  -o <file>       write output to file instead of stdout

Modes:
  -search only      search -db for term, redraw matching devices
  pipe only         extract devices from piped lines, redraw from -db
  pipe + -search    search piped content for term, redraw those devices from -db
                    (use this to chain/AND multiple searches)

Examples:
  dbredraw -db CampusDB.txt -search "vtp mode"
  dbredraw -db CampusDB.txt -search "vtp mode" -o subset.txt
  cat devices.txt | dbredraw -db CampusDB.txt
  dbredraw -db CampusDB.txt -search "vtp mode" | dbredraw -db CampusDB.txt -search "10.1.1.1"
`

func extractDevice(line string) string {
	idx := strings.IndexByte(line, ' ')
	if idx < 0 {
		return line
	}
	return line[:idx]
}

// buildDeviceSet collects unique non-empty device names from scanned lines.
func buildDeviceSet(sc *bufio.Scanner) (map[string]struct{}, error) {
	devices := make(map[string]struct{})
	for sc.Scan() {
		dev := extractDevice(sc.Text())
		if dev != "" {
			devices[dev] = struct{}{}
		}
	}
	return devices, sc.Err()
}

func writeSubset(dbFile, outputFile string, devices map[string]struct{}) error {
	f, err := os.Open(dbFile)
	if err != nil {
		return fmt.Errorf("open %s: %w", dbFile, err)
	}
	defer f.Close()

	var out *os.File
	if outputFile != "" {
		out, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("create %s: %w", outputFile, err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	writer := bufio.NewWriterSize(out, 1<<20)
	sc := bufio.NewScanner(bufio.NewReaderSize(f, 1<<20))
	sc.Buffer(make([]byte, 1<<20), 1<<20)

	for sc.Scan() {
		line := sc.Text()
		if _, ok := devices[extractDevice(line)]; ok {
			writer.WriteString(line)
			writer.WriteByte('\n')
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("scan db: %w", err)
	}
	return writer.Flush()
}

func run(dbFile, searchTerm, outputFile string, pipeMode bool) error {
	devices := make(map[string]struct{})

	// Determine the source to scan for device collection.
	// pipe + search  → search piped content (AND chaining)
	// pipe only      → extract all devices from piped lines
	// search only    → search the DB file
	if pipeMode {
		term := strings.ToLower(searchTerm)
		sc := bufio.NewScanner(bufio.NewReaderSize(os.Stdin, 1<<20))
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			line := sc.Text()
			if searchTerm == "" || strings.Contains(strings.ToLower(line), term) {
				dev := extractDevice(line)
				if dev != "" {
					devices[dev] = struct{}{}
				}
			}
		}
		if err := sc.Err(); err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
	} else {
		// Scan the DB file for lines matching the search term
		f, err := os.Open(dbFile)
		if err != nil {
			return fmt.Errorf("open %s: %w", dbFile, err)
		}
		term := strings.ToLower(searchTerm)
		sc := bufio.NewScanner(bufio.NewReaderSize(f, 1<<20))
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			line := sc.Text()
			if strings.Contains(strings.ToLower(line), term) {
				dev := extractDevice(line)
				if dev != "" {
					devices[dev] = struct{}{}
				}
			}
		}
		f.Close()
		if err := sc.Err(); err != nil {
			return fmt.Errorf("scan pass 1: %w", err)
		}
	}

	if len(devices) == 0 {
		fmt.Fprintln(os.Stderr, "No matching devices found.")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d matching device(s), rebuilding subset...\n", len(devices))
	return writeSubset(dbFile, outputFile, devices)
}

func main() {
	dbFile := flag.String("db", "", "path to the GatherDB file (required)")
	searchTerm := flag.String("search", "", "text to search for (case-insensitive)")
	outputFile := flag.String("o", "", "write output to file instead of stdout")

	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.Parse()

	if *dbFile == "" {
		fmt.Fprintln(os.Stderr, "error: -db is required")
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	// Detect pipe mode: stdin is not a terminal
	stat, _ := os.Stdin.Stat()
	pipeMode := (stat.Mode() & os.ModeCharDevice) == 0

	if !pipeMode && *searchTerm == "" {
		fmt.Fprintln(os.Stderr, "error: -search is required when not piping input")
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	if err := run(*dbFile, *searchTerm, *outputFile, pipeMode); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
