# Go module path - taken from backend/go.mod
%global import_path isc.org/stork

%global _unpackaged_files_terminate_build 1

Name:    stork
Version: 2.5.0
Release: alt1
Summary: ISC Stork - monitoring dashboard for Kea DHCP and BIND9
License: MPL-2.0
Group:   System/Servers
Url:     https://stork.isc.org

# Source0 - main tarball with sources and vendored Go deps
Source0: %name-%version.tar

BuildRequires(pre): rpm-build-golang
BuildRequires: golang >= 1.25.7

# Python/Sphinx needed to build man pages from documentation sources
BuildRequires: python3
BuildRequires: python3-module-sphinx

%description
ISC Stork monitoring application for ISC Kea DHCP and BIND9.

%package agent
Summary: ISC Stork Agent
Group:   System/Servers

%description agent
Stork agent monitors Kea DHCP and/or BIND9 processes.
Typically deployed on each machine running Kea or BIND9.

%package server
Summary: ISC Stork Server
Group:   System/Servers
# Requires PostgreSQL >= 10 with pgcrypto extension (postgresql-contrib)

%description server
ISC Stork Server provides centralized dashboard for Stork agents.
You typically need a single server in a network.

%prep
%setup -q
mkdir -p doc/user/man/hooks

%build
# --- Go build environment ---
# BUILDDIR - temporary directory for Go build cache
# IMPORT_PATH - module path from go.mod
# GOPATH - where Go looks for packages; %go_path is ALT system GOPATH
# GOFLAGS - -mod=vendor tells Go to use vendor/ without internet
export BUILDDIR="$PWD/.gopath"
export IMPORT_PATH="%import_path"
export GOPATH="$BUILDDIR:%go_path"
export GOFLAGS="-mod=vendor"

# LDFLAGS - embed build date and version into binary at link time
# These variables are read in backend/version.go
export BUILD_DATE=$(date -u +%%Y-%%m-%%d)
export LDFLAGS="-X '%import_path.BuildDate=${BUILD_DATE}' \
                -X '%import_path.Version=%version'"

# Build Go binaries and stork-code-gen (used to generate option defs) ---
# stork-code-gen is a Go tool located in backend/cmd/stork-code-gen
cd backend
%golang_prepare
%golang_build cmd/stork-code-gen
cd ..

# Generate standard DHCP option definitions ---
# These TypeScript files are generated from JSON + Go template
$BUILDDIR/bin/stork-code-gen std-option-defs \
    --input codegen/std_dhcpv4_option_def.json \
    --output webui/src/app/std-dhcpv4-option-defs.ts \
    --template webui/src/app/std-dhcpv4-option-defs.ts.template
$BUILDDIR/bin/stork-code-gen std-option-defs \
    --input codegen/std_dhcpv6_option_def.json \
    --output webui/src/app/std-dhcpv6-option-defs.ts \
    --template webui/src/app/std-dhcpv6-option-defs.ts.template

$BUILDDIR/bin/stork-code-gen std-option-defs \
    --input codegen/std_dhcpv4_option_def.json \
    --output backend/daemoncfg/kea/stdoptiondef4.go \
    --template backend/daemoncfg/kea/stdoptiondef4.go.template
$BUILDDIR/bin/stork-code-gen std-option-defs \
    --input codegen/std_dhcpv6_option_def.json \
    --output backend/daemoncfg/kea/stdoptiondef6.go \
    --template backend/daemoncfg/kea/stdoptiondef6.go.template

cd backend
%golang_build cmd/stork-agent cmd/stork-server cmd/stork-tool
cd ..

# Build man pages via Sphinx ---
# Output will be in doc/build/man/
sphinx-build-3 -M man doc/user doc/build -j 2

%install
export GOROOT=$(go env GOROOT)
export BUILDDIR="$PWD/.gopath"
export IMPORT_PATH="%import_path"
export GOPATH="$BUILDDIR:%go_path"
export IGNORE_SOURCES=1

%golang_install

# --- agent ---

install -Dm644 etc/agent.env \
    %buildroot%_sysconfdir/stork/agent.env

# systemd unit file for agent
install -Dm644 etc/isc-stork-agent.service \
    %buildroot%_unitdir/isc-stork-agent.service

# Agent man page
install -Dm644 doc/build/man/stork-agent.8 \
    %buildroot%_man8dir/stork-agent.8

# --- server ---

# Server config
install -Dm644 etc/server.env \
    %buildroot%_sysconfdir/stork/server.env

# versions.json - supported Kea/BIND versions config
install -Dm644 etc/versions.json \
    %buildroot%_sysconfdir/stork/versions.json

# systemd unit file for server
install -Dm644 etc/isc-stork-server.service \
    %buildroot%_unitdir/isc-stork-server.service

# Man pages for server and management tool
install -Dm644 doc/build/man/stork-server.8 \
    %buildroot%_man8dir/stork-server.8
install -Dm644 doc/build/man/stork-tool.8 \
    %buildroot%_man8dir/stork-tool.8

