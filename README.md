# Cattail

Cattail is an unofficial tailscale/headscale client using [Wails](https://wails.io) (Go + Vue3).

<div style="text-align: center;">
    <img src="frontend/src/assets/images/logo.png" />
</div>

# Installation

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
git clone https://github.com/nerdyslacker/cattail
cd cattail
make install
```

or download the binary from [Releases](https://github.com/nerdyslacker/cattail/releases) page.

# Features

- [x] Account switching
- [x] Detailed peer information
- [x] Tray menu for quick access
- [x] Copying of IP addresses/DNS name
- [ ] Pinging of peers
- [ ] Set control URL
- [ ] Adding tags
- [x] Exit node management
- [x] Allow LAN access
- [x] Accept routes
- [x] Run SSH
- [ ] Advertise routes
- [x] Toggle tailscale status
- [ ] Toggle taildrop status and change path
- [x] Sending files
- [x] Receiving files
- [ ] Notification on tailscale status change
- [ ] Notification on peer addition/removal
- [ ] Monitoring traffic

<div style="text-align: center;">
    <img src="_images/screenshot.png" />
</div>

# Credits 

* [dgrr/tailscale-client](https://github.com/dgrr/tailscale-client)
* [DeedleFake/trayscale](https://github.com/DeedleFake/trayscale)
* [tiny-craft/tiny-rdm](https://github.com/tiny-craft/tiny-rdm)
* [KSurzyn](https://github.com/KSurzyn) (for logo)