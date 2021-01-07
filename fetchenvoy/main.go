package main

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const envoyRepo = "envoyproxy/envoy"
const envoyPath = "/usr/local/bin/envoy"

var envoyArchs = []string{"arm64", "amd64"}

var log *zap.SugaredLogger

func init() {
	l, err := zap.NewDevelopment()
	log = l.Sugar()
	if err != nil {
		panic(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "fetchenvoy [image tag]",
	Short: "fetch envoy binaries from upstream containers",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := run(cmd.Context(), args[0])
		if err != nil {
			log.Errorw("failed", "error", err)
		}
		return err
	},
}

func main() {
	defer log.Sync() //nolint: errcheck
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context, tag string) error {
	log.Infow("fetching envoy binaries from docker images", "version", tag, "repo", envoyRepo)
	client, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("client error: %w", err)
	}

	var files []string
	for _, arch := range envoyArchs {
		// start a dummy container to grab binary from
		log.Infow("processing image", "arch", arch)
		id, err := doContainer(ctx, client, arch, tag)
		if err != nil {
			return err
		}

		// copy the binary out
		outFile := fmt.Sprintf("envoy-linux-%s", arch)
		err = doGetEnvoy(ctx, client, id, outFile)
		if err != nil {
			return err
		}

		// clean up
		err = client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{})
		if err != nil {
			log.Errorw("error removing container", "error", err)
			return err
		}

		files = append(files, outFile)
	}

	// list resulting files to stdout
	for _, file := range files {
		fmt.Printf("%s\n", file)
	}

	return nil
}

// extract the envoy binary to a local file path
func doGetEnvoy(ctx context.Context, client *client.Client, id string, dest string) error {
	log := log.With("id", id, "destination", dest)

	log.Info("copying envoy binary from container")
	tarFile, _, err := client.CopyFromContainer(ctx, id, envoyPath)
	if err != nil {
		log.Errorw("error copying from container", "error", err)
		return err
	}
	defer tarFile.Close() //nolint: err

	out, err := os.Create(dest)
	if err != nil {
		log.Errorw("could not create destination file", "error", err)
		return err
	}
	defer out.Close()

	// we get a tar format stream from CopyFromContainer
	t := tar.NewReader(tarFile)

	// this should be the envoy binary
	_, err = t.Next()
	if err != nil {
		log.Errorw("couldn't find envoy in archive", "error", err)
		return err
	}

	log.Infow("extracting envoy binary")
	readBytes, err := out.ReadFrom(t)
	if err != nil {
		log.Errorw("could not write binary out", "error", err)
		return err
	}
	log.Infow("copied envoy from container", "bytes", readBytes)

	return nil
}

// start and stop a container so we can copy files out of it afterwards
func doContainer(ctx context.Context, client *client.Client, arch string, tag string) (string, error) {
	log := log.With("arch", arch, "tag", tag)

	digest, err := getDigest(ctx, arch, tag)
	if err != nil {
		return "", err
	}

	image := fmt.Sprintf("%s@%s", envoyRepo, digest)
	log.Infow("pulling image", "image", image)
	imageBundle, err := client.ImagePull(ctx, image, types.ImagePullOptions{Platform: arch})
	if err != nil {
		log.Errorw("could not pull image digest", "image", image, "error", err)
		return "", err
	}
	defer imageBundle.Close() //nolint: err

	_, err = client.ImageLoad(ctx, imageBundle, true)
	if err != nil {
		log.Errorw("could not load image", "error", err)
		return "", err
	}

	config := container.Config{
		Image: image,
	}

	cnt, err := client.ContainerCreate(ctx, &config, nil, nil, &specs.Platform{Architecture: arch, OS: "linux"}, "")
	if err != nil {
		log.Errorw("could not create container", "error", err)
		return "", err
	}
	log.Infow("created container", "id", cnt.ID)

	err = client.ContainerStart(ctx, cnt.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Errorw("could not start container", "error", err)
		return "", err
	}

	log.Infow("started container", "id", cnt.ID)

	timeout := time.Duration(0)
	err = client.ContainerStop(ctx, cnt.ID, &timeout)
	if err != nil {
		log.Errorw("failed to stop container", "error", err)
		return "", err
	}

	log.Infow("stopped container", "id", cnt.ID)

	return cnt.ID, nil
}

// find the arch-specific manifest and digest from command line tools
// there doesn't seem to be a reasonable API for this
func getDigest(ctx context.Context, arch string, tag string) (string, error) {
	image := fmt.Sprintf("%s:%s", envoyRepo, tag)
	log := log.With("arch", arch, "image", image)
	log.Infow("searching upstream manifest", "image", image)

	cmd := exec.Command("docker", "manifest", "inspect", image)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorw("could not get manifest", "error", err, "output", out)
		return "", err
	}

	var fullManifest manifest
	err = json.Unmarshal(out, &fullManifest)
	if err != nil {
		log.Errorw("could not parse manifest", "error", err)
		return "", err
	}

	for _, m := range fullManifest.Manifests {
		if m.Platform.Architecture == arch {
			log.Infow("found manifest", "digest", m.Digest)
			return m.Digest, nil
		}
	}

	err = fmt.Errorf("no manifest for platform %s", arch)
	log.Errorw("manifest search failed", "error", err)
	return "", err
}
