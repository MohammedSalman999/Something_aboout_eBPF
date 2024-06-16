#include <uapi/linux/bpf.h>
#include <uapi/linux/if_ether.h>
#include <uapi/linux/ip.h>
#include <uapi/linux/tcp.h>
#include <bpf/bpf_helpers.h>

// Define a BPF map named drop_port, which is an array with a single entry.
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, int);
    __type(value, __u16);
} drop_port SEC(".maps");

// XDP program entry point
SEC("xdp")
int drop_tcp_port(struct xdp_md *ctx) {
    // Pointers to the start and end of the packet data
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;

    // Ethernet header pointer
    struct ethhdr *eth = data;

    // Check if Ethernet header is accessible
    if (eth + 1 > data_end)
        return XDP_PASS;  // Pass packet if header is not fully accessible

    // Verify packet is IPv4
    if (eth->h_proto != htons(ETH_P_IP))
        return XDP_PASS;  // Pass packet if not IPv4

    // IP header pointer
    struct iphdr *ip;
    ip = data + sizeof(*eth);

    // Check if IP header is accessible
    if (ip + 1 > data_end)
        return XDP_PASS;  // Pass packet if IP header is not fully accessible

    // Verify protocol is TCP
    if (ip->protocol != IPPROTO_TCP)
        return XDP_PASS;  // Pass packet if not TCP

    // TCP header pointer
    struct tcphdr *tcp;
    tcp = (void *)ip + sizeof(*ip);

    // Check if TCP header is accessible
    if (tcp + 1 > data_end)
        return XDP_PASS;  // Pass packet if TCP header is not fully accessible

    // Key for the BPF map (always 0 in this case as it's a single-entry array)
    int key = 0;

    // Lookup the TCP port number to drop from the BPF map
    __u16 *port;
    port = bpf_map_lookup_elem(&drop_port, &key);

    // If no port number found in the map, pass the packet
    if (!port)
        return XDP_PASS;

    // Drop the packet if the destination port matches the port number in the map
    if (tcp->dest == htons(*port))
        return XDP_DROP;

    // Pass the packet if the destination port does not match
    return XDP_PASS;
}

// GPL license
char _license[] SEC("license") = "GPL";
