# Kontainer

Kontainer is a mini container runtime completely written in Go.

It allows users to run simple Linux commands in an isolated environment, i.e. in a container on any Linux host.

Also provides Rootless Container Execution to users.

It makes use of many concepts including: 
- Linux Namespaces
- Control Groups
- System Calls
- Process Isolation
- Process Limiting
- ID Mapping

## Key Points
- No concept of images, cannot run arbitrary images
- Extremely basic in terms of functionalities, but covers all the concepts required for production grade containers
- Modeled after [runc](https://github.com/opencontainers/runc)
- Is not supposed to be a complete container engine, simply a low level container runtime

## Pre-Requisites
Since we are using various System Calls and Linux Kernal level concepts, a Linux Kernal is hence required for it's usage.

Environments that will work:
- Linux Host
- Virtual Machine running Linux
Because they have access to the Linux Kernal

Environments that will not work:
- Containers
- Windows
- Mac
Because they don't have access to the Linux Kernal

Even for the Linux Kernal based environments, ensure you are using latest versions, some system calls being used were not available/exposed in older kernal versions.

Ensure you have the following setup and added to PATH in your environment:
- Go (v1.22+)
- Make
- Curl

## Setup
- Clone the repo
- From the root of the repo, run:
  ```bash
  > make build-all
  ```
- This will do two things:
  - Build the Go "kontainer" binary and store it at `bin/kontainer`
  - Download the alpine root file system and unpack + save it at `rootfs/`, which is the filesystem being used within the container
- For cleaning things after being done playing with the tool, from the root of the repo, run:
  ```bash
  > make clean-all
  ```
- This will also do two things:
  - Remove the Go "kontainer" binary from `bin/kontainer`
  - Remove the unpacked alpine root file system from `rootfs/`

## Examples
You can run any basic Linux commands within the container which ships with the alpine root file system.
```bash
> bin/kontainer run <CMD>
```

- For listing files and directories at the root of the container:
  ```bash
  > bin/kontainer run ls
  ```
- For getting the user from within the container:
  ```bash
  > bin/kontainer run whoami
  ```
- For getting the hostname from within the container:
  ```bash
  > bin/kontainer run hostname
  ```
- For getting a shell within the container (Most Fun):
  ```bash
  > bin/kontainer run /bin/sh
  ```
  We can only get the `sh` shell here, since that's what the alpine root file system ships with, feel free to change the file system being used and/or adding other utilities/tools you'd want inside the container

## Further Improvements
- Network Namespace Isolation
- Use Cobra + Viper for the CLI and refactor the codebase
- Add more configurations/commands/options/features similar to [runc](https://github.com/opencontainers/runc)
