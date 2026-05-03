# dbredraw

A fast command-line utility for filtering and rebuilding Gather database (GatherDB) files. Given a search term, or a piped input that contains at a list of devices in the first column, it extracts all matching device names and redraws a complete database containing only those devices — producing a full, well-formed subset.

## How It Works

dbredraw runs in two passes:

1. **Discovery/Search** — finds all device names that match the search criteria (by searching the DB file, reading device names from stdin, or both)
2. **Redraw** — streams through the DB file a second time and creates a full new GatherDB containing only the matched devices

Because it streams rather than loading the entire file into memory, it handles large databases efficiently.

## Usage

```
dbredraw -db <GatherDB File> [-search <term>] [-o <outfile>]
grep <term> <GatherDB File> | dbredraw -db <GatherDB File> [-o <outfile>]
grep <term> <GatherDB File> | dbredraw -db <GatherDB File> 
```

### Flags

| Flag | Required | Description |
|------|----------|-------------|
| `-db <file>` | Yes | Path to the GatherDB file |
| `-search <term>` | Situational | Case-insensitive search term (required unless piping device names via stdin) |
| `-o <file>` | No | Write output to a file instead of stdout |

## Examples

**Search for all devices with a matching line, redraw to stdout:**
```bash
dbredraw -db GatherDB.txt -search "100G" 
```

**Write the subset to a file:**
```bash
dbredraw -db GatherDB.txt -search "100G" -o subset.txt
```

**Pipe a device list directly (no search term needed):**
```bash
cat devices.txt | dbredraw -db GatherDB.txt
```

**Chain searches with AND logic — pipe one result into another:**
```bash
dbredraw -db GatherDB.txt -search "100G" | dbredraw -db GatherDB.txt -search "10.1.1.1"
```
**Advanced Chain Example: Create a list of devices and their OS Versions that contain QSFP-100G SFPs**
```bash
grep "show inventory" GatherDB.txt | grep QSFP-100G | ./dbredraw -db GatherDB.txt | grep "show version"
```

In the advanced chain example above a GatherDB is searched for "show inventory" output. The "show inventory" is further filtered to only lines containing QSFP-100G. The filtered output is piped to the dbredraw. The dbredraw then creates a full subset database with only devices that have QSFP-100G. The resulting GatherDB is then streamed to a "show version" search. Finally, a list of devices with QSFP-100G and their OS versions is produced.  


## Building

Requires Go 1.21+. No external dependencies.

```bash
go build -o dbredraw main.go
```

## GatherDB Format

Each line in a GatherDB file begins with a device name (the first whitespace-separated token), followed by the data collected from that device. dbredraw uses this convention to group and filter lines by device.

## Compiling instructions in the notes.txt file

## Binaries avaliable in this repo
```
dbredraw-apl
dbredraw-apl_arm64
dbredraw-lnx32
dbredraw-lnx64
dbredraw-win32.exe
```
dbredraw-win64.exe
```
