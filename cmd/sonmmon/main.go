// +build linux

package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math"
	"math/big"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/cmd/cli/config"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/netutil"
	"github.com/sonm-io/core/util/xgrpc"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func guessHomeDir(log *zap.Logger) (string, error) {
	if home, err := guessHomeViaEnv(); err == nil {
		return home, nil
	} else {
		log.Debug("cannot guess home dir using env vars", zap.Error(err))
	}

	if home, err := guessHomeViaProc(); err == nil {
		return home, err
	} else {
		log.Debug("cannot guess home dir using node process stats", zap.Error(err))
	}

	return "", fmt.Errorf("failed to guess $HOME for user")
}

func guessHomeViaEnv() (string, error) {
	userName := os.Getenv("SONM_USER")
	if len(userName) == 0 {
		return "", fmt.Errorf("environment variable `SONM_USER` not set")
	}

	u, err := user.Lookup(userName)
	if err != nil {
		return "", fmt.Errorf("failed to lookup user: %v", err)
	}

	return u.HomeDir, nil
}

func guessHomeViaProc() (string, error) {
	spid, err := exec.Command("pgrep", "-x", "-o", "sonmnode").Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute `pgrep` command: %v", err)
	}

	pid := -1
	pid, err = strconv.Atoi(strings.TrimSpace(string(spid)))
	if err != nil {
		return "", fmt.Errorf("failed to parse pid from `%s`: %v", spid, err)
	}

	procPath := fmt.Sprintf("/proc/%d/cmdline", pid)
	st, err := os.Stat(procPath)
	if err != nil {
		return "", fmt.Errorf("cannot stat `%s`: %v", procPath, err)
	}

	stsys, ok := st.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("cannot get syscall.Stat_t")
	}

	u, err := user.LookupId(fmt.Sprintf("%d", stsys.Uid))
	if err != nil {
		return "", fmt.Errorf("cannot lookup user %d: %v\n", stsys.Uid, err)
	}

	return u.HomeDir, nil
}

func guessConfigPath(log *zap.Logger) (string, error) {
	home, err := guessHomeDir(log)
	if err != nil {
		// note: not sure with this solution, maybe we should instantly terminate.
		log.Info("failed to guess config path, fallback to user `sonm`", zap.Error(err))
		return "/home/sonm/.sonm/cli.yaml", nil
	}

	p := path.Join(home, util.HomeConfigDir, "cli.yaml")
	return p, nil
}

type WorkerStatus struct {
	Success   bool
	LastCheck time.Time
	Status    string `json:"status"`
	Uptime    string `json:"uptime"`
	IP        string `json:"ip"`
	Sold      struct {
		RAM     uint64 `json:"ram"`
		Storage uint64 `json:"storage"`
		NetIn   uint64 `json:"net_in"`
		NetOut  uint64 `json:"net_out"`
		GPUS    uint64 `json:"gpus"`
		CPUS    uint64 `json:"cpu"`
	} `json:"sold"`
	Worker     string        `json:"worker"`
	WorkerName string        `json:"worker_name"`
	Master     string        `json:"master"`
	Income     string        `json:"income"`
	CliExtra   []interface{} `json:"cliextra"`
}

func NewWorkerStatus() *WorkerStatus {
	return &WorkerStatus{
		Success:   false,
		LastCheck: time.Time{},
		Status:    "UNAVAILABLE",
		Uptime:    "0s",
		IP:        "127.0.0.1",
	}
}

