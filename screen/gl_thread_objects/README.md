# Instructions for components of this package

Functions within this directory should only be called within the main thread, 
which is running the event loop.

## What kind of functions go here?

The implementation detail code for geoms, etc, should go here for clarity.

Every function in this package is going to run in a single threaded context.
There is no need to implement concurrency management for variables in here.

## What kind of functions go in the screen directory?

Functions essential for control and setup relevant to being called by the event
loop, most often implemented within a closure.