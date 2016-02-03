package zk

import (
	"github.com/conductant/gohm/pkg/registry"
	"github.com/conductant/gohm/pkg/template"
)

// Note that the referenced implementation using the registry API
// already implements the Source function in a generic way.  The
// init function here is to bind the protocol and the implementation.
func init() {
	template.Register("zk", registry.Source)
}