func (w *WorkerStatus) update(ctx context.Context, cc *grpc.ClientConn, addr string) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	client := sonm.NewWorkerManagementClient(cc)
	md := metadata.MD{util.WorkerAddressHeader: []string{addr}}
	metaCtx := metadata.NewOutgoingContext(ctx, md)

	status, err := client.Status(metaCtx, &sonm.Empty{})
	if err != nil {
		w.Success = false
		return err
	}

	w.Success = true
	w.LastCheck = time.Now()
	w.Worker = status.EthAddr
	w.Master = status.Master.Unwrap().Hex()
	if !status.IsMasterConfirmed {
		w.Status = "AWAITING MASTER CONFIRMATION"
		return nil
	} else if !status.IsBenchmarkFinished {
		w.Status = "BENCHMARKING"
		return nil
	} else {
		w.Status = "WORKING"
	}

	askPlans, err := client.AskPlans(metaCtx, &sonm.Empty{})
	if err != nil {
		return err
	}

	devices, err := client.Devices(metaCtx, &sonm.Empty{})
	if err != nil {
		return err
	}

	var uptime string
	dur := time.Duration(status.Uptime) * time.Second
	if dur.Hours() > 1 {
		uptime = fmt.Sprintf("%dh %dm", int(dur.Hours()), int(dur.Minutes())%60)
	} else {
		if dur.Seconds() > 60 {
			uptime = fmt.Sprintf("%dm %ds", int(dur.Minutes()), int(dur.Seconds())%60)
		} else {
			uptime = fmt.Sprintf("%ds", int(dur.Seconds()))
		}
	}
	w.Uptime = uptime

	if pubIPs, err := netutil.GetPublicIPs(); err == nil && len(pubIPs) > 0 {
		w.IP = pubIPs[0].String()
	} else {
		if ips, err := netutil.GetAvailableIPs(); err == nil && len(ips) > 0 {
			w.IP = ips[0].String()
		}
	}

	income := big.NewInt(0)
	ramUsed := uint64(0)
	storageUsed := uint64(0)
	netInUsed := uint64(0)
	netOutUsed := uint64(0)
	gpuUsed := uint64(0)
	cpuUsed := uint64(0)

	for _, plan := range askPlans.GetAskPlans() {
		income = big.NewInt(0).Add(income, plan.GetPrice().PerSecond.Unwrap())
		ramUsed += plan.GetResources().GetRAM().GetSize().Bytes
		storageUsed += plan.GetResources().GetStorage().GetSize().Bytes
		netInUsed += plan.GetResources().GetNetwork().GetThroughputIn().GetBitsPerSecond()
		netOutUsed += plan.GetResources().GetNetwork().GetThroughputOut().GetBitsPerSecond()
		cpuUsed += plan.GetResources().GetCPU().GetCorePercents()
		gpuUsed += uint64(len(plan.GetResources().GetGPU().GetHashes()))
	}

	incomeHour := big.NewInt(0).Mul(income, big.NewInt(3600))
	w.Income = sonm.NewBigInt(incomeHour).ToPriceString()

	w.Sold.RAM = ramUsed * 100 / devices.GetRAM().GetDevice().GetTotal()
	w.Sold.Storage = storageUsed * 100 / devices.GetStorage().GetDevice().GetBytesAvailable()
	w.Sold.NetIn = netInUsed * 100 / devices.GetNetwork().GetIn()
	w.Sold.NetOut = netOutUsed * 100 / devices.GetNetwork().GetOut()
	w.Sold.CPUS = cpuUsed * 100 / uint64(devices.GetCPU().GetDevice().GetCores()*100)
	if len(devices.GetGPUs()) > 0 {
		w.Sold.GPUS = gpuUsed * 100 / uint64(len(devices.GetGPUs()))
	}

	return nil
}

func newClient(ctx context.Context, key *ecdsa.PrivateKey) (*grpc.ClientConn, error) {
	_, TLSConfig, err := util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %v", err)
	}

	creds := auth.NewWalletAuthenticator(util.NewTLS(TLSConfig), crypto.PubkeyToAddress(key.PublicKey))
	return xgrpc.NewClient(ctx, "127.0.0.1:15030", creds)
}

type displayCtl struct {
	X   int32
	Y   int32
	Rat float64

	surface    *sdl.Surface
	background *sdl.Surface

	window *sdl.Window
	png    *sdl.RWops
	font   *ttf.Font
	xorg   *exec.Cmd
}

func (ctl *displayCtl) drawText(x, y int32, text string, color sdl.Color) error {
	if len(text) == 0 {
		return nil
	}

	solid, err := ctl.font.RenderUTF8Solid(text, color)
	if err != nil {
		return err
	}
	defer solid.Free()

	var dstrect sdl.Rect
	if x < 0 {
		dstrect = sdl.Rect{
			X: int32(float64(-x)*ctl.Rat) - solid.W/2 - ctl.X,
			Y: int32(float64(y)*ctl.Rat) - ctl.Y,
			W: int32(float64(solid.W)),
			H: int32(float64(solid.H)),
		}
	} else {
		dstrect = sdl.Rect{
			X: int32(float64(x)*ctl.Rat) - ctl.X,
			Y: int32(float64(y)*ctl.Rat) - ctl.Y,
			W: int32(float64(solid.W)),
			H: int32(float64(solid.H)),
		}
	}

	return solid.BlitScaled(nil, ctl.surface, &dstrect)
}

