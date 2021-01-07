package main

type manifest struct {
	Digest    string
	Platform  platform
	Manifests []manifest
}
type platform struct {
	Architecture string
}
