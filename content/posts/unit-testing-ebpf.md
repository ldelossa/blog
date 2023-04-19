---
title: Unit Testing eBPF Programs
summary: A guide to effective eBPF unit testing
date: 2023-04-18
toc: true
---

## Unit Testing eBPF

Love it or hate it, writing unit tests is all but mandatory for your code. 

They provide a safety net when making changes, and give you that nice, warm
feeling when you see them all pass after a change.

While working on a Kernel patch I had to investigate writing unit tests for 
eBPF programs. 

Turns out, the Kernel developers have thought about this already and infrastructure 
exists to accomplish it. 

I'm going to provide a hands on example of unit testing a `TC` eBPF program.

In this test we want to confirm that looking up the route for a packet destined
for an external IP address results in selecting the default gateway.

We will have complete control over the network namespace which the test is ran 
in. 

If you have no idea what any of that means, don't worry, the concepts I will 
cover carry over to testing other types of eBPF programs.

## Testing environment

In this post I'm assuming you know how to compile your eBPF program utilizing
`clang`, `bpftool` and how to generate a `vmlinux.h` file. 

If you don't check out my other [post here](https://who.ldelossa.is/posts/bpf-zero-to-hero/)

With that said, we do need to level-set on your coding environment and the tools
we need in your coding environment to follow along. 

You must have:

* bpftool   - this, in additional to generating the vmlinux.h, will be used to 
generate a "skeleton" loader for your compiled eBPF program.
* clang     - we need this to compile eBPF programs
* make      - used to run my butchered up Makefile

You must also have `CAP_SYS_ADMIN` privileges on your machine, if you don't know
what that means, 99% of the time running as `root` will fill this requirement.

