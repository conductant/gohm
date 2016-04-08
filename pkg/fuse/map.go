package fuse

type mapbe struct {
	tree map[string]interface{}
}

func NewMapBackend(m map[string]interface{}) Backend {
	return &SimpleBackend{DirSource: &mapbe{tree: m}}
}

func (this *mapbe) Dir(path []string) (DirLike, error) {
	l := this
	for _, p := range path {
		if m, has := l.tree[p]; has {
			if ll, ok := m.(*mapbe); ok {
				l = ll
			}
		}
	}
	return l, nil
}

func (this *mapbe) GetDir(name string) (DirLike, error) {
	if m, has := this.tree[name]; has {
		if mm, ok := m.(*mapbe); ok {
			return mm, nil
		}
	}
	return nil, nil
}

func (this *mapbe) CreateDir(name string) (DirLike, error) {
	n := &mapbe{
		tree: map[string]interface{}{},
	}
	this.tree[name] = n
	return n, nil
}

func (this *mapbe) DeleteDir(name string) error {
	if _, has := this.tree[name]; has {
		delete(this.tree, name)
	}
	return nil
}

func (this *mapbe) Cursor() <-chan Entry {
	out := make(chan Entry)
	go func() {
		defer close(out)
		for k, v := range this.tree {
			entry := Entry{Name: k}

			switch v.(type) {
			case *mapbe:
				entry.Dir = true
			default:
				entry.Dir = false
			}

			out <- entry
		}
	}()
	return out
}

func (this *mapbe) Get(name string) ([]byte, error) {
	if v, has := this.tree[name]; has {
		if b, ok := v.([]byte); ok {
			return b, nil
		}
	}
	return nil, nil
}

func (this *mapbe) Put(name string, value []byte) error {
	this.tree[name] = value
	return nil
}

func (this *mapbe) Delete(name string) error {
	delete(this.tree, name)
	return nil
}
