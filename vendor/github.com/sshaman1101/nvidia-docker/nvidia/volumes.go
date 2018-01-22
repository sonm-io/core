// Copyright (c) 2015-2016, NVIDIA CORPORATION. All rights reserved.

package nvidia

import (
	"bufio"
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/NVIDIA/nvidia-docker/src/ldcache"
)

const (
	binDir   = "bin"
	lib32Dir = "lib"
	lib64Dir = "lib64"
)

type components map[string][]string

type volumeDir struct {
	name  string
	files []string
}

type VolumeInfo struct {
	Name         string
	Mountpoint   string
	MountOptions string
	Components   components
}

type Volume struct {
	*VolumeInfo

	Path    string
	Version string
	dirs    []volumeDir
}

type VolumeMap map[string]*Volume

type FileCloneStrategy interface {
	Clone(src, dst string) error
}

type LinkStrategy struct{}

func (s LinkStrategy) Clone(src, dst string) error {
	return os.Link(src, dst)
}

type LinkOrCopyStrategy struct{}

func (s LinkOrCopyStrategy) Clone(src, dst string) error {
	// Prefer hard link, fallback to copy
	err := os.Link(src, dst)
	if err != nil {
		err = Copy(src, dst)
	}
	return err
}

func Copy(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	fi, err := s.Stat()
	if err != nil {
		return err
	}

	d, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}

	if err := d.Chmod(fi.Mode()); err != nil {
		d.Close()
		return err
	}

	return d.Close()
}

func blacklisted(file string, obj *elf.File) (bool, error) {
	lib := regexp.MustCompile(`^.*/lib([\w-]+)\.so[\d.]*$`)
	glcore := regexp.MustCompile(`libnvidia-e?glcore\.so`)
	gldispatch := regexp.MustCompile(`libGLdispatch\.so`)

	if m := lib.FindStringSubmatch(file); m != nil {
		switch m[1] {

		// Blacklist EGL/OpenGL libraries issued by other vendors
		case "EGL":
			fallthrough
		case "GLESv1_CM":
			fallthrough
		case "GLESv2":
			fallthrough
		case "GL":
			deps, err := obj.DynString(elf.DT_NEEDED)
			if err != nil {
				return false, err
			}
			for _, d := range deps {
				if glcore.MatchString(d) || gldispatch.MatchString(d) {
					return false, nil
				}
			}
			return true, nil

		// Blacklist TLS libraries using the old ABI (!= 2.3.99)
		case "nvidia-tls":
			const abi = 0x6300000003
			s, err := obj.Section(".note.ABI-tag").Data()
			if err != nil {
				return false, err
			}
			return binary.LittleEndian.Uint64(s[24:]) != abi, nil
		}
	}
	return false, nil
}

func (v *Volume) Create(s FileCloneStrategy) (err error) {
	root := path.Join(v.Path, v.Version)
	if err = os.MkdirAll(root, 0755); err != nil {
		return
	}

	defer func() {
		if err != nil {
			v.Remove()
		}
	}()

	for _, d := range v.dirs {
		vpath := path.Join(root, d.name)
		if err := os.MkdirAll(vpath, 0755); err != nil {
			return err
		}

		// For each file matching the volume components (blacklist excluded), create a hardlink/copy
		// of it inside the volume directory. We also need to create soname symlinks similar to what
		// ldconfig does since our volume will only show up at runtime.
		for _, f := range d.files {
			obj, err := elf.Open(f)
			if err != nil {
				return fmt.Errorf("%s: %v", f, err)
			}
			defer obj.Close()

			ok, err := blacklisted(f, obj)
			if err != nil {
				return fmt.Errorf("%s: %v", f, err)
			}
			if ok {
				continue
			}

			l := path.Join(vpath, path.Base(f))
			if err := s.Clone(f, l); err != nil {
				return err
			}

			soname, err := obj.DynString(elf.DT_SONAME)
			if err != nil {
				return fmt.Errorf("%s: %v", f, err)
			}
			if len(soname) > 0 {
				l = path.Join(vpath, soname[0])
				if err := os.Symlink(path.Base(f), l); err != nil && !os.IsExist(err) {
					return err
				}
				// XXX Many applications (wrongly) assume that libcuda.so exists (e.g. with dlopen)
				// Hardcode the libcuda symlink for the time being.
				if strings.HasPrefix(soname[0], "libcuda") {
					l = strings.TrimRight(l, ".0123456789")
					if err := os.Symlink(path.Base(f), l); err != nil && !os.IsExist(err) {
						return err
					}
				}

				// XXX GLVND requires this symlink for indirect GLX support
				// It won't be needed once we have an indirect GLX vendor neutral library.
				if strings.HasPrefix(soname[0], "libGLX_nvidia") {
					l = strings.Replace(l, "GLX_nvidia", "GLX_indirect", 1)
					if err := os.Symlink(path.Base(f), l); err != nil && !os.IsExist(err) {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (v *Volume) Remove(version ...string) error {
	vv := v.Version
	if len(version) == 1 {
		vv = version[0]
	}
	return os.RemoveAll(path.Join(v.Path, vv))
}

func (v *Volume) Exists(version ...string) (bool, error) {
	vv := v.Version
	if len(version) == 1 {
		vv = version[0]
	}
	_, err := os.Stat(path.Join(v.Path, vv))
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func (v *Volume) ListVersions() ([]string, error) {
	dirs, err := ioutil.ReadDir(v.Path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	versions := make([]string, len(dirs))
	for i := range dirs {
		versions[i] = dirs[i].Name()
	}
	return versions, nil
}

func which(bins ...string) ([]string, error) {
	paths := make([]string, 0, len(bins))

	out, _ := exec.Command("which", bins...).Output()
	r := bufio.NewReader(bytes.NewBuffer(out))
	for {
		p, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if p = strings.TrimSpace(p); !path.IsAbs(p) {
			continue
		}
		path, err := filepath.EvalSymlinks(p)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func LookupVolumes(prefix, ver string, vi []VolumeInfo) (vols VolumeMap, err error) {
	cache, err := ldcache.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := cache.Close(); err == nil {
			err = e
		}
	}()

	vols = make(VolumeMap, len(vi))

	for i := range vi {
		vol := &Volume{
			VolumeInfo: &vi[i],
			Path:       path.Join(prefix, vi[i].Name),
			Version:    ver,
		}

		for t, c := range vol.Components {
			switch t {
			case "binaries":
				bins, err := which(c...)
				if err != nil {
					return nil, err
				}
				vol.dirs = append(vol.dirs, volumeDir{binDir, bins})
			case "libraries":
				libs32, libs64 := cache.Lookup(c...)
				vol.dirs = append(vol.dirs,
					volumeDir{lib32Dir, libs32},
					volumeDir{lib64Dir, libs64},
				)
			}
		}
		vols[vol.Name] = vol
	}
	return
}
