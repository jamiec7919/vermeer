package math

// Vec2 represents a 2D vector.
type Vec2 struct {
	X, Y float32
}

// Vec2Add returns the sum of 2 dimension vectors a and b.
func Vec2Add(a, b Vec2) Vec2 {
	var v Vec2

	v.X = a.X + b.X
	v.Y = a.Y + b.Y

	return v
}

// Vec2Sub returns the difference of 2 dimension vectors a and b.
func Vec2Sub(a, b Vec2) Vec2 {
	var v Vec2
	v.X = a.X - b.X
	v.Y = a.Y - b.Y

	return v
}

// Vec2Scale returns the 2-vector a scaled by s.
func Vec2Scale(s float32, a Vec2) Vec2 {
	var v Vec2

	v.X = a.X * s
	v.Y = a.Y * s

	return v
}

// Vec2Lerp linearly interpolates from a to b based on parameter t in [0,1].
func Vec2Lerp(a, b Vec2, t float32) Vec2 {
	var v Vec2

	v.X = (1.0-t)*a.X + t*b.X
	v.Y = (1.0-t)*a.Y + t*b.Y

	return v
}

// Vec2Dot computes the 2D dot product of vectors a and b.
func Vec2Dot(a, b Vec2) float32 {
	return a.X*b.X + a.Y*b.Y
}

// Vec2Length2 returns the squared length of vector a.
func Vec2Length2(a Vec2) float32 {
	return a.X*a.X + a.Y*a.Y
}

// Vec2Length returns the length of vector a.
func Vec2Length(a Vec2) float32 {
	return Sqrt(a.X*a.X + a.Y*a.Y)
}

// Vec2Mad returns the multiply-add:  a + s*b
func Vec2Mad(a, b Vec2, s float32) Vec2 {
	var v Vec2

	v.X = a.X + (b.X * s)
	v.Y = a.Y + (b.Y * s)

	return v
}

// Vec3Add3 adds 3 3-vectors a,b and c.
func Vec2Add3(a, b, c Vec2) Vec2 {
	var v Vec2

	v.X = a.X + b.X + c.X
	v.Y = a.Y + b.Y + c.Y

	return v
}
