package app

import "errors"

var errCyclicLifecycleDependencies = errors.New("cyclic lifecycle dependencies")

func orderLifecycleNames[N comparable](selected []N, dependencies func(N) []N, reverse bool) ([]N, error) {
	positions := make(map[N]int, len(selected))
	for _, name := range selected {
		if _, exists := positions[name]; exists {
			continue
		}
		positions[name] = len(positions)
	}

	indegree := make(map[N]int, len(selected))
	dependents := make(map[N][]N, len(selected))
	for _, name := range selected {
		for _, dependency := range dependencies(name) {
			if _, exists := positions[dependency]; !exists {
				continue
			}
			indegree[name]++
			dependents[dependency] = append(dependents[dependency], name)
		}
	}

	processed := make(map[N]bool, len(selected))
	ordered := make([]N, 0, len(selected))
	for len(ordered) < len(selected) {
		nextIndex := -1
		var nextName N
		for _, name := range selected {
			if processed[name] || indegree[name] != 0 {
				continue
			}
			index := positions[name]
			if nextIndex == -1 || index < nextIndex {
				nextIndex = index
				nextName = name
			}
		}
		if nextIndex == -1 {
			return nil, errCyclicLifecycleDependencies
		}

		processed[nextName] = true
		ordered = append(ordered, nextName)
		for _, dependent := range dependents[nextName] {
			indegree[dependent]--
		}
	}

	if reverse {
		for left, right := 0, len(ordered)-1; left < right; left, right = left+1, right-1 {
			ordered[left], ordered[right] = ordered[right], ordered[left]
		}
	}

	return ordered, nil
}
