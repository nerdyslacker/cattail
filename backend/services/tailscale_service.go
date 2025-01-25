package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cattail/backend/types"
	tsutils "cattail/backend/utils/ts"

	"github.com/dgrr/tl"
	"github.com/energye/systray"
	"github.com/gen2brain/beeep"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.design/x/clipboard"
	"tailscale.com/client/tailscale"
	"tailscale.com/client/tailscale/apitype"
	"tailscale.com/cmd/tailscale/cli"
	"tailscale.com/ipn"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/net/tsaddr"
	"tailscale.com/tailcfg"
	"tailscale.com/types/key"
)

type tailScaleService struct {
	ctx           context.Context
	client        tailscale.LocalClient
	fileMod       chan struct{}
	initClipboard sync.Once
	traySvc       *trayService
}

func TailScaleService() *tailScaleService {
	svc := &tailScaleService{}
	return svc
}

var iconPath = func() string {
	_, err := os.Stat("../../frontend/src/assets/images/icon.png")
	if err == nil {
		return "../../frontend/src/assets/images/icon.png"
	} else {
		home, _ := os.UserHomeDir()
		alterPath := filepath.Join(home, ".local", "share", "icons", "hicolor", "256x256", "apps", "com.cattail.png")

		_, err := os.Stat(alterPath)
		if err == nil {
			return alterPath
		}

		return ""
	}
}()

func Notify(format string, args ...interface{}) {
	beeep.Notify("Cattail", fmt.Sprintf(format, args...), iconPath)
}

func (tailSvc *tailScaleService) Startup(ctx context.Context) {
	tailSvc.ctx = ctx
	tailSvc.fileMod = make(chan struct{}, 1)

	Notify("Tailscale started")

	go tailSvc.watchFiles()
	go tailSvc.watchIPN()
	// go tailSvc.pingPeers()

	runtime.EventsOn(tailSvc.ctx, "file_upload", func(data ...interface{}) {
		fmt.Println(data)
	})

	// Init tray
	go tailSvc.traySvc.Start(func() { tailSvc.initTray() })

	// Refresh status
	tailSvc.Refresh()
}

func (tailSvc *tailScaleService) OnSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	secondInstanceArgs := secondInstanceData.Args

	runtime.WindowUnminimise(tailSvc.ctx)
	runtime.Show(tailSvc.ctx)
	go runtime.EventsEmit(tailSvc.ctx, "launchArgs", secondInstanceArgs)
}

func (tailSvc *tailScaleService) initTray() {
	online := tailSvc.GetStatus()

	if tailSvc.traySvc == nil {
		tailSvc.traySvc = TrayService(online)
	}

	tailSvc.setTrayActions()
}

func (tailSvc *tailScaleService) setTrayActions() {
	ts := tailSvc.traySvc
	ctx := tailSvc.ctx
	online := tailSvc.GetStatus()

	ts.ToggleStatusItem(online)

	ts.statusMenuItem.Click(func() {
		if ts.statusMenuItem.Checked() {
			ts.statusMenuItem.Uncheck()
			ts.statusMenuItem.SetTitle("Start")
			tailSvc.Stop()
		} else {
			ts.statusMenuItem.Check()
			ts.statusMenuItem.SetTitle("Stop")
			tailSvc.Start()
		}
	})

	ts.showMenuItem.Click(func() {
		runtime.WindowShow(ctx)
		ts.isWindowHidden = false
	})

	ts.quitMenuItem.Click(func() {
		ts.Stop()
		runtime.Quit(ctx)
	})

	systray.SetOnClick(func(menu systray.IMenu) {
		if ts.isWindowHidden {
			runtime.WindowShow(ctx)
			ts.isWindowHidden = false
		} else {
			runtime.WindowHide(ctx)
			ts.isWindowHidden = true
		}
	})
}

func (tailSvc *tailScaleService) PingPeers() {
	for {
		status, err := tailSvc.client.Status(tailSvc.ctx)
		if err != nil {
			log.Println("Getting client status", err)
			time.Sleep(time.Second * 10)
			continue
		}

		for _, nodeKey := range status.Peers() {
			peer := status.Peer[nodeKey]
			if len(peer.TailscaleIPs) == 0 {
				log.Printf("Peer %s doesn't have any IPs", peer.DNSName)
				continue
			}

			log.Printf("Pinging %s", peer.TailscaleIPs[0])

			ctx, cancelFn := context.WithCancel(tailSvc.ctx)
			done := make(chan struct{}, 1)

			go func() {
				select {
				case <-done:
				case <-time.After(time.Second * 5):
					cancelFn()
				}
			}()

			res, err := tailSvc.client.Ping(ctx, peer.TailscaleIPs[0], tailcfg.PingICMP)
			if err != nil {
				log.Printf("Unable to ping %s: %s\n", peer.TailscaleIPs[0], err)
			}

			done <- struct{}{}

			log.Println("Ping result", res)
		}

		time.Sleep(time.Second * 30)
	}
}

