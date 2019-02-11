// Public Domain (-) 2012 The Golly Authors.
// See the Golly UNLICENSE file for details.

// Package hash implements consistent hashing support.
package hash

import (
	"bytes"
	"crypto/sha512"
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
)

const defaultBucketsPerNode = 100

var crcTable = crc32.MakeTable(crc32.Castagnoli)

type bucket struct {
	id   uint32
	node string
}

type buckets []bucket

func (s buckets) Len() int           { return len(s) }
func (s buckets) Less(i, j int) bool { return s[i].id < s[j].id }
func (s buckets) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// Ring provides support for consistent hashing. The zero value for Ring is an
// empty Ring ready to use.
type Ring struct {
	buckets buckets
	mutex   sync.RWMutex
}

func (r *Ring) add(node string, numberOfBuckets int, refresh bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if refresh {
		r.remove(node)
	}
	i, hash := 0, sha512.New()
	for i < numberOfBuckets {
		hash.Write([]byte(node))
		d := hash.Sum(nil)
		m := numberOfBuckets - i
		if m > 64 {
			m = 64
		}
		for j := 0; j < m; j += 4 {
			r.buckets = append(
				r.buckets,
				bucket{
					id:   uint32(d[j+3]) | uint32(d[j+2])<<8 | uint32(d[j+1])<<16 | uint32(d[j])<<24,
					node: node,
				},
			)
			i += 1
		}
	}
	if refresh {
		sort.Sort(r.buckets)
	}
}

func (r *Ring) Add(node string) {
	r.add(node, defaultBucketsPerNode, true)
}

func (r *Ring) AddWithOpts(node string, numberOfBuckets int) {
	r.add(node, numberOfBuckets, true)
}

func (r *Ring) Find(key []byte) (string, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	c, n := crc32.Update(0, crcTable, key), len(r.buckets)
	if n == 0 {
		return "", false
	}
	i, j := 0, n
	for i < j {
		p := i + (j-i)/2
		if r.buckets[p].id < c {
			i = p + 1
		} else {
			j = p
		}
	}
	if i == n {
		i = 0
	}
	return r.buckets[i].node, true
}

func (r *Ring) FindMultiple(key []byte, count int) ([]string, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if count <= 0 {
		return nil, false
	}
	c, n := crc32.Update(0, crcTable, key), len(r.buckets)
	if n == 0 {
		return nil, false
	}
	i, j := 0, n
	for i < j {
		p := i + (j-i)/2
		if r.buckets[p].id < c {
			i = p + 1
		} else {
			j = p
		}
	}
	if i == n {
		i = 0
	}
	first := i
	found := []string{r.buckets[i].node}
outer:
	for count > len(found) {
		i += 1
		if i == n {
			i = 0
		}
		if i == first {
			break
		}
		next := r.buckets[i].node
		for _, node := range found {
			if node == next {
				continue outer
			}
		}
		found = append(found, next)
	}
	return found, true
}

func (r *Ring) remove(node string) {
	i := 0
	for {
		n := len(r.buckets) - 1
		if i > n {
			break
		}
		cur := r.buckets[i]
		if cur.node != node {
			i += 1
			continue
		}
		if i == n {
			r.buckets = r.buckets[:i]
			break
		} else {
			r.buckets = append(r.buckets[:i], r.buckets[i+1:]...)
		}
	}
}

func (r *Ring) Remove(node string) {
	r.mutex.Lock()
	r.remove(node)
	r.mutex.Unlock()
}

func (r *Ring) String() string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "[%d]ring{\n", r.buckets.Len())
	for _, bucket := range r.buckets {
		fmt.Fprintf(buf, "\t%d\t%s\n", bucket.id, bucket.node)
	}
	fmt.Fprint(buf, "}")
	return buf.String()
}

func NewRing(nodes ...string) *Ring {
	r := &Ring{}
	for _, node := range nodes {
		r.add(node, defaultBucketsPerNode, false)
	}
	sort.Sort(r.buckets)
	return r
}
