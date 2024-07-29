// Copyright (c) 2024 ESET
// See LICENSE file for redistribution.

package pattern

type parserSegmentItr struct {
	path    string
	idx     int
	counter int
}

func newParserSegmentItr(path string) parserSegmentItr {
	return parserSegmentItr{
		path:    path,
		idx:     0,
		counter: 0,
	}
}

func (it *parserSegmentItr) next() string {
	if it.idx >= len(it.path) {
		return ""
	}

	start := it.idx
	for ; it.idx < len(it.path); it.idx++ {
		if it.path[it.idx] == '/' {
			break
		}

		if it.path[it.idx] == '{' {
			for ; it.idx < len(it.path); it.idx++ {
				if it.path[it.idx] == '}' {
					break
				}
			}

			// reached end of string
			if it.idx == len(it.path) {
				break
			}
		}
	}

	segment := it.path[start:it.idx]
	it.idx++
	it.counter++
	return segment
}

func (it *parserSegmentItr) hasNext() bool {
	return it.idx < len(it.path)
}

func (it *parserSegmentItr) isFirst() bool {
	return it.counter == 1
}
