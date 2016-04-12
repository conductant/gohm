package fuse

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"errors"
	"golang.org/x/net/context"
	"os"
)

type Dir struct {
	fs *FS
	// path from root to this dir; empty for root dir
	path []string
}

var _ = fs.Node(&Dir{})

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0755
	return nil
}

var _ = fs.HandleReadDirAller(&Dir{})

func (d *Dir) ReadDirAll(c context.Context) ([]fuse.Dirent, error) {
	var res []fuse.Dirent
	err := d.fs.backend.View(c, func(ctx Context) error {
		b, err := ctx.Dir(d.path)
		if err != nil {
			return err
		}
		if b == nil {
			return errors.New("dir no longer exists")
		}
		for entry := range b.Cursor() {
			de := fuse.Dirent{
				Name: entry.Name,
			}
			if entry.Dir {
				de.Type = fuse.DT_Dir
			} else {
				de.Type = fuse.DT_File
			}
			res = append(res, de)
		}
		return nil
	})
	return res, err
}

var _ = fs.NodeStringLookuper(&Dir{})

func (d *Dir) Lookup(c context.Context, name string) (fs.Node, error) {
	var n fs.Node
	err := d.fs.backend.View(c, func(ctx Context) error {
		b, err := ctx.Dir(d.path)
		if err != nil {
			return err
		}
		if b == nil {
			return errors.New("dir no longer exists")
		}
		if child, err := b.GetDir(name); err != nil {
			return err
		} else if child != nil {
			// directory
			n = &Dir{
				fs:   d.fs,
				path: append(d.path, name),
			}
			return nil
		}
		if child, err := b.Get(name); err != nil {
			return err
		} else if child != nil {
			// file
			n = &File{
				dir:  d,
				name: name,
			}
			return nil
		}
		return fuse.ENOENT
	})
	if err != nil {
		return nil, err
	}
	return n, nil
}

var _ = fs.NodeMkdirer(&Dir{})

func (d *Dir) Mkdir(c context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	name := req.Name
	err := d.fs.backend.Update(c, func(ctx Context) error {
		b, err := ctx.Dir(d.path)
		if err != nil {
			return err
		}
		if b == nil {
			return errors.New("dir no longer exists")
		}
		if child, err := b.GetDir(name); err != nil {
			return err
		} else if child != nil {
			return fuse.EEXIST
		}
		if _, err := b.CreateDir(name); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	n := &Dir{
		fs:   d.fs,
		path: append(d.path, name),
	}
	return n, nil
}

var _ = fs.NodeCreater(&Dir{})

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {

	name := req.Name
	f := &File{
		dir:     d,
		name:    name,
		writers: 1,
		// file is empty at Create time, no need to set data
	}
	return f, f, nil
}

var _ = fs.NodeRemover(&Dir{})

func (d *Dir) Remove(c context.Context, req *fuse.RemoveRequest) error {
	name := req.Name
	return d.fs.backend.Update(c, func(ctx Context) error {
		b, err := ctx.Dir(d.path)
		if err != nil {
			return err
		}
		if b == nil {
			return errors.New("dir no longer exists")
		}

		switch req.Dir {
		case true:
			if d, err := b.GetDir(name); err != nil {
				return err
			} else if d == nil {
				return fuse.ENOENT
			}
			if err := b.DeleteDir(name); err != nil {
				return err
			}

		case false:
			if d, err := b.Get(name); err != nil {
				return err
			} else if d == nil {
				return fuse.ENOENT
			}
			if err := b.Delete(name); err != nil {
				return err
			}
		}
		return nil
	})
}

var _ = fs.NodeRenamer(&Dir{})

func (d *Dir) Rename(c context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	return d.fs.backend.Update(c, func(ctx Context) error {
		b, err := ctx.Dir(d.path)
		if err != nil {
			return err
		}
		data, err := b.Get(req.OldName)
		if err != nil {
			return err
		}
		if data == nil {
			return fuse.ENOENT
		}
		targetDir := b
		if d, ok := newDir.(*Dir); ok {
			dd, err := ctx.Dir(d.path)
			if err == nil {
				targetDir = dd
			} else {
				return err
			}
		}
		err = targetDir.Put(req.NewName, data)
		if err != nil {
			return err
		}
		return b.Delete(req.OldName)
	})
}

var _ = fs.NodeLinker(&Dir{})

func (d *Dir) Link(c context.Context, req *fuse.LinkRequest, old fs.Node) (fs.Node, error) {
	return nil, fuse.ENOSYS
}

var _ = fs.NodeSymlinker(&Dir{})

func (d *Dir) Symlink(c context.Context, req *fuse.SymlinkRequest) (fs.Node, error) {
	return nil, fuse.ENOSYS
}

var _ = fs.NodeReadlinker(&File{})

func (d *File) Readlink(c context.Context, req *fuse.ReadlinkRequest) (string, error) {
	return "", fuse.ENOSYS
}