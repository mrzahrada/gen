package main

import (
	"github.com/mrzahrada/gen/example/mutations"
	"github.com/mrzahrada/gen/pkg/gen"
)

func main() {

	svc, err := gen.New()
	if err != nil {
		panic(err)
	}

	svc.AddMutation(mutations.Mutation{})

	if err := svc.Build(); err != nil {
		panic(err)
	}

	if err := svc.Publish(); err != nil {
		panic(err)
	}

	// fmt.Println(svc.String())
}
