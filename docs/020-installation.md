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
sudo sh -c 'echo "deb https://apt.kuvasz.io stable main" > /etc/apt/sources.list.d/kuvasz.list'
wget --quiet -O - https://apt.kuvasz.io/gpg.key | sudo apt-key add -
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
gpgkey=https://rpm.kuvasz.io/gpg.key
sslverify=1
sslcacert=/etc/pki/tls/certs/ca-bundle.crt
EOF
```

### Install `kuvasz-streamer`

```bash
dnf update
dnf install kuvasz-streamer
```

## Install manually

1. Navigate to the [Releases Page](https://github.com/kuvasz-io/kuvasz-streamer/releases).

1. Scroll down to the Assets section under the version that you want to install.
1. Download the .tar,gz or .zip version needed.
1. Unzip the package contents.
1. Create the necessary config and map files
1. Run

