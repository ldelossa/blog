---
title: The gh.nvim guide
summary: This post serves the documentation for gh.nvim.
date: 2022-05-26
toc: true
---

# Introducing gh.nvim

[gh.nvim](https://github.com/ldelossa/gh.nvim) is a tool for working with the GitHub platform through Neovim. 

It was developed to alleviate the need to jump between editor and browser and 
because its author felt performing code reviews should be done inside the editor 
where all your language inspection tools exist. 

`gh.nvim` is a bit different then other GitHub plugins because it favors 
bringing all the code to your local machine and editing it as if it was just 
another branch you were working on. 

This allows seamless integration with LSP tools inside the editor and allows the 
review to make updates to the code in a natural way. 

In order to do this, `gh.nvim` utilizes both the `gh` and `git` CLI tools extensively. 

When `gh.nvim` opens a pull request it add the pull's remote to the local git 
repository, fetch the branch, and check it out. As you browse the code within 
`gh.nvim` the plugin will be checking out commits as you do. This allows you to 
visualize, run, debug, and inspect the code locally at every commit.

This blog post will be a constant "WIP" as features are added and shift around
during development. `gh.nvim` is new and has not reached a stable release yet.

## Getting started

### CLI dependencies

`gh.nvim` requires two external CLI tools to work. 
- `gh` - https://cli.github.com/
- `git` - https://git-scm.com/docs/gitcli

Ensure that the `gh` tool is configured to use the correct `git_protocol` for 
your usage. The `git_protocol` can be either "https" or "ssh".
If "ssh" is used ensure that your ssh-agent has the correct public keys added 
and is configured correctly. This is important since `gh.nvim` must pull the 
code locally and will fail if the correct protocols are not being used. 

When `gh.nvim` opens a timer is started. This timer will perform no action until 
an issue or a pull request is opened (so don't worry if you see this when you 
don't plan on using `gh.nvim` at the moment).

### litee.nvim

`gh.nvim` is implemented with help of [litee.nvim](https://github.com/ldelossa/litee.nvim) 
and is a Neovim plugin depedency. It must be installed for `gh.nvim` to work. 

`litee.nvim` provides the implementation of a unified "panel" similar to VSCode's 
and JetBrains panels. These panels allow components to be "registered" into them 
and can display multiple tools at once. 

This allows `gh.nvim` to work along side other `litee.nvim` plugins such as 
[litee-calltree.nvim](https://github.com/ldelossa/litee-calltree.nvim) and others. 

Therefore, `gh.nvim`'s panel is configured via `litee.nvim`'s configuration, just 
like all the other plugins. If you're looking at `gh.nvim` and wondering 
"how can I make the panel open on the top instead of right" then you'll want to 
look at `litee.nvim`'s configuration. Maybe a little confusing, but this allows 
the panel code to be shared between multiple plugins. 

### Configuration

For full configuration details you'll want to check out the plugin's docs using
(:h gh.nvim). 

At a minimum, you must call `litee.nvim`'s setup function followed by `gh.nvim`'s.

A configuration which utilizes all the defaults would be:

```lua
require('litee.lib').setup({
    -- this is where you configure details about your panel, such as 
    -- whether it toggles on the left, right, top, or bottom.
    -- leaving this blank will use the defaults.
    -- reminder: gh.nvim uses litee.lib to implement core portions of its UI.
})
require('litee.gh').setup({
    -- this is where you configure details about gh.nvim directly relating
    -- to GitHub integration.
})
```
Place the following function calls anywhere you setup your `neovim` plugins.

## Working with pull requests

### Opening a Pull Request

A pull request can be opened with "GHOpenPR". 

If no PR number is provided a `vim.ui.select` dialogue is presented. 

Typically, telescope and fzf.lua will override this providing a fuzzy searcher 
of the PRs.

![GHOpenPR](/ghopenpr.gif)

### Browsing Commits

Once a PR is opened the PR panel will toggle itself visible. 

Within this panel there will be a "commits" tree. 

By hitting (default) \<CR\> on a commit object the "Commits" panel will now open 
and `gh.nvim`'s diff-view will be presented with the first changed file of that commit. 
The underlying filesystem is also checked out to this commit so that running the 
code locally produces the exact functionality of the code at this commit. 

You can access the commit message by pressing (default) "d" on the commit node 
in the tree. 

If you'd like to comment on the commit itself, or the commit message popup is 
too small to read the entire commit message, you can hit (default) \<CR\> on 
the root commit object in the tree to open a commit conversation buffer in a new tab.

![GHCommit](/ghcommit.gif)

### Browsing files changed

If you prefer to browse at a higher level, a "Files Changed" tree exists as well. 

Anytime a file is opened the underlying repo is checked out to the HEAD of the 
pull request, providing an aggregated diff view of what has changed in the pull 
request. 

![GHFilesChanged](/ghfileschanged.gif)

### Working with review comments

The "Conversation" tree holds threads of conversations about areas of the code. 

By hitting (default) \<CR\> on a thread node the diffview will be toggled open and 
the thread will be displayed on the opposite side in which the thread refers to. 

From there you can use the opened thread window to perform actions on the 
conversation. By default \<Ctrl-a\> will open an actions dialogue when your cursor
is on top of a comment and \<Ctrl-s\> will submit any text you've typed in the bottom
text area.

You can close the thread window by issuing "GHToggleThread" on the same line as 
the open thread or on a blank line.

If multiple threads exist on a line you can use "GHNextThread" to cycle through 
them.

You can create your own thread on any line with a "+" sign in the sign column, 
using "GHCreateThread" 

You may hit (default) \<Ctrl-r\> to resolve the thread if you have permissions to 
do so.

![GHThreads](/ghthreads.gif)

### Working with pull request comments

Pull request comments are separate from review comments. Pull request comments
are not affiliated with a file diff. 

When the PR is initially opened an issue buffer is opened with the pull request's
comments. 

This buffer works just like the above thread buffer except you cannot resolve a
issue buffer.

If the issue buffer isn't present you can hit (default) \<CR\> on the root PR node
to open it in a new tab.

![GHPRIssues](/ghprissues.gif)

### Searching PRs

The "GHSearchPRs" command can be used to search for pull requests associated with
the repository Neovim is currently opened to.

Once the command is issued a `vim.ui.input` prompt will appear. 
This prompt will take further query strings as outlined [here](https://docs.github.com/en/search-github/searching-on-github/searching-issues-and-pull-requests):

If you do not specify any further query string all PRs will be listed, including
closed PRs. 
Be aware, this may exceed the "1000" PR list limit enforced by the GitHub API. 
You may want to filter with "is:open" to avoid this.

![GHSearchPRs](/ghsearchprs.gif)

## Working with Reviews

### Starting and submitting a review

A review can be started with the "GHStartReview". When this command completes
a new tree will appear indicating a pending review is in progress.

From this point on every thread comment that is made is made in "pending" state
which is also indicated in the conversations panel. 

If you close `neovim`, re-open it, and open the PR you'll be placed back into
your pending review. If you want to cancel the review use "GHDeleteReview". This
will drop any pending comments and remove the pending review state.

Once you've created all your comments you can use the "GHSubmitReview" command 
to submit it.

![GHReviews](/ghreviews.gif)

### Listing requested reviews

The "GHRequestedReview" command can be used to list pull requests your username
has explicitly been requested to review.

### Listing recently reviewed pull requests

The "GHRequestedReview" command will only show pull requests you've been requested
to review, and once you've submitted a review, it will no longer list the reviewed PR.

To list pull requests you've recently reviewed but are still open you may use the
"GHReviewed" command.

### Browsing by review comment

Submitted reviews can be opened with (default) \<CR\> in a new "Reviews" panel.
This panel aggregates review comments, further organizing threaded comments by
reviews.

### Immediately approving a review

The "GHApproveReview" command can be used to start a review and immediately approve
it. This is helpful if you do not plan on making any comments and would simply
like to approve the review with an optional comment.

## Working with Issues

Issues can be previewed, opened, commented on, and browsed.

The "GHOpenIssue" command works similarly to the "GHOpenPR" command. If no number
is provided to it, it will open a `vim.ui.select` dialogue.

Once an issue buffer is opened the buffer works exactly like the pull request
issue buffer we introduced earlier. 

Placing the cursor over any "#1234" formatted string will open a preview popup
of that issue and by default the keybinding "gd" will open the issue under the
cursor.

Issues are independent of a pull request being opened and are their own feature
in `gh.nvim`.

![GHOpenIssue](/ghissues.gif)

### Searching Issues

The "GHSearchIsssues" command can be used to search for issues on the GitHub platform.

Once the command is issued a `vim.ui.input` prompt will appear. 
This prompt will take further query strings as outlined [here](https://docs.github.com/en/search-github/searching-on-github/searching-issues-and-pull-requests):

If you do not specify any further query string an attempt to list all Issues on
GitHub will take place, and you will hit a GitHub limit.

You may want to filter with "repo:{owner}/{name}" and even further with "repo:{owner}/{name} search_term in:title"
to ensure you do not over extend your search.

![GHSearchIssues](/ghsearchissues.gif)

## Notifications

The "GHNotifications" command will open a buffer full of notifications for the 
current repository Neovim is opened to.

Within this buffer multiple notifications are presented. Each one can be opened
with (default) \<CR\>. Opening an issue will open an issue buffer while opening a
pull request will close any existing pull request and open the specified one.

You can preview the notification by pressing (default) "d" when the cursor is
over one. 

You can mark a notification as read, which removes it from the notification buffer.
You can also mark a notification as unsubscribed, which marks it as read and ignores
any further events other then a mention of your username.

![GHNotifications](/ghnotifications.gif)

## Completion

`gh.nvim` provides a completion function for auto-completing usernames and 
issues.

The completion is automatically registered in thread and issue buffers and can
optionally be registered in buffers with the "./git" string in their names. The
latter being useful if you use `neovim` as your `git` editor.

The completion function is registered to omnifunc with the default keybindings
being \<C-x\>\<C-o\>.

![GHCompletion](/ghcompletion.gif)
