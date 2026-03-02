//go:build mage
// +build mage

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
)

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
)

type tool struct {
	name        string
	versionArgs []string
	minVersion  string // from rakelib/00_init.rake where defined; empty = no check
	installHint string
	optional    bool
}

func CheckBuildDeps() error {
	buildTools := []tool{
		{"go", []string{"version"}, "1.25.7", "Install Go: https://go.dev/dl/ or distro (e.g. apt install golang-go)", false},
		{"node", []string{"--version"}, "20.20.0", "Install Node.js: https://nodejs.org/ or distro (e.g. apt install nodejs)", false},
		{"npm", []string{"--version"}, "11.9.0", "Comes with Node.js; ensure node is on PATH", false},
		{"npx", []string{"--version"}, "", "Comes with npm; ensure npm is on PATH", false},
		{"python3", []string{"--version"}, "", "Install Python 3: distro (e.g. apt install python3 python3-venv)", false},
		{"sphinx-build", []string{"--version"}, "", "pip install -r doc/src/requirements.txt or apt install python3-sphinx", false},
		{"swagger", []string{"version"}, "0.31.0", "go install github.com/go-swagger/go-swagger/cmd/swagger@v0.31.0", false},
		{"protoc", []string{"--version"}, "31.1", "Distro (e.g. apt install protobuf-compiler) or https://github.com/protocolbuffers/protobuf/releases", false},
		{"yamlinc", []string{"--version"}, "0.1.10", "npm install -g yamlinc", false},
		{"java", []string{"-version"}, "", "Distro (e.g. apt install openjdk-17-jre-headless)", false},
		{"wget", []string{"--version"}, "", "Distro (e.g. apt install wget)", false},
		{"unzip", []string{"-v"}, "", "Distro (e.g. apt install unzip)", false},
		{"git", []string{"--version"}, "", "Distro (e.g. apt install git)", false},
		{"tar", []string{"--version"}, "", "Usually preinstalled (GNU tar)", false},
		{"sed", []string{"--version"}, "", "Usually preinstalled", false},
		{"openssl", []string{"version"}, "", "Distro (e.g. apt install openssl)", false},
	}

	return checkDeps(buildTools)
}

func CheckDevDeps() error {
	devTools := []tool{
		{"chromedriver", []string{"--version"}, "", "Distro (e.g. apt install chromium-chromedriver)", false},
		{"playwright", []string{"--version"}, "", "npm install -g playwright", false},
		{"pytest", []string{"--version"}, "", "pip install pytest", false},
		{"pylint", []string{"--version"}, "", "pip install pylint", false},
		{"flake8", []string{"--version"}, "", "pip install flake8", false},
		{"black", []string{"--version"}, "", "pip install black", false},
		{"docker", []string{"--version"}, "", "Distro or https://docs.docker.com/get-docker/", true},
	}

	return checkDeps(devTools)
}

// CheckDeps detects required tools and their versions, reports all missing
// tools (does not stop at the first), and returns an error if any are missing.
// Green = found and version OK; Red = missing (required); Yellow = missing optional or version too old.
func checkDeps(tools []tool) error {
	type tool struct {
		name        string
		versionArgs []string
		minVersion  string // from rakelib/00_init.rake where defined; empty = no check
		installHint string
		optional    bool
	}

	type statusLine struct {
		line   string
		status string // "ok", "missing", "optional_missing", "old"
	}
	var lines []statusLine
	var missingRequired []string
	hints := make(map[string]string)

	for _, t := range tools {
		path, err := exec.LookPath(t.name)
		if err != nil {
			hints[t.name] = t.installHint
			if t.optional {
				lines = append(lines, statusLine{fmt.Sprintf("  %s: (not found) [optional]", t.name), "optional_missing"})
			} else {
				missingRequired = append(missingRequired, t.name)
				lines = append(lines, statusLine{fmt.Sprintf("  %s: missing", t.name), "missing"})
			}
			continue
		}
		rawVersion := getVersion(path, t.versionArgs)
		parsed := extractVersion(t.name, rawVersion)
		if t.minVersion != "" && parsed != "" {
			if compareVersion(parsed, t.minVersion) < 0 {
				lines = append(lines, statusLine{
					fmt.Sprintf("  %s: %s (%s) — need >= %s", t.name, path, rawVersion, t.minVersion),
					"old",
				})
				if !t.optional {
					missingRequired = append(missingRequired, t.name)
				}
				continue
			}
		}
		lines = append(lines, statusLine{fmt.Sprintf("  %s: %s (%s)", t.name, path, rawVersion), "ok"})
	}

	// Docker Compose: either "docker-compose" or "docker compose" plugin
	composePath, err := exec.LookPath("docker-compose")
	if err == nil {
		rawVersion := getVersion(composePath, []string{"--version"})
		lines = append(lines, statusLine{fmt.Sprintf("  docker-compose: %s (%s)", composePath, rawVersion), "ok"})
	} else {
		dockerPath, _ := exec.LookPath("docker")
		if dockerPath != "" {
			cmd := exec.Command("docker", "compose", "version")
			out, runErr := cmd.CombinedOutput()
			if runErr == nil {
				ver := strings.TrimSpace(string(out))
				if idx := strings.Index(ver, "\n"); idx > 0 {
					ver = ver[:idx]
				}
				lines = append(lines, statusLine{fmt.Sprintf("  docker compose: plugin (%s)", ver), "ok"})
			} else {
				lines = append(lines, statusLine{"  docker compose: missing [optional]", "optional_missing"})
				hints["docker compose"] = "Install Docker Compose plugin (docker compose version) or standalone docker-compose"
			}
		} else {
			lines = append(lines, statusLine{"  docker compose: missing [optional]", "optional_missing"})
			hints["docker compose"] = "Install Docker and Docker Compose plugin, or standalone docker-compose"
		}
	}

	// Report with colors
	fmt.Println("Dependency check:")
	for _, sl := range lines {
		switch sl.status {
		case "ok":
			fmt.Printf("%s%s%s\n", colorGreen, sl.line, colorReset)
		case "missing":
			fmt.Printf("%s%s%s\n", colorRed, sl.line, colorReset)
		case "optional_missing", "old":
			fmt.Printf("%s%s%s\n", colorYellow, sl.line, colorReset)
		default:
			fmt.Println(sl.line)
		}
	}
	if len(missingRequired) > 0 {
		fmt.Println()
		fmt.Printf("%sMissing required tools:%s\n", colorRed, colorReset)
		for _, name := range missingRequired {
			fmt.Printf("  %s%s%s\n", colorRed, name, colorReset)
			if h, ok := hints[name]; ok {
				fmt.Printf("    -> %s\n", h)
			}
		}
		fmt.Println()
		return fmt.Errorf("missing %d required tool(s): %s", len(missingRequired), strings.Join(missingRequired, ", "))
	}
	return nil
}

