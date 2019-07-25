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
)

type Banner struct {
	banner []byte
}

func NewBanner(ctx context.Context) (*Banner, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	la, err := load.AvgWithContext(ctx)
	if err != nil {
		return nil, err
	}

	du, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		return nil, err
	}

	mu, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	swap, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}

	banner := &Banner{}
	banner.AddLine(
		fmt.Sprintf("Welcome to %s %s (GNU/%s %s)",
			strings.Title(info.Platform),
			info.PlatformVersion,
			strings.Title(info.OS),
			info.KernelVersion,
		),
	)

	banner.AddLine("")
	banner.AddLine(fmt.Sprintf("  System information as of %s", time.Now().Format(time.UnixDate)))
	banner.AddLine("")
	banner.AddLine(fmt.Sprintf(`  System load:  %.1f`, la.Load1))
	banner.AddLine(fmt.Sprintf(`  Usage of /:   %.1f%% of %.2fGB`, du.UsedPercent, float64(du.Total)/1024/1024/1024))
	banner.AddLine(fmt.Sprintf(`  Memory usage: %.1f%%`, mu.UsedPercent))
	banner.AddLine(fmt.Sprintf(`  Swap usage:   %.1f%%`, swap.UsedPercent))
	banner.AddLine(fmt.Sprintf(`  Processes:    %d`, info.Procs))

	users, err := host.UsersWithContext(ctx)
	if err != nil {
		return nil, err
	}

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

	return banner, nil
}

func (m *Banner) AddLine(line string) {
	m.banner = append(m.banner, []byte(line)...)
	m.banner = append(m.banner, '\n')
}

func (m *Banner) String() string {
	return string(m.banner)
}
