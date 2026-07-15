package defer_contracts

import "context"

type FuncPost[T any, R any] func(ctx context.Context, scope T, ret R)
type Func[T any] func(ctx context.Context, scope T)

var enabledGlobal = false

func ContractsEnable(enabled bool) {
	enabledGlobal = enabled
}

// Contracts is a collection of contracts.
type Contracts[T any] struct {
	// contracts is the list of contracts.
	contracts []func(ctx context.Context, scope T)
	// contractsScope is the scope of the contracts.
	contractsScope T

	contractsEnabled bool
}

// NewContracts creates a new instance of a Contracts collection.
func NewContracts[T any](scope T, enabled bool) *Contracts[T] {
	return &Contracts[T]{
		contractsScope:   scope,
		contractsEnabled: enabled,
	}
}

// Add adds an invariant to the collection.
func (c *Contracts[T]) Add(fns ...func(context.Context, T)) {
	if !c.contractsEnabled || !enabledGlobal {
		return
	}
	c.contracts = append(c.contracts, fns...)
}

// Check runs the collection of contracts, executing functions in order and stopping on expired context.
func (c *Contracts[T]) Check(ctx context.Context, fns ...func(context.Context, T)) {
	if !c.contractsEnabled || !enabledGlobal {
		return
	}
	for _, fn := range fns {
		if ctx.Err() != nil {
			return
		}
		fn(ctx, c.contractsScope)
	}
	for _, fn := range c.contracts {
		if ctx.Err() != nil {
			return
		}
		fn(ctx, c.contractsScope)
	}
}

// Compose composes a list of invariant or pre-condition functions into a single function.
func Compose[T any](fns ...Func[T]) Func[T] {
	return func(ctx context.Context, scope T) {
		for _, fn := range fns {
			if ctx.Err() != nil {
				return
			}
			fn(ctx, scope)
		}
	}
}

// ComposePost composes a list of post-condition functions into a single function.
func ComposePost[T any, R any](fns ...FuncPost[T, R]) FuncPost[T, R] {
	return func(ctx context.Context, scope T, ret R) {
		for _, fn := range fns {
			if ctx.Err() != nil {
				return
			}
			fn(ctx, scope, ret)
		}
	}
}
