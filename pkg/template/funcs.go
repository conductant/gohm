package template

import (
	"fmt"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

var (
	NullTemplate string = ""

	lock      sync.Mutex
	userFuncs = map[string]func(context.Context) interface{}{}
)

func RegisterFunc(name string, generator func(context.Context) interface{}) {
	lock.Lock()
	defer lock.Unlock()
	userFuncs[name] = generator
}

func DefaultFuncMap(ctx context.Context) template.FuncMap {
	fm := template.FuncMap{}
	for k, v := range userFuncs {
		fm[k] = v(ctx)
	}
	return fm
}

func MergeFuncMaps(a, b template.FuncMap) template.FuncMap {
	merged := template.FuncMap{}
	for k, v := range a {
		merged[k] = v
	}
	for k, v := range b {
		merged[k] = v
	}
	return merged
}

func init() {
	RegisterFunc("host", ParseHost)
	RegisterFunc("port", ParsePort)
	RegisterFunc("inline", ContentInline)
	RegisterFunc("file", ContentToFile)
	RegisterFunc("sh", ExecuteShell)
}

func ParseHost(ctx context.Context) interface{} {
	return func(hostport string) (string, error) {
		host, _, err := net.SplitHostPort(hostport)
		return host, err
	}
}

func ParsePort(ctx context.Context) interface{} {
	return func(hostport string) (string, error) {
		_, port, err := net.SplitHostPort(hostport)
		return port, err
	}
}

// Fetch the url and write content inline
// ex) {{ inline "http://file/here" }}
func ContentInline(ctx context.Context) interface{} {
	return func(uri string) (string, error) {
		data := ContextGetTemplateData(ctx)
		applied, err := Apply(uri, data)
		if err != nil {
			return NullTemplate, err
		}
		url := string(applied)
		content, err := Source(ctx, url)
		if err != nil {
			return NullTemplate, err
		}
		return string(content), nil
	}
}

// Fetch the url and write content to temp file or given filepath.  File mode also an option.
// ex) {{ file "http://file/here" "/path/to/file" "0644" }}
func ContentToFile(ctx context.Context) interface{} {
	return func(uri string, opts ...string) (string, error) {
		data := ContextGetTemplateData(ctx)
		applied, err := Apply(uri, data)
		if err != nil {
			return NullTemplate, err
		}
		url := string(applied)
		content, err := Source(ctx, url)
		if err != nil {
			return NullTemplate, err
		}

		// Write to local file and return the path, unless the
		// path is provided.
		destination := os.TempDir()
		// the destination path is given
		if len(opts) >= 1 {
			destination = opts[0]
			// We support variables inside the function argument
			p, err := Apply(destination, data)
			if err != nil {
				return NullTemplate, err
			}
			destination = string(p)
			switch {
			case strings.Index(destination, "~") > -1:
				// expand tilda
				destination = strings.Replace(destination, "~", os.Getenv("HOME"), 1)
			case strings.Index(destination, "./") > -1:
				// expand tilda
				destination = strings.Replace(destination, "./", os.Getenv("PWD")+"/", 1)
			}
		}
		// Default permission unless it's provided
		var perm os.FileMode = 0644
		if len(opts) >= 2 {
			permString := opts[1]
			perm = fileModeFromString(permString)
		}
		fpath := destination
		parent := filepath.Dir(fpath)
		fi, err := os.Stat(parent)
		if err != nil {
			switch {
			case os.IsNotExist(err):
				err = os.MkdirAll(parent, 0777)
				if err != nil {
					return NullTemplate, err
				}
			default:
				return NullTemplate, err
			}
		}
		// read again after we created the directories
		fi, err = os.Stat(fpath)
		if err == nil && fi.IsDir() {
			// build the name because we provided only a directory path
			fpath = filepath.Join(destination, filepath.Base(string(url)))
		}

		err = ioutil.WriteFile(fpath, []byte(content), perm)
		glog.Infoln("Written", len([]byte(content)), " bytes to", fpath, "perm=", perm.String(), "Err=", err)
		if err != nil {
			return NullTemplate, err
		}
		return fpath, nil
	}
}

func fileModeFromString(perm string) os.FileMode {
	if len(perm) < 4 {
		perm = fmt.Sprintf("%04v", perm)
	}
	fm := new(os.FileMode)
	fmt.Sscanf(perm, "%v", fm)
	return *fm
}
