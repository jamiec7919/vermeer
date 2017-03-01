Quickstart
==========

Vermeer is meant to be run as a command line program.  After installation you should be able to (in a terminal/command prompt) type 'vermeer <file>.vnf' and the render will start. If you haven't specified a maximum number of iterations then Ctrl-C will result in the last frame being completed and 
written out to any output nodes.

(preview is disabled in v0.3.0) Preview - A window will pop up.  After a short while 
the first iteration should appear and then gradually improve. Closing the window should have the same effect as Ctrl-C (you may need to be patient after clicking close, the final iteration will be completed before the application exits). 

Command line parameters
-----------------------

The vermeer command takes a few parameters.

maxiter=-1
  Specify that the render should run for only this many iterations. Defaults to -1 which means 'run until closed'.

Structure of a .vnf
-------------------

Over the next few versions there is planned exporters for popular 3D packages to make exporting and invoking vermeer as painless as possible.  At the present however the only way is by hand.

A vnf file is simply a text file containing a series of nodes.  Nodes describe various aspects of the
render job, the scene, image parameters, materials and sampling rates.  In a vnf nodes are created with::

  <NodeName> {
	Param1 <value>
	Param2 <value>

  }

Some parameters have implicit types, others may need to be specified as the parameter can take multiple
types (e.g. constant colours or texture files).  Acceptable types will be listed along with the parameter.

Comments may be useful in a vnf file and are introduced with the # character.  Anything after the # is 
ignored until the end of the line.

Parameter Types
---------------

A parameter may have one of the following types:

- Bool
- Float
- Int
- String
- Colour
- Matrix4
- Point
- Vec2
- Vec3

And also arrays of a subset of these types (Matrix4, Vec2, Point, Vec3, String, Int).  When specifying an array this is often for motion blur keys and hence need both the count of element types and count of motion keys.  All elements of one key are then listed, followed by the next key and so on.  For matrix, string and int arrays this doesn't apply as the matrix has a fixed number of elements per key and string and int arrays don't change over a frame.

Available Nodes
---------------

- Globals_
- PolyMesh_
- ShaderStd_
- DebugShader_
- Camera_
- DiskLight_
- SphereLight_
- QuadLight_
- TriLight_
- OutputHDR_
- AiryFilter_
- GaussFilter_
- Proc_

Globals
+++++++

Globals contains parameters that affect the whole render process.

::

  Globals {
	XRes 1024
	YRes 1024
  }

XRes
  Width of image in pixels. Int.

YRes
  Height of image in pixels.  Int.

MaxGoRoutines 
  Maximum number of goroutines (essentially screen tile/buckets) to execute simultaneously.  As Go multiplexes
  goroutines into system threads it can be helpful to have slightly more goroutines than threads to avoid wasting time
  waiting on texture locks.

.. _polymesh-def:

PolyMesh
++++++++

The PolyMesh is the default mesh type::

  PolyMesh {
	UV 1 4 vec2 0 0 1 0 1 1 0 1
	UVIdx 4 int 3 2 1 0

    Normals 1 4 vec3 0 0 1 0 0 1 0 0 1 0 0 1
	Verts 2 4 point -1 0.5 0  -1 1 0   0 1 0  0 0.5 0    -1 0.53 0  -1 1.03 0   0 1.03 0  0 0.53 0 
	PolyCount 1 int 4
	FaceIdx 4 int 3 2 1 0
	Transform 1 matrix 1 0 0 0 
	             0 1 0 0
	             0 0 1 0 
	             0 0 0 1
    Shader "mtl2"
    CalcNormals 1
  }

UV
  Primary texture/surface coordinate parameter.  Motion keyed vec2 array.

UVIdx
  Primary texture/surface index array. Operates similar to the FaceIdx array. Int array

Normals
  Vertex normal array.  Motion keyed vec3 array.

Verts
  Vertex position array. Motion keyed point array.

PolyCount
  Each entry in this array represents a polygon in the mesh, the number specifies the number of sides. 
  If this parameter is missing the Polymesh is assumed to be a triangle mesh. Int array.

FaceIdx
  Each entry in this array indexes into the Verts array.  The PolyCount array determines the meaning
  of this array, each polygon will take a certain number of indices as specified in the PolyCount.  Int Array.

