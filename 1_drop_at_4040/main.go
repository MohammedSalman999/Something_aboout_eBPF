package main

import (
    "fmt"
    "os"
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/perf"
    "github.com/vishvananda/netlink"
)

// Define the structure of the BPF map (must match the eBPF program)
const mapName = "port_map"

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: sudo go run main.go <port>")
        os.Exit(1)
    }

    port := os.Args[1]
    var portNum uint16
    _, err := fmt.Sscanf(port, "%d", &portNum)
    if err != nil {
        fmt.Printf("Invalid port number: %s\n", port)
        os.Exit(1)
    }

    // Load the BPF object
    bpfObjects := struct {
        PortMap *ebpf.Map `ebpf:"port_map"`
    }{}
    spec, err := ebpf.LoadCollectionSpec("drop_tcp_port.o")
    if err != nil {
        fmt.Printf("Error loading collection spec: %v\n", err)
        os.Exit(1)
    }

    if err := spec.LoadAndAssign(&bpfObjects, nil); err != nil {
        fmt.Printf("Error loading and assigning collection: %v\n", err)
        os.Exit(1)
    }

    defer bpfObjects.PortMap.Close()

    // Update the BPF map with the new port number
    key := uint32(0)
    if err := bpfObjects.PortMap.Put(key, portNum); err != nil {
        fmt.Printf("Error updating BPF map: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Successfully updated port to %d\n", portNum)

    // Attach the XDP program to an interface (example: eth0)
    ifName := "eth0"
    link, err := netlink.LinkByName(ifName)
    if err != nil {
        fmt.Printf("Error getting link by name: %v\n", err)
        os.Exit(1)
    }

    err = netlink.LinkSetXdpFd(link, bpfObjects.PortMap.FD(), netlink.XDP_FLAGS_SKB_MODE)
    if err != nil {
        fmt.Printf("Error attaching XDP program to %s: %v\n", ifName, err)
        os.Exit(1)
    }

    fmt.Printf("Successfully attached XDP program to %s\n", ifName)
}
