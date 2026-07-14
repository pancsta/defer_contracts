# defer_contracts

`defer_contracts` is an 80-line Golang library for [programming-by-contract](https://en.wikipedia.org/wiki/Design_by_contract) via pure code and generics, instead of godoc code annotations and pre-processing. It also has "sticky" contracts (invariants), which remain throughout the datum's lifecycle. It has 0 dependencies and relies on language features instead (`defer`, named returns, `context`, and `panic`).

Supported:

- invariants (inherited / sticky)
- pre-condition functions
- post-condition functions (via `defer` and named `return`)
- cancelation on `context.Err()`

## Examples

### Example Function

```go
func someFunc(ctx context.Context, d *Data) (ret string, err error) {
	// check pre-condition
	d.Check(ctx, checkFooNotEmpty)
	// check post-condition
	defer d.Check(ctx, func(ctx context.Context, scope *Data) {
		// check named return
		checkRetDoubleFoo(ctx, scope, ret)
	})

	// function body
	ret = d.Foo + d.Foo

	// named return
	return ret, nil
}
```

### Example Contract

```go
// invariant or pre-condition
func checkFooNotEmpty(_ context.Context, scope *Data) {
	if scope.Foo == "" {
		panic(fmt.Errorf("field Foo is empty"))
	}
}

// post-condition
func checkRetDoubleFoo(_ context.Context, scope *Data, ret string) {
	if len(ret) != 2*len(scope.Foo) {
		panic(fmt.Errorf("return value len isn't a double of field Foo: %s", ret))
	}
}
```

### Example Init

```go
package main

import dc "github.com/pancsta/defer_contracts"

type Data struct {
	// embed Contracts for *Data
	*dc.Contracts[*Data]

	Foo string
}

func main() {
	// global switch
	dc.ContractsEnable(true)
	// init datatype
	d := &Data{
		Foo: "ab",
	}
	// init contracts
	d.Contracts = NewContracts(d, true)
	// add contracts option 1
	d.Add(checkFooEven)
	// add contracts option 2
	d.Contracts.Add(checkFooEven)
}
```

### Runnable Example

See [`./dc_test`](./dc_test.go) for a runnable example and [`./integartion/govy_test.go`](./integartion/govy_test.go)
for another one integrating with the [nobl9/govy](https://github.com/nobl9/govy) validation library.

## How To Use

While contracts can be defined anywhere (prod code and tests) they should only be executed in tests. The support for `context` makes it async-compatible and checking will stop in case of a canceled context. Besides data validation, you can check with a "source of truth", like a database or a [state machine](https://github.com/pancsta/asyncmachine-go). The more contracts attached to an instance, the more unwanted scenarios are covered. The execution is not parallel, and the contracts' slice is not thread-safe. Using [zog](https://github.com/Oudwins/zog) for validation is also recommended.

## Alternatives

- https://github.com/chavacava/dbc4go
- https://github.com/Parquery/gocontracts
