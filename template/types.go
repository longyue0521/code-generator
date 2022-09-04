package template

type ServiceMethod struct {
	Name         string
	Path         string
	ReqTypeName  string
	RespTypeName string
}

type ServiceDefinition struct {
	Package string
	Name    string
	Methods []ServiceMethod
}

func (s ServiceDefinition) GenName() string {
	return s.Name + "Gen"
}
