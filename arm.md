# ARM64

This guide explains how to build envoy for arm64.

1. Create an AWS arm64 instance (ie `m6g.medium`). Make sure the root device is at least 20GB in size.
2. Install docker:

   ```bash
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh
   sudo usermod -aG docker ubuntu
   ```

   Logout/login again.

3. Clone envoy repo:

   ```bash
   mkdir -p $HOME/src/github.com/envoyproxy
   git clone https://github.com/envoyproxy/envoy.git $HOME/src/github.com/envoyproxy/envoy
   cd $HOME/src/github.com/envoyproxy/envoy
   ```

4. Build:

   ```bash
   ci/run_envoy_docker.sh 'ci/do_ci.sh bazel.release.server_only'
   ```

   This will take several hours to complete.

5. Once done the binary is stored in /home/ubuntu/src/github.com/envoyproxy/envoy/build_release_stripped/envoy.

- https://github.com/envoyproxy/envoy/issues/1861
