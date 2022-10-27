package videostreams

import (
	"fmt"
)

// Operations

type Operation string

const Nop = ""

func Operationf(formatString string, params ...interface{}) Operation {
	return Operation(fmt.Sprintf(formatString, params...))
}

// Operation chaining

type OpChain struct {
	Op     Operation
	source *OpChain
}

func (o *OpChain) With(op Operation) *OpChain {
	if op == Nop {
		return o
	}
	return &OpChain{
		Op:     op,
		source: o,
	}
}

func (o *OpChain) Linearization() []Operation {
	if o.source == nil {
		return []Operation{o.Op}
	}
	return append(o.source.Linearization(), o.Op)
}
