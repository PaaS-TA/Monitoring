// +build !linux

package chown

func Chown(path string, uid, gid int) error {
	panic("not supported on this OS")
}
