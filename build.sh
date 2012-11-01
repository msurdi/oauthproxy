#!/bin/bash -e
BUILD_DIR="build"
DIST_DIR="dist"
VERSION="0.1"
mkdir -p $BUILD_DIR
mkdir -p $DIST_DIR
for os in "linux" "darwin";do
	for arch in "amd64" "386";do
		echo "building ${os}-${arch} binary..."
		GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -o $BUILD_DIR/oauthproxy-${os}-${arch}
		echo "Creating package for ${os}-${arch}"
		tar cvzf $DIST_DIR/oauthproxy-${VERSION}-${os}-${arch}.tar.gz LICENSE README.md *.pem oauthproxy.conf.example $BUILD_DIR/oauthproxy-${os}-${arch} 
	done
done

