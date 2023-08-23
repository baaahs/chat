# Camp Server

This describes the setup of the camp server which runs a mqtt server along with the UI for displaying the sheep location. It is rather similar to the chromebook setup, but starts with a RPI 4 instead of a chromebook.

## Prerequisite setup

Start with a Raspberry Pi 4 Model B and a 32GB micro SD card rathed V30, HC, A1 or better. I'm using a SanDisk Extreme card. If you have previously used the SD card you may need to first initialize it so macOS (or windows) is happy before you can use the RPI installer. I initialized it as ExFat since this is going to get overwritten by the installer anyway.

It's worth noting at this point that you probably want to know the information from this tutorial about setting up a RPI in "kiosk mode". More about this later on. <https://www.raspberrypi.com/tutorials/how-to-use-a-raspberry-pi-in-kiosk-mode/>

Install the official RPI imager from <https://www.raspberrypi.com/software/>

Run the imager software and select the Raspberry Pi OS (64-bit) image, which at the time of this writing is a port of Debian Bullseye and is compatible is Raspberry Pi 3/4/400 released 2023-05-03. Later versions are probably fine, this is just what is current in August of 2023.

Before flashing, customize the image.

   * Set a hostname of camp-console.local
   * Enable SSH using password authentication
   * Set a username and password of baaahs/notsecure
   * Wireless LAN does not need to be configured (although probably yes, see below)
   * Set the locale to "America/Denver" and keyboard layout of "us"
   * Play sound and eject media as you wish
   * **Disable** telemetry (this thing is going to be in the dessert)

**Note:** To get this application to work I had to open System Preferences, go to "Privacy and Security" and then manually add "Full Disk Access" for the Raspberry Pi Imager. This means it has both "Full Disk Access" as well as "Files and Folders". While it still prompts to be allowed this access, this seemed necessary.

These directions may not be entirely sufficient, but the basic idea is you need to get that SD card written with the RPI image. You may need to use more manual ways of doing this.

**WIFI**



---

At this point you should be able to insert the SD card into the RPI, connect the RPI via ethernet, and let the thing boot. Both the red and green indicator lights on the RPI should light. The green one should flash occasionally as it boots. Eventually the network indicator light should start showing network activity.

