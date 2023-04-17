---
title: Enough (x86_64) Assembly To Be Dangerous
summary: Discover the key assembly concepts that are essential to comprehend C. 
date: 2023-04-16
toc: true
mermaid: true
---

## Assembly?!?!

I know what you're thinking, "who would write assembly in 2023?" and, I totally
agree with the sentiment. 

No one should be writing assembly in 2023 and we should all trust our compilers.

However, if you choose to work in a systems programming context, or you want to
get more familiar with C, learning a least a bit of assembly can help fill the
gap. 

## Following Along

Since assembly programming occurs so close to the CPU its very easy to follow
along with this post.

I will be on Linux and will be utilizing the [System V Application Binary Interface](https://refspecs.linuxbase.org/elf/x86_64-abi-0.99.pdf)

I will also expect you to have the following dependencies:
* yasm
* gcc
* gdb

These dependencies should be in your package manager of choice. 

## The Basics

Learning Intel's x64 assembly is a daunting task. 
The instruction manual is, literally, thousands of pages and broken into sections.

Luckily, its not our goal to become assembly experts.
Instead, we just want a basic understanding of what is happening under the hood
when you compile and run your C programs.

Also luckily, the road to this understanding isn't too long. 

Let's go over the basics of assembly programming.

## Registers

Register's are a small memory area, separate from your computer's RAM, which your
CPU prefers to work with. 

Why you ask? Because, these registers are extremely quick for the CPU to access.

So much so, that some assembly instructions **require** that values be in specific
registers to operate on them.

Since we are focusing on x64 assembly we have 64bit registers available to us.

What's fun about this, is each 64bit register can be accessed as a 32, 16, or
8 bit registers with a specific name. Welcome to Intel's incredible backwards 
compatibility model, which supports all the way back to the 8086 processor, 
a 16bit processor.

Here's a table of all the registers.

| Register | Size (bits) | Description |
| --- | --- | --- |
| RAX | 64 | Accumulator register |
| EAX | 32 | Lower 32 bits of RAX |
| AX | 16 | Lower 16 bits of RAX |
| AH | 8 | High 8 bits of AX |
| AL | 8 | Low 8 bits of AX |
| RBX | 64 | Base register |
| EBX | 32 | Lower 32 bits of RBX |
| BX | 16 | Lower 16 bits of RBX |
| BH | 8 | High 8 bits of BX |
| BL | 8 | Low 8 bits of BX |
| RCX | 64 | Counter register |
| ECX | 32 | Lower 32 bits of RCX |
| CX | 16 | Lower 16 bits of RCX |
| CH | 8 | High 8 bits of CX |
| CL | 8 | Low 8 bits of CX |
| RDX | 64 | Data register |
| EDX | 32 | Lower 32 bits of RDX |
| DX | 16 | Lower 16 bits of RDX |
| DH | 8 | High 8 bits of DX |
| DL | 8 | Low 8 bits of DX |
| RSI | 64 | Source index register |
| ESI | 32 | Lower 32 bits of RSI |
| SI | 16 | Lower 16 bits of RSI |
| RDI | 64 | Destination index register |
| EDI | 32 | Lower 32 bits of RDI |
| DI | 16 | Lower 16 bits of RDI |
| RBP | 64 | Base pointer register |
| EBP | 32 | Lower 32 bits of RBP |
| BP | 16 | Lower 16 bits of RBP |
| RSP | 64 | Stack pointer register |
| ESP | 32 | Lower 32 bits of RSP |
| SP | 16 | Lower 16 bits of RSP |
| R8 | 64 | General purpose register |
| R9 | 64 | General purpose register |
| R10 | 64 | General purpose register |
| R11 | 64 | General purpose register |
| R12 | 64 | General purpose register |
| R13 | 64 | General purpose register |
| R14 | 64 | General purpose register |
| R15 | 64 | General purpose register |

I wouldn't try to memorize all these, but rather, just refer back to them when
you need to remember.

There are general purpose registers, and registers which more often then not,
serve an important purpose during your code's execution. A quick summary of 
these special registers are:

`RIP` - The instruction pointer, it points to the **next** instruction the CPU
should run. Manipulating this register effectively makes your program "jump" to
different parts of the code you write.

`RSP` - The stack pointer, it maintains a pointer to the top of stack area 
during your program's execution. We'll dig into this quite a bit in the next
section.

`RBP` - The stack frame base pointer, it maintains a pointer to the start of the
current stack frame. While technically optional, its very handy to utilize as 
it allows debuggers to understand the call stack during a series of function calls.

### Register Usage

While there are some specialized usages of registers, when writing code of your
own, you'll be reading and writing to these registers to accomplish some goal.

I won't dig too much into [addressing modes](https://en.wikipedia.org/wiki/Addressing_mode)
or the deep details of [yasm syntax](https://www.tortall.net/projects/yasm/manual/html/manual.html#nasm-language)
so we'll just go over basic reading and writing to registers.

When reading and writing to registers, you can write from: 
* memory => register
* register => register
* register => memory 

However, you cannot write from memory => memory, you need to use a temporary
register for this.

Now, lets write, compile, and run our first assembly code to demonstrate data 
movement.
You'll see some new and interesting things such as the `section` keyword, and the
`main` label.
If your C and ELF sixth sense is tingling, you're not off, we'll discuss the 
relationship between these assembly concepts, C's main function, and ELF.

```asm
; move.asm
section .data
    a db 0xA; 10 in decimal
    b db 0x0
section .text
global main
main:
    mov r10, [a]    ; derefence `a` and place that value into r10
    mov r11, r10    ; move r10 into r11, r11=10
    add r11, 10     ; add 10 to r11, r11=20
    mov [a], r11;   ; dereference `a` and store r11, *a=20

    mov r10, b      ; place the address `b` into r10
    mov [r10], r11  ; dereference r10, which points to `b` and set it to r11 which equals 20
    nop
```

Compile and link this file like so:
```shell
yasm -f elf64 -g dwarf2 mov.asm -o mov.o
gcc -o mov mov.o
```

Next, lets debug this file with `gdb`, I'll annotate the `nexti` directives with
the comments of the assembly instructions it correlates to.

```shell
$ gdb mov
(gdb) break main
(gdb) run
(gdb) nexti ; derefence `a` and place that value into r10
(gdb) info registers r10
r10            0xa                 10
(gdb) nexti ; move r10 into r11, r11=10
(gdb) info registers r11
r11            0xa                 10
(gdb) nexti ; add 10 to r11, r11=20
(gdb) info registers r11
r11            0x14                20
(gdb) nexti ; dereference `a` and store r11, *a=20
(gdb) print (int)a
$1 = 20
(gdb) nexti ; place the address `b` into r10
info registers r10
r10            0x40401d            4210717
(gdb) nexti ; dereference r10, which points to `b` and set it to r11 which equals 20
(gdb) print (int)b
$2 = 20
```

We stumble into our first assembly-to-C concept overlap with this example.
Notice how I use the term "dereference" quite a bit here, that's on purpose.

If you squint at this example you'll start to realize that we are dealing with
pointers here. 

In assembly, almost all variables are what we consider in C, a pointer.

You can then "dereference" these pointers to follow them to memory, which is 
done with the `[]` characters.

Take this scenario in C:
```c
int *x;
*x = 10;
```

This is exactly the same as:
```
mov [r10], r11  ; dereference r10, which points to `b` and set it to r11 which equals 20
```
It's just that r10 is a **register** holding a pointer, and we dereference the
**register** to the memory location and write the value within `r11` at that location.


