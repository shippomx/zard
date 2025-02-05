package dbresolver

import (
	"math/rand"
	"sync/atomic"
)

type Policy interface {
	Resolve([]Instance) Instance
}

func PolicyFrom(policy string) Policy {
	switch policy {
	case "random":
		return &RandomPolicy{}
	case "roundRobin":
		fallthrough
	default:
		return &RoundRobinPolicy{next: new(uint32)}
	}
}

type RandomPolicy struct {
}

func (*RandomPolicy) Resolve(instances []Instance) Instance {
	return instances[rand.Intn(len(instances))]
}

type RoundRobinPolicy struct {
	next *uint32
}

func (r *RoundRobinPolicy) Resolve(instances []Instance) Instance {
	n := atomic.AddUint32(r.next, 1)
	return instances[(int(n)-1)%len(instances)]
}
