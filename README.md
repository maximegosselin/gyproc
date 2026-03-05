# Gyproc

Gyproc runs multiple commands in parallel and multiplexes their output into a single newline-delimited JSON stream. Each event is tagged with a `seq` identifier, allowing the client to track the individual progress of every command, regardless of the order in which output arrives.

## Use cases

- **Batch file processing** — run a worker per file and follow each one's progress independently
- **Multi-service migrations** — execute database migrations across services in parallel and know exactly which one succeeds or fails
- **Long-running tasks** — run tasks concurrently and let the client aggregate their individual progress in real time

Gyproc is language-agnostic: any client that can read a stream of JSON lines can consume its output.

No pre-built binaries are provided. Clone the repository and compile from source.

## Build

```sh
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o gyproc .

# Linux (arm64)
GOOS=linux GOARCH=arm64 go build -o gyproc .

# macOS (amd64)
GOOS=darwin GOARCH=amd64 go build -o gyproc .

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o gyproc .

# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o gyproc.exe .
```

## Usage

### Defining commands

Pass commands via a file with `--file`:

```sh
gyproc --file commands.txt
```

Or via stdin:

```sh
echo -e "echo hello\necho world" | gyproc
```

Lines starting with `#` are ignored.

### Events

Each event is a JSON object on its own line. Fields:

| Field     | Type   | Description                                      |
|-----------|--------|--------------------------------------------------|
| `seq`     | int    | Sequential ID assigned to the command            |
| `event`   | string | Event type                                       |
| `pid`     | int    | Process ID (`run`, `out`, `exit`)                |
| `command` | string | Command string (`ack`)                           |
| `message` | string | Output content or error message (`out`, `fail`)  |
| `code`    | int    | Exit code (`exit`)                               |
| `time`    | string | ISO 8601 timestamp                               |

Event types:

| Event  | When                                                    |
|--------|---------------------------------------------------------|
| `ack`  | Command received and queued                             |
| `run`  | Command started                                         |
| `out`  | Process wrote to stdout or stderr (one event per write) |
| `exit` | Command finished                                        |
| `fail` | Command could not be started                            |

### Concurrency limit

Use `--limit` to cap the number of commands running simultaneously:

```sh
gyproc --file commands.txt --limit 4
```

### Signals

SIGINT (CTRL+C) and SIGTERM are forwarded to all running child processes.

### Consuming the stream

Filter events by type or by command using `jq`:

```sh
# Follow the output of command seq 2 only
gyproc --file commands.txt | jq 'select(.seq == 2 and .event == "out") | .message'

# Watch for failures
gyproc --file commands.txt | jq 'select(.event == "fail")'

# Print exit codes for all commands
gyproc --file commands.txt | jq 'select(.event == "exit") | {seq, code}'
```

### Example

`commands.txt`:
```
# this is a comment
sleep 5
nonexistentcommand
sleep 2
```

```sh
$ gyproc --file commands.txt --limit 2
{"seq":1,"event":"ack","command":"sleep 5","time":"2026-01-15T10:00:00.000000000Z"}
{"seq":2,"event":"ack","command":"nonexistentcommand","time":"2026-01-15T10:00:00.000000000Z"}
{"seq":3,"event":"ack","command":"sleep 2","time":"2026-01-15T10:00:00.000000000Z"}
{"seq":1,"event":"run","time":"2026-01-15T10:00:00.000000000Z"}
{"seq":2,"event":"run","time":"2026-01-15T10:00:00.000000000Z"}
{"seq":2,"event":"fail","message":"could not start process: exec: \"nonexistentcommand\": executable file not found in $PATH","time":"2026-01-15T10:00:00.000000000Z"}
{"seq":3,"event":"run","time":"2026-01-15T10:00:00.000000000Z"}
{"seq":3,"event":"exit","pid":5679,"code":0,"time":"2026-01-15T10:00:02.000000000Z"}
{"seq":1,"event":"exit","pid":5678,"code":0,"time":"2026-01-15T10:00:05.000000000Z"}
```