Transform
  Transform into worldspace. Transform motion blur is supported by providing multiple matrices which are
  interpolated. Matrix4 array.

Shader
  The shaders to use.  String array.

ShadeIdx
  (optional) Index into shader array for each face.

CalcNormals
  Specify whether to calculate vertex normals.

ShaderStd
+++++++++

The ShaderStd node is the default shader and consists of a multi-layered physical model using an OrenNayar model for diffuse and Microfacet GGX models for the specular and transmission components. It also supports
mirror reflection and perfect transmission with SpecularRoughness set to 0. 

As an example::

  ShaderStd {
	Name "material1"
	DiffuseRoughness float 0.5
	Spec1Roughness float 0.6

	DiffuseStrength float 0
	Spec1Strength float 1

	DiffuseColour rgbtex "maps/cuadricula.jpg"
	Spec1Colour rgb 0.9 0.9 0.9

	IOR float 1.5

	Spec1FresnelModel "Metal"
	Spec1FresnelRefl rgb 0.6 0.6 0.6
	Spec1FresnelEdge rgb 0.95 0.95 0.95
  }


Name
  Every shader material must have a name as this is referred to by other nodes.

DiffuseRoughness 
  Roughness of the diffuse part. float, may be textured.

Spec1Roughness
  Roughness of the specular part. float, may be textured.

DiffuseStrength
  The weight of the diffuse component. float, may be textured.

Spec1Strength
  The weight of the specular part. float, may be textured.

TransStrength
  The weight of the transmissive part (set to 0 for no transmission). float, may be textured.

DiffuseColour
 The colour of the diffuse part.  Colour, may be textured.

Spec1Colour
  The colour of the specular part. Colour, may be textured.

TransColour
  The colour of the transmissive part.  Colour, may be textured.

TransThin
  Boolean value controlling whether the surface should be considered 'thin'.  Thin materials
  don't bend rays according to index of refraction but do still affect with colour and absorbtion.
  This is mostly useful for glass windows modelled as single polygons.

IOR
  Index of refraction.  Float, may be textured.

Spec1FresnelMode
  There are two fresnel modes, "Dielectric" (default) and "Metal".  String.

Spec1FresnelRefl
  For the metal mode this is the usual reflectivity colour.  Colour, may be textured.

Spec1FresnelEdge
  For the metal mode this is the edge tint.  Colour, may be textured.

DebugShader
+++++++++

DebugShader is a simple shader for debugging, a single colour is returned for any surface/lighting combo::

  DebugShader {
  Name "material1"

  Colour rgbtex "maps/cuadricula.jpg"
  }


Name
  Every shader material must have a name as this is referred to by other nodes.

Colour
  The colour to use (may be textured).

Camera
++++++

The camera node creates a camera in the scene.  Cameras support depth of field and frame motion.

::

  Camera {
	Name "camera"
	Type "LookAt"
	Roll 2 1 float 0 0.1
	From 2 1 point 0 0.85 4 0 0.85 4 
	# From 1 1 point 0 0.85 4
	To 1 1 point 0 0.85 -1
	#From 0 0.85 4
	Radius 0.0
	Focal  3.5
	Fov 35
	Up 0 1 0
  }

Name
  The default camera should be called "camera" and if there is no camera called this then rendering will fail.

Type
  Currently only LookAt is supported.

Roll
  For LookAt cameras this specifies the rotation (in radians) around the z axis after the lookat calculation is performed.  Similar effects can be achieved with the Up parameter but Roll is easier to control and animate.  Motion keyed Float array.

From
  For LookAt cameras this specifies the location of the eye. Motion keyed Point array.

To 
  For LookAt cameras this specifies the target location.  Motion keyed Point array.

Radius
  This is the radius of the aperture. 0 for a pinhole camera, make larger to enable DOF.  Float.

Focal
  Length along the z axis to the focal plane (the plane of perfect focus).

Fov
  Field of view in degrees. Float.

Up
  Assist vector for calculating LookAt, should point in a different direction to the line formed between From and To and specify the world 'up' direction for the camera.  Vec3.

DiskLight
+++++++++

The DiskLight node creates a flat circular disk light in the scene::

  DiskLight {
	Name "light01"
	Shader "lightmtl"
	P 0 1.57 0
	LookAt 0 0 0
	Up 0 0 1
	Radius 0.15
  Samples 1
  }

