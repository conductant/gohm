Current status

File open response flag, `fuse.OpenDirectIO`, seems to make hard link work,
but it causes git to return bus error.

When not using this flag, `git init` succeeds, but `git commit` would fail
on invalid object.
