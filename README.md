# Envoy Binaries

For some platforms, envoy is only distributed via docker images.

This repo contains:

- a utility to extract envoy binaries from upstream container images
- releases to house the envoy binary assets in an easily accessible place

# Releasing

When a release tag is created, the envoy binaries will be automatically extracted from the matching `envoyproxy/envoy` image tag and uploaded to the release.

