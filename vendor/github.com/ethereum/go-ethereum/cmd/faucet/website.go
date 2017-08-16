// Code generated by go-bindata.
// sources:
// faucet.html
// DO NOT EDIT!

package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _faucetHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x59\x6d\x6f\xdc\x36\x12\xfe\xec\xfc\x8a\xa9\x2e\xad\x77\x61\x4b\xb2\xe3\x20\x2d\xd6\xd2\x16\x41\x9a\x4b\x7b\x38\xb4\x45\x9b\xe2\xae\x68\x8b\x03\x25\xcd\x4a\x8c\x29\x52\x25\x87\xbb\xde\x1a\xfb\xdf\x0f\x24\x25\xad\x76\x6d\xa7\xb9\x4b\xf3\x61\x23\x92\x33\xcf\xbc\x51\xf3\x22\x67\x9f\x7c\xf5\xdd\xab\xb7\x3f\x7f\xff\x1a\x1a\x6a\xc5\xf2\x49\xe6\xfe\x03\xc1\x64\x9d\x47\x28\xa3\xe5\x93\x93\xac\x41\x56\x2d\x9f\x9c\x9c\x64\x2d\x12\x83\xb2\x61\xda\x20\xe5\x91\xa5\x55\xfc\x45\xb4\x3f\x68\x88\xba\x18\x7f\xb7\x7c\x9d\x47\xff\x8e\x7f\x7a\x19\xbf\x52\x6d\xc7\x88\x17\x02\x23\x28\x95\x24\x94\x94\x47\xdf\xbc\xce\xb1\xaa\x71\xc2\x27\x59\x8b\x79\xb4\xe6\xb8\xe9\x94\xa6\x09\xe9\x86\x57\xd4\xe4\x15\xae\x79\x89\xb1\x5f\x9c\x03\x97\x9c\x38\x13\xb1\x29\x99\xc0\xfc\x32\x5a\x3e\x71\x38\xc4\x49\xe0\xf2\xee\x2e\xf9\x16\x69\xa3\xf4\xcd\x6e\xb7\x80\x37\x9c\xbe\xb6\x05\xfc\x9d\xd9\x12\x29\x4b\x03\x89\xa7\x16\x5c\xde\x40\xa3\x71\x95\x47\x4e\x67\xb3\x48\xd3\xb2\x92\xef\x4c\x52\x0a\x65\xab\x95\x60\x1a\x93\x52\xb5\x29\x7b\xc7\x6e\x53\xc1\x0b\x93\xd2\x86\x13\xa1\x8e\x0b\xa5\xc8\x90\x66\x5d\x7a\x95\x5c\x25\x9f\xa7\xa5\x31\xe9\xb8\x97\xb4\x5c\x26\xa5\x31\x11\x68\x14\x79\x64\x68\x2b\xd0\x34\x88\x14\x41\xba\xfc\xff\xe4\xae\x94\xa4\x98\x6d\xd0\xa8\x16\xd3\xe7\xc9\xe7\xc9\x85\x17\x39\xdd\x7e\xbf\x54\x27\xd6\x94\x9a\x77\x04\x46\x97\x1f\x2c\xf7\xdd\xef\x16\xf5\x36\xbd\x4a\x2e\x93\xcb\x7e\xe1\xe5\xbc\x33\xd1\x32\x4b\x03\xe0\xf2\xa3\xb0\x63\xa9\x68\x9b\x3e\x4b\x9e\x27\x97\x69\xc7\xca\x1b\x56\x63\x35\x48\x72\x47\xc9\xb0\xf9\x97\xc9\x7d\x2c\x86\xef\x8e\x43\xf8\x57\x08\x6b\x55\x8b\x92\x92\x77\x26\x7d\x96\x5c\x7e\x91\x5c\x0c\x1b\xf7\xf1\xbd\x00\x17\x34\x27\xea\x24\x59\xa3\x26\x5e\x32\x11\x97\x28\x09\x35\xdc\xb9\xdd\x93\x96\xcb\xb8\x41\x5e\x37\xb4\x80\xcb\x8b\x8b\x4f\xaf\x1f\xda\x5d\x37\x61\xbb\xe2\xa6\x13\x6c\xbb\x80\x95\xc0\xdb\xb0\xc5\x04\xaf\x65\xcc\x09\x5b\xb3\x80\x80\xec\x0f\x76\x5e\x66\xa7\x55\xad\xd1\x98\x5e\x58\xa7\x0c\x27\xae\xe4\xc2\xdd\x28\x46\x7c\x8d\x0f\xd1\x9a\x8e\xc9\x7b\x0c\xac\x30\x4a\x58\xc2\x23\x45\x0a\xa1\xca\x9b\xb0\xe7\x5f\xe3\xa9\x11\xa5\x12\x4a\x2f\x60\xd3\xf0\x9e\x0d\xbc\x20\xe8\x34\xf6\xf0\xd0\xb1\xaa\xe2\xb2\x5e\xc0\x8b\xae\xb7\x07\x5a\xa6\x6b\x2e\x17\x70\xb1\x67\xc9\xd2\xc1\x8d\x59\x1a\x32\xd6\x93\x93\xac\x50\xd5\xd6\xc7\xb0\xe2\x6b\x28\x05\x33\x26\x8f\x8e\x5c\xec\x33\xd1\x01\x81\x4b\x40\x8c\xcb\xe1\xe8\xe0\x4c\xab\x4d\x04\x5e\x50\x1e\x05\x25\xe2\x42\x11\xa9\x76\x01\x97\x4e\xbd\x9e\xe5\x08\x4f\xc4\xa2\x8e\x2f\x9f\x0d\x87\x27\x59\x73\x39\x80\x10\xde\x52\xec\xe3\x33\x46\x26\x5a\x66\x7c\xe0\x5d\x31\x58\xb1\xb8\x60\xd4\x44\xc0\x34\x67\x71\xc3\xab\x0a\x65\x1e\x91\xb6\xe8\xee\x11\x5f\xc2\x34\xef\x0d\x69\xef\xa5\xa5\x06\xa5\xb3\x93\xb0\xea\x93\x20\x1c\xc3\xd6\x9c\x1a\x5b\xc4\x4c\xd0\xa3\xe0\x59\xda\x5c\x0e\x26\xa5\x15\x5f\xf7\x1e\x99\x3c\x1e\x39\xe7\x71\xfb\xbf\x80\xfe\x41\xad\x56\x06\x29\x9e\xb8\x63\x42\xcc\x65\x67\x29\xae\xb5\xb2\xdd\x78\x7e\x92\xf9\x5d\xe0\x55\x1e\xd5\xdc\x50\x04\xb4\xed\x7a\xdf\x45\xa3\x49\x4a\xb7\xb1\x0b\x9d\x56\x22\x82\x4e\xb0\x12\x1b\x25\x2a\xd4\x79\xd4\xfb\xe4\x0d\x37\x04\x3f\xfd\xf0\x4f\xe8\x03\xcc\x65\x0d\x5b\x65\x35\xbc\xa6\x06\x35\xda\x16\x58\x55\xb9\xcb\x9d\x24\xc9\x44\xb6\xbf\xe9\xf7\xb5\x8b\x0b\x92\x7b\xaa\x93\xac\xb0\x44\x6a\x24\x2c\x48\x42\x41\x32\xae\x70\xc5\xac\x20\xa8\xb4\xea\x2a\xb5\x91\x31\xa9\xba\x76\x05\x31\x58\x10\x98\x22\xa8\x18\xb1\xfe\x28\x8f\x06\xda\x21\x28\xcc\x74\xaa\xb3\x5d\x1f\x96\xb0\x89\xb7\x1d\x93\x15\x56\x2e\x94\xc2\x60\xb4\x7c\xc3\xd7\x08\x2d\x06\x5b\x4e\x8e\x23\x5d\x32\x8d\x14\x4f\x41\x1f\x88\x74\x50\x26\x98\x04\xfd\xbf\xcc\x8a\x01\x69\x34\xa1\x45\x69\xe1\x60\x15\x6b\x97\x85\xa2\xe5\xdd\x9d\x66\xb2\x46\x78\xca\xab\xdb\x73\x78\xca\x5a\x65\x25\xc1\x22\x87\xe4\xa5\x7f\x34\xbb\xdd\x01\x3a\x40\x26\xf8\x32\x63\xef\x7b\x19\x40\xc9\x52\xf0\xf2\x26\x8f\x88\xa3\xce\xef\xee\x1c\xf8\x6e\x77\x0d\x77\x77\x7c\x05\x4f\x93\x1f\xb0\x64\x1d\x95\x0d\xdb\xed\x6a\x3d\x3c\x27\x78\x8b\xa5\x25\x9c\xcd\xef\xee\x50\x18\xdc\xed\x8c\x2d\x5a\x4e\xb3\x81\xdd\xed\xcb\x6a\xb7\x73\x3a\xf7\x7a\xee\x76\x90\x3a\x50\x59\xe1\x2d\x3c\x4d\xbe\x47\xcd\x55\x65\x20\xd0\x67\x29\x5b\x66\xa9\xe0\xcb\x9e\xef\xd0\x49\xa9\x15\xfb\xfb\x92\xba\x0b\x33\x5e\x6d\xff\xa6\x78\x55\xa7\x9a\x3e\x70\xf1\xeb\x78\xd4\xbe\xbf\x0f\x86\x13\xde\xe0\x36\x8f\xee\xee\xa6\xbc\xfd\x69\xc9\x84\x28\x98\xf3\x4b\x30\x6d\x64\xfa\x03\xdd\x3d\x5d\x73\xe3\x3b\xaf\xe5\xa0\xc1\x5e\xed\x0f\x7c\x93\x8f\xd2\x1c\xa9\x6e\x01\x57\xcf\x26\x39\xee\xa1\x97\xfc\xc5\xd1\x4b\x7e\xf5\x20\x71\xc7\x24\x0a\xf0\xbf\xb1\x69\x99\x18\x9e\xfb\xb7\x65\xf2\xf2\x1d\x33\xc5\x2e\xa3\x8f\xaa\x8d\x95\xe1\xe2\x1a\xd4\x1a\xf5\x4a\xa8\xcd\x02\x98\x25\x75\x0d\x2d\xbb\x1d\xab\xe3\xd5\xc5\xc5\x54\x6f\xd7\x31\xb2\x42\xa0\x4f\x28\x1a\x7f\xb7\x68\xc8\x8c\x89\x24\x1c\xf9\x5f\x97\x4f\x2a\x94\x06\xab\x23\x6f\x38\x89\xce\xb5\x9e\x6a\x12\xfa\xd1\x99\x0f\xea\xbe\x52\x6a\x2c\x38\x53\x35\x7a\xe8\x49\x6d\x8c\x96\x19\xe9\x3d\xdd\x49\x46\xd5\xff\x54\x30\xb4\x6b\x08\x1f\xab\x17\x21\xa3\x39\xdb\x3b\x44\x1d\xba\x11\x77\x65\xc1\x2f\xb3\x94\xaa\x8f\x90\xec\x2e\x61\xc1\x0c\x7e\x88\x78\xdf\x17\xec\xc5\xfb\xe5\xc7\xca\x6f\x90\x69\x2a\x90\x3d\x5e\xd2\x26\x0a\xac\xac\xac\x26\xf6\xfb\xdc\xf9\xb1\x0a\x58\xc9\xd7\xa8\x0d\xa7\xed\x87\x6a\x80\xd5\x5e\x85\xb0\x3e\x54\x21\x4b\x49\xbf\xff\xae\x4d\x17\x7f\xd1\xcb\xfd\x67\x0d\xcc\xd5\xf2\x6b\xb5\x81\x4a\xa1\x01\x6a\xb8\x01\xd7\x7e\x7c\x99\xa5\xcd\xd5\x48\xd2\x2d\xdf\xba\x03\xef\x54\x58\x85\x0e\x84\x1b\xd0\x56\xfa\xca\xab\x24\x50\x83\x87\xcd\x8b\x0c\x4f\x09\xbc\x55\xae\x01\x5c\xa3\x24\x68\x99\xe0\x25\x57\xd6\x00\x2b\x49\x69\x03\x2b\xad\x5a\xc0\xdb\x86\x59\x43\x0e\xc8\xa5\x0f\xb6\x66\x5c\xf8\x77\xc9\x87\x14\x94\x06\x56\x96\xb6\xb5\xae\x81\x95\x35\xa0\x54\xb6\x6e\x7a\x5d\x48\x41\x28\x4c\x42\xc9\x7a\xd4\xc7\x74\xac\x05\x46\xc4\xca\x1b\x73\x0e\x43\x56\x00\xa6\x11\x88\x63\xe5\xb8\xfa\x3e\x82\x95\xa5\x2f\x66\x09\xbc\x94\x5b\x25\x11\x1a\xb6\xf6\x8a\x1c\x11\x40\xcb\xb6\x03\x50\xaf\xd7\x86\x53\xc3\x83\xe1\x1d\xea\xd6\x4d\x24\x15\x08\xde\x72\x32\x49\x96\x76\x53\xdf\xa9\x43\xd6\x73\x30\xbc\xed\xc4\x16\x4a\x8d\x8c\x10\x18\x64\xec\x68\x98\x74\xad\x51\x12\x7a\x3a\x3f\x8e\x44\x40\x4c\xd7\x6e\x54\xff\x0f\x2b\x94\xa5\x45\x21\x98\xbc\x71\xad\xc2\xd8\x0e\xb9\xb2\xe6\x95\x7a\xb8\x11\x82\x8e\x19\xa7\x21\x97\xa4\xbc\xd2\xfd\x6c\x6e\x60\xe6\x56\x2b\x2e\xd0\x8f\xef\xfe\x1e\xc8\x53\x67\xb1\x9b\xb1\xe6\xe7\x50\xaa\x6e\x1b\xb8\x3d\x9f\x53\xcd\xf8\xde\x6b\x84\x62\x85\x5a\x23\x84\xc6\xae\x50\xb7\xc0\x64\x05\x2b\xae\x11\xd8\x86\x6d\x3f\x81\x9f\x95\x85\x92\x49\x20\xcd\xca\x9b\x20\xdb\x6a\xed\x2e\x44\x87\xd2\x25\xfd\x7d\x88\x0a\x14\x6a\xe3\x49\x02\xda\x8a\xa3\xf0\xf1\x32\x88\xd0\xa8\x0d\xb4\xb6\xf4\x06\xba\x40\xa1\x3b\xd8\x30\x4e\x60\x25\x71\x11\xec\x26\xab\x25\x94\xaa\xc5\x83\x28\xdc\xab\xda\x19\xb6\xcb\xb7\xce\xee\x7b\x97\x79\xac\xb7\xa0\xf1\x55\x20\x87\x4e\x2b\xc2\xd2\x0d\x46\xc0\x6a\xc6\xa5\x71\x76\xfa\x38\x63\xfb\x01\xf5\x78\x7c\xea\x1f\xf6\x93\xa8\x3f\x4e\x53\x78\x23\x54\xc1\x04\xac\x5d\x96\x29\x84\x7b\x11\x15\xb8\x96\xf7\xc0\x5b\x86\x18\x59\x03\x6a\xe5\x77\x83\xe6\x8e\x7f\xcd\xb4\xbb\xed\xd8\x76\x04\x79\x3f\x47\xb9\x3d\x83\x7a\xdd\x4f\x87\x6e\xe9\x7a\xae\x70\xde\x0b\xfd\x0a\x57\x5c\x86\xa0\xae\xac\x0c\xe6\x51\xc3\x08\x42\x17\x62\x80\xf9\x60\x83\xd5\x02\xfa\x48\x07\xc8\x51\x80\xa7\x83\x7c\x64\x9f\xdd\xf3\x73\xff\xd0\xfb\x68\xde\xcf\x81\x01\x26\x31\x28\xab\xd9\x3f\x7e\xfc\xee\xdb\xc4\x90\xe6\xb2\xe6\xab\xed\xec\xce\x6a\xb1\x80\xa7\xb3\xe8\x6f\x7e\x3c\x98\xff\x72\xf1\x5b\xb2\x66\xc2\xe2\xb9\x37\x60\xe1\x7f\xef\x89\x39\x87\xfe\x71\x01\x87\x12\x77\xf3\xf9\xf5\xc3\x2d\xdb\xa4\xc3\xd4\x68\x90\x66\x8e\x70\x8c\xe4\xee\xfa\xd0\x49\x0c\x5a\xa4\x46\xf9\xbb\xa8\xb1\x54\x52\x62\x49\x60\x3b\x25\x7b\x9f\x80\x50\xc6\x0c\x8e\xd9\x53\x4c\x7c\x33\x18\xcf\x57\x30\x1b\xc2\xf5\x29\x3c\x83\x3c\x87\x8b\xe1\xac\xf7\x0c\xe4\x20\x71\x03\xff\xc2\xe2\x47\x55\xde\x20\xcd\xa2\x8d\x71\x69\x21\x82\x33\x10\xaa\x64\x0e\x2f\x69\x94\x21\x38\x83\x28\x65\x1d\x8f\xe6\x61\x9a\xde\x81\x6b\x91\xff\x1c\xec\x83\xb0\xc2\xf7\x86\xa0\xe9\xd9\x59\xb8\x36\x43\xe8\x94\x6c\xd1\x18\x56\xe3\xd4\x42\x9f\xe5\x47\x53\x9c\x23\x5a\x53\x43\x0e\x3e\xc4\x1d\xd3\x06\x03\x49\xe2\x3a\x8b\x5e\x8a\x77\x87\x27\xcb\x73\x90\x56\x88\x91\xff\x44\xa3\x7b\x99\x7b\xb2\xdd\x93\x03\xf2\x24\x24\xe1\x4f\xf2\x1c\x5c\x99\x75\x31\xaa\xf6\x9c\xee\xfa\x84\x86\x60\x9e\xb8\x4a\xbf\xe7\x98\x8f\x70\xf7\xd0\xb0\xfa\x33\x38\xac\x8e\xf1\xb0\x7a\x04\xd0\xf7\x5f\xef\xc3\x0b\xfd\xda\x04\xce\x6f\x3c\x82\x26\x6d\x5b\xa0\x7e\x1f\x5c\xe8\xbf\x7a\x38\xef\xea\x6f\x24\x4d\x78\xcf\xe1\xf2\xc5\xfc\x11\x74\xd4\x5a\x3d\x0a\x2e\x15\x6d\x67\x77\x82\x6d\x5d\xd5\x81\x53\x52\xdd\x2b\xdf\x2e\x9d\x9e\x83\x93\xb5\x80\x11\xe1\xdc\x0f\xc2\x0b\x38\xf5\xab\xd3\xdd\x23\xd2\x8c\x2d\x4b\x57\x8f\x3e\x46\x5e\x8f\x31\x4a\xec\xd7\x8f\xca\x1c\xeb\xcb\x81\x50\xf8\xec\x33\xb8\x77\x7a\x78\x05\xdd\x1d\xee\x0b\x25\xe4\x10\x45\x3d\xfc\xc9\x4a\x69\x98\xb9\x43\x9e\x5f\x5c\x03\xcf\xa6\x30\x89\x40\x59\x53\x73\x0d\xfc\xec\x6c\x8f\x74\x32\xc0\x9c\xe5\x10\xb9\x89\x20\xa3\x6a\xe9\x3b\xb3\xd0\xbe\xfd\x1a\xb9\x09\xb0\xd6\xca\xca\x6a\xe1\x52\xee\xec\x74\xdf\x0c\x4c\xfa\x80\xb3\x03\x95\x7f\xe1\xbf\x25\xd6\xa0\xf6\x95\xfb\x0c\xa2\xa4\x93\xf5\x97\x7e\x6e\x7c\xf1\xfc\x74\x7e\x0d\x7b\x4c\x3f\x4d\x2e\xa0\x74\xb3\xd5\x35\x84\xf9\xc4\x77\x89\x30\x4e\x56\x7e\x55\x28\x5d\xa1\x8e\x35\xab\xb8\x35\x0b\x78\xde\xdd\x5e\xff\x3a\x4c\x9e\xbe\x97\xf5\x7a\x77\x1a\x97\x0f\xe9\x32\xb4\x4b\x67\x10\x65\xa9\x23\x1a\x58\x46\x2b\xa7\x5f\x0d\xe1\x81\x2e\x1c\xc6\x6f\x7a\xfd\x7e\xcb\xab\x4a\xa0\x53\xc2\x0b\x0c\x1f\x5f\x2b\xab\x7d\xe2\x9a\x85\xf5\xec\x58\x0f\xe2\x2d\xce\x13\x2b\xf9\xed\x6c\x1e\xf7\x34\xc3\xfa\x1c\x4e\x8d\xcb\xcf\x95\x39\x9d\x27\x8d\x6d\x99\xe4\x7f\xe0\xcc\xb5\xf4\xf3\xa0\xb7\xd3\xd8\xf5\xe9\x63\xb4\x77\x93\x17\x6d\x9c\x31\xe7\x49\x43\xad\x98\x45\x19\xf9\x2f\x93\x4e\xb9\x31\xc4\x1e\x25\x6c\x1f\xde\xc8\xdd\x61\x0e\x2d\x85\x32\x78\x54\x23\xc0\x20\xbd\xe5\x2d\x2a\x4b\xb3\xb1\x8e\x9c\xbb\xb9\xf7\x62\x7e\x0d\xbb\xfd\x07\xdc\x34\x85\xd7\xc6\x4d\x12\xdc\x34\xc0\x60\x83\x85\xf1\xf9\x1d\x7a\x1e\x5f\xce\x43\xd9\x7e\xf9\xfd\x37\x93\xd2\x3d\xa2\xce\xbc\x72\xe3\x07\xec\x87\xea\xe4\x83\x5f\xcc\x37\x9b\x4d\x52\x2b\x55\x8b\xf0\xad\x7c\x2c\xa4\xae\x7a\x24\xef\xdc\xb8\x6a\xb6\xb2\x84\x0a\x57\xa8\x97\x13\xf8\xbe\xba\x66\x69\xf8\x96\x9b\xa5\xe1\xef\x54\xff\x0d\x00\x00\xff\xff\x71\x50\x77\xf3\xb8\x1a\x00\x00")

func faucetHtmlBytes() ([]byte, error) {
	return bindataRead(
		_faucetHtml,
		"faucet.html",
	)
}

func faucetHtml() (*asset, error) {
	bytes, err := faucetHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "faucet.html", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"faucet.html": faucetHtml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"faucet.html": &bintree{faucetHtml, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
