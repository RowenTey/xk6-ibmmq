package ibmmq

import (
	"github.com/RowenTey/xk6-ibmmq/ibmmq"
	"go.k6.io/k6/js/modules"
)

const importPath = "k6/x/ibmmq"

func init() {
	modules.Register(importPath, new(ibmmq.RootModule))
}
