# Envoy Binaries

For some platforms, envoy is only distributed via docker images.

This repo contains:

- a utility to extract envoy binaries from upstream container images
- releases to house the envoy binary assets in an easily accessible place

# Releasing

When a release tag is created, the envoy binaries will be automatically extracted from the matching `envoyproxy/envoy` image tag and uploaded to the release.

# Building Envoy for darwin/arm64

Currently there's no way to run a github action on darwin/arm64, so the build must be done manually:

1. Clone envoyproxy/envoy
   ```bash
   git clone git@github.com:envoyproxy/envoy
   ```
1. Checkout the release tag
   ```bash
   git checkout v1.23.0
   ```
1. Build
   ```bash
   bazelisk build -c "opt" //source/exe:envoy-static --config=sizeopt
   ```
1. Wait a **very** long time.
1. Copy binary
   ```bash
   cp -f bazel-bin/source/exe/envoy-static ~/Downloads/envoy-darwin-arm64
   ```
1. Produce shasum
   ```bash
   shasum -a 256 ~/Downloads/envoy-darwin-arm64 > ~/Downloads/envoy-darwin-arm64.sha256
   ```
1. Drag & Drop the files to the release
