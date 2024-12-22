# Instructions for components of this package

Functions within this directory should only be called by the main thread, which
is running the event loop.

When the event loop receives a function (closure), that function will be
specific to an object whose code
resides in the screen package. That code, specific to the action, will call out
to functions in this package.

## What kind of functions go here?

The implementation detail code for geoms, etc, should go here for clarity.

## What kind of functions go in the screen directory?

Functions essential for control and setup relevant to being called by the event
loop, most often implemented within a closure.