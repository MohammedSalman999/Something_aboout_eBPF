package main

import (
    "flag"  // Package for parsing command-line flags
    "fmt"   // Package for formatted I/O
    "log"   // Package for logging
    "os"    // Package provides a platform-independent interface to operating system functionality
    "os/signal" // Package for signal handling
    "syscall"   // Package provides low-level access to the operating system

    "github.com/cilium/ebpf"    // Package for interacting with eBPF
    "github.com/cilium/ebpf/link"   // Package for interacting with network interfaces
    "github.com/cilium/ebpf/loader" // Package for loading eBPF programs
)

// Define the default filename for the eBPF object file
const (
    bpfFileName = "tcp_drop_port.o"
    defaultPort = 4040  // Default port number to drop TCP packets
)

func main() {
    // Parse command-line flags to allow setting the port number
    port := flag.Uint("port", defaultPort, "TCP port to drop (default 4040)")
    flag.Parse()

    // Remove the memory lock limit to allow loading eBPF programs
    if err := rlimit.RemoveMemlock(); err != nil {
        log.Fatalf("failed to remove memlock: %v", err)
    }

    // Read the contents of the eBPF object file
    objBytes, err := bpfObjects.ReadFile(bpfFileName)
    if err != nil {
        log.Fatalf("failed to read eBPF object file: %v", err)
    }

    // Load the eBPF program from the object file bytes
    spec, err := ebpf.LoadCollectionSpecFromReader(objBytes)
    if err != nil {
        log.Fatalf("failed to load eBPF program: %v", err)
    }

    // Create a new eBPF collection
    coll, err := ebpf.NewCollection(spec)
    if err != nil {
        log.Fatalf("failed to create eBPF collection: %v", err)
    }
    defer coll.Close()

    // Retrieve the drop_tcp_port eBPF program from the collection
    prog := coll.Programs["drop_tcp_port"]
    if prog == nil {
        log.Fatalf("could not find eBPF program: drop_tcp_port")
    }

    // Attach the XDP program to the network interface
    xdpLink, err := link.AttachXDP(link.XDPOptions{
        Program:   prog,
        Interface: "eth0",  // Replace with the appropriate network interface name
    })
    if err != nil {
        log.Fatalf("failed to attach XDP program: %v", err)
    }
    defer xdpLink.Close()

    // Prepare the key and value for updating the BPF map with the port number
    key := uint32(0)
    value := uint16(*port)

    // Retrieve the drop_port BPF map from the collection
    portMap := coll.Maps["drop_port"]
    if portMap == nil {
        log.Fatalf("could not find eBPF map: drop_port")
    }

    // Update the drop_port BPF map with the specified port number
    if err := portMap.Update(&key, &value, 0); err != nil {
        log.Fatalf("failed to update eBPF map: %v", err)
    }

    // Print success message with the configured port number
    fmt.Printf("Successfully loaded eBPF program. Dropping TCP packets on port %d\n", *port)

    // Handle signals to gracefully detach eBPF program on termination
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs

    fmt.Println("Received signal, detaching eBPF program and exiting")
}
