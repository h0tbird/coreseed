## Burning iPXE into Intel ROM

You can use `lspci -nn` to find the relevant IDs of your network card:
```
lspci -nn | grep 'Ethernet controller'
00:19.0 Ethernet controller [0200]: Intel Corporation 82579LM Gigabit Network Connection [8086:1502]
02:00.0 Ethernet controller [0200]: Intel Corporation 82574L Gigabit Network Connection [8086:10d3]
```

Build the iPXE ROM:
```
git clone git://git.ipxe.org/ipxe.git && cd ipxe/src
make bin/808610d3.rom EMBED=/path/to/menu.ipxe
cp bin/808610d3.rom /media/pendrive
```

Download `BootUtil.exe`:
```
wget http://downloadmirror.intel.com/19186/eng/PREBOOT.EXE
file-roller -e /tmp PREBOOT.EXE
cp /tmp/APPS/BootUtil/DOS/BootUtil.exe /media/pendrive
```
