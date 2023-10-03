---
title: Fun with Sway's Programmable Workspaces
summary: Use Sway's workspaces to automate your development environment
date: 2023-10-03
toc: true
---

I'm a big fan of the Sway compositor.

I actually moved to Sway after I raged-quit Gnome for its crap multi-monitor 
support.

In Sway you have 'workspaces' which you can assign to particular monitors.
This means when you unplug from the monitors and plug back in your workspace 
layouts are restored, something Gnome has not figured out how to do yet.

You can also name these workspaces and this got my A.D.D, automation-obsessed
brain thinking, what else can we do with these names?

## Automatically opening a terminal

Are you tired of this monotonous sequence of events:
1. Create a workspace
2. Open a terminal
3. `cd` to the directory you'll be working in

Probably not, but I was.

I thought it would be cool if creating a workspace also automatically opened
a terminal window to the repository you were going to hack on.

Maybe its possible to have to have the workspace's name drive opening a terminal
to a specific working directory.

The first issue tho, is long paths, right?

It would pretty annoying to open a legacy `GOPATH` project residing at
`~/git/gopath/src/github.com/ldelossa/blog` if the workspace's name needed to 
reflect the terminal's working directory.

If we squint at this problem tho, POSIX shells have a solution to this built in,
the [CDPATH](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/cd.html)
environment variable. 

This allows us to define `CDPATH=~/git/gopath/src/github.com` and now if you 
were to just type 'cd ldelossa/blog' the shell would try 
`cd ~/git/gopath/src/github.com/ldelossa/blog` and we'd be happy.

With `CDPATH` as an example we can come up with the following solution then
1. Create a workspace named after the relative path of the directory 
2. Take the relative path and postfix it onto each value in `CDPATH`
3. If the fully qualified path exists, open a terminal to that directory. 

Sounds good to me, lets now break this down into what we need to script.

### Creating a workspace easily

