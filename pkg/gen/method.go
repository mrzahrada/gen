package gen

import (
	"crypto/md5"
	"fmt"
	"log"
	"path"
	"reflect"
	"sort"
	"strings"
)

type AssetType string

const (
	CommandType  = AssetType("Command")
	QueryType    = AssetType("Query")
	MutationType = AssetType("Mutation")
	FunctionType = AssetType("Function")
)

type Asset interface {
	Type() AssetType
	Name() string
	Key() string
	PackageName() string

	SetBuildPath(path string)
	BuildPath() string
	SetS3Key(key string)
	S3Key() string
}

type Mutation struct {
	ServiceType reflect.Type
	Methods     []*Method
	s3Key       string
	buildPath   string
}

func isMutation(method reflect.Method) bool {
	if !strings.HasPrefix(method.Name, "On") || len(method.Name) <= 2 {
		return false
	}

	if method.Type.NumIn() != 3 || method.Type.NumOut() != 1 {
		return false
	}

	if !method.Type.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return false
	}

	event := method.Type.In(2)

	if event.Kind() == reflect.Ptr {
		event = event.Elem()
	}

	if method.Name[2:] != event.Name() {
		return false
	}

	return true
}

func (m *Mutation) Type() AssetType {
	return MutationType
}

func (m *Mutation) SetBuildPath(path string) {
	m.buildPath = path
}

func (m *Mutation) SetS3Key(key string) {
	m.s3Key = key
}

func (m *Mutation) BuildPath() string {
	return m.buildPath
}

func (m *Mutation) S3Key() string {
	return m.s3Key
}

func (m Mutation) Name() string {
	return m.ServiceType.Name()
}

func (m Mutation) EventTypes() []reflect.Type {
	events := []reflect.Type{}
	for _, method := range m.Methods {
		events = append(events, method.Event)
	}
	return events
}

func (m Mutation) Events() []string {
	events := m.EventTypes()
	result := []string{}

	for _, event := range events {

		name := event.String()
		if event.Kind() == reflect.Ptr {
			name = event.Elem().String()
		}

		result = append(result, name)
	}

	return result
}

func (m Mutation) EventNames() []string {
	events := m.EventTypes()
	result := []string{}

	for _, event := range events {

		name := event.Name()
		if event.Kind() == reflect.Ptr {
			name = event.Elem().Name()
		}

		result = append(result, name)
	}

	return result
}

func (m *Mutation) Add(method reflect.Method) error {

	log.Println("Method:", method)
	log.Println("ServiceType:", m.ServiceType)
	log.Println("Event:", method.Type.In(2))
	log.Println("Event package:", method.Type.In(2).String())
	log.Println("")

	m.Methods = append(m.Methods, &Method{
		Method:      method,
		ServiceType: m.ServiceType,
		Event:       method.Type.In(2),
	})
	return nil
}

func (m Mutation) PackageName() string {
	return path.Base(m.Package())
}

func (m Mutation) Package() string {
	return m.ServiceType.PkgPath()
}

func (m Mutation) Key() string {
	return fmt.Sprintf("%x",
		md5.Sum([]byte(fmt.Sprintf("%s.%s",
			m.Package(),
			m.Name(),
		))))
}

// TODO: sort imports
func (m Mutation) Imports() []string {
	imports := map[string]struct{}{}
	for _, method := range m.Methods {
		for _, i := range method.Imports() {
			imports[i] = struct{}{}
		}
	}

	defaultImports := []string{"context", m.ServiceType.PkgPath()}
	for _, i := range defaultImports {
		imports[i] = struct{}{}
	}

	delete(imports, "")
	result := []string{}
	for i := range imports {
		result = append(result, i)
	}
	sort.Strings(result)
	return result
}

type Command struct {
	*Method
}

func (c *Command) Type() AssetType {
	return CommandType
}

type Query struct {
	*Method
}

func (c *Query) Type() AssetType {
	return QueryType
}

type Method struct {
	ServiceType reflect.Type
	Method      reflect.Method
	Event       reflect.Type
	s3Key       string
	buildPath   string
}

func (m *Method) SetBuildPath(path string) {
	m.buildPath = path
}

func (m *Method) SetS3Key(key string) {
	m.s3Key = key
}

func (m *Method) BuildPath() string {
	return m.buildPath
}

func (m *Method) S3Key() string {
	return m.s3Key
}

func (m *Method) Name() string {
	return m.Method.Name
}
func (m *Method) PackageName() string {
	return path.Base(m.Package())
}

func (m *Method) Package() string {
	pkg := m.ServiceType.PkgPath()
	if m.ServiceType.Kind() == reflect.Ptr {
		pkg = m.ServiceType.Elem().PkgPath()
	}
	return pkg
}

func (m *Method) Inputs() []reflect.Type {
	result := []reflect.Type{}
	for i := 0; i < m.Method.Type.NumIn(); i++ {
		t := m.Method.Type.In(i)
		if t == m.ServiceType {
			continue
		}
		result = append(result, t)
	}
	return result
}

func (m *Method) Outputs() []reflect.Type {
	result := []reflect.Type{}
	for i := 0; i < m.Method.Type.NumOut(); i++ {
		t := m.Method.Type.Out(i)
		result = append(result, t)
	}
	return result
}

func (m *Method) Imports() []string {

	types := []reflect.Type{m.ServiceType}
	types = append(types, m.Inputs()...)
	types = append(types, m.Outputs()...)

	imports := map[string]struct{}{}
	for _, t := range types {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		imports[t.PkgPath()] = struct{}{}
	}
	delete(imports, "")
	result := []string{}
	for i := range imports {
		result = append(result, i)
	}
	sort.Strings(result)
	return result
}

func (m *Method) String() string {
	inputs := []string{}
	for _, in := range m.Inputs() {
		inputs = append(inputs, in.String())
	}
	outputs := []string{}
	for _, out := range m.Outputs() {
		outputs = append(outputs, out.String())
	}
	return fmt.Sprintf("func(%s) (%s)",
		strings.Join(inputs, ","),
		strings.Join(outputs, ","),
	)
}

func (m *Method) Key() string {
	//return m.Package() + "." + m.Name()
	return fmt.Sprintf("%x",
		md5.Sum([]byte(fmt.Sprintf("%s.%s",
			m.Package(),
			m.Name(),
		))))
}
