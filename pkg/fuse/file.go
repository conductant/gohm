package fuse

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
	"golang.org/x/net/context"
	"sync"
	"syscall"
)

type File struct {
	dir  *Dir
	name string

	mu sync.Mutex
	// number of write-capable handles currently open
	writers uint
	// only valid if writers > 0
	data []byte
}

var _ = fs.Node(&File{})
var _ = fs.Handle(&File{})

// load calls fn inside a View with the contents of the file. Caller
// must make a copy of the data if needed.
func (f *File) load(c context.Context, fn func([]byte)) error {
	err := f.dir.fs.backend.View(c, func(ctx Context) error {
		b, err := ctx.Dir(f.dir.path)
		if err != nil {
			return err
		}
		v, err := b.Get(f.name)
		if err != nil {
			return err
		}
		if v == nil {
			return fuse.ESTALE
		}
		fn(v)
		return nil
	})
	return err
}

func (f *File) Attr(c context.Context, a *fuse.Attr) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	a.Mode = 0644
	a.Size = uint64(len(f.data))
	if f.writers == 0 {
		// not in memory, fetch correct size.
		// Attr can't fail, so ignore errors
		_ = f.load(c, func(b []byte) { a.Size = uint64(len(b)) })
	}
	return nil
}

var _ = fs.NodeOpener(&File{})

func (f *File) Open(c context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	if req.Flags.IsReadOnly() {
		// we don't need to track read-only handles
		return f, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.writers == 0 {
		// load data
		fn := func(b []byte) {
			f.data = append([]byte(nil), b...)
		}
		if err := f.load(c, fn); err != nil {
			return nil, err
		}
	}

	f.writers++
	return f, nil
}

var _ = fs.HandleReleaser(&File{})

func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	if req.Flags.IsReadOnly() {
		// we don't need to track read-only handles
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.writers--
	if f.writers == 0 {
		f.data = nil
	}
	return nil
}

var _ = fs.HandleReader(&File{})

func (f *File) Read(c context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	fn := func(b []byte) {
		fuseutil.HandleRead(req, resp, b)
	}
	if f.writers == 0 {
		f.load(c, fn)
	} else {
		fn(f.data)
	}
	return nil
}

var _ = fs.HandleWriter(&File{})

const maxInt = int(^uint(0) >> 1)

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// expand the buffer if necessary
	newLen := req.Offset + int64(len(req.Data))
	if newLen > int64(maxInt) {
		return fuse.Errno(syscall.EFBIG)
	}
	if newLen := int(newLen); newLen > len(f.data) {
		f.data = append(f.data, make([]byte, newLen-len(f.data))...)
	}

	n := copy(f.data[req.Offset:], req.Data)
	resp.Size = n
	return nil
}

var _ = fs.HandleFlusher(&File{})

func (f *File) Flush(c context.Context, req *fuse.FlushRequest) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.writers == 0 {
		// Read-only handles also get flushes. Make sure we don't
		// overwrite valid file contents with a nil buffer.
		return nil
	}

	err := f.dir.fs.backend.Update(c, func(ctx Context) error {
		b, err := ctx.Dir(f.dir.path)
		if err != nil {
			return err
		}
		return b.Put(f.name, f.data)
	})
	if err != nil {
		return err
	}
	return nil
}

var _ = fs.NodeSetattrer(&File{})

func (f *File) Setattr(c context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if req.Valid.Size() {
		if req.Size > uint64(maxInt) {
			return fuse.Errno(syscall.EFBIG)
		}
		newLen := int(req.Size)
		switch {
		case newLen > len(f.data):
			f.data = append(f.data, make([]byte, newLen-len(f.data))...)
		case newLen < len(f.data):
			f.data = f.data[:newLen]
		}
	}
	return nil
}

var _ = fs.NodeFsyncer(&File{})

// Here we don't do anything and just handle the request with no error.
// This is because we have a simple Put semantic that writes all the data in a single
// call.  So there's no flushing or marking a commit.
func (f *File) Fsync(c context.Context, req *fuse.FsyncRequest) error {
	return nil
}