func (ctl *displayCtl) Close() error {
	if ctl.window != nil {
		ctl.window.Destroy()
	}

	if ctl.png != nil {
		ctl.png.Close()
	}

	if ctl.font != nil {
		ctl.font.Close()
	}

	ttf.Quit()
	sdl.Quit()

	if ctl.xorg != nil {
		ctl.xorg.Process.Kill()
	}

	return nil
}

func initGraphics(log *zap.Logger) (*displayCtl, error) {
	ctl := &displayCtl{}

	if os.Getenv("DISPLAY") == "" {
		ctl.xorg = exec.Command("Xorg", ":0", "vt7", "-noreset", "-br", "-quiet")
		err := ctl.xorg.Start()
		if err != nil {
			return ctl, fmt.Errorf("failed to start Xord: %v", err)
		}

		time.Sleep(2 * time.Second)
		os.Setenv("DISPLAY", ":0")
		log.Info("X server started")
	}

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return ctl, fmt.Errorf("failed to init SDL: %v", err)
	}

	if err := ttf.Init(); err != nil {
		return ctl, fmt.Errorf("failed to init fonts: %v", err)
	}

	mode, err := sdl.GetDesktopDisplayMode(0)
	if err != nil {
		return ctl, fmt.Errorf("failed to obtain display mode: %v", err)
	}

	window, err := sdl.CreateWindow("SONM Status", 0, 0, mode.W, mode.H, sdl.WINDOW_SHOWN)
	if err != nil {
		return ctl, fmt.Errorf("failed to create window: %v", err)
	}
	ctl.window = window

	_, err = sdl.ShowCursor(0)
	if err != nil {
		log.Info("failed to hide cursor", zap.Error(err))
	}

	surface, err := window.GetSurface()
	if err != nil {
		return ctl, fmt.Errorf("failed to get window surface: %v", err)
	}
	ctl.surface = surface

	if err := surface.FillRect(nil, 0); err != nil {
		log.Info("failed to fill surface", zap.Error(err))
	}

	ctl.png = sdl.RWFromFile("image.png", "rb")
	ctl.background, err = img.LoadPNGRW(ctl.png)
	if err != nil {
		return ctl, fmt.Errorf("failed to load background image: %v", err)
	}

	rat0 := float64(ctl.background.H) / float64(ctl.background.W)
	rat1 := float64(mode.H) / float64(mode.W)
	ratInv := 1.0

	if rat0 > rat1 {
		ratInv = float64(ctl.background.H) / float64(mode.H)
	} else {
		ratInv = float64(ctl.background.W) / float64(mode.W)
	}

	rat := 1.0 / (math.Floor(ratInv*5.0) / 5.0)
	x := int32((float64(ctl.background.W)*rat - float64(mode.W)) / 2)
	y := int32((float64(ctl.background.H)*rat - float64(mode.H)) / 2)

	ctl.Rat = rat
	ctl.X = x
	ctl.Y = y

	font, err := ttf.OpenFont("TerminusTTFWindows-4.46.0.ttf", int(34*rat))
	if err != nil {
		return ctl, fmt.Errorf("failed to open font: %v", err)
	}
	ctl.font = font

	return ctl, nil
}

