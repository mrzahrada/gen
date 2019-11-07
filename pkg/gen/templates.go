package gen

var resolverTmpl = `// DO NOT EDIT! Generated code
package main

import (
	"github.com/aws/aws-lambda-go/lambda"{{ range .Imports}}
	"{{- .}}"{{end}}
)

func main() {
	lambda.Start(func() {{ .String }} {
		svc, err := {{ .PackageName }}.New()
		if err != nil {
			panic(err)
		}
		return svc.{{ .Name }}
	}())
}
`

var handlerTmpl = `// DO NOT EDIT! Generated code
package main

import(
	"github.com/aws/aws-lambda-go/lambda"{{ range .Imports}}
	"{{- .}}"{{end}}
)
func main() {
	lambda.Start({{.Name}}())
}
`

var mutationTmpl = `
// DO NOT EDIT! Generated code.
package main

import (
	"log"
	"github.com/mrzahrada/es"
    lambdaevents "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"{{ range .Imports}}
	"{{- .}}"{{end}}
)

func main() {
	svc, err := {{ .PackageName }}.New()
	if err != nil {
		panic(err)
	}
	lambda.Start(func() func(ctx context.Context, event lambdaevents.KinesisEvent) error {
		h := handler{
			svc: svc,
			es: es.NewUnmarshaler({{range .Events}} 
				{{.}}{},{{end}}
			),
		}
		return h.on
	}())
}
{{ $events := .Events }}	
type service interface{
	Push(context.Context) error{{range $index, $element := .EventNames}} 
	On{{.}}(context.Context, *{{ index $events $index }}) error {{end}}
}

type handler struct {
	svc service
	es  *es.Unmarshaler
}

func (h handler) on(ctx context.Context, input lambdaevents.KinesisEvent) error {

	for _, record := range input.Records {
		data := record.Kinesis.Data
		event, err := h.es.Unmarshal(data)

		if err != nil {
			if err == es.ErrUnknownEventType {
				log.Printf("uknown event: %s", string(data))
				continue
			}
			return err
		}
		if err := h.call(ctx, event); err != nil {
			return err
		}
	}
	return h.svc.Push(ctx)
}

func (h handler) call(ctx context.Context, input interface{}) error {
	var err error
	switch v := input.(type) { {{range $index, $element := .EventNames}} 
    case *{{ index $events $index }}:
        err = h.svc.On{{$element}}(ctx, v){{end}}
	default:
		// log.Printf("[ERROR] missing handler for %+v. event: %+v", v, input)
	}

	return err
	// {{ .EventNames}}
}
`

var mutationTmplOld = `
// DO NOT EDIT! Generated code.
package main

import (
	"log"
	"bitbucket.org/wereldo/shipments/pkg/es"
    "github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"{{ range .Imports}}
	"{{- .}}"{{end}}
)

func main() {
	svc, err := {{ .PackageName }}.New()
	if err != nil {
		panic(err)
	}
	lambda.Start(func() func(ctx context.Context, event events.KinesisEvent) error {
		h := handler{
			svc: svc,
			es: es.NewUnmarshaler({{range .EventNames}} 
				shipments.{{.}}{},{{end}}
			),
		}
		return h.on
	}())
}

type service interface{
	Push(context.Context) error{{range .EventNames}} 
	On{{.}}(context.Context, *shipments.{{.}}) error {{end}}
}

type handler struct {
	svc service
	es  *es.Unmarshaler
}

func (h handler) on(ctx context.Context, input events.KinesisEvent) error {

	for _, record := range input.Records {
		data := record.Kinesis.Data
		event, err := h.es.Unmarshal(data)

		if err != nil {
			if err == es.ErrUnknownEventType {
				log.Printf("uknown event: %s", string(data))
				continue
			}
			return err
		}
		if err := h.call(ctx, event); err != nil {
			return err
		}
	}
	return h.svc.Push(ctx)
}

// generate
func (h handler) call(ctx context.Context, input interface{}) error {

	var err error
	switch v := input.(type) { {{range .EventNames}} 
    case *shipments.{{.}}:
        err = h.svc.On{{.}}(ctx, v){{end}}
	default:
		// log.Printf("[ERROR] missing handler for %+v. event: %+v", v, input)
	}

	return err
}
`
