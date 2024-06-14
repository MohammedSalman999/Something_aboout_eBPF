#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <bpf/bpf_helpers.h>

// Define a BPF map to store the port number we want to filter
struct bpf_map_def SEC("maps") port_map = {
    .type = BPF_MAP_TYPE_ARRAY,
    .key_size = sizeof(uint32_t),
    .value_size = sizeof(uint16_t),
    .max_entries = 1,
};

// Main XDP (eXpress Data Path) program
SEC("xdp")
int drop_tcp_port(struct xdp_md *ctx) {
    uint32_t key = 0; // Key for looking up the port number in the map
    uint16_t *port = bpf_map_lookup_elem(&port_map, &key); // Lookup the port number
    if (!port) {
        // If port is not found in the map, pass the packet
        return XDP_PASS;
    }

    // Get pointers to the packet's data and data_end
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;

    // Parse Ethernet header
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end) {
        // If the Ethernet header exceeds the packet bounds, pass the packet
        return XDP_PASS;
    }

    // Check if the packet is an IPv4 packet
    if (eth->h_proto != bpf_htons(ETH_P_IP)) {
        // If not IPv4, pass the packet
        return XDP_PASS;
    }

    // Parse IP header
    struct iphdr *ip = (struct iphdr *)(eth + 1);
    if ((void *)(ip + 1) > data_end) {
        // If the IP header exceeds the packet bounds, pass the packet
        return XDP_PASS;
    }

    // Check if the packet is a TCP packet
    if (ip->protocol != IPPROTO_TCP) {
        // If not TCP, pass the packet
        return XDP_PASS;
    }

    // Calculate the length of the IP header (ip->ihl is in 32-bit words)
    int ip_hdr_len = ip->ihl * 4;
    // Parse TCP header
    struct tcphdr *tcp = (struct tcphdr *)((void *)ip + ip_hdr_len);
    if ((void *)(tcp + 1) > data_end) {
        // If the TCP header exceeds the packet bounds, pass the packet
        return XDP_PASS;
    }

    // Check if the destination port matches the port number in the map
    if (tcp->dest == bpf_htons(*port)) {
        // If the destination port matches, drop the packet
        return XDP_DROP;
    }

    // If the destination port does not match, pass the packet
    return XDP_PASS;
}

// Define the license for the eBPF program
char _license[] SEC("license") = "GPL";

// Default port to drop if map is not configured from userspace
__attribute__((constructor)) static void default_port(void) {
    uint32_t key = 0;
    uint16_t default_port = 4040; // Default port number
    bpf_map_update_elem(&port_map, &key, &default_port, BPF_ANY);
}
