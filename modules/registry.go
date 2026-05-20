package modules

import (
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
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Priority < sorted[i].Priority {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	for _, mod := range sorted {
		log.Printf("Loading module: %s", mod.Name)
		mod.Load()
	}
}
