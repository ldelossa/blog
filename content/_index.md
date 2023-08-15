# About Me

My name is Louis DeLosSantos and I'm currently a software engineer at Isovalent focusing on Cilium,
eBPF and Linux kernel networking.

My interests stretch across infrastructure, systems programming, networking, distributed systems and application architectures.

I've worked at a lot of places, built a lot of things, and enjoy sharing my experience along the way.

# Open Source

I'm an OSS enthusiast. 
Most of my career has been developing and contributing code in an open and shareable environment. 
Below are some of my open source projects.

## [ClairCore](https://github.com/quay/claircore)
I'm the initial designer and creator of the ClairCore libraries.

These libraries, written in Go, provide facilities for the static analysis of container security vulnerabilities. 

## [ClairV4](https://github.com/quay/clair)
I'm the initial designer of the latest edition of Clair.

You can see some of my talks about Clair here:

- [Container Security with Clair](https://youtu.be/AhdPC_d0Lso)
- [Inside the Indexer](https://youtu.be/pEAU6E1rZWo)

## [nvim-ide](https://github.com/)

A full featured IDE layer for Neovim. 

Heavily inspired by VSCode.

## [gh.nvim](https://github.com/ldelossa/gh.nvim)

A fully featured GitHub integration for Neovim utilizing the `litee.nvim` framework.

## [litee.nvim](https://github.com/ldelossa/litee.nvim)

Litee.nvim is a framework, written in Lua, for building plugins. 

There are several plugins reachable from the above link's readme. 

Most plugins deal with porting over feature present in VSCode that Neovim lacks.

# Kernel Contributions

- [bpf: utilize table ID in bpf_fib_lookup helper](https://lore.kernel.org/bpf/20230505-bpf-add-tbid-fib-lookup-v2-0-0a31c22c748c@gmail.com/)

This change allows a routing table ID to be used in the eBPF fib lookup helper.
When a fib lookup occurs we can now say which table we should perform the lookup
from. 
This essentially pushes policy routing, akin to the `ip rule` functionality, into
the `tc` layer of the Kernel.

# Talks

## [Simplifying and Making the Network Programmable with Kubernetes and SRv6 (Kubecon)](https://www.youtube.com/watch?v=ncYG-wScuL8&t=1s)

I co-presented this talk at Kubecon NA 2022. 

An overview of my work with Cilium and SRv6 L3VPN. 

## [Inside the Indexer (RedHat Commons)](https://www.youtube.com/watch?v=pEAU6E1rZWo&t=3s)

An overview of Clair's Indexer architecture. 