// extractVersion returns a comparable version string from raw tool output, or "" if not parsed.
func extractVersion(toolName, raw string) string {
	raw = strings.TrimSpace(raw)
	// Go: "go version go1.25.7 linux/amd64"
	if toolName == "go" {
		if idx := strings.Index(raw, "go"); idx >= 0 {
			rest := raw[idx+2:]
			return firstSemver(rest)
		}
	}
	// Node: "v22.22.0" or "22.22.0"
	if toolName == "node" || toolName == "npm" || toolName == "npx" {
		return firstSemver(strings.TrimPrefix(raw, "v"))
	}
	// Swagger: "version 0.31.0"
	if toolName == "swagger" {
		return firstSemver(raw)
	}
	// Protoc: "libprotoc 3.25.1" or "libprotoc 31.1"
	if toolName == "protoc" {
		return firstSemver(raw)
	}
	// yamlinc and others: first semver in line
	return firstSemver(raw)
}

var semverRe = regexp.MustCompile(`(\d+(?:\.\d+)*(?:\.\d+)?)`)

func firstSemver(s string) string {
	m := semverRe.FindStringSubmatch(s)
	if len(m) >= 1 {
		return m[1]
	}
	return ""
}

// compareVersion returns -1 if a < b, 0 if a == b, 1 if a > b. Compares numeric segments.
func compareVersion(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		var an, bn int
		if i < len(as) {
			an, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			bn, _ = strconv.Atoi(bs[i])
		}
		if an < bn {
			return -1
		}
		if an > bn {
			return 1
		}
	}
	return 0
}

func getVersion(path string, args []string) string {
	cmd := exec.Command(path, args...)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	_ = cmd.Run()
	combined := strings.TrimSpace(out.String() + "\n" + errOut.String())
	firstLine := strings.SplitN(combined, "\n", 2)[0]
	return strings.TrimSpace(firstLine)
}

// GenProto generates Go code from backend/api/agent.proto (isc.org/stork/api).
// Requires protoc, protoc-gen-go, and protoc-gen-go-grpc on PATH.
func GenProto() error {
	root, err := repoRoot()
	if err != nil {
		return err
	}
	apiDir := filepath.Join(root, "backend", "api")
	cmd := exec.Command("protoc", "--proto_path=.", "--go_out=.", "--go-grpc_out=.", "agent.proto")
	cmd.Dir = apiDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("protoc agent.proto: %w", err)
	}
	fmt.Println("Generated backend/api/agent.pb.go and agent_grpc.pb.go")
	return nil
}

// BuildTool builds the stork-tool backend binary (CLI, no API).
// Depends on GenProto so that isc.org/stork/api is available.
// Output: backend/cmd/stork-tool/stork-tool
func BuildTool() error {
	mg.Deps(GenProto)
	root, err := repoRoot()
	if err != nil {
		return err
	}
	backendDir := filepath.Join(root, "backend")
	outBinary := filepath.Join(backendDir, "cmd", "stork-tool", "stork-tool")
	buildDate := time.Now().Format("2006-01-02 15:04")
	ldflags := fmt.Sprintf("-X 'isc.org/stork.BuildDate=%s'", buildDate)
	cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", outBinary, "./cmd/stork-tool")
	cmd.Dir = backendDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build stork-tool: %w", err)
	}
	fmt.Printf("Built %s (BuildDate=%s)\n", outBinary, buildDate)
	return nil
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "backend", "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repository root not found (no backend/go.mod)")
		}
		dir = parent
	}
}
