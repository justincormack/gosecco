package compiler

import "golang.org/x/sys/unix"

type shift struct {
	position int
}

func (c *compilerContext) isLongJump(jumpSize int) bool {
	return jumpSize > c.maxJumpSize
}

func hasLongJump(index int, jts, jfs map[int]int) bool {
	// Using the unshifted index to look up positions in jts and jfs is
	// only safe if we're iterating backwards. Otherwise we would have to
	// fix up the positions in the maps as well and that would be fugly.

	if _, ok := jts[index]; ok {
		return true
	}
	if _, ok := jfs[index]; ok {
		return true
	}
	return false
}

func fixupWithShifts(pos, add int, shifts []shift) int {
	to := pos + add + 1
	currentAdd := add
	for _, s := range shifts {
		if s.position > pos && s.position <= to {
			currentAdd++
			to++
		}
	}
	return currentAdd
}

func (c *compilerContext) fixupJumps() {
	maxIndexWithLongJump := -1
	jtLongJumps := make(map[int]int)
	jfLongJumps := make(map[int]int)

	for l, at := range c.labels.allLabels() {
		for _, pos := range c.jts.allJumpsTo(l) {
			jumpSize := (at - pos) - 1
			if c.isLongJump(jumpSize) {
				if maxIndexWithLongJump < pos {
					maxIndexWithLongJump = pos
				}
				jtLongJumps[pos] = jumpSize
			} else {
				c.result[pos].Jt = uint8(jumpSize)
			}
		}

		for _, pos := range c.jfs.allJumpsTo(l) {
			jumpSize := (at - pos) - 1
			if c.isLongJump(jumpSize) {
				if maxIndexWithLongJump < pos {
					maxIndexWithLongJump = pos
				}
				jfLongJumps[pos] = jumpSize
			} else {
				c.result[pos].Jf = uint8(jumpSize)
			}
		}

		for _, pos := range c.uconds.allJumpsTo(l) {
			c.result[pos].K = uint32((at - pos) - 1)
		}
	}

	// This is an optimization. Please don't comment away.
	if maxIndexWithLongJump == -1 {
		return
	}

	shifts := []shift{}

	currentIndex := maxIndexWithLongJump
	for currentIndex > -1 {
		current := c.result[currentIndex]

		if isConditionalJump(current) && hasLongJump(currentIndex, jtLongJumps, jfLongJumps) {
			hadJt := c.handleJTLongJumpFor(currentIndex, jtLongJumps, jfLongJumps, &shifts)
			c.handleJFLongJumpFor(currentIndex, jfLongJumps, hadJt, &shifts)
		} else {
			if isUnconditionalJump(current) {
				c.result[currentIndex].K = uint32(fixupWithShifts(currentIndex, int(c.result[currentIndex].K), shifts))
			} else {
				hadJt := c.shiftJt(currentIndex, &shifts)
				c.shiftJf(hadJt, currentIndex, &shifts)
			}
		}
		currentIndex--
	}
}

func (c *compilerContext) handleJTLongJumpFor(currentIndex int, jtLongJumps map[int]int, jfLongJumps map[int]int, shifts *[]shift) bool {
	hadJt := false
	if jmpLen, ok := jtLongJumps[currentIndex]; ok {
		jmpLen = fixupWithShifts(currentIndex, jmpLen, *shifts)
		hadJt = true

		newJf := int(c.result[currentIndex].Jf) + 1
		if c.isLongJump(newJf) {
			// Simple case, we can just add it to the long jumps for JF:
			jfLongJumps[currentIndex] = newJf
		} else {
			c.result[currentIndex].Jf = uint8(newJf)
		}

		shifts = c.insertJumps(currentIndex, jmpLen, 0, shifts)
	}
	return hadJt
}

func (c *compilerContext) handleJFLongJumpFor(currentIndex int, jfLongJumps map[int]int, hadJt bool, shifts *[]shift) {
	if jmpLen, ok := jfLongJumps[currentIndex]; ok {
		jmpLen = fixupWithShifts(currentIndex, jmpLen, *shifts)
		var incr int
		shifts, incr, jmpLen = c.increment(hadJt, jmpLen, currentIndex, shifts)
		shifts = c.insertJumps(currentIndex, jmpLen, incr, shifts)
	}
}

func (c *compilerContext) increment(hadJt bool, jmpLen, currentIndex int, shifts *[]shift) (*[]shift, int, int) {
	incr := 0
	if hadJt {
		c.result[currentIndex+1].K++
		incr++
		jmpLen--
	} else {
		newJt := int(c.result[currentIndex].Jt) + 1
		if c.isLongJump(newJt) {
			// incr in this case doesn't seem to do much, all tests pass when it is changed to 0
			shifts = c.insertJumps(currentIndex, newJt, incr, shifts)
			incr++
		} else {
			c.result[currentIndex].Jt = uint8(newJt)
		}
	}
	return shifts, incr, jmpLen
}

func (c *compilerContext) shiftJf(hadJt bool, currentIndex int, shifts *[]shift) {
	newJf := fixupWithShifts(currentIndex, int(c.result[currentIndex].Jf), *shifts)
	if c.isLongJump(newJf) {
		var incr int
		shifts, incr, _ = c.increment(hadJt, 0, currentIndex, shifts)
		shifts = c.insertJumps(currentIndex, newJf, incr, shifts)
	} else {
		c.result[currentIndex].Jf = uint8(newJf)
	}
}

func (c *compilerContext) shiftJt(currentIndex int, shifts *[]shift) bool {
	hadJt := false
	newJt := fixupWithShifts(currentIndex, int(c.result[currentIndex].Jt), *shifts)
	if c.isLongJump(newJt) {
		hadJt = true

		// Jf doesn't need to be modified here, because it will be fixed up with the shifts. Hopefully correctly...
		shifts = c.insertJumps(currentIndex, newJt, 0, shifts)
	} else {
		c.result[currentIndex].Jt = uint8(newJt)
	}
	return hadJt
}

func (c *compilerContext) insertJumps(currentIndex, pos, incr int, shifts *[]shift) *[]shift {
	c.insertUnconditionalJump(currentIndex+1+incr, pos)
	c.result[currentIndex].Jf = uint8(incr)
	*shifts = append(*shifts, shift{currentIndex + 1 + incr})
	return shifts
}

func (c *compilerContext) hasPreviousUnconditionalJump(from int) bool {
	return c.uconds.hasJumpFrom(from)
}

func insertSockFilter(sfs []unix.SockFilter, ix int, x unix.SockFilter) []unix.SockFilter {
	return append(
		append(
			append([]unix.SockFilter{}, sfs[:ix]...), x), sfs[ix:]...)
}

func (c *compilerContext) insertUnconditionalJump(from, k int) {
	x := unix.SockFilter{Code: OP_JMP_K, K: uint32(k)}
	c.result = insertSockFilter(c.result, from, x)
}

func (c *compilerContext) shiftJumps(from int, hasPrev bool) {
	incr := 1
	if hasPrev {
		incr = 2
	}
	c.shiftJumpsBy(from, incr)
}

func (c *compilerContext) shiftJumpsBy(from, incr int) {
	c.jts.shift(from, incr)
	c.jfs.shift(from, incr)
	c.uconds.shift(from, incr)
	c.labels.shiftLabels(from, incr)
}

func (c *compilerContext) fixUpPreviousRule(from int, positiveJump bool) {
	if positiveJump {
		c.result[from].Jt = 0
		c.result[from].Jf = 1
	} else {
		c.result[from].Jt = 1
		c.result[from].Jf = 0
	}
}