func main() {
	log, _ := zap.NewDevelopment()
	log.Info("starting sonmmon service")

	cfgPath, err := guessConfigPath(log)
	if err != nil {
		log.Error("failed to guess find proper config path", zap.Error(err))
		return
	}

	log.Sugar().Debugf("using `%s` as config path", cfgPath)
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Error("failed to load config", zap.Error(err))
		return
	}

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		log.Error("failed to open keystore", zap.Error(err))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc, err := newClient(ctx, key)
	if err != nil {
		log.Error("failed to create client connection", zap.Error(err))
		return
	}

	worker := NewWorkerStatus()

	ctl, err := initGraphics(log)
	defer ctl.Close()
	if err != nil {
		log.Error("failed to init graphics", zap.Error(err))
		return
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return cmd.WaitInterrupted(ctx)
	})

	for {
		if ctx.Err() != nil {
			break
		}

		if err := ctl.background.BlitScaled(nil, ctl.surface, &sdl.Rect{
			X: -ctl.X,
			Y: -ctl.Y,
			W: int32(float64(ctl.background.W) * ctl.Rat),
			H: int32(float64(ctl.background.H) * ctl.Rat),
		}); err != nil {
			log.Info("failed to BlitScaled", zap.Error(err))
			continue
		}
		ctl.window.UpdateSurface()

		black := sdl.Color{0, 0, 0, 255}
		white := sdl.Color{255, 255, 255, 255}
		red := sdl.Color{255, 0, 0, 255}

		if worker.Success {
			ctl.drawText(-1020, 190, worker.Status, white)
		} else {
			ctl.drawText(-1020, 190, "UNABLE TO GET WORKER STATUS", red)
			if !worker.LastCheck.IsZero() {
				builder := strings.Builder{}
				builder.WriteString("STATUS RECEIVED LAST ")
				builder.WriteString(time.Now().Sub(worker.LastCheck).Truncate(1 * time.Second).String())
				builder.WriteString(" AGO.")
				ctl.drawText(-1020, 190+40, builder.String(), red)
				ctl.drawText(-1020, 190+40*2, "DISPLAYING LATEST RETRIEVED VALUES", red)
			}
		}

		ctl.drawText(273, 421, "Worker address", black)
		if worker.WorkerName != "" {
			ctl.drawText(273, 469, fmt.Sprintf("%s (%s)", worker.Worker, worker.WorkerName), black)
		} else {
			ctl.drawText(273, 469, worker.Worker, black)
		}

		ctl.drawText(273, 589, "Master address", black)
		ctl.drawText(273, 636, worker.Master, black)
		ctl.drawText(220, 938, worker.IP, white)
		ctl.drawText(980, 938, worker.Uptime, white)
		ctl.drawText(1375, 938, fmt.Sprintf("%s USD/h", worker.Income), white)
		ctl.drawText(120, 1100, "Resources sold", black)
		ctl.drawText(120, 1150, fmt.Sprintf("%9s %9s %9s %9s %9s %9s", "GPU", "CPU", "RAM", "Storage", "Net In", "Net Out"), black)
		ctl.drawText(120, 1200,
			fmt.Sprintf("%8d%% %8d%% %8d%% %8d%% %8d%% %8d%%",
				worker.Sold.GPUS,
				worker.Sold.CPUS,
				worker.Sold.RAM,
				worker.Sold.Storage,
				worker.Sold.NetIn,
				worker.Sold.NetOut,
			),
			black)

		if err := ctl.window.UpdateSurface(); err != nil {
			log.Info("failed to update surface", zap.Error(err))
			continue
		}

		for {
			if ctx.Err() != nil {
				break
			}

			// todo: maybe move to goroutine and wait for signal asynchronously?
			event := sdl.WaitEventTimeout(5000)
			if event == nil {
				break
			}

			switch t := event.(type) {
			case *sdl.KeyboardEvent:
				// handle alt+F1
				if t.State == 0 && t.Keysym.Sym == 0x4000003a && t.Keysym.Mod == 256 {
					log.Info("detected tty switch event")
					tty, err := os.OpenFile("/dev/tty7", syscall.O_RDWR, 0666)
					if err != nil {
						log.Info("failed to open tty", zap.Error(err))
						continue
					}

					// switch to TTY1
					syscall.Syscall(syscall.SYS_IOCTL, tty.Fd(), 0x5606, 1)
					tty.Close()
				}
			case *sdl.QuitEvent:
				log.Info("quit event received")
				return
			}
		}

		if err := worker.update(ctx, cc, cfg.WorkerAddr); err != nil {
			log.Warn("failed to update worker status", zap.Error(err))
			continue
		}
	}

	wg.Wait()
	log.Info("exiting")
}

// vim: ai:ts=8:sw=8:noet:syntax=go
