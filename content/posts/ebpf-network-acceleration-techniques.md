---
title: eBPF Network Acceleration Techniques
summary: An exploration of how eBPF can accelerate the Linux network stack.
date: 2023-04-12
toc: true
mermaid: true
---

In my current role at Isovalent I've had a lot of exposure to various eBPF 
networking techniques. 

As a former network engineer, this is one of my favorite topics in computer 
science. 

Since eBPF is a moving target and its development is usually done behind 
mailing lists, I wanted to make a post which inventories and explores the 
various eBPF can be used to accelerate portions of the Linux network stack. 

I'll use the term "accelerate" loosely here to mean, anytime we can use eBPF to
reduce the amount of time spent in the kernel when delivering or sending 
network packets.

Examples in this post assume you are familiar with building eBPF programs and
attaching them to various points in the network stack.

## eBPF Device Redirection

The first technique we will explore is the ability to redirect packets to a 
particular device. 

The usecase can be explained in a diagram:

{{< mermaid >}}
flowchart LR
    lan["LAN"] ---|Packet| inf1
    subgraph host network namespace
        inf1["eth0\n[tc ingress]"]
        nets["network stack"]
        inf2["veth1.1"]
        inf1 ---> nets
        nets ---> inf2
        inf1 -.->|eBPF Redirect| inf2
    end
{{< /mermaid >}}

In the above diagram, a packet comes from a LAN and hits the `eth0` interface
in the host network namespace.

Lets assume the incoming packet is destined for another network namespace on
this host, which is a common case for container orchestration software such
as Kubernetes.

Traditionally, the packet will be processed by the `eth0` network device and 
then handed to the `IP` layer to determine if the packet should be delivered
locally or forwarded.
In our example, the packet would be forwarded to `veth1.1` where the destination
network namespace resides.

We can accelerate this process by hoping over the network stack, where the packet
is handed to the protocol handler, and redirect the packet directly to the interface
which connects the host network namepace with the destination network namespace.

The eBPF program in this example is placed on the `TC` ingress hook.
This hook runs after delivering the packet to any taps, such as tcpdump, but
before providing the packet to the protocol handler, such as `ip_rcv`.

Let's look at the eBPF helper functions which implements this:

```c
/*
 * bpf_redirect
 *
 * 	Redirect the packet to another net device of index *ifindex*.
 * 	This helper is somewhat similar to **bpf_clone_redirect**\
 * 	(), except that the packet is not cloned, which provides
 * 	increased performance.
 *
 * 	Except for XDP, both ingress and egress interfaces can be used
 * 	for redirection. The **BPF_F_INGRESS** value in *flags* is used
 * 	to make the distinction (ingress path is selected if the flag
 * 	is present, egress path otherwise). Currently, XDP only
 * 	supports redirection to the egress interface, and accepts no
 * 	flag at all.
 *
 * 	The same effect can also be attained with the more generic
 * 	**bpf_redirect_map**\ (), which uses a BPF map to store the
 * 	redirect target instead of providing it directly to the helper.
 *
 * Returns
 * 	For XDP, the helper returns **XDP_REDIRECT** on success or
 * 	**XDP_ABORTED** on error. For other program types, the values
 * 	are **TC_ACT_REDIRECT** on success or **TC_ACT_SHOT** on
 * 	error.
 */
static long (*bpf_redirect)(__u32 ifindex, __u64 flags) = (void *) 23;
```

As you can see, the eBPF helper wants the ifindex of the device to redirect the
packet to. 

Flags will determine if the packet is redirected to the ingress or egress path
of the interface.

Here's a simple eBPF program which performs the redirection.
For the full context checkout [this code](https://github.com/ldelossa/ebpf-net/tree/main/bpf_redirect)

```c
SEC("tc")
int tc_ingress_socket_redirect(struct __sk_buff *skb) {
        struct bpf_fib_lookup rt = {0};

        bpf_printk("performing skb redirect\n");

        bpf_skb_pull_data(skb, 0);

        struct bpf_sock_tuple tuple = {0};

        int ret = tuple_extract_skb(skb, &tuple);
        if (ret != 1) {
                bpf_printk("failed to extract tuple at layer: %d\n", ret*1);
                return TC_ACT_OK;
        }

        // do a fib lookup on the destination
        rt.family = AF_INET;
        rt.ifindex = skb->ifindex;
        rt.ipv4_src = tuple.ipv4.saddr;
        rt.ipv4_dst = tuple.ipv4.daddr;

        ret = bpf_fib_lookup(skb, &rt, sizeof(struct bpf_fib_lookup), 0);
        if (ret != 0) {
                bpf_printk("fib lookup failed: %d\n", ret);
                return TC_ACT_SHOT;
        }

        // perform eBPF redirect
        bpf_printk("redirecting to interface: %d\n", rt.ifindex);
        ret = bpf_redirect(rt.ifindex, 0);
        if (ret != TC_ACT_REDIRECT) {
                bpf_printk("BPF redirect failed: %d\n", ret);
        }
        return ret;
};
```

This program a bit contrived but it will do a FIB lookup for the destination 
IPv4 address and then perform a redirection to the interface returned by this
lookup.