func (tailSvc *tailScaleService) UploadFile(dnsName string) {
	status, err := tailSvc.client.Status(tailSvc.ctx)
	if err != nil {
		panic(err)
	}

	peers := status.Peers()

	i := tl.SearchFn(peers, func(nodeKey key.NodePublic) bool {
		peer := status.Peer[nodeKey]
		return peer.DNSName == dnsName
	})
	if i == -1 {
		return
	}

	peer := status.Peer[peers[i]]

	filename, err := runtime.OpenFileDialog(tailSvc.ctx, runtime.OpenDialogOptions{
		DefaultDirectory: func() string {
			dir, _ := os.UserHomeDir()
			return dir
		}(),
	})
	if err != nil {
		panic(err)
	}

	if len(filename) == 0 {
		return
	}

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	stat, _ := file.Stat()

	err = tailSvc.client.PushFile(tailSvc.ctx, peer.ID, stat.Size(), stat.Name(), file)
	if err != nil {
		log.Printf("error uploading file to %s: %s\n", dnsName, err)
	}

	Notify("File %s sent to %s", stat.Name(), dnsName)
}

func (tailSvc *tailScaleService) AcceptFile(filename string) {
	dir, err := runtime.OpenDirectoryDialog(tailSvc.ctx, runtime.OpenDialogOptions{
		DefaultDirectory: func() string {
			dir, _ := os.UserHomeDir()
			return dir
		}(),
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		tailSvc.RemoveFile(filename)
	}()

	r, _, err := tailSvc.client.GetWaitingFile(tailSvc.ctx, filename)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	dstPath := filepath.Join(dir, filename)
	file, err := os.Create(dstPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, _ = io.Copy(file, r)

	Notify("Downloaded %s to %s", filename, dstPath)
}

func (tailSvc *tailScaleService) RemoveFile(filename string) {
	log.Printf("Removing file %s\n", filename)

	err := tailSvc.client.DeleteWaitingFile(tailSvc.ctx, filename)
	if err != nil {
		log.Printf("Removing file: %s: %s\n", filename, err)
	}

	tailSvc.fileMod <- struct{}{}
}

func (tailSvc *tailScaleService) CurrentAccount() string {
	current, _, err := tailSvc.client.ProfileStatus(tailSvc.ctx)
	if err != nil {
		panic(err)
	}

	return current.Name
}

func (tailSvc *tailScaleService) SetExitNode(dnsName string) {
	status, err := tailSvc.client.Status(tailSvc.ctx)
	if err != nil {
		panic(err)
	}

	peers := status.Peers()

	i := tl.SearchFn(peers, func(nodeKey key.NodePublic) bool {
		peer := status.Peer[nodeKey]
		return peer.DNSName == dnsName
	})
	if i == -1 {
		return
	}

	peer := status.Peer[peers[i]]

	prefs := &ipn.MaskedPrefs{
		Prefs:         ipn.Prefs{},
		ExitNodeIPSet: true,
		ExitNodeIDSet: true,
	}

	if !peer.ExitNode {
		success := false
		ipsToTry := []string{
			peer.DNSName,
			peer.HostName,
		}

		for _, ip := range peer.TailscaleIPs {
			ipsToTry = append(ipsToTry, ip.String())
		}

		for _, host := range ipsToTry {
			log.Printf("Exit node as %s\n", host)

			err = prefs.SetExitNodeIP(host, status)
			if err != nil {
				log.Printf("Setting exit node as %s: %s\n", host, err)
				continue
			}

			success = true
			break
		}

		if !success {
			runtime.EventsEmit(tailSvc.ctx, "exit_node_connect")
			return
		}
	}

	_, err = tailSvc.client.EditPrefs(tailSvc.ctx, prefs)
	if err != nil {
		log.Println(err)
	}

	runtime.EventsEmit(tailSvc.ctx, "exit_node_connect")

	if peer.ExitNode {
		Notify("Removed exit node %s", peer.DNSName)
	} else {
		Notify("Using %s as exit node", peer.DNSName)
	}
}

func (tailSvc *tailScaleService) AdvertiseExitNode(dnsName string) {
	status, err := tailSvc.client.Status(tailSvc.ctx)
	if err != nil {
		panic(err)
	}

	if status.Self.DNSName != dnsName {
		return
	}

	curPrefs, err := tailSvc.client.GetPrefs(tailSvc.ctx)
	if err != nil {
		panic(err)
	}

	isAdvertise := curPrefs.AdvertisesExitNode()

	prefs := &ipn.MaskedPrefs{
		Prefs: ipn.Prefs{
			AdvertiseRoutes: append([]netip.Prefix{},
				tsaddr.AllIPv4(), tsaddr.AllIPv4(),
			),
		},
		AdvertiseRoutesSet: true,
	}

	prefs.SetAdvertiseExitNode(!isAdvertise)

	// if current settings is advertise, then remove
	if isAdvertise {
		prefs.Prefs.AdvertiseRoutes = nil
	}

	_, err = tailSvc.client.EditPrefs(tailSvc.ctx, prefs)
	if err != nil {
		log.Println(err)
	}

	runtime.EventsEmit(tailSvc.ctx, "advertise_exit_node_done")

	if isAdvertise {
		Notify("Removed advertising node")
	} else {
		Notify("Advertising as exit node")
	}
}

func (tailSvc *tailScaleService) AdvertiseRoutes(routes string) error {
	curPrefs, err := tailSvc.client.GetPrefs(tailSvc.ctx)

	if err != nil {
		panic(err)
	}

	exit := curPrefs.AdvertisesExitNode()

	
	if strings.TrimSpace(routes) == "" {
		curPrefs.AdvertiseRoutes = nil
	} else {
		ipStrings := strings.Split(routes, ",")
		var prefixes []netip.Prefix
		for _, ipStr := range ipStrings {
			ipStr = strings.TrimSpace(ipStr)
			prefix, err := netip.ParsePrefix(ipStr)
			if err != nil {
				log.Println(err)
				return nil
			}
			prefixes = append(prefixes, prefix)
		}

		curPrefs.AdvertiseRoutes = prefixes
	}

	curPrefs.SetAdvertiseExitNode(exit)

	_, err = tailSvc.client.EditPrefs(tailSvc.ctx, &ipn.MaskedPrefs{
		Prefs:              *curPrefs,
		AdvertiseRoutesSet: true,
	})

	if err != nil {
		log.Println(err)
	}

	return nil
}

func (tailSvc *tailScaleService) AllowLANAccess(allow bool) error {
	prefs := ipn.Prefs{
		ExitNodeAllowLANAccess: allow,
	}

	_, err := tailSvc.client.EditPrefs(tailSvc.ctx, &ipn.MaskedPrefs{
		Prefs:                     prefs,
		ExitNodeAllowLANAccessSet: true,
	})

	if err != nil {
		log.Println(err)
	}

	if allow {
		Notify("LAN access has been granted.")
	} else {
		Notify("LAN access has been restricted.")
	}

	return nil
}

func (tailSvc *tailScaleService) AcceptRoutes(accept bool) error {
	prefs := ipn.Prefs{
		RouteAll: accept,
	}

	_, err := tailSvc.client.EditPrefs(tailSvc.ctx, &ipn.MaskedPrefs{
		Prefs:       prefs,
		RouteAllSet: true,
	})

	if err != nil {
		log.Println(err)
	}

	if accept {
		Notify("All routes acceptance is enabled.")
	} else {
		Notify("All routes acceptance is disabled.")
	}

	return nil
}

func (tailSvc *tailScaleService) RunSSH(run bool) error {
	prefs := ipn.Prefs{
		RunSSH: run,
	}

	_, err := tailSvc.client.EditPrefs(tailSvc.ctx, &ipn.MaskedPrefs{
		Prefs:     prefs,
		RunSSHSet: true,
	})

	if err != nil {
		log.Println(err)
	}

	if run {
		Notify("SSH access has been enabled.")
	} else {
		Notify("SSH access has been disabled.")
	}

	return nil
}

func (tailSvc *tailScaleService) SetControlURL(controlURL string) error {
	curPrefs, err := tailSvc.client.GetPrefs(tailSvc.ctx)

	if err != nil {
		panic(err)
	}

	curPrefs.ControlURL = controlURL

	err = tailSvc.client.Start(tailSvc.ctx, ipn.Options{
		UpdatePrefs: curPrefs,
	})

	if err != nil {
		log.Println(err)
	}

	return nil
}

func (tailSvc *tailScaleService) CopyClipboard(s string) {
	tailSvc.initClipboard.Do(func() {
		if err := clipboard.Init(); err != nil {
			panic(err)
		}
	})
	log.Printf("Copying \"%s\" to the clipboard\n", s)
	clipboard.Write(clipboard.FmtText, []byte(s))
}

func (tailSvc *tailScaleService) Accounts() []string {
	current, all, err := tailSvc.client.ProfileStatus(tailSvc.ctx)
	if err != nil {
		panic(err)
	}

	names := tl.Filter(
		tl.Map(all, func(profile ipn.LoginProfile) string {
			return profile.Name
		}),
		func(name string) bool {
			return name != current.Name
		},
	)

	return names
}

func (tailSvc *tailScaleService) Self() types.Peer {
	log.Printf("Requesting self")

	status, err := tailSvc.client.Status(tailSvc.ctx)
	if err != nil {
		log.Printf("Requesting self: %s\n", err)
		return types.Peer{}
	}

	curPrefs, err := tailSvc.client.GetPrefs(tailSvc.ctx)
	if err != nil {
		panic(err)
	}

	self := status.Self
	peer := convertPeer(self, curPrefs)

	peer.ExitNodeOption = curPrefs.AdvertisesExitNode()
	peer.AllowLANAccess = curPrefs.ExitNodeAllowLANAccess
	peer.AcceptRoutes = curPrefs.RouteAll
	peer.RunSSH = curPrefs.RunSSH

	return peer
}

func (tailSvc *tailScaleService) Files() []types.File {
	files, err := tailSvc.client.AwaitWaitingFiles(tailSvc.ctx, time.Second)
	if err != nil {
		log.Println(err)
		return nil
	}

	return tl.Map(files, func(file apitype.WaitingFile) types.File {
		return types.File{
			Name: file.Name,
			Size: file.Size,
		}
	})
}

func (tailSvc *tailScaleService) Namespaces() []types.Namespace {
	status, err := tailSvc.client.Status(tailSvc.ctx)
	if err != nil {
		log.Printf("requesting instance: %s\n", err)
		return nil
	}

	curPrefs, err := tailSvc.client.GetPrefs(tailSvc.ctx)

	if err != nil {
		panic(err)
	}

	res := make([]types.Namespace, 0)

	for _, nodeKey := range status.Peers() {
		tsPeer := status.Peer[nodeKey]
		_, namespace := splitPeerNamespace(tsPeer.DNSName)

		peer := convertPeer(tsPeer, curPrefs)

		i := tl.SearchFn(res, func(a types.Namespace) bool {
			return namespace == a.Name
		})
		if i == -1 {
			res = append(res, types.Namespace{
				Name: namespace,
				Peers: []types.Peer{
					peer,
				},
			})
		} else {
			res[i].Peers = append(res[i].Peers, peer)
		}
	}

	return res
}

func (tailSvc *tailScaleService) SwitchTo(account string) {
	current, all, err := tailSvc.client.ProfileStatus(tailSvc.ctx)
	if err != nil {
		panic(err)
	}

	if account == current.Name {
		return
	}

	all = tl.Filter(all, func(profile ipn.LoginProfile) bool {
		return profile.Name == account
	})
	if len(all) == 0 {
		log.Printf("Profile %s not found\n", account)
		return
	}

	log.Printf("Profile %s", all[0].ID)
	tailSvc.client.SwitchProfile(tailSvc.ctx, all[0].ID)

	Notify("Switched to account: %s", account)
}

func (tailSvc *tailScaleService) GetStatus() bool {
	st, err := tailSvc.client.Status(tailSvc.ctx)

	if err != nil {
		return false
	}

	status := &tsutils.Status{
		Status: st,
	}
	online := status.Online()

	return online
}

func (tailSvc *tailScaleService) UpdateStatus(previousOnlineStatus bool) bool {
	if tailSvc == nil {
		return false
	}

	if tailSvc.traySvc == nil {
		return false
	}

	online := tailSvc.GetStatus()

	if online != previousOnlineStatus {
		tailSvc.traySvc.setStatus(online)
		runtime.EventsEmit(tailSvc.ctx, "tailscale:status-changed", online)
	}

	tailSvc.traySvc.ToggleStatusItem(online)

	return online
}

func (tailSvc *tailScaleService) Start() error {
	st, err := tailSvc.client.Status(tailSvc.ctx)

	if err != nil {
		return err
	}

	status := &tsutils.Status{
		Status: st,
	}

	if status.NeedsLogin() {
		result, err := runtime.MessageDialog(tailSvc.ctx, runtime.MessageDialogOptions{
			Type:          runtime.QuestionDialog,
			Title:         "Login Required",
			Message:       "Open a browser to authenticate with Tailscale?",
			DefaultButton: "Yes",
			CancelButton:  "No",
		})

		if err != nil {
			return err
		}

		if result == "Yes" {
			runtime.BrowserOpenURL(tailSvc.ctx, status.Status.AuthURL)
		}

		return nil
	}

	Notify("Tailscale started")
	return cli.Run([]string{"up"})
}

func (tailSvc *tailScaleService) Stop() error {
	Notify("Tailscale stopped")
	return cli.Run([]string{"down"})
}

func (tailSvc *tailScaleService) Refresh() {
	var previousStatus bool

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				currentStatus := tailSvc.UpdateStatus(previousStatus)
				previousStatus = currentStatus
				runtime.EventsOn(tailSvc.ctx, "wails:window:hide", func(optionalData ...interface{}) {
					tailSvc.traySvc.isWindowHidden = true
				})
			case <-tailSvc.ctx.Done():
				return
			}
		}
	}()
}

