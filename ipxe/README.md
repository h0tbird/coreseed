## Setup iPXE using a boot USB key

Build the iPXE ROM:
```
git clone git://git.ipxe.org/ipxe.git && cd ipxe/src
make bin/ipxe.usb EMBED=${GOPATH}/src/github.com/katosys/kato/ipxe/menu.ipxe
sudo dd if=bin/ipxe.usb of=/dev/sdX
sudo eject /dev/sdX
```

## Intel AMT

http://node-1:16992 (54:BE:F7:88:25:C5)  
http://node-2:16992 (4C:72:B9:26:25:97)
