package gen

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"

	"github.com/fatih/color"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

var progressTmpl = ` {{string . "operation" | green}} {{string . "method_name" | green}} {{ bar . "▕" "█" (cycle . "░" "▒" "▓" ) "░" "▏"}} {{percent .}} {{counters .}}`

type Handler func()

type Service struct {
	Commands []*Command
	Mutation *Mutation
	Queries  []*Query

	dir string

	cfg *CDKConfig
}

// New Service
func New() (*Service, error) {
	cfg, dir, err := findConfig()
	if err != nil {
		return nil, err
	}
	fmt.Println(cfg.Context.Bucket)
	dirAbs, err := filepath.Abs(path.Join(dir, "cdk.out"))
	if err != nil {
		return nil, err
	}

	return &Service{
		Commands: []*Command{},
		Queries:  []*Query{},
		dir:      dirAbs,
		cfg:      cfg,
	}, nil
}

// AddCommands -
func (svc *Service) AddCommands(input interface{}) {
	v := reflect.TypeOf(input)
	for i := 0; i < v.NumMethod(); i++ {
		method := &Method{
			Method:      v.Method(i),
			ServiceType: v,
		}

		svc.Commands = append(svc.Commands, &Command{method})
	}
}

// AddQueries -
func (svc *Service) AddQueries(input interface{}) {
	v := reflect.TypeOf(input)
	for i := 0; i < v.NumMethod(); i++ {
		method := &Method{
			Method:      v.Method(i),
			ServiceType: v,
		}
		svc.Queries = append(svc.Queries, &Query{method})
	}
}

func (svc *Service) add(t AssetType, input interface{}) {

	v := reflect.TypeOf(input)
	for i := 0; i < v.NumMethod(); i++ {
		method := &Method{
			Method:      v.Method(i),
			ServiceType: v,
		}
		svc.Queries = append(svc.Queries, &Query{method})
	}
}

// AddMutation -
// TODO: mutation has following structure: OnEvent(context, Event) error
func (svc *Service) AddMutation(input interface{}) error {

	if svc.Mutation != nil {
		return errors.New("mutation already exists")
	}
	v := reflect.TypeOf(input)
	svc.Mutation = &Mutation{
		ServiceType: v,
		Methods:     []*Method{},
	}
	for i := 0; i < v.NumMethod(); i++ {
		if !isMutation(v.Method(i)) {
			continue
		}
		if err := svc.Mutation.Add(v.Method(i)); err != nil {
			return err
		}
	}
	return nil
}

// AddFunction -
// trigger should be plain method which is called in lambda.Start()
func (svc *Service) AddFunction(handler Handler) error {
	return errors.New("not implemented")
}

func (svc *Service) Initialize() error {
	dir, err := filepath.Abs(svc.dir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		fmt.Println(dir, "exists")
		return nil
	}

	if err := os.Mkdir(dir, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (svc *Service) Clean() error {
	return os.RemoveAll(svc.dir)
}

func (svc *Service) Build() error {

	assets := svc.assets()

	p := mpb.New(
		mpb.WithWidth(60),
	)

	bar := p.AddBar(int64(len(assets)), mpb.BarStyle("[=>-|"),
		mpb.PrependDecorators(
			decor.Name("building"),
		),
	)

	for _, asset := range assets {
		tmpl, err := getTemplate(asset.Type())
		if err != nil {
			return err
		}
		fmtMethod, err := generate(tmpl, asset)
		if err != nil {
			return err
		}
		zipPath, err := buildMain(svc.dir, asset.Key(), fmtMethod)
		if err != nil {
			return err
		}
		asset.SetBuildPath(zipPath)
		bar.Increment()
	}
	p.Wait()
	return nil
}

func (svc *Service) Publish() error {
	assets := svc.assets()

	uploader, err := NewUploader(svc.cfg.Context.Bucket, "assets/")
	if err != nil {
		return err
	}
	return uploader.Upload(assets)
}

func (svc *Service) assets() []Asset {
	result := []Asset{}

	for _, cmd := range svc.Commands {
		result = append(result, cmd)
	}

	for _, q := range svc.Queries {
		result = append(result, q)
	}

	if svc.Mutation != nil {
		result = append(result, svc.Mutation)
	}

	return result
}

// TODO: check existance
func (svc *Service) Write(name string) error {
	p := path.Join(svc.dir, name)
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	if _, err = f.WriteString(svc.String()); err != nil {
		return err
	}

	color.Yellow("✅ generate output file: %s", p)
	return nil
}

func (svc *Service) Config() *Config {
	cfg := &Config{
		ServiceName: svc.cfg.Context.Name,
		Bucket:      svc.cfg.Context.Bucket,
		Commands:    []ConfigMethod{},
		Queries:     []ConfigMethod{},
	}

	for _, command := range svc.Commands {
		cfg.Commands = append(cfg.Commands, ConfigMethod{
			Name:    command.Name(),
			S3Key:   command.S3Key(),
			Handler: "main.out",
			Runtime: "GO1.X",
		})
	}

	for _, query := range svc.Queries {
		cfg.Queries = append(cfg.Queries, ConfigMethod{
			Name:    query.Name(),
			S3Key:   query.S3Key(),
			Handler: "main.out",
			Runtime: "GO1.X",
		})
	}

	if svc.Mutation != nil {
		cfg.Mutation = &ConfigMethod{
			Name:    "Mutation",
			S3Key:   svc.Mutation.S3Key(),
			Handler: "main.out",
			Runtime: "GO1.X",
		}
		cfg.Events = svc.Mutation.EventNames()
	}

	return cfg
}

func (svc *Service) String() string {
	cfg := svc.Config()
	data, _ := json.MarshalIndent(cfg, "", " ")
	return string(data)
}

func (svc *Service) Deploy() error {
	return errors.New("not implemented")
}
