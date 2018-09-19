// +build linux

package btrfs

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"

	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
)

type API interface {
	QuotaEnable(ctx context.Context, path string) error
	QuotaExists(ctx context.Context, qgroupID string, path string) (bool, error)
	QuotaCreate(ctx context.Context, qgroupID string, path string) error
	QuotaDestroy(ctx context.Context, qgroupID string, path string) error
	QuotaLimit(ctx context.Context, sizeInBytes uint64, qgroupID string, path string) error
	QuotaAssign(ctx context.Context, src string, dst string, path string) error
	QuotaRemove(ctx context.Context, src string, dst string, path string) error
	GetQuotaID(ctx context.Context, path string) (string, error)
}

var _ API = btrfsCLI{}

func NewAPI() (API, error) {
	if _, err := exec.LookPath("btrfs"); err != nil {
		return nil, fmt.Errorf("btrfs executable is not found. Install btrfs-progs")
	}
	return btrfsCLI{}, nil
}

type btrfsCLI struct{}

func (btrfsCLI) QuotaEnable(ctx context.Context, path string) error {
	output, err := exec.CommandContext(ctx, "btrfs", "quota", "enable", path).Output()
	if err != nil {
		log.G(ctx).Error("failed to create to enable btrfs quota", zap.String("path", path), zap.Error(err), zap.ByteString("output", output))
		return err
	}
	return nil
}

func (btrfsCLI) QuotaExists(ctx context.Context, qgroupID string, path string) (bool, error) {
	output, err := exec.CommandContext(ctx, "btrfs", "qgroup", "show", path).Output()
	if err != nil {
		log.G(ctx).Error("failed to lookup quota", zap.String("path", path), zap.String("quota", qgroupID), zap.Error(err))
	}

	return lookupQuotaInShowOutput(output, qgroupID)
}

func lookupQuotaInShowOutput(output []byte, qgroupID string) (bool, error) {
	qgroupIDBytes := []byte(qgroupID)
	r := bytes.NewReader(output)
	scanner := bufio.NewScanner(r)
	foundHeader := false
	endOfHeaderToken := []byte("--------") // next line after qgroupid
	for scanner.Scan() {
		line := scanner.Bytes()
		if foundHeader {
			pos := bytes.IndexByte(line, ' ')
			if pos != -1 && bytes.Equal(qgroupIDBytes, line[:pos]) {
				return true, nil
			}
		} else {
			foundHeader = bytes.HasPrefix(line, endOfHeaderToken)
		}
	}
	return false, scanner.Err()
}

func (btrfsCLI) GetQuotaID(ctx context.Context, path string) (string, error) {
	output, err := exec.CommandContext(ctx, "btrfs", "qgroup", "show", "-f", path).Output()
	if err != nil {
		log.G(ctx).Error("failed to lookup groupid for a btrfs subvolume", zap.String("path", path), zap.Error(err))
		return "", err
	}
	return lookupIDForSubvolumeWithPath(output)
}

func lookupIDForSubvolumeWithPath(output []byte) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	foundHeader := false
	endOfHeaderToken := []byte("--------")
	for scanner.Scan() {
		line := scanner.Bytes()
		if foundHeader {
			if pos := bytes.IndexByte(line, ' '); pos != -1 {
				return string(line[:pos]), nil
			}
			return "", errors.New("malformed format")
		}
		foundHeader = bytes.HasPrefix(line, endOfHeaderToken)
	}
	if scanner.Err() != nil {
		return "", scanner.Err()
	}
	return "", errors.New("not found")
}

func (btrfsCLI) QuotaCreate(ctx context.Context, qgroupID string, path string) error {
	output, err := exec.CommandContext(ctx, "btrfs", "qgroup", "create", qgroupID, path).Output()
	if err != nil {
		log.G(ctx).Error("failed to create btrfs qgroup", zap.String("path", path), zap.String("qgroupid", qgroupID), zap.Error(err), zap.ByteString("output", output))
		return err
	}
	return nil
}

func (btrfsCLI) QuotaDestroy(ctx context.Context, qgroupID string, path string) error {
	output, err := exec.CommandContext(ctx, "btrfs", "qgroup", "destroy", qgroupID, path).Output()
	if err != nil {
		log.G(ctx).Error("failed to destroy btrfs qgroup", zap.String("path", path), zap.String("qgroupid", qgroupID), zap.Error(err), zap.ByteString("output", output))
		return err
	}
	return nil
}

func (btrfsCLI) QuotaLimit(ctx context.Context, sizeInBytes uint64, qgroupID string, path string) error {
	output, err := exec.CommandContext(ctx, "btrfs", "qgroup", "limit", strconv.FormatUint(sizeInBytes, 10), qgroupID, path).Output()
	if err != nil {
		log.G(ctx).Error("failed to limit btrfs qgroup", zap.String("qgroupid", qgroupID), zap.Error(err), zap.ByteString("output", output))
		return err
	}
	return nil
}

func (btrfsCLI) QuotaAssign(ctx context.Context, src string, dst string, path string) error {
	output, err := exec.CommandContext(ctx, "btrfs", "qgroup", "assign", "--rescan", src, dst, path).CombinedOutput()
	if err != nil {
		if bytes.HasPrefix(output, []byte("WARNING: quotas may be inconsistent, rescan needed")) {
			_, err = exec.CommandContext(ctx, "btrfs", "quota", "rescan", "-w", path).Output()
			return err
		}
		log.G(ctx).Error("failed to assign btrfs qgroup", zap.String("src", src), zap.String("dst", dst), zap.Error(err), zap.ByteString("output", output))
		return err
	}
	return nil
}

func (btrfsCLI) QuotaRemove(ctx context.Context, src string, dst string, path string) error {
	output, err := exec.CommandContext(ctx, "btrfs", "qgroup", "remove", src, dst, path).CombinedOutput()
	if err != nil {
		log.G(ctx).Error("failed to remove btrfs qgroup", zap.String("src", src), zap.String("dst", dst), zap.Error(err), zap.ByteString("output", output))
		return err
	}
	return nil
}