func (tailSvc *tailScaleService) watchFiles() {
	prevFiles := 0

	for {
		select {
		case <-time.After(time.Second * 10):
		case <-tailSvc.fileMod:
		}

		files, err := tailSvc.client.AwaitWaitingFiles(tailSvc.ctx, time.Second)
		if err != nil {
			log.Println(err)
		}

		if len(files) != prevFiles {
			prevFiles = len(files)

			for _, file := range files {
				Notify("File %s available", file.Name)
			}

			runtime.EventsEmit(tailSvc.ctx, "update_files")
		}
	}
}

func (tailSvc *tailScaleService) watchIPN() {
	for {
		watcher, err := tailSvc.client.WatchIPNBus(tailSvc.ctx, 0)
		if err != nil {
			log.Printf("loading IPN bus watcher: %s\n", err)
			time.Sleep(time.Second)
			continue
		}

		for {
			not, err := watcher.Next()
			if err != nil {
				log.Printf("Watching IPN Bus: %s\n", err)
				break
			}

			if not.FilesWaiting != nil {	
				tailSvc.fileMod <- struct{}{}
			}

			if not.State != nil {
				if *not.State == ipn.Running {
					runtime.EventsEmit(tailSvc.ctx, "app_running")
				} else {
					runtime.EventsEmit(tailSvc.ctx, "app_not_running")
				}
			}

			runtime.EventsEmit(tailSvc.ctx, "update_all")

			log.Printf("IPN bus update: %v\n", not)
		}
	}
}

