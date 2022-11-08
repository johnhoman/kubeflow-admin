package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/alecthomas/kong"
	"sigs.k8s.io/yaml"
)

const wrapperTemplate = `
package {{ .package }}

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)
{{ $typeName := .typeName }}

const (
	err{{ .typeName }}Kind    = "cannot convert to {{ .typeName }}, invalid kind"
{{- range $attribute := .attributes }}
    errSet{{ $attribute.name }} = "Could not set attribute {{ $attribute.name}} value on kind {{ $typeName }}"
    errGet{{ $attribute.name }} = "Could not get attribute {{ $attribute.name}} value from kind {{ $typeName }}"
{{- end }}
)


type {{ .typeName }} struct {
	obj map[string]any
}

func (w *{{ .typeName }}) ToUnstructured() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: w.obj}
}

func (w *{{ .typeName }}) UnstructuredContent() map[string]any {
	return w.obj
}

{{ range $attribute := .attributes }}
func (w *{{ $typeName }}) Get{{ $attribute.name }}() ({{ $attribute.type }}, error) {
    {{- if eq $attribute.type "string" }}
    const def = ""
	{{ $attribute.nameLower }}, ok, err := unstructured.NestedString(w.obj, {{- range $item := $attribute.path }}"{{ $item }}", {{end }} "{{ $attribute.nameLower }}")
    {{ end }}
    {{- if eq $attribute.type "bool" }}
    const def = false
	{{ $attribute.nameLower }}, ok, err := unstructured.NestedBool(w.obj, {{ range $item := $attribute.path }}"{{ $item }}", {{end -}} "{{ $attribute.nameLower }}")
    {{ end }}
	if err != nil {
		return def, errors.Wrap(err, errGet{{ $attribute.name }})
	}
    if !ok {
        return def, nil
    }
    return {{ $attribute.nameLower }}, nil
}

func (w *{{ $typeName }}) Set{{ $attribute.name }}({{ $attribute.nameLower }} {{ $attribute.type }}) error {
	err := unstructured.SetNestedField(w.obj, {{ $attribute.nameLower }}, {{ range $item := $attribute.path }}"{{ $item }}", {{end }} "{{ $attribute.nameLower }}")
	if err != nil {
		return errors.Wrap(err, errSet{{ $attribute.name }})
	}
    return nil
}
{{ end }}

func New{{ .typeName }}FromUnstructured(u *unstructured.Unstructured) (*{{ .typeName }}, error) {
	if u.GroupVersionKind() != GroupVersion.WithKind({{ .typeName }}Kind) {
		return nil, errors.New(err{{ .typeName }}Kind)
	}
	return &{{ .typeName }}{obj: u.Object}, nil
}

func NewUnstructured{{ .typeName }}() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": GroupVersion.String(),
		"kind":       {{ .typeName }}Kind,
	}}
}

var (
	{{ .typeName }}Kind      = "{{ .typeName }}"
	{{ .typeName }}GroupKind = GroupVersion.WithKind({{ .typeName }}Kind).GroupKind()
)
`

var cli struct {
	Config string `help:"yaml file containing the type spec"`
}

func main() {
	ctx := kong.Parse(&cli)

	tmpl, err := template.New("").Parse(wrapperTemplate)
	ctx.FatalIfErrorf(err, "cannot parse template")

	data := make(map[string]any)

	f, err := os.Open(cli.Config)
	ctx.FatalIfErrorf(err, "could not open config file")

	raw, err := ioutil.ReadAll(f)
	ctx.FatalIfErrorf(err, "could not read config file")

	raw, err = yaml.YAMLToJSON(raw)
	ctx.FatalIfErrorf(err, "could not convert yaml to json")
	ctx.FatalIfErrorf(json.Unmarshal(raw, &data))
	ctx.FatalIfErrorf(tmpl.Execute(os.Stdout, data), "failed to execute template")
}
