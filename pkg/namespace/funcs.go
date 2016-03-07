package namespace

import (
	"fmt"
	"github.com/conductant/gohm/pkg/template"
	"golang.org/x/net/context"
	net "net/url"
)

// Export registry-related template functions

func init() {
	template.RegisterFunc("exists", ExistsTemplateFunc)
	template.RegisterFunc("get", GetTemplateFunc)
	template.RegisterFunc("list", ListTemplateFunc)
}

func registryAndPath(ctx context.Context, url string) (Registry, Path, error) {
	reg, err := Dial(ctx, url)
	if err != nil {
		return nil, EmptyPath, err
	}
	parsed, err := net.Parse(url)
	if err != nil {
		return nil, EmptyPath, err
	}
	return reg, FromUrl(parsed), nil
}

// Template function that returns a list of members as string
func ListTemplateFunc(ctx context.Context) interface{} {
	return func(url string) ([]*net.URL, error) {
		reg, path, err := registryAndPath(ctx, url)
		if err != nil {
			return nil, err
		}
		// To make this compatible we need to include the registry id url as prefix
		list, err := reg.List(path)
		if err != nil {
			return nil, err
		}
		out := make([]*net.URL, len(list))
		for i, v := range list {
			fullUrl := new(net.URL)
			*fullUrl = reg.Id()
			fullUrl.Path = v.String()
			out[i] = fullUrl
		}
		return out, nil
	}
}

// Template function that returns a True or False that a path exists
func ExistsTemplateFunc(ctx context.Context) interface{} {
	return func(url string) (bool, error) {
		reg, path, err := registryAndPath(ctx, url)
		if err != nil {
			return false, err
		}
		return reg.Exists(path)
	}
}

// Template function that returns a string at the path/url
func GetTemplateFunc(ctx context.Context) interface{} {
	return func(url interface{}) (string, error) {
		var urlStr string
		switch url := url.(type) {
		case string:
			urlStr = url
		default:
			urlStr = fmt.Sprintf("%v", url)
		}
		reg, path, err := registryAndPath(ctx, urlStr)
		if err != nil {
			return err.Error(), err
		}
		buff, _, err := reg.Get(path)
		if err != nil {
			return err.Error(), err
		}
		return string(buff), nil
	}
}