After sometime you should be able to ping camp-console.local from your desktop machine (everything is assuming you're doing this with a modern mac that does MDNS etc.)

Once you can ping the RPI, you should be able to ssh using the credentials supplied during the creation of the SD card image.

    ssh baaahs@camp-console.local

In the past the first thing you would have wanted to do is run `sudo raspi-config` and select `6 Advanced Options` and `A1 Expand filesystem` to make sure that the filesystem is using the full SD card. If you do this you need to reboot the system with `sudo reboot now` - Note that a reboot may take a minute or so.

This no longer seems necessary though. Here is the session immediately upon first boot including running this expansion manually:

---

    ➜  camp-server git:(master) ✗ ssh baaahs@camp-console.local
    The authenticity of host 'camp-console.local (fe80::9095:20a0:2614:493b%en0)' can't be established.
    ED25519 key fingerprint is SHA256:up3Wy2N71vA+uxCENDpXbiVvUoBgdKZGa6mn6pJtlqc.
    This key is not known by any other names
    Are you sure you want to continue connecting (yes/no/[fingerprint])? yes
    Warning: Permanently added 'camp-console.local' (ED25519) to the list of known hosts.
    baaahs@camp-console.local's password: 
    Linux camp-console 6.1.21-v8+ #1642 SMP PREEMPT Mon Apr  3 17:24:16 BST 2023 aarch64

    The programs included with the Debian GNU/Linux system are free software;
    the exact distribution terms for each program are described in the
    individual files in /usr/share/doc/*/copyright.

    Debian GNU/Linux comes with ABSOLUTELY NO WARRANTY, to the extent
    permitted by applicable law.
    Last login: Tue May  2 21:23:55 2023

    Wi-Fi is currently blocked by rfkill.
    Use raspi-config to set the country before use.

    baaahs@camp-console:~ $ df -h
    Filesystem      Size  Used Avail Use% Mounted on
    /dev/root        29G  3.4G   25G  12% /
    devtmpfs        1.7G     0  1.7G   0% /dev
    tmpfs           1.9G     0  1.9G   0% /dev/shm
    tmpfs           759M  1.2M  758M   1% /run
    tmpfs           5.0M  4.0K  5.0M   1% /run/lock
    /dev/mmcblk0p1  255M   33M  223M  13% /boot
    tmpfs           380M   20K  380M   1% /run/user/1000
    baaahs@camp-console:~ $ mount
    /dev/mmcblk0p2 on / type ext4 (rw,noatime)
    devtmpfs on /dev type devtmpfs (rw,relatime,size=1678476k,nr_inodes=419619,mode=755)
    proc on /proc type proc (rw,relatime)
    sysfs on /sys type sysfs (rw,nosuid,nodev,noexec,relatime)
    securityfs on /sys/kernel/security type securityfs (rw,nosuid,nodev,noexec,relatime)
    tmpfs on /dev/shm type tmpfs (rw,nosuid,nodev)
    devpts on /dev/pts type devpts (rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=000)
    tmpfs on /run type tmpfs (rw,nosuid,nodev,size=777080k,nr_inodes=819200,mode=755)
    tmpfs on /run/lock type tmpfs (rw,nosuid,nodev,noexec,relatime,size=5120k)
    cgroup2 on /sys/fs/cgroup type cgroup2 (rw,nosuid,nodev,noexec,relatime,nsdelegate,memory_recursiveprot)
    pstore on /sys/fs/pstore type pstore (rw,nosuid,nodev,noexec,relatime)
    bpf on /sys/fs/bpf type bpf (rw,nosuid,nodev,noexec,relatime,mode=700)
    systemd-1 on /proc/sys/fs/binfmt_misc type autofs (rw,relatime,fd=29,pgrp=1,timeout=0,minproto=5,maxproto=5,direct)
    debugfs on /sys/kernel/debug type debugfs (rw,nosuid,nodev,noexec,relatime)
    sunrpc on /run/rpc_pipefs type rpc_pipefs (rw,relatime)
    tracefs on /sys/kernel/tracing type tracefs (rw,nosuid,nodev,noexec,relatime)
    mqueue on /dev/mqueue type mqueue (rw,nosuid,nodev,noexec,relatime)
    configfs on /sys/kernel/config type configfs (rw,nosuid,nodev,noexec,relatime)
    fusectl on /sys/fs/fuse/connections type fusectl (rw,nosuid,nodev,noexec,relatime)
    /dev/mmcblk0p1 on /boot type vfat (rw,relatime,fmask=0022,dmask=0022,codepage=437,iocharset=ascii,shortname=mixed,errors=remount-ro)
    tmpfs on /run/user/1000 type tmpfs (rw,nosuid,nodev,relatime,size=388540k,nr_inodes=97135,mode=700,uid=1000,gid=1000)
    baaahs@camp-console:~ $ raspi-conf
    -bash: raspi-conf: command not found
    baaahs@camp-console:~ $ raspi-config
    Script must be run as root. Try 'sudo raspi-config'
    baaahs@camp-console:~ $ sudo -i

    Wi-Fi is currently blocked by rfkill.
    Use raspi-config to set the country before use.

    root@camp-console:~# raspi-config 
    grep: /sys/class/leds/led0/trigger: No such file or directory
    grep: /sys/class/leds/led0/trigger: No such file or directory
    root@camp-console:~# raspi-config 

    Welcome to fdisk (util-linux 2.36.1).
    Changes will remain in memory only, until you decide to write them.
    Be careful before using the write command.


    Command (m for help): Disk /dev/mmcblk0: 29.72 GiB, 31914983424 bytes, 62333952 sectors
    Units: sectors of 1 * 512 = 512 bytes
    Sector size (logical/physical): 512 bytes / 512 bytes
    I/O size (minimum/optimal): 512 bytes / 512 bytes
    Disklabel type: dos
    Disk identifier: 0x1747c873

    Device         Boot  Start      End  Sectors  Size Id Type
    /dev/mmcblk0p1        8192   532479   524288  256M  c W95 FAT32 (LBA)
    /dev/mmcblk0p2      532480 62333951 61801472 29.5G 83 Linux

    Command (m for help): Partition number (1,2, default 2): 
    Partition 2 has been deleted.

    Command (m for help): Partition type
       p   primary (1 primary, 0 extended, 3 free)
       e   extended (container for logical partitions)
    Select (default p): Partition number (2-4, default 2): First sector (2048-62333951, default 2048): Last sector, +/-sectors or +/-size{K,M,G,T,P} (532480-62333951, default 62333951): 
    Created a new partition 2 of type 'Linux' and of size 29.5 GiB.
    Partition #2 contains a ext4 signature.

    Command (m for help): 
    Disk /dev/mmcblk0: 29.72 GiB, 31914983424 bytes, 62333952 sectors
    Units: sectors of 1 * 512 = 512 bytes
    Sector size (logical/physical): 512 bytes / 512 bytes
    I/O size (minimum/optimal): 512 bytes / 512 bytes
    Disklabel type: dos
    Disk identifier: 0x1747c873

    Device         Boot  Start      End  Sectors  Size Id Type
    /dev/mmcblk0p1        8192   532479   524288  256M  c W95 FAT32 (LBA)
    /dev/mmcblk0p2      532480 62333951 61801472 29.5G 83 Linux

    Command (m for help): The partition table has been altered.
    Syncing disks.

    root@camp-console:~#    

----------

And then here is after rebooting

    baaahs@camp-console:~ $ df -h
    Filesystem      Size  Used Avail Use% Mounted on
    /dev/root        29G  3.4G   25G  12% /
    devtmpfs        1.7G     0  1.7G   0% /dev
    tmpfs           1.9G     0  1.9G   0% /dev/shm
    tmpfs           759M  1.2M  758M   1% /run
    tmpfs           5.0M  4.0K  5.0M   1% /run/lock
    /dev/mmcblk0p1  255M   32M  224M  13% /boot
    tmpfs           380M   24K  380M   1% /run/user/1000

Notice that `/dev/root` is still 3.4G used of 29G so it looks like the current disk image handles resizing on first start automatically on it's own.

## WiFi & Calling Home

Since the camp server should wifi via Starlink, let's configure that using `raspi-config` since that wasn't done in the install. Importantly you have to set the world region in "System Settings" / "Wireless LAN" - but also SSID and password of course.

Not putting my credentials in this doc, but suffice it to say I'm configuring the camp-msg server with the credentials for the Starlink unit I sent to playa so maybe it can have both camp network on ethernet and wifi internet directly, with like, no additional setup? Maybe?

I screwed up the first time I tried to set the WiFi region because I fat fingered a menu - I mean come on. Awesome. Anyway, if it's more than the first time you need the same utility `raspi-config` and a different menu path: Localisation Options --> Change Wi-fi Country --> US United States --> OK --> Finish.

Now, a ssh tunnel has been configured as described in [calling home](calling-home.md) so that if this machine ever is able to make it to the regular internet (which it should be able to do via starlink now) then a person can reach back into it via ssh. Neat huh?

## Update

Always good for an update

    sudo -i
    apt update
    apt full-upgrade
    reboot




# Kiosk Mode

This really comes down to using `xdotool` if one wants to send events to X (we large do not? I don't think? Maybe to full screen??) and then `unclutter` to get rid of the mouse.

Before that, let's edit `/etc/xdg/lxsession/LXDE-pi/autostart` so that the only thing in it is the `@pcmanfm` command which provides the desktop file manager. We'll probably nuke this later. Restart the desktop with a `killall lxsession`. 