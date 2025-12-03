package intentbuilder

import (
	"github.com/ethereum/go-ethereum/common"
)

func (c *l2Configurator) WithL1MNT(address common.Address) {
	c.builder.intent.Chains[c.chainIndex].L1MNT = address
}
