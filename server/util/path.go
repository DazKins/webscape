package util

import "webscape/server/math"

type Path struct {
	nodes []math.Vec2
}

func (p *Path) Reversed() Path {
	for i := 0; i < len(p.nodes)/2; i++ {
		p.nodes[i], p.nodes[len(p.nodes)-i-1] = p.nodes[len(p.nodes)-i-1], p.nodes[i]
	}
	return *p
}

func (p *Path) Pop() math.Vec2 {
	node := p.nodes[0]
	p.nodes = p.nodes[1:]
	return node
}

func (p *Path) Size() int {
	return len(p.nodes)
}

func (p *Path) Append(node math.Vec2) {
	p.nodes = append(p.nodes, node)
}
