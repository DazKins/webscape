package collision

import (
	"errors"
	"math/rand"
	"testing"
	"webscape/server/math"
)

type gridWorld struct {
	size    int
	blocked map[math.Vec2]bool
	checks  int
}

func (w *gridWorld) GetStaticWall(x, y int) bool {
	w.checks++
	return x < 0 || y < 0 || x >= w.size || y >= w.size || w.blocked[math.Vec2{X: x, Y: y}]
}

func TestPathWithinRangeMatchesShortestReachableGoal(t *testing.T) {
	// Compare against a simple exhaustive Dijkstra reference on deterministic
	// obstacle layouts, including unreachable goals and diagonal corner rules.
	random := rand.New(rand.NewSource(7))
	for range 80 {
		w := &gridWorld{size: 8, blocked: make(map[math.Vec2]bool)}
		for x := 0; x < 8; x++ {
			for y := 0; y < 8; y++ {
				w.blocked[math.Vec2{X: x, Y: y}] = random.Intn(4) == 0
			}
		}
		start, target := math.Vec2{}, math.Vec2{X: 7, Y: 7}
		delete(w.blocked, start)
		checker := Checker{World: w}
		for radius := 0; radius <= 4; radius++ {
			want := referencePathCost(checker, start, target, radius, 8)
			path, err := checker.GetPathWithinRange(start, target, radius)
			if want < 0 {
				if !errors.Is(err, ErrNoPath) {
					t.Fatalf("unreachable: %v", err)
				}
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			current, cost := start, 0
			for next := path.Pop(); next != nil; next = path.Pop() {
				if !checker.CanStep(current, *next) {
					t.Fatalf("invalid step %v -> %v", current, *next)
				}
				cost += 10000
				if current.X != next.X && current.Y != next.Y {
					cost++
				}
				current = *next
			}
			if cost != want {
				t.Fatalf("range %d: cost=%d want=%d", radius, cost, want)
			}
		}
	}
}

func referencePathCost(c Checker, from, target math.Vec2, radius, size int) int {
	costs := map[math.Vec2]int{from: 0}
	visited := map[math.Vec2]bool{}
	for {
		best := math.Vec2{}
		bestCost := int(^uint(0) >> 1)
		for p, cost := range costs {
			if !visited[p] && cost < bestCost {
				best, bestCost = p, cost
			}
		}
		if bestCost == int(^uint(0)>>1) {
			return -1
		}
		distance := abs(best.X-target.X) + abs(best.Y-target.Y)
		if radius == 0 && distance == 0 || radius > 0 && distance > 0 && distance <= radius {
			return bestCost
		}
		visited[best] = true
		for x := max(0, best.X-1); x <= min(size-1, best.X+1); x++ {
			for y := max(0, best.Y-1); y <= min(size-1, best.Y+1); y++ {
				next := math.Vec2{X: x, Y: y}
				if !c.CanStep(best, next) {
					continue
				}
				cost := bestCost + 10000
				if best.X != x && best.Y != y {
					cost++
				}
				if previous, ok := costs[next]; !ok || cost < previous {
					costs[next] = cost
				}
			}
		}
	}
}

func TestPathSearchIsDirectedAndBounded(t *testing.T) {
	w := &gridWorld{size: 256}
	checker := Checker{World: w}
	path, err := checker.GetPathWithinRange(math.Vec2{}, math.Vec2{X: 200, Y: 200}, 4)
	if err != nil || path.Size() != 198 {
		t.Fatalf("size=%d error=%v", path.Size(), err)
	}
	if w.checks > 2000 {
		t.Fatalf("search explored too broadly: %d checks", w.checks)
	}
	// An unblocked target enclosed by walls forces an exhaustive search, but
	// should terminate at the expansion cap instead of traversing the map.
	w.blocked = map[math.Vec2]bool{}
	for x := 199; x <= 201; x++ {
		for y := 199; y <= 201; y++ {
			if x != 200 || y != 200 {
				w.blocked[math.Vec2{X: x, Y: y}] = true
			}
		}
	}
	w.checks = 0
	_, err = checker.GetPath(math.Vec2{}, math.Vec2{X: 200, Y: 200})
	if !errors.Is(err, ErrPathSearchLimit) {
		t.Fatalf("error=%v", err)
	}
	if w.checks > 8*MaxPathExpandedNodes+1 {
		t.Fatalf("unbounded search: %d checks", w.checks)
	}
}

func BenchmarkPathWithinAttackRange(b *testing.B) {
	checker := Checker{World: &gridWorld{size: 256}}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := checker.GetPathWithinRange(math.Vec2{}, math.Vec2{X: 100, Y: 100}, 4); err != nil {
			b.Fatal(err)
		}
	}
}

func TestCachedDiagonalStepRequiresClearSideTiles(t *testing.T) {
	w := &gridWorld{size: 3, blocked: map[math.Vec2]bool{}}
	c := Checker{World: w}
	from, to := math.Vec2{}, math.Vec2{X: 1, Y: 1}
	if !c.CanStep(from, to) {
		t.Fatal("clear diagonal was blocked")
	}
	w.blocked[math.Vec2{X: 1, Y: 0}] = true
	if c.CanStep(from, to) {
		t.Fatal("cached diagonal cut a blocked corner")
	}
}
