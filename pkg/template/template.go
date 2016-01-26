package template

import (
	"bytes"
	"golang.org/x/net/context"
	"hash/fnv"
	"strconv"
	"strings"
	"text/template"
)

type templateDataContextKey int

var (
	TemplateDataContextKey templateDataContextKey = 1
)

func ContextPutTemplateData(ctx context.Context, data interface{}) context.Context {
	return context.WithValue(ctx, TemplateDataContextKey, data)
}
func ContextGetTemplateData(ctx context.Context) interface{} {
	return ctx.Value(TemplateDataContextKey)
}

func GetKeyForTemplate(tmpl []byte) string {
	hash := fnv.New64a()
	hash.Write(tmpl)
	return strconv.FormatUint(hash.Sum64(), 16)
}

func Apply(tmpl string, data interface{}, funcs ...template.FuncMap) ([]byte, error) {
	fm := template.FuncMap{}
	for _, opt := range funcs {
		fm = MergeFuncMaps(fm, opt)
	}
	t := template.New(GetKeyForTemplate([]byte(tmpl))).Funcs(fm)
	t, err := t.Parse(tmpl)
	if err != nil {
		return nil, err
	}
	var buff bytes.Buffer
	err = t.Execute(&buff, data)
	return buff.Bytes(), err
}

func Execute(ctx context.Context, uri string, funcs ...template.FuncMap) ([]byte, error) {
	data := ContextGetTemplateData(ctx)
	fm := DefaultFuncMap(ctx)
	for _, opt := range funcs {
		fm = MergeFuncMaps(fm, opt)
	}

	url := uri
	if applied, err := Apply(uri, data, fm); err != nil {
		return nil, err
	} else {
		url = string(applied)
	}

	body := NullTemplate
	switch {
	case strings.Index(url, "func://") == 0:
		if f, has := fm[url[len("func://"):]]; has {
			if ff, ok := f.(func() string); ok {
				body = ff()
			} else {
				return nil, ErrBadTemplateFunc
			}
		} else {
			return nil, ErrMissingTemplateFunc
		}
	default:
		if bytes, err := Source(ctx, url); err != nil {
			return nil, err
		} else {
			body = string(bytes)
		}
	}
	return Apply(body, data, fm)
}
