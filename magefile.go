//go:build mage
// +build mage

package main

import (
	"path"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh"
)

const (
	storkServerExe          = "stork-server"
	storkAgentExe           = "stork-agent"
	storkToolExe            = "stork-tool"
	storkCodeGenExe         = "stork-code-gen"
	dhcpV4OptionsJSON       = "codegen/std_dhcpv4_option_def.json"
	dhcpV6OptionsJSON       = "codegen/std_dhcpv6_option_def.json"
	dhcpV4OptionsGo         = "backend/daemoncfg/kea/stdoptiondef4.go"
	dhcpV4OptionsGoTemplate = "backend/daemoncfg/kea/stdoptiondef4.go.template"
	dhcpV6OptionsGo         = "backend/daemoncfg/kea/stdoptiondef6.go"
	dhcpV6OptionsGoTemplate = "backend/daemoncfg/kea/stdoptiondef6.go.template"
	combinedOpenAPIYAML     = "api/swagger.yaml"
	dhcpV4OptionsTS         = "webui/src/app/std-dhcpv4-option-defs.ts"
	dhcpV4OptionsTSTemplate = "webui/src/app/std-dhcpv4-option-defs.ts.template"
	dhcpV6OptionsTS         = "webui/src/app/std-dhcpv6-option-defs.ts"
	dhcpV6OptionsTSTemplate = "webui/src/app/std-dhcpv6-option-defs.ts.template"
	generatedWebBackend     = "webui/src/app/backend"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

// Run `go build`.
func go_build(cwd string, output string, flags ...string) error {
	args := []string{
		"-C",
		cwd,
		"build",
		"-o",
		output,
	}
	for _, flag := range flags {
		args = append(args, flag)
	}
	return sh.Run("go", args...)
}

// Build the backend executables and the TypeScript frontend.
func Build() error {
	mg.Deps(BuildBackend, BuildFrontend)
	return nil
}

// Build the TypeScript frontend.
func BuildFrontend() error {
	mg.Deps(InstallNodeDeps, GenerateFrontendCode)
	return sh.Run("npm", "exec", "--prefix", "webui", "--", "ng", "build", "--configuration", "production")
}

// Build all three backend executables (agent, server, tool).
func BuildBackend() error {
	mg.Deps(BuildServer, BuildTool, BuildAgent)
	return nil
}

// Build the Stork server.
func BuildServer() error {
	mg.Deps(InstallGoDeps, GenerateBackendCode)
	return go_build("backend", "stork-server", "./cmd/stork-server")
}

// Build the Stork Tool.
func BuildTool() error {
	mg.Deps(InstallGoDeps, GenerateBackendCode)
	return go_build("backend", "stork-tool", "./cmd/stork-tool")
}

// Build the Stork Agent.
func BuildAgent() error {
	mg.Deps(InstallGoDeps, GenerateBackendCode)
	return go_build("backend", "stork-agent", "./cmd/stork-agent")
}

func BuildStorkCodeGen() error {
	mg.Deps(InstallGoDeps)
	return go_build("backend", "stork-code-gen", "./cmd/stork-code-gen")
}

func prettier_format(file string) error {
	mg.Deps(InstallNodeDeps)
	return sh.Run("npm", "exec", "--prefix", "webui", "--", "prettier", "--config", ".prettierrc", "--write", file)
}

func GenerateOptionDefs(input, output, template string) error {
	mg.Deps(BuildStorkCodeGen)
	err := sh.Run(
		"backend/stork-code-gen",
		"std-option-defs",
		"--input",
		input,
		"--output",
		output,
		"--template",
		template,
	)
	if err != nil {
		return err
	}
	return prettier_format(output)
}

func GoGenerateBackend() error {
	return sh.Run("go", "-C", "backend", "generate", "./...")
}

// Perform all of the Go code generation.
func GenerateBackendCode() error {
	mg.Deps(
		mg.F(
			GenerateOptionDefs,
			dhcpV4OptionsJSON,
			dhcpV4OptionsGo,
			dhcpV4OptionsGoTemplate,
		),
		mg.F(
			GenerateOptionDefs,
			dhcpV6OptionsJSON,
			dhcpV6OptionsGo,
			dhcpV6OptionsGoTemplate,
		),
		GoGenerateBackend,
	)
	return nil
}

// Install the Go module dependencies.
func InstallGoDeps() error {
	return sh.Run("go", "-C", "backend", "mod", "tidy")
}

// Install the frontend NPM dependencies.
func InstallNodeDeps() error {
	return sh.Run("npm", "clean-install", "--prefer-offline", "--prefix", "webui")
}

// Merge all of the OpenAPI specification files into one file, as required by
// every OpenAPI tool.
func MergeOpenAPISpecs() error {
	mg.Deps(InstallNodeDeps)
	return sh.Run("npm", "--prefix", "webui", "exec", "--", "yamlinc", "-o", combinedOpenAPIYAML, "api/swagger.in.yaml")
}

// java :(
func FrontendOpenAPIGenerator() error {
	mg.Deps(MergeOpenAPISpecs)
	return sh.Run(
		"java",
		"-jar",
		"tools/openapi-generator-cli.jar",
		"generate",
		"-i",
		combinedOpenAPIYAML,
		"-g",
		"typescript-angular",
		"-o",
		generatedWebBackend,
		"--additional-properties",
		"snapshot=true,ngVersion=10.1.5,modelPropertyNaming=camelCase",
	)
}

// Generate front-end code: OpenAPI layer and DHCP option definitions.
func GenerateFrontendCode() error {
	mg.Deps(
		FrontendOpenAPIGenerator,
		mg.F(
			GenerateOptionDefs,
			dhcpV4OptionsJSON,
			dhcpV4OptionsTS,
			dhcpV4OptionsTSTemplate,
		),
		mg.F(
			GenerateOptionDefs,
			dhcpV6OptionsJSON,
			dhcpV6OptionsTS,
			dhcpV6OptionsTSTemplate,
		),
	)
	return nil
}

// Clean up generated files
func Clean() {
	backendFiles := []string{
		storkServerExe,
		storkAgentExe,
		storkToolExe,
		storkCodeGenExe,
	}
	for _, f := range backendFiles {
		sh.Rm(path.Join("backend", f))
	}
	otherFiles := []string{
		dhcpV4OptionsGo,
		dhcpV6OptionsGo,
		combinedOpenAPIYAML,
		dhcpV4OptionsTS,
		dhcpV6OptionsTS,
		generatedWebBackend,
	}
	for _, f := range otherFiles {
		sh.Rm(f)
	}
}
