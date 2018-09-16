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

const (
	btrfsQuotaDestroy = iota
	btrfsQuotaCreate
)

type btrfsCreateFlag C.__u64

func (m btrfsCreateFlag) asInt() C.__u64 {
	return C.__u64(m)
}

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
		return fmt.Errorf("failed to search qgroup for %s: %v", path, errno.Error())
	}
	sh := (*C.struct_btrfs_ioctl_search_header)(unsafe.Pointer(&args.buf))
	if sh._type != C.BTRFS_QGROUP_STATUS_KEY {
		return fmt.Errorf("invalid qgroup search header type for %s: %v", path, sh._type)
	}
	return nil
}

func (btrfsNativeAPI) QuotaExists(ctx context.Context, qgroupID string, path string) (bool, error) {
	var args C.struct_btrfs_ioctl_search_args
	args.key.tree_id = C.BTRFS_QUOTA_TREE_OBJECTID
	args.key.max_type = C.BTRFS_QGROUP_INFO_KEY
	args.key.min_type = C.BTRFS_QGROUP_INFO_KEY
	args.key.max_objectid = C.__u64(math.MaxUint64)
	args.key.max_offset = C.__u64(math.MaxUint64)
	args.key.max_transid = C.__u64(math.MaxUint64)
	args.key.nr_items = 4096

	qgroupid, err := parseQgroupID(qgroupID)
	if err != nil {
		return false, err
	}

	dir, err := openDir(path)
	if err != nil {
		return false, err
	}
	defer closeDir(dir)

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_TREE_SEARCH,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return false, fmt.Errorf("failed to search qgroup for %s: %v", path, errno.Error())
	}

	var sh C.struct_btrfs_ioctl_search_header
	shSize := unsafe.Sizeof(sh)
	buf := (*[1<<31 - 1]byte)(unsafe.Pointer(&args.buf[0]))[:C.BTRFS_SEARCH_ARGS_BUFSIZE]

	for i := C.uint(0); i < args.key.nr_items; i++ {
		sh = (*(*C.struct_btrfs_ioctl_search_header)(unsafe.Pointer(&buf[0])))
		buf = buf[shSize:]
		if sh._type == C.BTRFS_QGROUP_INFO_KEY && uint64(sh.offset) == qgroupid {
			return true, nil
		}
		buf = buf[sh.len:]
	}

	return false, nil
}

func (b btrfsNativeAPI) QuotaCreate(ctx context.Context, qgroupID string, path string) error {
	return b.quotaCreateOrDestroy(qgroupID, path, btrfsQuotaCreate)
}

func (b btrfsNativeAPI) QuotaDestroy(ctx context.Context, qgroupID string, path string) error {
	return b.quotaCreateOrDestroy(qgroupID, path, btrfsQuotaDestroy)
}

func (btrfsNativeAPI) quotaCreateOrDestroy(qgroupID string, path string, create btrfsCreateFlag) error {
	dir, err := openDir(path)
	if err != nil {
		return err
	}
	defer closeDir(dir)

	var args C.struct_btrfs_ioctl_qgroup_create_args
	args.create = create.asInt()

	qgroupid, err := parseQgroupID(qgroupID)
	if err != nil {
		return err
	}
	args.qgroupid = C.__u64(qgroupid)
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_QGROUP_CREATE,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("failed to create qgroup %s for %s: %v", qgroupID, path, errno.Error())
	}

	return nil
}

func (btrfsNativeAPI) QuotaLimit(ctx context.Context, sizeInBytes uint64, qgroupID string, path string) error {
	dir, err := openDir(path)
	if err != nil {
		return err
	}
	defer closeDir(dir)

	qgroupid, err := parseQgroupID(qgroupID)
	if err != nil {
		return err
	}

	var args C.struct_btrfs_ioctl_qgroup_limit_args
	args.lim.max_exclusive = C.__u64(sizeInBytes)
	args.lim.flags = C.BTRFS_QGROUP_LIMIT_MAX_EXCL
	args.qgroupid = C.__u64(qgroupid)
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_QGROUP_LIMIT,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("failed to limit qgroup %s for %s: %v", qgroupID, path, errno.Error())
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

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_QGROUP_ASSIGN,
		uintptr(unsafe.Pointer(&args)))
	// https://github.com/kdave/btrfs-progs/blob/23df5de0d07028f31f017a0f80091ba158980742/cmds-qgroup.c#L108
	if errno == 0 {
		return nil
	} else if errno < 0 {
		return fmt.Errorf("failed to assign qgroup for %s: %v", dir, errno.Error())
	}
	// errno > 0 -> rescan
	var qargs C.struct_btrfs_ioctl_quota_rescan_args
	_, _, errno = unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_QUOTA_RESCAN, uintptr(unsafe.Pointer(&qargs)))
	if errno < 0 {
		return fmt.Errorf("rescan failed, quotas may be inconsistent %s: %v", dir, errno)
	}
	return nil
}

func (btrfsNativeAPI) GetQuotaID(ctx context.Context, path string) (string, error) {
	dir, err := openDir(path)
	if err != nil {
		return "", err
	}
	defer closeDir(dir)
	// https://github.com/kdave/btrfs-progs/blob/7faaca0d9f78f7162ae603231f693dd8e1af2a41/cmds-qgroup.c#L387
	var xargs C.struct_btrfs_ioctl_ino_lookup_args
	xargs.treeid = 0
	xargs.objectid = C.BTRFS_FIRST_FREE_OBJECTID
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_INO_LOOKUP,
		uintptr(unsafe.Pointer(&xargs)))
	if errno != 0 {
		return "", fmt.Errorf("failed to search qgroup at %s: %v", path, errno.Error())
	}

	var args C.struct_btrfs_ioctl_search_args
	args.key.tree_id = C.BTRFS_QUOTA_TREE_OBJECTID
	args.key.max_type = C.BTRFS_QGROUP_INFO_KEY
	args.key.min_type = C.BTRFS_QGROUP_INFO_KEY
	args.key.max_objectid = C.__u64(math.MaxUint64)
	args.key.max_offset = C.__u64(math.MaxUint64)
	args.key.max_transid = C.__u64(math.MaxUint64)
	args.key.nr_items = 4096

	_, _, errno = unix.Syscall(unix.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_TREE_SEARCH,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return "", fmt.Errorf("failed to search qgroup at %s: %v", path, errno.Error())
	}

	var sh C.struct_btrfs_ioctl_search_header
	shSize := unsafe.Sizeof(sh)
	buf := (*[1<<31 - 1]byte)(unsafe.Pointer(&args.buf[0]))[:C.BTRFS_SEARCH_ARGS_BUFSIZE]

	for i := C.uint(0); i < args.key.nr_items; i++ {
		sh = (*(*C.struct_btrfs_ioctl_search_header)(unsafe.Pointer(&buf[0])))
		buf = buf[shSize:]
		if sh._type == C.BTRFS_QGROUP_INFO_KEY && xargs.treeid == sh.offset {
			return fmt.Sprintf("%d/%d", sh.objectid, sh.offset), nil
		}
		buf = buf[sh.len:]
	}

	// return btrfsCLI{}.GetQuotaID(ctx, path)
	return "", fmt.Errorf("not found")
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
		return nil, fmt.Errorf("can't open dir")
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
