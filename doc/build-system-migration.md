# Build System Migration: Mage

This document outlines how to replace the current Rake-based build with **Mage** (Go): one build system that (1) uses system dependencies only, (2) checks and aborts with clear messages instead of auto-installing, and (3) avoids adding Ruby.

## What the current Rake system does

- **Dependency “management”** (`rakelib/00_init.rake`): Downloads and installs into `tools/`:
  - **Node**: Node.js tarball, npm, npx, `yamlinc` (npm global)
  - **Go**: Go tarball, go-swagger binary, protoc, protoc-gen-go, protoc-gen-go-grpc, golangci-lint, shellcheck, tparse, go-junit-report, gocover-cobertura, mockgen, delve, nfpm, govulncheck
  - **Ruby**: Bundler + Danger (Gemfile in `rakelib/init_deps/danger/`)
  - **Python**: venv in `tools/python`, sphinx-build, pytest, pylint, flake8, black, flask (via pip-compile lockfiles per Python version)
  - **Other**: openapi-generator-cli.jar (wget)
- **Codebase / codegen** (`10_codebase.rake`): File lists, code-gen binary, swagger merge (yamlinc), backend swagger server gen (goswagger).
- **Build** (`20_build.rake`): Docs (sphinx), frontend (`npx ng build`), backend (agent, server, tool).
- **Dev** (`30_dev.rake`): Lint, test, storybook, run server, etc.
- **Dist** (`40_dist.rake`): Packaging (nfpm, etc.).
- **Docker** (`stork.Dockerfile`): Uses `rake prepare` and `rake build:...` in stages.

The replacement uses **Mage**: build logic in Go under `magefiles/`. Targets run from the repository root (e.g. `mage -d magefiles check`, `mage -d magefiles buildTool`).

## Design principles

1. **No auto-install**: Use only what’s on the system (or in the container). If something is missing, print what’s missing and how to install it, then exit with non-zero.
2. **Single entry point**: `mage -d magefiles <target>`. Use `check` to verify dependencies and report all missing tools with versions; use build targets for server, agent, tool, UI, docs.
3. **Language**: Go only for the build system (same as backend).

## Mage usage

From the repository root:

- **Check dependencies (all tools, report all missing):**  
  `mage -d magefiles checkDeps`
- **Generate protobuf Go code (isc.org/stork/api):**  
  `mage -d magefiles genProto`  
  Requires `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` on PATH.
- **Build stork-tool (backend CLI, no API):**  
  `mage -d magefiles buildTool`  
  Runs `genProto` first, then builds `backend/cmd/stork-tool/stork-tool`.

Additional targets (build server/agent, UI, docs, etc.) can be added as the migration progresses.

## Migration path (high level)

1. **Document required tools**  
   From `rakelib/00_init.rake` and `rakelib/init_deps`, list every tool and (where relevant) version. Keep this in README and in the Mage `check` target (install hints).

2. **Implement check and build targets**  
   - `check`: detect all required tools and their versions; report **all** missing tools with install hints, then exit non-zero if any are missing.  
   - Build targets: e.g. `buildTool`, then `buildServer`, `buildAgent`, `buildUI`, `buildDocs`, etc., calling `go`, `npm`, `sphinx-build`, etc. directly. No downloading; assume tools are on `PATH`.

3. **Replace Rake usage**  
   In developer docs and README, replace `rake prepare` with “install the following tools: …” and `rake check` with `mage -d magefiles check`. Replace `rake build:tool` with `mage -d magefiles buildTool`, etc. Keep the same **outputs** (e.g. `backend/cmd/stork-tool/stork-tool`, `webui/dist/...`) so the rest of the repo and CI don’t need to change.

4. **Docker**  
   In `docker/images/stork.Dockerfile`, stop using `rake prepare` and `rake build:...`. Install required tools in the image (Node, Go, Python, goswagger, protoc, etc.) via `apt-get` or multi-stage copies, then run Mage (e.g. `mage -d magefiles build` or per-component targets).

5. **Remove Rake**  
   Once CI and Docker use Mage, delete `Rakefile`, `rakelib/`, and any Ruby-specific bits. Optionally keep a minimal `Rakefile` that forwards to Mage for a transition period.

6. **Optional: per-subproject delegation**  
   Build targets can `cd` into `backend/` or `webui/` and run `go build` or `npm run build`, keeping the magefiles thin.

## Required tools summary (for check target)

These should be on `PATH` (or documented as optional where applicable). The Mage `checkDeps` target reports presence and version for each, and lists all missing with install hints. Where the Rake build defines a minimum version (`rakelib/00_init.rake`), that minimum is enforced (go, node, npm, swagger, protoc, yamlinc); others are not version-checked.

| Tool | Purpose | How to install (example) |
|------|--------|---------------------------|
| go | Backend, codegen, many Go tools | distro / go.dev |
| node | Frontend | distro / nodejs.org |
| npm | Frontend | comes with Node |
| npx | Angular CLI, yamlinc | comes with npm |
| python3 | Docs, tests/sim | distro |
| sphinx-build | Docs | `pip install -r doc/src/requirements.txt` or distro |
| swagger (goswagger) | API codegen | `go install github.com/go-swagger/go-swagger/cmd/swagger@v0.31.0` |
| protoc | (if used) | distro or release binary |
| yamlinc | Swagger merge | `npm install -g yamlinc` or use from webui node_modules |
| java | openapi-generator (if used) | distro |
| docker, docker compose | System tests / demo | distro |

Optional / CI-only: golangci-lint, shellcheck, pytest, flake8, black, etc. The `checkDeps` target can distinguish “required” vs “optional” so only required tools cause a non-zero exit when missing.

**Perl** is not required for building Stork. It is only used in the Rake release workflow (`rake notes`) to transform release notes from the wiki (e.g. stripping markdown link syntax). If you need to prepare release notes the same way without Rake, you would need Perl (or an equivalent script) for that step only.
