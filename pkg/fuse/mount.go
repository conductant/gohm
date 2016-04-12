package fuse

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"io"
	"os"
)

type handle struct {
	io.Closer

	conn *fuse.Conn
}

func (this *handle) Close() error {
	if this.conn != nil {
		return this.conn.Close()
	}
	return nil
}

// Blocks and serve requests.  Close the stop channel will cause this to return
// which will along the say unmount the mountpoint and release resources.
func Serve(mountpoint string, b Backend, stop <-chan interface{}) error {
	closer, err := Mount(mountpoint, b)
	if err != nil {
		return err
	}

	<-stop

	Unmount(mountpoint)
	return closer.Close()
}

// Mount does not block.  It's up to the caller to block by reading on a channel, etc.
func Mount(mountpoint string, b Backend) (io.Closer, error) {
	var perm os.FileMode = 0644
	if err := os.MkdirAll(mountpoint, perm); err != nil {
		return nil, err
	}

	c, err := fuse.Mount(mountpoint)
	if err != nil {
		return nil, err
	}

	go func() {
		fs.Serve(c, &FS{backend: b})
	}()
	return &handle{conn: c}, nil
}

func Unmount(mountpoint string) error {
	return fuse.Unmount(mountpoint)
}
