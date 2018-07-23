// +build linux,btrfsnative

package btrfs

/*
#include <dirent.h>
#include <stdlib.h>
#include <btrfs/ioctl.h>
#include <btrfs/ctree.h>
*/
import "C"

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

type btrfsNativeAPI struct{}

var _ API = btrfsNativeAPI{}

func (btrfsNativeAPI) QuotaEnable(ctx context.Context, path string) error {
	dir, err := openDir(path)
	if err != nil {
		return err
	}
	defer closeDir(dir)

	var args C.struct_btrfs_ioctl_search_args
	args.key.tree_id = C.BTRFS_QUOTA_TREE_OBJECTID
	args.key.min_type = C.BTRFS_QGROUP_STATUS_KEY
	args.key.max_type = C.BTRFS_QGROUP_STATUS_KEY
	args.key.max_objectid = C.__u64(math.MaxUint64)
	args.key.max_offset = C.__u64(math.MaxUint64)
	args.key.max_transid = C.__u64(math.MaxUint64)
	args.key.nr_items = 4096

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_TREE_SEARCH,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("Failed to search qgroup for %s: %v", path, errno.Error())
	}
	sh := (*C.struct_btrfs_ioctl_search_header)(unsafe.Pointer(&args.buf))
	if sh._type != C.BTRFS_QGROUP_STATUS_KEY {
		return fmt.Errorf("Invalid qgroup search header type for %s: %v", path, sh._type)
	}
	return nil
}

func (btrfsNativeAPI) QuotaExists(ctx context.Context, qgroupID string, path string) (bool, error) {
	return false, fmt.Errorf("NOT IMPLEMENTED NATIVE API CALL")
}

func (b btrfsNativeAPI) QuotaCreate(ctx context.Context, qgroupID string, path string) error {
	return b.quotaCreateOrDestroy(qgroupID, path, true)
}

func (b btrfsNativeAPI) QuotaDestroy(ctx context.Context, qgroupID string, path string) error {
	return b.quotaCreateOrDestroy(qgroupID, path, false)
}

func (btrfsNativeAPI) quotaCreateOrDestroy(qgroupID string, path string, create bool) error {
	dir, err := openDir(path)
	if err != nil {
		return err
	}
	defer closeDir(dir)

	var args C.struct_btrfs_ioctl_qgroup_create_args
	if create {
		args.create = 1
	} else {
		args.create = 0
	}
	qgroupid, err := parseQgroupID(qgroupID)
	if err != nil {
		return err
	}
	args.qgroupid = C.__u64(qgroupid)
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_QGROUP_CREATE,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("Failed to create qgroup for %s: %v", dir, errno.Error())
	}

	return nil
}

func (btrfsNativeAPI) QuotaLimit(ctx context.Context, sizeInBytes uint64, qgroupID string, path string) error {
	dir, err := openDir(path)
	if err != nil {
		return err
	}
	defer closeDir(dir)

	var args C.struct_btrfs_ioctl_qgroup_limit_args
	args.lim.max_exclusive = C.__u64(sizeInBytes)
	args.lim.flags = C.BTRFS_QGROUP_LIMIT_MAX_EXCL
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_QGROUP_LIMIT,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("Failed to limit qgroup for %s: %v", dir, errno.Error())
	}

	return nil
}

func (b btrfsNativeAPI) QuotaAssign(ctx context.Context, src string, dst string, path string) error {
	return b.quotaAssignOrRemove(src, dst, path, true)
}

func (b btrfsNativeAPI) QuotaRemove(ctx context.Context, src string, dst string, path string) error {
	return b.quotaAssignOrRemove(src, dst, path, false)
}

func (btrfsNativeAPI) quotaAssignOrRemove(src string, dst string, path string, assign bool) error {
	srcU64, err := parseQgroupID(src)
	if err != nil {
		return err
	}
	dstU64, err := parseQgroupID(dst)
	if err != nil {
		return err
	}
	var args C.struct_btrfs_ioctl_qgroup_assign_args
	if assign {
		args.assign = 1
	} else {
		args.assign = 0
	}
	args.src = C.__u64(srcU64)
	args.dst = C.__u64(dstU64)

	dir, err := openDir(path)
	if err != nil {
		return err
	}
	defer closeDir(dir)

	r1, r2, errno := unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_QGROUP_ASSIGN,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("Failed to assign qgroup for %s: %v", dir, errno.Error())
	}
	fmt.Printf("%v %v %s\n", r1, r2, errno)
	// TODO: add rescan in case of error!!!
	// https://github.com/kdave/btrfs-progs/blob/23df5de0d07028f31f017a0f80091ba158980742/cmds-qgroup.c#L108

	return nil
}

func (btrfsNativeAPI) GetQuotaID(ctx context.Context, path string) (string, error) {
	return "", fmt.Errorf("NOT IMPLEMENTED NATIVE API CALL")
}

// https://github.com/oldcap/btrfs-progs/blob/master/qgroup.c#L1223
func parseQgroupID(p string) (uint64, error) {
	pos := strings.IndexByte(p, '/')
	if pos == -1 {
		return strconv.ParseUint(p, 10, 64)
	}

	level, err := strconv.ParseUint(p[:pos], 10, 64)
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseUint(p[pos+1:], 10, 64)
	if err != nil {
		return 0, err
	}

	return (level << 48) | id, nil
}

func free(p *C.char) {
	C.free(unsafe.Pointer(p))
}

func openDir(path string) (*C.DIR, error) {
	Cpath := C.CString(path)
	defer free(Cpath)

	dir := C.opendir(Cpath)
	if dir == nil {
		return nil, fmt.Errorf("Can't open dir")
	}
	return dir, nil
}

func closeDir(dir *C.DIR) {
	if dir != nil {
		C.closedir(dir)
	}
}

func getDirFd(dir *C.DIR) uintptr {
	return uintptr(C.dirfd(dir))
}