I will also assume you're on Linux, which you may think is obvious statement,
but [it's a dwindling assumption](https://github.com/microsoft/ebpf-for-windows)

Okay, one last assumption, you have `libbpf` installed correctly and clang/gcc can
locate it and compile your eBPF programs.

## Introducing the BPF_PROG_RUN command

The core functionality we want to focus on for unit testing eBPF program is a
eBPF command called "BPF_PROG_RUN". 

This command was renamed from "BPF_PROG_TEST_RUN" and this identifier maybe used
interchangeable.

A "command" is an enum value which can be passed to the `bpf` sys-call exposed
by Linux. 

However, `libbpf` usually wraps the `bpf` sys-call usage for convenience and 
sanity checking. 

Therefore, we'll focus on using `libbpf`'s wrapper around the "BPF_PROG_RUN" 
command, `bpf_test_run_opts`

Let's take a look at it's forward declaration:
https://elixir.bootlin.com/linux/v6.2.11/source/tools/lib/bpf/bpf.h#L454
```c
struct bpf_test_run_opts {
	size_t sz; /* size of this struct for forward/backward compatibility */
	const void *data_in; /* optional */
	void *data_out;      /* optional */
	__u32 data_size_in;
	__u32 data_size_out; /* in: max length of data_out
			      * out: length of data_out
			      */
	const void *ctx_in; /* optional */
	void *ctx_out;      /* optional */
	__u32 ctx_size_in;
	__u32 ctx_size_out; /* in: max length of ctx_out
			     * out: length of cxt_out
			     */
	__u32 retval;        /* out: return code of the BPF program */
	int repeat;
	__u32 duration;      /* out: average per repetition in ns */
	__u32 flags;
	__u32 cpu;
	__u32 batch_size;
};
#define bpf_test_run_opts__last_field batch_size

LIBBPF_API int bpf_prog_test_run_opts(int prog_fd,
				      struct bpf_test_run_opts *opts);
```

If we were to look at the implementation, we'd find that `bpf_prog_test_run_opts`
simply copies the provided `opts` to its a structure the Kernel will own, does
some sanity checking on the `opts` structure, and then calls the `bpf` sys-call
directly.

The arguments to the `libbpf` function takes an eBPF program file descriptor 
and a `opts` structure.

The eBPF program file descriptor represents an eBPF program which is loaded into
the Kernel, we'll demonstrate a convenient way of obtaining this file descriptor
later in this post.

The `opts` structure provides both mock data and options to the function. 
While some fields say 'optional' we will learn that it really depends on the 
eBPF program type you are testing, whether these fields are optional or not.

The important fields we'll utilize in this post are:

`sz` is always required, and its simply set to `sizeof(bpf_test_run_opts)`.

`data_in, data_size_in` allows you to provide mock data to the `ctx` that is 
passed into your eBPF program, in the case of a "TC" program, a mock IPv4 packet.

`ctx_in, ctx_size_in` allows you to pass in a mock ctx, in the case of a "TC" program, 
a mock `__sk_buff` structure, which is eBPF's representation of the Kernel's 
socket buffer.

## Test Case and Skeleton Loader

With the introduction of `bpf_test_run_opts` out of the way, lets start writing
our eBPF test case. 

We will also use `bpftool` to generate a skeleton loader, which is a header file
with functions for loading our compiled eBPF program into the Kernel and giving
us a handle to the loaded program. 

This handle can be used to obtain the file descriptor to the loaded eBPF program, 
and interact with it during the Kernel's runtime.

Our test case's goal is to ensure that a packet sourced from the host, destined
for an external node, selects the default route, and is forwarded to the correct
interface.

To test this we will be utilizing the eBPF helper `bpf_fib_lookup`. 

We don't need to understand how this helper works in details, suffice it to say
that we provide in the source and destination of a incoming packet, and it 
returns to us an interface, if any, that the packet would be forwarded to.

In our test case, we want to see the aforementioned interface be the default
gateway for the network namespace.

Our test packet will be sourced from `127.0.0.1` and its destination will be
`8.8.8.8`.

Since we are running a unit test, no data will actually be sent, and no side
effects outside of the host will occur.

Keep in mind, this test a bit contrived to show off a few features of the testing
infrastructure, and we lean more towards "demonstration" then "practicality".

Okay, so lets examine our test eBPF program:

`fib_lookup.bpf.c`
```c
#include "../vmlinux.h"
#include <bpf/bpf_helpers.h>

#define TC_ACT_OK		0
#define TC_ACT_SHOT		2
#define TC_ACT_REDIRECT		7

#define AF_INET		        2	/* Internet IP Protocol 	*/

struct bpf_fib_lookup fib_params = {0};

int fib_lookup_ret = 0;

SEC("tc")
int fib_lookup(struct __sk_buff *skb)
{
        struct iphdr *ip = 0;

        bpf_printk("performing FIB lookup\n");

        bpf_printk("fib lookup original ret: %d\n", fib_lookup_ret);

	    fib_lookup_ret = bpf_fib_lookup(skb, &fib_params, sizeof(fib_params),
					0);

        bpf_printk("fib lookup ret: %d\n", fib_lookup_ret);

	return TC_ACT_OK;
}

char _license[] SEC("license") = "GPL";
```

As you can see the test is very simple. 

We import the necessary headers and then we define two global variables, setting
them both to zero.

By defining these variables as global and setting them to zero, they actually 
become available to userspace via our skeleton.

Let's use the following Makefile to compile and generate a skeleton for this
eBPF program.

`Makefile`
```make
CFLAGS += -g3 \
          -Wall

LIBS = bpf

all: fib_lookup.bpf.o fib_lookup.skel.h

fib_lookup.bpf.o: fib_lookup.bpf.c
	clang -target bpf -Wall -O2 -g -c $<

fib_lookup.skel.h: fib_lookup.bpf.o
	bpftool gen skeleton $< > $@

test: test.c
	gcc $(CFLAGS) -l$(LIBS) -o $@ $<

.PHONY:
clean:
	rm -rf fib_lookup.bpf.o
	rm -rf fib_lookup.skel.h
	rm -rf test
```

Ignore the `test` binary for now, we'll write our test runner in the next 
section.

If we inspect the file "fib_lookup.skel.h" we come across the interesting
structure.

`fib_lookup.skel.h`
```c
struct fib_lookup_bpf {
	struct bpf_object_skeleton *skeleton;
	struct bpf_object *obj;
	struct {
		struct bpf_map *bss;
		struct bpf_map *rodata;
	} maps;
	struct {
		struct bpf_program *fib_lookup;
	} progs;
	struct {
		struct bpf_link *fib_lookup;
	} links;
	struct fib_lookup_bpf__bss {
		struct bpf_fib_lookup fib_params;
		int fib_lookup_ret;
	} *bss;
	struct fib_lookup_bpf__rodata {
	} *rodata;

#ifdef __cplusplus
	static inline struct fib_lookup_bpf *open(const struct bpf_object_open_opts *opts = nullptr);
	static inline struct fib_lookup_bpf *open_and_load();
	static inline int load(struct fib_lookup_bpf *skel);
	static inline int attach(struct fib_lookup_bpf *skel);
	static inline void detach(struct fib_lookup_bpf *skel);
	static inline void destroy(struct fib_lookup_bpf *skel);
	static inline const void *elf_bytes(size_t *sz);
#endif /* __cplusplus */
};
```

This is the handle to our loaded eBPF program which the skeleton loader returns
to us when we call:
```c
static inline struct fib_lookup_bpf *
fib_lookup_bpf__open_and_load(void)
```
In the same file.

The interesting bit here is:
```c
struct fib_lookup_bpf__bss {
    struct bpf_fib_lookup fib_params;
    int fib_lookup_ret;
} *bss;
```

Notice, we get access to our global zero initialized variables in the `bss` 
field.

This allows a userspace program to load the eBPF program, retrieve the handle 
to it, and then both "inject" and "read" values from globals before and after
`bpf_test_run_opts` is called.

This is exactly what our test runner is going to do.

## Writing the test runner

As eluded to above, we want our test runner to do the following:

* Load our eBPF test program into the Kernel, getting a handle to the 
`fib_lookup_bpf` structure defined in `bpf_lookup.skel.h`
* Inject a mock `bpf_fib_lookup` parameter structure into the test before 
its ran
* utilize `libpf`'s `bpf_test_run_opts` function to run our test in userspace
* read the resulting `fib_lookup_bpf` and `fib_lookup_ret` to determine if the
default gateway was used.

We have control of the network namespace the test runs in, so we can hard-code
the interface ID (ifindex) which represents the default gateway, making our 
test runner a bit simpler.

Let's take a look at the test runner:

`test.c`
```c
#include <bpf/bpf.h>
#include <bpf/libbpf.h>
#include <stdio.h>
#include <bpf/bpf_endian.h>
#include "fib_lookup.skel.h"
#include "net/ethernet.h"
#include "linux/ip.h"
#include "netinet/tcp.h"

#define TARGET_IFINDEX 2

// in our test, we only care that the packet is the correct size,
// since our test does not touch any packet data.
char v4_pkt[(sizeof(struct ethhdr) + sizeof(struct iphdr) + sizeof(struct tcphdr))];

// create an empty skb as mock data, our tests do not touch any skb fields.
struct __sk_buff skb = {0};

int main (int argc, char *argv[]) {
        struct fib_lookup_bpf *skel;
        int prog_fd, err = 0;

        // define our BPF_PROG_RUN options with our mock data.
        struct bpf_test_run_opts opts = {
                // required, or else bpf_prog_test_run_opts will fail
                .sz = sizeof(struct bpf_test_run_opts),
                // data_in will wind up being ctx.data
                .data_in = &v4_pkt,
                .data_size_in = sizeof(v4_pkt),
                // ctx is an skb in this case
                .ctx_in = &skb,
                .ctx_size_in = sizeof(skb)
        };

        // load our fib lookup test program into the Kernel and return our
        // skeleton handle to it.
        skel = fib_lookup_bpf__open_and_load();
        if (!skel) {
                printf("[error]: failed to open and load skeleton: %d\n", err);
                return -1;
        }

        // inject our test parameters into the fib lookup parameter, this primes
        // our test.
        skel->bss->fib_lookup_ret = -1;
        skel->bss->fib_params.family = AF_INET;
        skel->bss->fib_params.ipv4_src = 0x100007f;
        skel->bss->fib_params.ipv4_dst = 0x8080808;
        skel->bss->fib_params.ifindex = 1;

        // get the prog_fd from the skeleton, and run our test.
        prog_fd = bpf_program__fd(skel->progs.fib_lookup);
        err = bpf_prog_test_run_opts(prog_fd, &opts);
        if (err != 0) {
                printf("[error]: bpf test run failed: %d\n", err);
                return -2;
        }

        // check global variables for response
        if (skel->bss->fib_lookup_ret != 0) {
                printf("[FAIL]: fib lookup returned: %d", skel->bss->fib_lookup_ret);
                return -1;
        }

        if (skel->bss->fib_params.ifindex != TARGET_IFINDEX) {
                printf("[FAIL]: fib lookup did not choose default gw interface: %d", skel->bss->fib_params.ifindex);
                return -1;
        }

        printf("[PASS]: ifindex %d\n", skel->bss->fib_params.ifindex);
        return 0;
}
```

Let's update our Makefile to build our test runner as well.

```make
CFLAGS += -g3 \
          -Wall

LIBS = bpf

all: fib_lookup.bpf.o fib_lookup.skel.h test

fib_lookup.bpf.o: fib_lookup.bpf.c
	clang -target bpf -Wall -O2 -g -c $<

fib_lookup.skel.h: fib_lookup.bpf.o
	bpftool gen skeleton $< > $@

test: test.c
	gcc $(CFLAGS) -l$(LIBS) -o $@ $<

.PHONY:
clean:
	rm -rf fib_lookup.bpf.o
	rm -rf fib_lookup.skel.h
	rm -rf test
```

And finally lets provide a script which sets up a network namespace for this 
test runner to work inside, and runs the test.

```bash
#!/bin/bash
NETNS_NAME="netns-1"
n='sudo ip netns'
nexec="sudo ip netns exec $NETNS_NAME"

function setup_netns() {
    # add 'netns-1' network namespace where we'll
    # run our test.
    $n add $NETNS_NAME

    # setup loopback
    $nexec ip addr add 127.0.0.1 dev lo

    # setup a dummy interface which can route to the default gw, and 
    # setup a route to the default gw.
    $nexec ip link add name eth0 type dummy
    $nexec ip link set up eth0
    $nexec ip addr add 192.168.1.10/24 dev eth0
    $nexec ip route add default via 192.168.1.11

    # since 192.168.1.11 doesn't actually exist, create a perm arp-table entry 
    # for it, allowing fib lookup to succeed.
    $nexec ip neigh add 192.168.1.11 dev eth0 lladdr "0F:0F:0F:0F:0F:0F" nud permanent
}

function teardown_netns() {
    $n del $NETNS_NAME
}

setup_netns
$nexec ./test
teardown_netns
```

Now when we run this script we get the following output:
```shell
[PASS]: ifindex 2
```

## Summing it up

Lets summarize the major points in this post. 

A eBPF program can define global variables which can be modified, both before and
after, a userspace test run of the program.

The `BPF_PROG_RUN` command can be used run your eBPF program in user space,
which is wrapped by the `bpf_prog_test_run_opts()` function in `libbpf`.

Once the eBPF program is compiled into an object file, you can generate a skeleton
loader with `bpftool`, this skeleton loader will load your eBPF program into the 
Kernel, and also provide your userspace program access to the global variables
mentioned above.

Finally, you can write a userspace test runner which sets the global variables
of the loaded eBPF program before the test, and reads them after, allowing you
to determine if the eBPF program performed the actions you intended.
