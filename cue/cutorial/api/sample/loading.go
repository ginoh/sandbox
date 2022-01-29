package main

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/load"
)

func main() {
	ctx := cuecontext.New()

	//entrypoints := []string{"hello.cue"}
	entrypoints := []string{"format.cue"}

	// Load Cue files into Cue build.Instances slice
	// the second arg is a configuration object, we'll see this later
	bis := load.Instances(entrypoints, nil)
	for _, bi := range bis {
		// check for errors on the instance
		// these are typically parsing errors
		if bi.Err != nil {
			fmt.Println("Error during load:", bi.Err)
			continue
		}

		// Use cue.Context to turn build.Instance to cue.Instance
		value := ctx.BuildInstance(bi)
		if value.Err() != nil {
			fmt.Println("Error during build:", value.Err())
			continue
		}

		// print the error
		fmt.Println("root value:", value)

		// Validate the value
		err := value.Validate()
		if err != nil {
			fmt.Println("Error during validate:", err)
			continue
		}

		// Generate an AST
		//   try out different options
		syn := value.Syntax(
			cue.Final(),         // close structs and lists
			cue.Concrete(false), // allow incomplete values
			cue.Definitions(false),
			cue.Hidden(true),
			cue.Optional(true),
			cue.Attributes(true),
			cue.Docs(true),
		)

		// Pretty print the AST, returns ([]byte, error)
		bs, err := format.Node(
			syn,
			format.TabIndent(false),
			format.UseSpaces(2),
			// format.Simplify(),
		)
		if err != nil {
			fmt.Println("Error during format:", err)
			continue
		}

		fmt.Println(string(bs))
	}
}
