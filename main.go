package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
)

const usage = `cos-sync <file> [--bucket <alias>] [-q] [--md]

Upload a single file to Tencent COS and print its public URL.
Object key format: 20060102_150405_<original-filename> (local timezone).

Flags (may appear before or after the file path):
  --bucket <alias>   bucket alias from ~/.cos-sync/config.yaml (default: config 'default')
  -q, --quiet        print only the resulting URL on stdout (for piping)
  -m, --md           print the URL as Markdown image syntax: ![](url)
  --version          print version and exit
  -h, --help         show this help

The result is automatically copied to the system clipboard when a clipboard is
available (requires xclip / xsel / wl-copy on Linux, native on macOS / Windows);
a status line is printed to stderr on success, silent on failure.

Config: ~/.cos-sync/config.yaml — see config.example.yaml for a template.

Exit codes: 0 success, 1 runtime error (open/upload), 2 config/usage error.
`

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	bucketAlias, quiet, md, showVersion, positional, err := parseArgs(args)
	if err != nil {
		if errors.Is(err, errHelp) {
			fmt.Print(usage)
			return 0
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		fmt.Fprint(os.Stderr, usage)
		return 2
	}

	if showVersion {
		fmt.Println(version)
		return 0
	}

	if len(positional) != 1 {
		fmt.Fprintln(os.Stderr, "error: expected exactly one file path, got", len(positional))
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	filePath := positional[0]

	cfg, cfgPath, err := LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}

	resolvedAlias, bc, err := cfg.Resolve(bucketAlias)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", fmt.Errorf("open %s: %w", filePath, err))
		return 1
	}
	defer file.Close()

	key := time.Now().Format("20060102_150405") + "_" + filepath.Base(filePath)

	uploader, err := NewUploader(bc)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}

	if err := uploader.Upload(context.Background(), key, file); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	pubURL := PublicURL(uploader.BucketName(), uploader.Region(), key)
	payload := pubURL
	if md {
		payload = fmt.Sprintf("![](%s)", pubURL)
	}

	switch {
	case md:
		fmt.Println(payload)
	case quiet:
		fmt.Println(pubURL)
	default:
		fmt.Fprintf(os.Stderr, "config: %s\n", cfgPath)
		fmt.Fprintf(os.Stderr, "bucket: %s\n", resolvedAlias)
		fmt.Printf("Uploaded → %s\n", pubURL)
	}

	if err := clipboard.WriteAll(payload); err == nil {
		fmt.Fprintln(os.Stderr, "clipboard: 已复制到剪贴板")
	}
	return 0
}

var errHelp = errors.New("help requested")

func parseArgs(args []string) (bucketAlias string, quiet, md, showVersion bool, positional []string, err error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			return "", false, false, false, nil, errHelp
		case a == "-q" || a == "--quiet":
			quiet = true
		case a == "-m" || a == "--md" || a == "--markdown":
			md = true
		case a == "--version":
			showVersion = true
		case a == "--bucket" || a == "-bucket":
			if i+1 >= len(args) {
				return "", false, false, false, nil, fmt.Errorf("%s requires a value", a)
			}
			bucketAlias = args[i+1]
			i++
		case strings.HasPrefix(a, "--bucket="):
			bucketAlias = strings.TrimPrefix(a, "--bucket=")
		case strings.HasPrefix(a, "-"):
			return "", false, false, false, nil, fmt.Errorf("unknown flag: %s", a)
		default:
			positional = append(positional, a)
		}
	}
	return
}
