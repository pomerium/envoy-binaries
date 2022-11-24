#!/bin/sh
# downloads a MacOS Envoy from homebrew
# takes `v1.x.y` as an argument

set -e

if [ $# -eq 0 ]; then
    echo "expected an envoy version"
    exit 1
fi

# homebrew doesn't use v, just a number triplet
_version=${1:1}

# AMD64 is not supported beyond 12.6
oras pull ghcr.io/homebrew/core/envoy:${_version} --platform "darwin/amd64:macOS 12.6"
tar --extract \
    --file envoy--${_version}.monterey.bottle.tar.gz \
    --directory . \
    --strip-components=3 \
    envoy/${_version}/bin/envoy
mv -f envoy envoy-darwin-amd64

oras pull ghcr.io/homebrew/core/envoy:${_version} --platform "darwin/arm64:macOS 13"
tar --extract \
    --file envoy--${_version}.arm64_ventura.bottle.tar.gz \
    --directory . \
    --strip-components=3 \
    envoy/${_version}/bin/envoy
mv -f envoy envoy-darwin-arm64

shasum -a 256 envoy-darwin-amd64 > envoy-darwin-amd64.sha256
shasum -a 256 envoy-darwin-arm64 > envoy-darwin-arm64.sha256