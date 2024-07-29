// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package pattern

type segmentItr struct {
	path            string
	idx             int
	captureStartIdx int
}

func newSegmentItr(path string) segmentItr {
	// skip leading slash to make sure the first segment is not empty
	startIdx := 0
	if path != "" && path[0] == '/' {
		startIdx = 1
	}

	return segmentItr{
		path: path,
		idx:  startIdx,
	}
}

func (it *segmentItr) next() string {
	if it.idx >= len(it.path) {
		return ""
	}

	start := it.idx
	for ; it.idx < len(it.path); it.idx++ {
		if it.path[it.idx] == '/' {
			break
		}
	}

	segment := it.path[start:it.idx]
	it.idx++
	return segment
}

func (it *segmentItr) hasNext() bool {
	return it.idx < len(it.path)
}

func (it *segmentItr) startCapture() {
	it.captureStartIdx = it.idx
}

func (it *segmentItr) capture() string {
	if it.idx == 0 {
		return ""
	}

	if it.idx >= len(it.path) {
		return it.path[it.captureStartIdx:]
	}

	capturedVal := it.path[it.captureStartIdx : it.idx-1]
	return capturedVal
}

func (it *segmentItr) skipToEnd() {
	it.idx = len(it.path)
}
