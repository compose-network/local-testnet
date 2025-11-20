package blockscout

import "context"

type Blockscout struct {
}

func New() *Blockscout {
	return &Blockscout{}
}

func (b *Blockscout) Run(ctx context.Context) error {
	return nil
}
