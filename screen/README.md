# Description

Screen implements an event driven rendering model. There is a go routine that
locks the main thread, as is required by
OpenGL. All gl library calls must be issued from the main thread, so the event
loop is the only thread that will call
OGL routines.

The event loop is optimized to achieve event driver rendering. It does not spin
busy waiting to aggressively render
frames, instead it waits to be notified of a change in object or scene, then
re-renders the whole scene at that time. As
a consequence, an app using this system will seem idle when mouse/keyboard or
timer events aren't firing.

## Similarity to a game engine

Because this is event driven, we can sustain many windows sharing the event
loop, doing many things simultaneously.
Objects can be updating and sending their changes continuously, and the event
loop shouldn't be overrun with the
changes.

## Organization

The Screen package controls the setup of the windows and the management of the
objects that will be drawn in them. Under
the screen directory we have the main_gl_thread_object_actions package that does
what it says, code there interacts with
the OGL pipeline to draw objects. This and the screen directory are the only
places we should see opengl library calls.

## Rendering approach, scene setup, world coordinates

Right now we only have a 2D world, but at some point this library will be
extended to support 3D rendering.

The 2D world coordinates within the OpenGL drawing context are all normalized to
a X range and Y range through an
Orthographic projection. This implies that if we enter an X range and Y range
that are not square, say Y range bigger
than X range, the objects rendered into this space will be "squished" into the
screen coordinates, so that an square
box in X and Y world coordinates will appear as an enlongated rectangle:

```
            xmin,ymax   xmax,ymax
            -----------------
            |               |
            |               |
            |               |
            |               |
            -----------------
            xmin,ymin   xmax,ymin
```

when (ymax-ymin) >> (xmax-xmin) looks like:

```
            xmin,ymax                       xmax,ymax
            -------------------------------------
            |                                   |
            |                                   |
            |                                   |
            |                                   |
            -------------------------------------
            xmin,ymin                       xmax,ymin
```

This is the expected behavior. If you want the screen objects to look square
when they are in the "real" world, you
should enter a square range pair when configuring a Screen.

The window pixel width and height don't change the OGL orthographic projection.
If you enter a square world range, the
objects rendered into a screen will be square regardless of the window
dimensions.