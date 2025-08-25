package ibmmq

import (
	"github.com/grafana/sobek"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
)

type (
	Ibmmq struct {
		vu      modules.VU
		exports *sobek.Object
	}

	// RootModule is the global module instance that will create Client
	// instances for each VU.
	RootModule struct{}

	// ModuleInstance represents an instance of the JS module.
	ModuleInstance struct {
		*Ibmmq
	}
)

// Ensure the interfaces are implemented correctly
var (
	_ modules.Instance = &ModuleInstance{}
	_ modules.Module   = &RootModule{}
)

// New returns a pointer to a new RootModule instance
func New() *RootModule {
	return &RootModule{}
}

// NewModuleInstance implements the modules.Module interface and returns
// a new instance for each VU.
func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	runtime := vu.Runtime()

	// Create a new Ibmmq module.
	moduleInstance := &ModuleInstance{
		Ibmmq: &Ibmmq{
			vu:      vu,
			exports: runtime.NewObject(),
		},
	}

	mustExport := func(name string, value interface{}) {
		if err := moduleInstance.exports.Set(name, value); err != nil {
			common.Throw(runtime, err)
		}
	}

	mustExport("Producer", moduleInstance.ProducerJs)

	return moduleInstance
}

// Exports implements the modules.Instance interface and returns
// the exports of the JS module.
func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{
		Default: mi.Ibmmq.exports,
	}
}
