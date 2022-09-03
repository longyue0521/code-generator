package anthor

import (
	"io"
	"text/template"
)

// 这部分和课堂的很像，但是有一些地方被我改掉了
const serviceTpl = `package {{.Package}}

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type {{.GenName}} struct {
    Endpoint string
    Path string
	Client http.Client
}
{{ range $method := .Methods }}
func (s *UserServiceGen) {{$method.Name}}(ctx context.Context, req *{{$method.ReqTypeName}}) (*{{$method.RespTypeName}}, error) {
	url := s.Endpoint + s.Path + "{{$method.Path}}"
	bs, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	body := &bytes.Buffer{}
	body.Write(bs)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	httpResp, err := s.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	bs, err = ioutil.ReadAll(httpResp.Body)
	resp := &{{$method.RespTypeName}}{}
	err = json.Unmarshal(bs, resp)
	return resp, err
}
{{end}}
`

func Gen(writer io.Writer, def ServiceDefinition) error {
	tpl, _ := template.New("service").Parse(serviceTpl)
	// 还可以进一步调用 format.Source 来格式化生成代码
	return tpl.Execute(writer, def)
}
