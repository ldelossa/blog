---
title: From Zero To BPF Hero.
summary: Learn how to run BPF programs with libbpf.
date: 2021-04-02
toc: true
---

BPF development in the Linux kernel is occurring at a
rapid pace.
This makes finding up-to-date documentation and instructional
material difficult.

Even today, the examples in the book "Linux Observability with BPF" is a bit
out of date (but still very useful).

Since learning new Linux features always involves chasing
a moving target, I want to take a "teach a person to fish"
approach.

I'll share with you how I worked backwards from libbpf and
examples to get a working BPF program as of Linux 5.11.10
kernel.

## Setting Up Your Environment

Personally, I do my kernel hacking on a virtual
machine.

KVM is my hypervisor of choice for obvious reasons.

I found installing a fresh stable kernel avoids many headaches.
Doing this will ensure all the headers and other tools we need are installed or buildable.

Let's run through a quick kernel install.

Grab the latest stable kernel from www.kernel.org (5.11.10 as of this post).
Both a tar-ball of the source or a `git clone` of the stable branch will work.

Once its on your machine of choice build and install it.

```shell
$ make oldconfig 
$ make 
$ make install_modules
$ make install
$ make headers_install
```

If this is the first time you're building and installing a fresh kernel
expect it to fail, as you do not have the required dependencies.

You can google away most of these issues if you're on Ubuntu or Fedora.
As a tip, jot down what dependencies you need to apt or dnf install for next time.

Once the final `make headers_install` command runs without
issues you'll want to reboot and then run `uname -r` to confirm
you're on the vanilla stable kernel (it should simply return 5.11.10).

A fresh kernel install ensure all headers we need to build the
latest and greatest BPF programs are installed.

## Libbpf

The latest way of working with BPF programs is libbpf.

This library is located at `/tools/lib/bpf`. This is a library
that has replaced other tools such as BCC.

Build and install this library by changing directories to it
and running:
```shell
$ make
$ make install
```

As a spot check, confirm `pkg-config` can locate this library:
```shell
$ pkg-config --exists libbpf
$ $?
$ 0
```

Our linker will be able to find the BPF headers fine now.

## Hello World

The following BPF program is stolen from the book
"Linux Observability with BPF" book, but modified it to work with the
latest libbpf and related header files.
I don't take any credit for this.

Lets use this "hello world" program to demonstrate how to
build and run a BPF program with libbpf.

Copy this code into a hello_world.c file:

```c
#include <bpf/bpf_helpers.h>

SEC("tracepoint/syscalls/sys_enter_execve")
int bpf_prog(void *ctx) {
    char msg[] = "Hello, BPF World!";
    bpf_trace_printk(msg, sizeof(msg));
    return 0;
}

char _license[] SEC("license") = "GPL";
```

You'll notice we pull in the "bpf/bpf_helpers.h" file to
obtain the "SEC" macro and the "bpf_trace_printk" function.

"SEC" writes some information into an ELF section, which
the kernel will use to understand where this BPF program
attaches.

In our case, we are attaching it the execve system
call tracepoint.

Everytime execve is called our BPF program
will run.

Go and try to build this program, it will fail, on purpose
to demonstrated another step we need to accomplish.

```shell
$ clang -O1 -target bpf -c hello_world.c -o hello_world.o
```

You'll see a lot of errors about undefined types.

The issue is, the header `<bpf/bpf_helpers.h>` uses
types defined by the kernel.

We need a way to forward declare these types.

Time to meet a new friend "bpftool".

## bpftool

Inside the kernel source a tool exists at `tools/bpf/bpftool`
This tool has a magic power, it can export a header file
with **all** the types used within the kernel.

When we installed our kernel with BPF support a special
file was exposed in the sys virtual directory.

This file can be parsed to generate the required header
file mentioned above.

To build the tool 'cd' into the source directory and
do a quick:

```shell
$ make 
$ make install
```

This should install the tool to `/usr/local/sbin/bpftool` by
default.

You can generate the header file, cononically named "vmlinux.h"
by issuing the following command:

```shell
bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h
```

We can now link this header into our hello_world.c program.

```c
#include "../vmlinux.h" /* all kernel types */
#include <bpf/bpf_helpers.h>

SEC("tracepoint/syscalls/sys_enter_execve")
int bpf_prog(void *ctx) {
    char msg[] = "Hello, BPF World!";
    bpf_trace_printk(msg, sizeof(msg));
    return 0;
}

char _license[] SEC("license") = "GPL";
```

Attempting to compile this program again should work.

## A BPF loader

Getting a BPF program to run with libbpf takes a few steps.

I'll demonstrate with code you can copy into a "loader.c" file.

```c
#include <bpf/libbpf.h>
#include <fcntl.h>
#include <stdint.h>
#include <stdio.h>
#include <unistd.h>

int main(int argc, char *argv[]) {
struct bpf_object *obj = bpf_object__open("./hello_world.o");
if (!obj) {
  printf("failed to open bpf object file, %p\n", obj);
  return -1;
}
printf("created bpf object, %p\n", obj);

if (bpf_object__load(obj)) {
  printf("failed to load bpf object into kernel.\n");
  return -1;
}
printf("loaded bpf object into kernel\n");

struct bpf_program *prog = bpf_program__next(NULL, obj);
if (!prog) {
  printf("failed to query for bpf program in loaded object\n");
  return -1;
}
printf("extracted bpf program name: %s section name: %s\n",
       bpf_program__name(prog), bpf_program__section_name(prog));

struct bpf_link *link = bpf_program__attach(prog);
if (!link) {
  printf("failed to attach bpf program\n");
  return -1;
}
printf("bpf program is now running");

getchar();
return 0;
}
```

As you can see the phases are:

- open: parses the elf object file extracting BPF programs (funcions with a SEC() macro declared above it).
- load: runs each BPF program against the verifier and if it passes loads it into the kernel.
- attach: attaches a specific BPF program to the target specified by the SEC macro.

## A skeleton loader

Above is how to run a program with no boiler-plate code.

There's an easier way if you don't mind some magic.

The bpftool can generate a skeleton header which can be linked
into another loader program that takes care of the nitty-gritty we
demonstrated above.

Run the following command targeting our hello_world.o object file.

```shell
bpftool gen skeleton hello_world.o > hello_world.skel.h
```

You can explore this file yourself, but here is an example
of using the skeleton to reduce the boilerplate above.

Copy this code into "skeleton_loader.c".

```c
#include "hello_world.skel.h"
#include <stdio.h>

int main(int argc, char **argv) {
struct hello_world *hw = hello_world__open_and_load();
printf("created hello_world skeleton program\n");

int attached = hello_world__attach(hw);
if (attached) {
  printf("failed to attach hello world program\n");
}
printf("hello world program running");
getchar();
}
```

Nice and simple.

## Wrapping it up

It goes without saying, we are only scratching the surface of BPF.

In subsequent posts I'll be covering mount points, BPF maps, and exposing data
from BPF programs.

### Sources

[1] https://facebookmicrosites.github.io/bpf/blog/2020/02/20/bcc-to-libbpf-howto-guide.html

[2] https://docs.cilium.io/en/stable/bpf/ 

[3] Calavera, D., & Fontana, L. (2019). Linux Observability with BPF: Advanced Programming for Performance Analysis and Networking (1st ed.). O’Reilly Media.
