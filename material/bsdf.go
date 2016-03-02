package material

import (
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

type BSDF interface {

	// Compute the PDF for the given omega in (shade) and out vector omega_o
	PDF(shade *SurfacePoint, omega_i, omega_o m.Vec3) float64

	// Sample an outbound direction for the incoming direction in shade.Omega.
	// Return core.ErrNoSample if unable to compute a sample for the given configuration
	Sample(shade *SurfacePoint, omega_i m.Vec3, rnd *rand.Rand, omega_o *m.Vec3, rho *Spectrum, pdf *float64) error

	// Evaluate the BSDF function for the given parameters.  Spectrum is stored in out.
	// Does not divide by the PDF, this should be computed seperately.
	Eval(shade *SurfacePoint, omega_i, omega_o m.Vec3, rho *Spectrum) error

	// Returns true if function includes a dirac delta (e.g. specular reflection)
	IsDelta(shade *SurfacePoint) bool

	// Probability that the path with the shade point as the end should continue.
	ContinuationProb(shade *SurfacePoint) float64
}

type LayeredBSDF struct {
	Layers  []BSDF
	Weights []float64
}

/*
type WeightedBSDF struct {
	BSDF    []BSDF
	Weights []ScalarSampler
}

func (b *WeightedBSDF) PDF(shade *ShadePoint, omega_o m.Vec3) float64 {
	return 0
}

func (b *WeightedBSDF) Sample(shade *ShadePoint, rnd *rand.Rand, out *DirectionalSample) error {
	return nil
}

func (b *WeightedBSDF) Eval(shade *ShadePoint, omega_o m.Vec3, out *colour.Spectrum) error { return nil }
*/
