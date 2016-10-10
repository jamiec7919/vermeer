package core

import (
	m "github.com/jamiec7919/vermeer/math"
)

// The Geom interface represents all of the basic intersectable shapes.
type Geom interface {
	// Trace returns true if ray hits something. Usually this is the first hit along the ray
	// unless ray.Type &= RayTypeShadow.
	Trace(*Ray, *ShaderContext) bool

	// MotionKeys returns the number of motion keys for this primitive.
	MotionKeys() int

	// Bounds returns the world space bounding volume for the given time.
	Bounds(float32) m.BoundingBox
}