Okay, I'm going to cheat a bit here, I already scripted a lot of Sway stuff
into [Rofi](https://github.com/davatorium/rofi) GUI menus. 

Rofi is just a general 'selector' GUI that has some baked in functionality, but
you can think of it as a "pipe in a list, select an option, return that option"
kinda program.

Check out my [Sway-Rofi-Scripts](https://github.com/ldelossa/sway-rofi-scripts)
repository, even if you hate Rofi you can just steal the bash scripts. 

Anyway, I already have a Rofi script which I can hit `alt+n` in my Sway 
configuration to create a new workspace. 

### Using CDPATH to open terminal

Let's actually take a look at my Rofi script which creates a workspace and
opens a terminal to the appropriate directory.

```bash
#!/bin/bash
TERM=kitty
theme_overrides="listview { enabled: false;} num-rows { enabled: false;} num-filtered-rows { enabled: false;} case-indicator { enabled: false;} textbox-num-sep { enabled: false;}"

workspace=$(rofi -p "New workspace" -dmenu -theme-str "$theme_overrides")
[[ -z $workspace ]] && exit
swaymsg workspace $workspace
IFS=':' read -ra cdpaths <<< "$CDPATH"
for path in "${cdpaths[@]}"; do
    echo "$path/$workspace"
    if [[ -d "$path/$workspace" ]]; then
        swaymsg exec "$TERM --detach --working-directory \"$path/$workspace\""
        exit
    fi
done
```

Alright, I know your eyes are bleeding, I hate looking at bash, but that 
is partly because I don't take the time to write "pretty" bash. 

Regardless, this script accomplishes what we set out to, lets walk thru it.

The first couple lines don't matter too much, we set our terminal to kitty, which
is my preferred terminal, and do some theming. 

Next, we use `rofi` to prompt for a workspace name and exit if we get an empty
string.

```bash
workspace=$(rofi -p "New workspace" -dmenu -theme-str "$theme_overrides")
[[ -z $workspace ]] && exit
```

Next, we create the workspace, easy enough

```bash
swaymsg workspace $workspace
```

And finally we append each path in `CDPATH` to our workspace name, and if a 
directory is found, we open a terminal directly to that fully qualified 
path.


```bash
IFS=':' read -ra cdpaths <<< "$CDPATH"
for path in "${cdpaths[@]}"; do
    echo "$path/$workspace"
    if [[ -d "$path/$workspace" ]]; then
        swaymsg exec "$TERM --detach --working-directory \"$path/$workspace\""
        exit
    fi
done
```

That little `IFS` bit is such a funny bashism. 

Its a global which tells the bash interpretor 
"use this character as a delimiter", so the `read -ra` call creates an array
by considering each value between the `:` characters in `CDPATH` as elements.

It's probably a well known trick at this point, but the uninitiated are probably 
saying "WTF is that?" right now.

## Opening terminals in an existing workspace

Of course I wasn't happy with just the above being sorted out.

My next irk was when I needed to open another terminal I'd have to again `cd`
into the directory the workspace was focused on.

There are ways around this but they require you to be inside an active terminal
and 'spawn' another one.

This breaks down if you're focused inside a GUI application but want a quick 
terminal. 

I wanted a global Sway shortcut which looks at the workspace's name and tries
to open a terminal in that directory first, before simply opening one at the 
terminal's default working directory.

### A shell opening script

In `Sway` you can bind keyboard shortcuts to script execution.

So, what we can do is build a script which performs the following task:
1. Get the current Sway workspace
2. Postfix it to every path prefix in `CDPATH`
3. Check if it's a directory, if so open terminal to this directory
4. If not, open a terminal to the default working directory.

Given the example just before, this should be a piece of toast. 

```bash
#!/bin/bash

workspace=$(swaymsg -t get_workspaces -r | jq -r -c '.[] | select(.focused == true) | .name')

IFS=':' read -ra cdpaths <<< "$CDPATH"
for path in "${cdpaths[@]}"; do
    echo "$path/$workspace"
    if [[ -d "$path/$workspace" ]]; then
        swaymsg exec "$TERM --detach --working-directory \"$path/$workspace\""
        exit
    fi
done
swaymsg exec "$TERM --detach"
```

The above is real fun, because we get to use `jq` to extract the currently 
focused workspace.


```bash
workspace=$(swaymsg -t get_workspaces -r | jq -r -c '.[] | select(.focused == true) | .name')
```

This lists all the workspaces using `swaymsg` then iterates over them using `jq`,
selects the workspace that has its `focused` field set to true, and then returns
its name.

We then just do the same `CDPATH` searching we did in the earlier example.

## Plugging this into Sway Configuration

We need to make these scripts run in Sway, we have a few concerns here.

First, we need to get $CDPATH into our Sway config. 

Usually, `CDPATH` would be set in a `.bashrc` or a `.zshrc`, however Sway has
no idea about your shell's environment. 

There are some fancy ways to plumb environment variables into Sway's environment,
however these usually involve launching Sway via systemd, and IIUC this is not
the "common" setup. 
At least, I opted not to do this.

So, I simply just added it at the top of my sway.conf


```txt
set $cdpaths '/home/louis/git/gopath/src/'
```

Now, we can call our scripts with this variable


```txt
bindsym $mod+Return exec 'CDPATH=$cdpaths TERM=$term ~/.config/sway/shell_open.sh'
```

In this configuration line I bind 'alt+return' to our script above and plumb
our environment variables into the script's call.

The same is done for our workspace creation script, but in my config, I just 
call my rofi script:


```txt
bindsym $mod+n exec 'CDPATH=$cdpaths ~/git/sway/sway-rofi-scripts/sway-new-workspace.sh'
```

## End Automation

I wanted to write this post was to show how powerful exporting state from 
the compositor can be for scripting. 

This post should help get your gears turning. 
With the ability to introspect details of your compositor's running state 
you can really take automation as far as your imagination warrants you.

If you want to see any of the concepts we talked about in further details you 
can check out my [sway dotfiles]("https://github.com/ldelossa/dotfiles/tree/master/config/sway").
