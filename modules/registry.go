package modules

import (
	"slices"

	log "github.com/sirupsen/logrus"
)

type ModuleEntry struct {
	Name     string
	Load     func()
	Priority int
}

var moduleRegistry []ModuleEntry

func RegisterModule(name string, load func(), priority ...int) {
	p := 0
	if len(priority) > 0 {
		p = priority[0]
	}
	moduleRegistry = append(moduleRegistry, ModuleEntry{Name: name, Load: load, Priority: p})
}

func loadAllModules() {
	sorted := make([]ModuleEntry, len(moduleRegistry))
	copy(sorted, moduleRegistry)
	slices.SortFunc(sorted, func(a, b ModuleEntry) int {
		return a.Priority - b.Priority
	})
	for _, mod := range sorted {
		log.Printf("Loading module: %s", mod.Name)
		mod.Load()
	}
}
