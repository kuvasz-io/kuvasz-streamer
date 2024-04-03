---
layout: page
title: Installation
permalink: /installation/
nav_order: 20
---
# Installation

## Install on DEB based systems (Debian, Ubuntu, Kali, Raspbian, ...)

### Install kuvasz.io APT repository if it is not installed

```bash
sudo mkdir -m 0755 -p /etc/apt/keyrings/
wget -O- https://apt.kuvasz.io/kuvasz.gpg | gpg --dearmor | sudo tee /etc/apt/keyrings/kuvasz.gpg > /dev/null 
sudo chmod 644 /etc/apt/keyrings/kuvasz.gpg
echo "deb [signed-by=/etc/apt/keyrings/kuvasz.gpg] https://apt.kuvasz.io stable main" | sudo tee /etc/apt/sources.list.d/kuvasz.list
sudo chmod 644 /etc/apt/sources.list.d/kuvasz.list
```

### Install `kuvasz-streamer`

```bash
sudo apt-get update
sudo apt-get install kuvasz-streamer
```

## Install on RPM based systems (RHEL/OEL/RockyLinux/...)

### Install kuvasz.io RPM repository if it is not already installed

```bash
sudo cat <<EOF > /etc/yum.repos.d/kuvasz.repo
[kuvasz]
name=Kuvasz.io
baseurl=https://rpm.kuvasz.io
enabled=1
gpgcheck=1
gpgkey=https://rpm.kuvasz.io/RPM-GPG-KEY-kuvasz
sslverify=1
sslcacert=/etc/pki/tls/certs/ca-bundle.crt
EOF
```

### Install `kuvasz-streamer`

```bash
sudo dnf install -y kuvasz-streamer
```

## Install manually

1. Navigate to the [Releases Page](https://github.com/kuvasz-io/kuvasz-streamer/releases).
1. Scroll down to the Assets section under the version that you want to install.
1. Download the .tar,gz or .zip version needed.
1. Unzip the package contents.
1. Create the necessary config and map files
1. Run

## Build from source

Building from source assumes you are on Ubuntu 22.04 LTS

### Install dependencies
Minimal requirements are `Make` and `git`, but you will also need PostgreSQL client for testing.

```bash
sudo apt install build-essential git postgresql postgresql-contrib
```
### Install Go and tools

`kuvasz-streamer` requires Go 1.22 or higher. Install Go and GoReleaser using snaps, then install `staticcheck` from source and `golangci-lint` binary from its repository. Finally, add the local Go bin directory to the PATH.

```bash
sudo snap install go --channel=1.22/stable --classic
sudo snap install goreleaser --classic
go install honnef.co/go/tools/cmd/staticcheck@latest
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.56.2
export PATH=${PATH}:$(go env GOPATH)/bin
```

### Clone repository

Clone repo from GitHub

```bash
git clone https://github.com/kuvasz-io/kuvasz-streamer.git
cd kuvasz-streamer
```

### Build

This step will download all dependencies and build the binary for the underlying architecture

```bash
make build
```

Run code checks

This will run `staticcheck` and `golangci-lint` and `go vet` on the code to ensure it is clean.

```bash
make check
```

Build packages

This will build RPMs, DEBs and tarballs for all supported architectures.
Create a GPG key for signing the packages then export it to a file before running the `goreleaser` command.

```bash
gpg --generate-key
gpg --output ${HOME}/private.pgp --armor --export-secret-key <email address used to create key>
export NFPM_DEFAULT_PASSPHRASE=<passphrase>
make release
```

## Run test suite

The test suite relies on Docker to set up instances of all supported version of PostgreSQL 
and on Robot Framework to run end-to-end tests for all the supported features.

### Install Docker

First, install Docker following the instructions [here](https://docs.docker.com/engine/install/).
Then start it with

```bash
sudo systemctl enable --now docker
```

### Install pip

Install the `pip` package manager and Postgres driver

```bash
sudo apt install python3-pip
```

### Install Robot Framework

Then use `pip` to install Robot Framework and its dependencies

```bash
pip3 install psycopg2-binary robotframework robotframework-databaselibrary
```

### Run the test suit

```bash
make test
```
