package defer_contracts

import (
	"context"
	"fmt"
	"testing"
)

func init() {
	// enable the global switch
	ContractsEnable(true)
}

func ExampleInit() {
	// enable the global switch
	ContractsEnable(true)
	// define a datatype
	type Example struct {
		// embed Contracts for *Data
		*Contracts[*Example]

		Foo string
	}
	// init datatype
	d := &Example{
		Foo: "ab",
	}
	// init contracts
	d.Contracts = NewContracts(d, true)
	// define a contract
	contract := func(_ context.Context, scope *Example) {
		if len(scope.Foo)%2 != 0 {
			// fail a contract
			panic("field Foo len not even")
		}
	}
	// register a contract
	d.Add(contract)
}

type Data struct {
	// embed Contracts for *Data
	*Contracts[*Data]

	Foo string
}

// checkFooEven is a contract for [Data.Foo] that will panic if the length of Foo is not even.
func checkFooEven(_ context.Context, scope *Data) {
	if len(scope.Foo)%2 != 0 {
		panic("field Foo len not even")
	}
}

// checkFooNotEmpty checks if the Foo field is not empty.
func checkFooNotEmpty(_ context.Context, scope *Data) {
	if scope.Foo == "" {
		panic(fmt.Errorf("field Foo is empty"))
	}
}

// checkRetDoubleFoo checks if the return value is a double of the Foo field.
func checkRetDoubleFoo(_ context.Context, scope *Data, ret string) {
	if len(ret) != 2*len(scope.Foo) {
		panic(fmt.Errorf("return value len isn't a double of field Foo: %s", ret))
	}
}

func TestBasic(t *testing.T) {
	ctx := context.TODO()

	// init datatype
	d := &Data{
		Foo: "ab",
	}
	// init contracts
	d.Contracts = NewContracts(d, true)
	// register a contract
	d.Add(checkFooEven)

	// pass all 3 conditions
	_, _ = doubleFooOk(ctx, d)

	recovered1 := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				recovered1 = true
			}
		}()
		// fail on post-condition
		_, _ = doubleFooBad(ctx, d)
	}()
	if !recovered1 {
		t.Fatal("expected panic doubleFooBad()")
	}

	// corrupt Foo here
	d.Foo = "a"
	recovered2 := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				recovered2 = true
			}
		}()
		// fail on contract (inherited)
		_, _ = doubleFooOk(ctx, d)
	}()
	if !recovered2 {
		t.Fatal("expected panic from doubleFooOk()")
	}

}

// doubleFooOk doubles the Foo field correctly.
func doubleFooOk(ctx context.Context, d *Data) (ret string, err error) {
	d.Check(ctx, checkFooNotEmpty)
	defer d.Check(ctx, func(ctx context.Context, scope *Data) {
		checkRetDoubleFoo(ctx, scope, ret)
	})

	// op
	ret = d.Foo + d.Foo

	return ret, nil
}

// doubleFooBad fails to double the Foo field correctly.
func doubleFooBad(ctx context.Context, d *Data) (ret string, err error) {
	d.Check(ctx, checkFooNotEmpty)
	defer d.Check(ctx, func(ctx context.Context, scope *Data) {
		checkRetDoubleFoo(ctx, scope, ret)
	})

	// op
	ret = d.Foo + d.Foo + "_too_long"

	return ret, nil
}

// TODO TestContext
// TODO TestComposition
