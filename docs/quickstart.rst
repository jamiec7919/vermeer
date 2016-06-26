Quickstart
==========

Vermeer is meant to be run as a command line program.  After installation you should be able to (in a terminal/command prompt) type 'vermeer <file>.vnf' and a window will pop up.  After a short while 
the first iteration should appear and then gradually improve.  If you haven't specified a maximum
number of iterations then closing the window should result in the last frame being completed and 
written out to any output nodes (you may need to be patient after clicking close, the final iteration will be completed before the application exits). 

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

And also arrays of a subset of these types (Matrix4, Vec2, Point, Vec3).  When specifying an array this is often for motion blur keys and hence need both the count of element types and count of motion keys.  All elements of one key are then listed, followed by the next key and so on.  For matrix arrays this doesn't apply as the matrix has a fixed number of elements per key.

Available Nodes
---------------

- Globals_
- Meshfile_
- Polymesh_
- Material_
- Camera_
- DiskLight_
- OutputHDR_

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

Meshfile
++++++++

Meshfile nodes represent a triangular mesh loaded from a file.  This node is temporary as conversion
to Polymesh will eventually be preferred.  The system will load and .mtl files and will
translate materials as best as possible into the default shader.  Materials defined in vnf files
will override those with the same name in the .mtl libraries::

  Meshfile {
	Name "mesh03"
	Filename "gopher.obj"
	Transform matrix 0.1 0 0 0.5
	          0 0.1 0 0.5
	          0 0 0.1 2
	          0 0 0 1

	CalcNormals 0
  }

Name
  You should give each node a recognizable name to aid debugging.

Filename
  Filename of the mesh to load.  Only Alias Wavefront .obj files are supported currently.

Transform
  Specifies a transformation matrix to apply after loading.  Matrix4.

CalcNormals
  Boolean value controlling whether vertex normals should be calculated (e.g. if not supplied in the
  model file).

.. _polymesh-def:

Polymesh
++++++++

::

  PolyMesh {
	UV 1 4 vec2 0 0 1 0 1 1 0 1
	UVIdx 4 int 3 2 1 0

    Normals 1 4 vec3 0 0 1 0 0 1 0 0 1 0 0 1
	Verts 2 4 point -1 0.5 0  -1 1 0   0 1 0  0 0.5 0    -1 0.53 0  -1 1.03 0   0 1.03 0  0 0.53 0 
	PolyCount 1 int 4
	FaceIdx 4 int 3 2 1 0
	ModelToWorld 1 matrix 1 0 0 0 
	             0 1 0 0
	             0 0 1 0 
	             0 0 0 1
    Material "mtl2"
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

ModelToWorld
  Transform into worldspace.  Matrix4 array.

Material
  The material to use.  String.

CalcNormals
  Specify whether to calculate vertex normals.

Material
++++++++

The Material node is the default shader and consists of a multi-layered physical model using an OrenNayar model for diffuse and Microfacet GGX models for the specular and transmission components. It also supports
mirror reflection and perfect transmission with SpecularRoughness set to 0. 

As an example::

  Material {
	Name "material1"
	Roughness float 0.5
	SpecularRoughness float 0.6

	DiffuseStrength float 0
	SpecularStrength float 1

	Kd rgbtex "maps/cuadricula.jpg"
	Ks rgb 0.9 0.9 0.9

	IOR float 1.5
	TransStrength float 0
	Kt rgb 0.9 0.9 0.9
	TransThin 0

	Spec1FresnelModel "Metal"
	Spec1FresnelRefl rgb 0.6 0.6 0.6
	Spec1FresnelEdge rgb 0.95 0.95 0.95
  }


Name
  Every shader material must have a name as this is referred to by other nodes.

Roughness 
  Roughness of the diffuse part. float, may be textured.

SpecularRoughness
  Roughness of the specular part. float, may be textured.

DiffuseStrength
  The weight of the diffuse component. float, may be textured.

SpecularStrength
  The weight of the specular part. float, may be textured.

TransStrength
  The weight of the transmissive part (set to 0 for no transmission). float, may be textured.

Kd
 The colour of the diffuse part.  Colour, may be textured.

Ks
  The colour of the specular part. Colour, may be textured.

Kt
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
	Material "lightmtl"
	P 0 1.57 0
	LookAt 0 0 0
	Up 0 0 1
	Radius 0.15
  }

Name
  You should give the node a recognizable name to aid debugging.

Material
  Specify the material shader to use. String.

P
  Position of the centre of the disk.  Point.

LookAt
  Point in space that the disk will be oriented towards.  The disk will be formed in the plane perpendicular to the line between P and LookAt and located such that P is on the plane.  Point.

Up
  Unit vector assist.  Should point in a direction other than the lookat line.  Will be deprecated as can be calculated.  Vec3.

Radius
  Radius of the disk in world units.

OutputHDR
+++++++++

The OutputHDR node instructs the renderer to output a Radiance HDR file of the given name, it
only takes one parameter::

  OutputHDR {
	Filename "myfile.hdr"
  }
