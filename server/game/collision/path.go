package collision

import (
	"container/heap"
	"errors"
	"webscape/server/math"
	"webscape/server/util"
)

// Bound work even for unreachable destinations in very large connected worlds.
const MaxPathExpandedNodes = 16384

var ErrNoPath = errors.New("no path found")
var ErrPathSearchLimit = errors.New("path search expansion limit reached")

func (c Checker) GetPath(from, to math.Vec2) (util.Path, error) {
	return c.GetPathWithinRange(from, to, 0)
}

// GetPathWithinRange finds a path to any walkable tile in the Manhattan range
// around the target. A nonzero range excludes the target's own tile.
func (c Checker) GetPathWithinRange(from, target math.Vec2, distance int) (util.Path, error) {
	if distance < 0 {
		return util.Path{}, ErrNoPath
	}
	if distance == 0 && from != target && c.IsBlocked(target.X, target.Y) {
		return util.Path{}, ErrNoPath
	}
	frontier := &pathFrontier{}
	costs := map[math.Vec2]int{from: 0}
	parents := make(map[math.Vec2]math.Vec2)
	heap.Push(frontier, pathNode{position: from, estimate: pathHeuristic(from, target, distance)})
	expanded := 0
	for frontier.Len() > 0 {
		node := heap.Pop(frontier).(pathNode)
		if costs[node.position] != node.cost {
			continue
		}
		dx, dy := abs(node.position.X-target.X), abs(node.position.Y-target.Y)
		if (distance == 0 && dx+dy == 0) || (distance > 0 && dx+dy > 0 && dx+dy <= distance) {
			path := util.Path{}
			for current := node.position; current != from; current = parents[current] {
				path.Append(current)
			}
			return path.Reversed(), nil
		}
		if expanded >= MaxPathExpandedNodes {
			return util.Path{}, ErrPathSearchLimit
		}
		expanded++
		// Cache each neighboring tile once, including the orthogonal tiles needed
		// to prevent cutting corners on diagonal steps.
		var blocked [3][3]bool
		for x := -1; x <= 1; x++ {
			for y := -1; y <= 1; y++ {
				if x != 0 || y != 0 {
					blocked[x+1][y+1] = c.IsBlocked(node.position.X+x, node.position.Y+y)
				}
			}
		}
		for x := -1; x <= 1; x++ {
			for y := -1; y <= 1; y++ {
				if x == 0 && y == 0 || blocked[x+1][y+1] {
					continue
				}
				diagonal := x != 0 && y != 0
				if diagonal && (blocked[x+1][1] || blocked[1][y+1]) {
					continue
				}
				next := math.Vec2{X: node.position.X + x, Y: node.position.Y + y}
				cost := node.cost + 10000
				if diagonal {
					cost++
				} // Preserve the slight preference for orthogonal steps.
				if previous, ok := costs[next]; ok && previous <= cost {
					continue
				}
				costs[next], parents[next] = cost, node.position
				heap.Push(frontier, pathNode{position: next, cost: cost, estimate: cost + pathHeuristic(next, target, distance)})
			}
		}
	}
	return util.Path{}, ErrNoPath
}

// A lower bound for eight-way movement into a Manhattan goal region.
func pathHeuristic(from, target math.Vec2, distance int) int {
	dx, dy := abs(from.X-target.X), abs(from.Y-target.Y)
	return max(0, max(dx, dy)-distance, (dx+dy-distance+1)/2) * 10000
}

// CanStep also validates the side tiles of a cached diagonal step.
func (c Checker) CanStep(from, to math.Vec2) bool {
	dx, dy := abs(from.X-to.X), abs(from.Y-to.Y)
	if dx > 1 || dy > 1 || dx+dy == 0 || c.IsBlocked(to.X, to.Y) {
		return false
	}
	return dx == 0 || dy == 0 || (!c.IsBlocked(from.X, to.Y) && !c.IsBlocked(to.X, from.Y))
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

type pathNode struct {
	position       math.Vec2
	cost, estimate int
}
type pathFrontier []pathNode

func (q pathFrontier) Len() int { return len(q) }
func (q pathFrontier) Less(i, j int) bool {
	if q[i].estimate != q[j].estimate {
		return q[i].estimate < q[j].estimate
	}
	if q[i].cost != q[j].cost {
		return q[i].cost > q[j].cost
	}
	if q[i].position.X != q[j].position.X {
		return q[i].position.X < q[j].position.X
	}
	return q[i].position.Y < q[j].position.Y
}
func (q pathFrontier) Swap(i, j int)   { q[i], q[j] = q[j], q[i] }
func (q *pathFrontier) Push(value any) { *q = append(*q, value.(pathNode)) }
func (q *pathFrontier) Pop() any {
	last := len(*q) - 1
	value := (*q)[last]
	*q = (*q)[:last]
	return value
}
