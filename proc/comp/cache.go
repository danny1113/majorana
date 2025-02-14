package comp

import (
	"fmt"
	"strings"
)

type LRUCache struct {
	numberOfLines int
	lineLength    int
	cacheLength   int
	lines         []Line
}

type Line struct {
	Boundary [2]int32
	Data     []int8
}

func (l Line) String() string {
	return fmt.Sprintf("(%d-%d): %v", l.Boundary[0], l.Boundary[1], l.Data)
}

func (l Line) get(addr int32) (int8, bool) {
	if addr >= l.Boundary[0] && addr < l.Boundary[1] {
		return l.Data[addr-l.Boundary[0]], true
	}
	return 0, false
}

func (l Line) set(addr int32, value int8) {
	l.Data[addr-l.Boundary[0]] = value
}

func NewLRUCache(lineLength int, cacheLength int) *LRUCache {
	if cacheLength%lineLength != 0 {
		panic("cache length should be a multiple of the line length")
	}
	return &LRUCache{
		numberOfLines: cacheLength / lineLength,
		lineLength:    lineLength,
		cacheLength:   cacheLength,
	}
}

func (c *LRUCache) Get(addr int32) (int8, bool) {
	for i, l := range c.lines {
		if v, exists := l.get(addr); exists {
			c.lines = append(append([]Line{l}, c.lines[:i]...), c.lines[i+1:]...)
			return v, exists
		}
	}
	return 0, false
}

func (c *LRUCache) Write(addr int32, data []int8) {
	for _, l := range c.lines {
		if _, exists := l.get(addr); exists {
			for i, v := range data {
				l.set(addr+int32(i), v)
			}
			return
		}
	}
	panic("cache line doesn't exist")
}

func (c *LRUCache) PushLine(addr int32, data []int8) []int8 {
	newLine := Line{
		Boundary: [2]int32{addr, addr + int32(c.lineLength)},
		Data:     data,
	}

	c.lines = append([]Line{newLine}, c.lines...)
	if len(c.lines) > c.numberOfLines {
		c.lines = c.lines[:c.numberOfLines]
		// Return the evicted line
		return c.lines[len(c.lines)-1].Data
	}
	return nil
}

func (c *LRUCache) Lines() []Line {
	return c.lines
}

func (c *LRUCache) String() string {
	res := make([]string, 0, len(c.lines))
	for _, line := range c.lines {
		res = append(res, line.String())
	}
	return strings.Join(res, "\n")
}
