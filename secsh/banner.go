package secsh

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

type Banner struct {
	banner []byte
}

func NewBanner(ctx context.Context, log *zap.SugaredLogger) *Banner {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		log.Warnw("failed to obtain host info", zap.Error(err))
	}

	la, err := load.AvgWithContext(ctx)
	if err != nil {
		log.Warnw("failed to obtain load average info", zap.Error(err))
	}

	du, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		log.Warnw("failed to obtain disk info", zap.Error(err))
	}

	mu, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		log.Warnw("failed to obtain virtual memory info", zap.Error(err))
	}

	swap, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		log.Warnw("failed to obtain swap memory info", zap.Error(err))
	}

	platform := ""
	platformVersion := ""
	os := ""
	kernelVersion := ""

	if info != nil {
		platform = info.Platform
		platformVersion = info.PlatformVersion
		os = info.OS
		kernelVersion = info.KernelVersion
	}

	banner := &Banner{}
	banner.AddLine(
		fmt.Sprintf("Welcome to %s %s (GNU/%s %s)",
			strings.Title(platform),
			platformVersion,
			strings.Title(os),
			kernelVersion,
		),
	)

	banner.AddLine("")
	banner.AddLine(fmt.Sprintf("  System information as of %s", time.Now().Format(time.UnixDate)))
	banner.AddLine("")
	if la != nil {
		banner.AddLine(fmt.Sprintf(`  System load:  %.1f`, la.Load1))
	}
	if du != nil {
		banner.AddLine(fmt.Sprintf(`  Usage of /:   %.1f%% of %.2fGB`, du.UsedPercent, float64(du.Total)/1024/1024/1024))
	}
	if mu != nil {
		banner.AddLine(fmt.Sprintf(`  Memory usage: %.1f%%`, mu.UsedPercent))
	}
	if swap != nil {
		banner.AddLine(fmt.Sprintf(`  Swap usage:   %.1f%%`, swap.UsedPercent))
	}
	if info != nil {
		banner.AddLine(fmt.Sprintf(`  Processes:    %d`, info.Procs))
	}

	users, err := host.UsersWithContext(ctx)
	if err != nil {
		log.Warnw("failed to obtain users info", zap.Error(err))
	}

	if users != nil {
		if len(users) > 0 {
			banner.AddLine("")
			banner.AddLine(`  Users Logged In: `)
			for _, user := range users {
				banner.AddLine(
					fmt.Sprintf("   * %s %s (%s on %s)",
						user.User,
						time.Unix(int64(user.Started), 0).Format(time.UnixDate),
						user.Host,
						user.Terminal,
					),
				)
			}
		}
	}

	return banner
}

func (m *Banner) AddLine(line string) {
	m.banner = append(m.banner, []byte(line)...)
	m.banner = append(m.banner, '\n')
}

func (m *Banner) String() string {
	return string(m.banner)
}