func convertPeer(status *ipnstate.PeerStatus, prefs *ipn.Prefs) types.Peer {
	peerName, _ := splitPeerNamespace(status.DNSName)
	return types.Peer{
		ID:             string(status.ID),
		DNSName:        status.DNSName,
		Name:           peerName,
		ExitNode:       status.ExitNode,
		ExitNodeOption: status.ExitNodeOption,
		Online:         status.Online,
		OS:             status.OS,
		Addrs:          status.Addrs,
		Created:        status.Created,
		LastSeen:       status.LastSeen,
		LastWrite:      status.LastWrite,
		Routes: func() []string {
			if status.PrimaryRoutes == nil {
				return nil
			}

			return tl.Map(status.PrimaryRoutes.AsSlice(), func(prefix netip.Prefix) string {
				return prefix.String()
			})
		}(),
		IPs: tl.Map(status.TailscaleIPs, func(ip netip.Addr) string {
			return ip.String()
		}),
		AllowedIPs: func() []string {
			if status.AllowedIPs == nil {
				return nil
			}

			return tl.Map(status.AllowedIPs.AsSlice(), func(prefix netip.Prefix) string {
				return prefix.String()
			})
		}(),
		AdvertisedRoutes: func() []string {
			if prefs.AdvertiseRoutes == nil {
				return nil
			}

			return tl.Map(prefs.AdvertiseRoutes, func(prefix netip.Prefix) string {
				return prefix.String()
			})
		}(),
	}
}

func splitPeerNamespace(dnsName string) (peerName, namespace string) {
	names := strings.Split(dnsName, ".")
	namespace = strings.Join(names[1:], ".")
	peerName = names[0]
	return peerName, namespace
}