# Frontend - built Angular app
mkdir -p %buildroot%_datadir/stork/www
cp -a webui/dist/stork/browser/. \
    %buildroot%_datadir/stork/www/

# Example configs for nginx and grafana
install -Dm644 etc/nginx-stork.conf \
    %buildroot%_datadir/stork/examples/nginx-stork.conf

install -Dm644 grafana/kea-dhcp4.json \
    %buildroot%_datadir/stork/examples/grafana/kea-dhcp4.json
install -Dm644 grafana/kea-dhcp6.json \
    %buildroot%_datadir/stork/examples/grafana/kea-dhcp6.json
install -Dm644 grafana/bind9-resolver.json \
    %buildroot%_datadir/stork/examples/grafana/bind9-resolver.json

# stork-code-gen is a build-time tool only, not needed in the final package
rm %buildroot%_bindir/stork-code-gen

###########################################################################
# Install/remove scripts for stork-agent
# $1 - RPM argument: 1 = install, 2 = upgrade, 0 = removal
###########################################################################

%post agent
if [ "$1" -eq 1 ]; then
    # Fresh install: create system user and directories

    home_dir=/var/lib/stork-agent

    # Directories for TLS certificates and auth tokens
    mkdir -p "${home_dir}/certs"
    mkdir -p "${home_dir}/tokens"
    chmod 700 "${home_dir}/certs"
    chmod 700 "${home_dir}/tokens"

    # Create system user if it doesn't exist
    if ! getent passwd stork-agent > /dev/null; then
        useradd --system --home-dir "${home_dir}" stork-agent
    fi

    # Add to named group (BIND9) to be able to read its config files
    if getent group named > /dev/null; then
        usermod -aG named stork-agent
    fi

    # Add to kea group to be able to read its config files
    if getent group _kea > /dev/null; then
        usermod -aG _kea stork-agent
    fi

    chown -R stork-agent "${home_dir}"

elif [ "$1" -gt 1 ]; then
    # Upgrade: restart service if it was running
    if command -v systemctl > /dev/null; then
        status=$(systemctl is-system-running || true)
        case "$status" in
            running|degraded|maintenance)
                if [ "$(systemctl is-active isc-stork-agent || true)" = "active" ]; then
                    systemctl restart isc-stork-agent
                fi
                ;;
        esac
    fi
fi

%preun agent
if [ "$1" -eq 0 ]; then
    # Removal: stop and disable service
    if command -v systemctl > /dev/null; then
        status=$(systemctl is-system-running || true)
        case "$status" in
            running|degraded|maintenance)
                systemctl disable isc-stork-agent
                systemctl stop isc-stork-agent
                ;;
        esac
    fi
    # Remove user from extra groups (named, kea)
    usermod -G "" stork-agent || true
fi

%postun agent
if [ "$1" -eq 0 ]; then
    # After removal: delete system user
    userdel stork-agent > /dev/null || true
fi

###########################################################################
# Install/remove scripts for stork-server
###########################################################################

%post server
if [ "$1" -eq 1 ]; then
    # Fresh install: create system user
    if ! getent passwd stork-server > /dev/null; then
        useradd --system --base-dir /var/lib stork-server
    fi

elif [ "$1" -gt 1 ]; then
    # Upgrade: restart service if it was running
    if command -v systemctl > /dev/null; then
        status=$(systemctl is-system-running || true)
        case "$status" in
            running|degraded|maintenance)
                if [ "$(systemctl is-active isc-stork-server || true)" = "active" ]; then
                    systemctl restart isc-stork-server
                fi
                ;;
        esac
    fi
fi

%preun server
if [ "$1" -eq 0 ]; then
    # Removal: stop and disable service
    if command -v systemctl > /dev/null; then
        status=$(systemctl is-system-running || true)
        case "$status" in
            running|degraded|maintenance)
                systemctl disable isc-stork-server
                systemctl stop isc-stork-server
                ;;
        esac
    fi
fi

%postun server
if [ "$1" -eq 0 ]; then
    # After removal: delete system user
    userdel stork-server > /dev/null || true
fi


%files

%files agent
%_bindir/stork-agent
%config(noreplace) %_sysconfdir/stork/agent.env
%_unitdir/isc-stork-agent.service
# Man page (glob captures compressed variant .8.gz)
%_man8dir/stork-agent.8*

%files server
# Server and DB management tool binaries
%_bindir/stork-server
%_bindir/stork-tool
%config(noreplace) %_sysconfdir/stork/server.env
# versions.json - updated with package, not noreplace
%config %_sysconfdir/stork/versions.json
%_unitdir/isc-stork-server.service
%_man8dir/stork-server.8*
%_man8dir/stork-tool.8*
# Frontend static files and example configs
%_datadir/stork/


%changelog
* Tue Jun 18 2026 Semyon Knyazev <samael@altlinux.org> 2.5.0-alt1
- Initial build for ALT Linux
