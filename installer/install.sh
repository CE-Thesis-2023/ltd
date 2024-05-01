#!/bin/bash

GITHUB_USERNAME='ttn-nguyen42'
GITHUB_PASSWORD='Ttng132465'

GO_DOWNLOAD=https://go.dev/dl/go1.21.4.linux-amd64.tar.gz
GO_BIN=/usr/local/go/bin/go

# Is git is not installed, install git
if ! [ -x "$(command -v git)" ]; then
    echo 'Warning: git is not installed. Installing git' >&2
    sudo apt-get install git
    exit 1
else
    echo 'git is installed'
fi

# Is wget installed, install wget
if ! [ -x "$(command -v wget)" ]; then
    echo 'Warning: wget is not installed. Installing wget' >&2
    sudo apt-get install wget
    exit 1
else
    echo 'wget is installed'
fi

# Is Docker installed, install Docker
if ! [ -x "$(command -v docker)" ]; then
    echo 'Warning: Docker is not installed. Installing Docker' >&2
    install_docker
    exit 1
else
    echo 'Docker is installed'
fi

# Is go installed, install go
if ! [ -x "$(command -v go)" ]; then
    echo 'Warning: go is not installed. Installing go' >&2
    install_go
    exit 1
else
    echo 'go is installed'
fi

# Pull the Local Transcoder Device repository
git clone "https://$GITHUB_USERNAME:$GITHUB_PASSWORD@github.com/CE-Thesis-2023/ltd.git" ./ltd
git clone "https://$GITHUB_USERNAME:$GITHUB_PASSWORD@github.com/CE-Thesis-2023/backend.git" ./backend
ROOT=$(pwd)

cd $ROOT/backend
$GO_BIN mod download
cd $ROOT

cd $ROOT/ltd
$GO_BIN mod download
$GO_BIN build -o main src/main.go
mv main ..
cp -r opengate ..
cp -r configs.json ..
cd $ROOT

rm -rf ltd
rm -rf backend

function install_docker() {
    # Add Docker's official GPG key:
    sudo apt-get update
    sudo apt-get install ca-certificates curl
    sudo install -m 0755 -d /etc/apt/keyrings
    sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
    sudo chmod a+r /etc/apt/keyrings/docker.asc

    # Add the repository to Apt sources:
    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" |
        sudo tee /etc/apt/sources.list.d/docker.list >/dev/null
    sudo apt-get update

    sudo apt-get install docker-ce \
        docker-ce-cli \
        containerd.io docker-buildx-plugin \
        docker-compose-plugin
}

function install_go() {
    COMPRESSED_FILE='go.tar.gz'
    wget -O $COMPRESSED_FILE $GO_DOWNLOAD
    tar -xf $COMPRESSED_FILE
    sudo mv go /usr/local
    # Is GO_BIN is installed
    if ! [ -x "$(command -v go)" ]; then
        echo 'Error: Installing go failed' >&2
        exit 1
    fi
}
