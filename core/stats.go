// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"sync/atomic"
	"time"
)

// RenderStats collects stats for the whole render.
type RenderStats struct {
	Duration                 time.Duration
	RayCount, ShadowRayCount uint64
	start                    time.Time
}

// String returns string representation of stats.
func (s RenderStats) String() string {
	return fmt.Sprintf("%v	%v	%v/%v", s.Duration, float64(s.RayCount)/(1000000.0*s.Duration.Seconds()), s.RayCount, s.ShadowRayCount)
}

// incRayCount atomically increases the ray count.
func (s *RenderStats) incRayCount() {
	atomic.AddUint64(&s.RayCount, 1)
}

// incShadowRayCount atomically increases the shadow ray count.
func (s *RenderStats) incShadowRayCount() {
	atomic.AddUint64(&s.ShadowRayCount, 1)
}

// begin starts stat collection.
func (s *RenderStats) begin() {
	s.RayCount = 0
	s.ShadowRayCount = 0
	s.start = time.Now()
}

// end stops stat collection
func (s *RenderStats) end() {
	s.Duration = time.Since(s.start)
}
