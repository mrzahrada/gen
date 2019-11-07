package gen

type ConfigMethod struct {
	Name    string `json:"name"`
	S3Key   string `json:"s3Key"`
	Handler string `json:"handler"`
	Runtime string `json:"runtime"`
}

// Config -
// TODO: some meta about commit version
type Config struct {
	ServiceName string         `json:"service"`
	Bucket      string         `json:"bucket"`
	Events      []string       `json:"events,omitempty"`
	Commands    []ConfigMethod `json:"commands,omitempty"`
	Queries     []ConfigMethod `json:"queries,omitempty"`
	Mutation    *ConfigMethod  `json:"mutations,omitempty"`
	Functions   []ConfigMethod `json:"functions,omitempty"`
}

type CDKConfig struct {
	App     string     `json:"app"`
	Context CDKContext `json:"context"`
}

type CDKContext struct {
	Name   string `json:"name"`
	Bucket string `json:"bucket"`
}
