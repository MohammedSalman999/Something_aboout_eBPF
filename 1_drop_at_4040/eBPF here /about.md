This project demonstrates how to use eBPF to drop TCP packets destined for a specific port on a Linux system. The provided eBPF program will be attached to the XDP (eXpress Data Path) hook point to filter network packets at a very early stage in the networking stack.

Prerequisites
Linux Kernel: Ensure you are running a modern Linux kernel that supports eBPF and XDP.
Development Tools: Install clang, llvm, libbpf-dev, and kernel headers.
Steps to Compile and Load the eBPF Program
Install Dependencies

Open a terminal and run the following command to install the necessary packages:

bash
Copy code
sudo apt-get update
sudo apt-get install clang llvm libbpf-dev linux-headers-$(uname -r)
Save the eBPF Program

Save the following code in a file named tcp_drop_port.c:

c
Copy code
// tcp_drop_port.c
#include <uapi/linux/bpf.h>
#include <uapi/linux/if_ether.h>
#include <uapi/linux/ip.h>
#include <uapi/linux/tcp.h>
#include <bpf/bpf_helpers.h>

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, int);
    __type(value, __u16);
} drop_port SEC(".maps");

SEC("xdp")
int drop_tcp_port(struct xdp_md *ctx) {
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    struct ethhdr *eth = data;
    struct iphdr *ip;
    struct tcphdr *tcp;
    __u16 *port;
    int key = 0;

    if (eth + 1 > data_end)
        return XDP_PASS;

    if (eth->h_proto != htons(ETH_P_IP))
        return XDP_PASS;

    ip = data + sizeof(*eth);
    if (ip + 1 > data_end)
        return XDP_PASS;

    if (ip->protocol != IPPROTO_TCP)
        return XDP_PASS;

    tcp = (void *)ip + sizeof(*ip);
    if (tcp + 1 > data_end)
        return XDP_PASS;

    port = bpf_map_lookup_elem(&drop_port, &key);
    if (!port)
        return XDP_PASS;

    if (tcp->dest == htons(*port))
        return XDP_DROP;

    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";
Compile the eBPF Program

Use clang to compile the eBPF program:

bash
Copy code
clang -O2 -target bpf -c tcp_drop_port.c -o tcp_drop_port.o
Create the BPF Map and Set the Drop Port

Create the BPF map and set the port you want to drop (e.g., port 80):

bash
Copy code
sudo bpftool map create /sys/fs/bpf/drop_port type array key 4 value 2 entries 1 name drop_port
sudo bpftool map update name drop_port key hex 00 00 00 00 value hex 00 50 # Port 80 in hex (0x0050)
Attach the eBPF Program to the Network Interface

Attach the compiled eBPF program to your network interface (replace eth0 with your actual interface name):

bash
Copy code
sudo ip link set dev eth0 xdp obj tcp_drop_port.o
Verify the Program is Loaded

Check if the eBPF program and map are loaded correctly:

bash
Copy code
sudo bpftool prog
sudo bpftool map
Monitor Dropped Packets

Use ip -s link to monitor the network interface statistics and verify that packets are being dropped:

bash
Copy code
ip -s link show eth0
Script to Automate the Process
To make it easier, you can use the following script to compile, load the eBPF program, and set the drop port:

bash
Copy code
#!/bin/bash

# Variables
INTERFACE="eth0"
BPF_OBJECT="tcp_drop_port.o"
MAP_PATH="/sys/fs/bpf/drop_port"
DROP_PORT=80

# Compile the eBPF program
clang -O2 -target bpf -c tcp_drop_port.c -o $BPF_OBJECT

# Create the BPF map
sudo bpftool map create $MAP_PATH type array key 4 value 2 entries 1 name drop_port

# Convert the port to hex and update the map
PORT_HEX=$(printf "%04x" $DROP_PORT | sed 's/../& /g')
sudo bpftool map update name drop_port key hex 00 00 00 00 value hex $PORT_HEX

# Attach the XDP program to the network interface
sudo ip link set dev $INTERFACE xdp obj $BPF_OBJECT

echo "eBPF program loaded and attached to $INTERFACE, dropping TCP packets on port $DROP_PORT."
Notes
Make sure to replace eth0 with your actual network interface name.
The DROP_PORT variable should be set to the port number you want to drop.
This example uses port 80 (HTTP) for demonstration purposes.