// main serves as the package for "Kontainer"
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// main serves as the entrypoint for "Kontainer"
// Kontainer CLI only supports one command which is "run"
// Even though "child" command is available, it's not intended to be run directly by users
func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child": // Not intended for direct use by users
		child()
	default:
		panic(fmt.Sprintf("Unexpected Command -> %s\n", os.Args[1]))
	}
}

// run is the only user facing command supported by `kontainer`
// It handles all the system calls and cgroup setups needed for running a container but doesn't run it
// It is only responsible for setting up the container environment
// Then calls "child()" by reexecuting `kontainer` for acutally running the container
func run() {
	fmt.Printf("Running Command Parent (pid: %d) (uid: %d) (gid: %d) -> %v\n", os.Getpid(), os.Getuid(), os.Getgid(), os.Args[2:])

	// Sets up a command which will reinvoke `kontainer` with the `child` command and the passed params from the user
	// Also sets up IO for the command
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Setting up required system calls to be made with the invocation of the `child` command
	// The child process will run in an environment with the defined `Cloneflags` which are essentially different kinds of namespaces
	// Namespaces govern what a process can "see", essentially jailing/isolating the process/container
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER, // More types of namespaces can be used here for more isolation
		Unshareflags: syscall.CLONE_NEWNS, // Prevents the child process/container from sharing the view of it's file system with the parent process/host
		UidMappings: []syscall.SysProcIDMap{ // UID mappings for allowing rootless container execution
			{
				ContainerID: 0,
				HostID:      os.Getuid(), // Whichever UID the user invokes `kontainer` with will be given root within the container
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{ // GID mappings for allowing rootless container execution
			{
				ContainerID: 0,
				HostID:      os.Getgid(), // Whichever GID the user invokes `kontainer` with will be given root within the container
				Size:        1,
			},
		},
	}

	// Control groups face permission issues when invoked by non-root users, hence we only enable it for root user invocation
	// Control groups govern what a process can "do" and to what "limits", essentially restricting resource usage for the process/container
	// Control groups are managed via files/folders
	if os.Getuid() == 0 {
		setupCG()
		defer teardownCG()
	}

	must(cmd.Run()) // Execute the command, i.e. the child process/container for running the user command
}

// child is the function responsible for actually running the container
// It is invoked by "run()" which sets up the relevant namespaces and control groups
// This is also why this function shouldn't be directly invoked by user, as this needs to be run in a confined/setup environment which is done via "run()"
// It sets up the internals to complete the container and then runs the user given command
func child() {
	fmt.Printf("Running Command Child (pid: %d) (uid: %d) (gid: %d) -> %v\n", os.Getpid(), os.Getuid(), os.Getgid(), os.Args[2:])

	// Changing the hostname of the container
	syscall.Sethostname([]byte("kontainer"))

	// Changing the root directory within the container to a different path to prevent any references to the host root, hence preventing any "break-outs" from within the container
	syscall.Chroot("./rootfs")
	syscall.Chdir("/") // Explicitly setting the present working directory to the new root directory

	// Mounting the `proc` directory allowing for process management since this is where all the data related to processes are stored, without this `ps` will not work
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	// Setting up the user defined command to finally be executed within the fully initialized container with proper IO configured
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Creating Kontainer\n")
	must(cmd.Run()) // Run the user command in the container
	fmt.Printf("Exiting Kontainer\n")

	// Unmount the `proc` directory otherwise once the container exits, this directory would still show as mounted/under use
	syscall.Unmount("/proc", 0)
}

// setupCG is responsible for setting up the relevant control groups configurations
// It writes the configuration to the `/sys/fs/cgroup/` directory on the host (and hence requires root access, which is why non-root users cannot use control group features)
// Only adds cgroup for `pids`
func setupCG() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	os.Mkdir(filepath.Join(pids, "kontainer"), 0755) // Creating cgroup for managing processes for "kontainer" -> `/sys/fs/cgroup/pids/kontainer`
	must(os.WriteFile(filepath.Join(pids, "kontainer/pids.max"), []byte("20"), 0700)) // Setting the maximum limit of processes that can be spawned by this cgroup to 20
	must(os.WriteFile(filepath.Join(pids, "kontainer/notify_on_release"), []byte("1"), 0700)) // Enabled this feature, which notifies about the completion of processes in thie cgroup
	must(os.WriteFile(filepath.Join(pids, "kontainer/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700)) // Write the current PID, so the kernal considers it as a part of the `kontainer` cgroup
}

// teardownCG is responsible for cleaning up the relevant control groups configurations
// Only removes cgroup for `pids`
func teardownCG() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	os.Remove(filepath.Join(pids, "kontainer")) // Deleting cgroup for managing processes for "kontainer" -> `/sys/fs/cgroup/pids/kontainer`
}

// must is a utility function for handling critical errors
func must(err error) {
	if err != nil {
		panic(err)
	}
}