Name
  You should give the node a recognizable name to aid debugging.

Shader
  Specify the material shader to use. String.

P
  Position of the centre of the disk.  Point.

LookAt
  Point in space that the disk will be oriented towards.  The disk will be formed in the plane perpendicular to the line between P and LookAt and located such that P is on the plane.  Point.

Up
  Unit vector assist.  Should point in a direction other than the lookat line.  Will be deprecated as can be calculated.  Vec3.

Radius
  Radius of the disk in world units.

Samples
  Number of samples to take from this light.  This value is squared to give actual number taken. Default is 1.

SphereLight
+++++++++

The SphereLight node creates a sphere light in the scene::

  SphereLight {
  Name "light01"
  Shader "lightmtl"
  P 0 1.57 0
  Radius 0.15
  Samples 2
  }

Name
  You should give the node a recognizable name to aid debugging.

Shader
  Specify the material shader to use. String.

P
  Position of the centre of the sphere.  Point.

Radius
  Radius of the disk in world units.

Samples
  Number of samples to take from this light.  This value is squared to give actual number taken. Default is 1.

QuadLight
+++++++++

The QuadLight node creates a quadrilateral light in the scene.  Quad is formed from the points [P, P+U, P+U+V, P+V]::

  QuadLight {
  Name "light01"
  Shader "lightmtl"
  P 0 1.57 0
  U 0.15 0 1
  V 1 0 0.15
  Samples 2
  }

Name
  You should give the node a recognizable name to aid debugging.

Shader
  Specify the material shader to use. String.

P
  Position of the first point of the quad.  Point.

U
  Vector representing first side of quad.

V
  Vector representing other side of quad.

Samples
  Number of samples to take from this light.  This value is squared to give actual number taken. Default is 1.

TriLight
+++++++++

The TriLight node creates a triangular light in the scene::

  TriLight {
  Name "light01"
  Shader "lightmtl"
  P0 0 1.57 0
  P1 0.15 0 1
  P2 1 0 0.15
  Samples 2
  }

Name
  You should give the node a recognizable name to aid debugging.

Shader
  Specify the material shader to use. String.

P0
  Position of the first point of the triangle.  Point.

P1
  Position of the second point of the triangle.  Point.

P2
  Position of the third point of the triangle.  Point.

Samples
  Number of samples to take from this light.  This value is squared to give actual number taken. Default is 1.

OutputHDR
+++++++++

The OutputHDR node instructs the renderer to output a Radiance HDR file of the given name, it
only takes one parameter::

  OutputHDR {
	Filename "myfile.hdr"
  }

AiryFilter
+++++++++

The AiryFilter node represents a pixel filter based on the Airy disk::

  AiryFilter {
  Name "filter1"
  Res 61
  Width 4
  }

Name
  You should give the filter a name to aid debugging.

Width
  Filter support width in pixels.  4 is a decent starting point.

Res
  Res is the resolution of the pre-computed importance sampling CDF inversion.  A value of 61 is reasonable but for extremely
  high number of iterations it might be worth increasing this.  

GaussFilter
+++++++++

The GaussFilter node represents a pixel filter based on the 2D Gaussian::

  GaussFilter {
  Name "filter1"
  Res 61
  Width 4
  }

Name
  You should give the filter a name to aid debugging.

Width
  Filter support width in pixels.  4 is a decent starting point.

Res
  Res is the resolution of the pre-computed importance sampling CDF inversion.  A value of 61 is reasonable but for extremely
  high number of iterations it might be worth increasing this.  

Proc
++++++

Procedure node.

 Proc {
  Name "proc1"
  Handler "wfobj"
  Data "amodel.obj"
  BMin 1 1 point -100 -100 -100
  BMax 1 1 point 100 100 100
  Transform 1 matrix 1 0 0 0
                     0 1 0 0
                     0 0 1 0
                     0 0 0 1
 }

Name
  Name for the Proce node.

Handler
  Which handler to use (currently 'wfobj' or 'vnf').

Data
  Data string passed into handler init function (usually filename of model to load).

BMin
  Point array for bounding box min.

BMax
  Point array for bounding box max.

Transform
  Matrix array for world space transform.