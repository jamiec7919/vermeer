Introduction
============

  Well, in regular fossilisation, flesh and bone turn to minerals. Realising that, it was a simple matter to reverse the process.
    - Prof. H.J. Farnsworth (ToI ii)

The Vermeer system essentially simulates light - realising that light beams leave emitters (lights) and bounce around floors, walls, coffee cups (and tables) and eventually hits you (or the cat) in the eyes it is a simple matter to reverse the process.  

Instead of tracing light beams from the lights we trace them
from the eye, using some sampling procedure within each image pixel and for depth-of-field.  Where they
hit the scene we evaluate a surface (or, if in some medium a volume) shader which can trace new rays
and sample the light sources in order to determine the colour at that point.

Geometry, cameras and light sources are usually specified by a modelling package (e.g. Blender or Autodesk Maya).
Vermeer doesn't yet contain any exporters (it is only at v0.3.0!) but these are planned as soon as possible.

Animation is a key requisite of production rendering.  In order to smooth out animation for films it is necessary to simulate a small amount of motion blur (this reduces what is known as temporal aliasing) - Vermeer currently supports this for :ref:`polymesh-def` motion and deformation and camera motion.  

How it works
------------

Each frame that is rendered goes through three phases: PreRender, Render and PostRender.

PreRender collects all of the nodes and allows them to perform any setup (loading/caching data files, making links between them etc.).

Render performs the actual rendering.  It generates camera ray samples, traces them through the scene and executes any appropriate shaders to generate colours for the samples.  These samples are then filtered into the output buffer.  The render phase can either be run until you stop it or with a maximum number of iterations.

PostRender performs any post processing (e.g. tone mapping/gamma correction) and outputs any files.

Progressive Rendering
---------------------

Vermeer is a progressive renderer. This means that an image goes through a set of iterations, each one improving the quality.  Due to the spectral sampling
at least 4 iterations are required to reduce colour noise acceptably.  The quasi-monte carlo sampling procedure perfoms careful stratification of samples across iterations which means that surprisingly few are required for (e.g.) soft shadows and glossy reflections.

A render can be left as long as desired, scenes will converge at different rates depending on geometry and lighting configuration and shaders in use.  It is also possible to tweak the number of samples taken
per light source in order to account for difficult lighting situations.