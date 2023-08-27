# The Chromebook Girls

We have purchased 3 chromebooks for the initial chat implementation. One is meant as a backup, which is good because one of the 3 that arrived is apparently enterprise locked still and we might not be able to defeat this before playa in 2019.

This directory exists to collect notes about their specific configuration. The goal is to do as much of this remotely via ansible so that it is easily repeatable.

That being said, we need to start with a base. In our case that is getting [GalliumOS](https://galliumos.org) up and running. So far for the two machines that have successfully been able to go into developer mode, I've managed to get Gallium running on at least one of them that I'm using for the test bed.

The [Official Dell Dissassembly Guide for 3120 Chromebooks](https://downloads.dell.com/Manuals/all-products/esuprt_laptop/esuprt_chromebook/chromebook-11-3120_User%27s%20Guide3_en-us.pdf).

After removing the write protect screw from the motherboard (and in our case leaving it out so that we don't have to deal with this in the future), the Mr Chromebox [Firmware Utility Script](https://mrchromebox.tech/#fwscript) can be used to install the *Full ROM Firmware*. The reason for using this one is that it gets rid of all the Chrome stuff at the beginning. Does this mean we can't go back to ChromeOS? Probably. Do we care? Probably not. Becky was a rough girl who's write protect screw was stripped out probably by the refurbishing people, so after grinding at it a little with a dremel it was decided that popping a nearby resistor off the board was a better option. Don't mess with Texas.

Running this script requires a root shell and then executing the super safe, just download this from the internet and execute as root, I'm sure nothing could ever go wrong script like so:

    cd; curl -LO https://mrchromebox.tech/firmware-util.sh && sudo bash firmware-util.sh

After installing the full ROM, the firmware is now a UEFI firmware. So if you previously installed Gallium using the legacy mode, the previous installation won't boot. To avoid this, the recommendation is going to be to:

* Remove the firmware WP screw before anything else
* Get the machine into developer mode where you can get a terminal in ChromeOS
* From that root shell, install the full ROM Mr Chromebox firmware - which moves you to UEFI land
* Now put GalliumOS on a USB drive (I hade to use a USB 2.0 one instead of a 3.0 one). From the UEFI boot Manager you should be able to browse to that drive and find the boot64.efi file which launches the live GalliumOS image
* From this live image, now that your machine is all properly UEFI'd, install GalliumOS

Installing GalliumOS
====================

Create a USB install drive (recommend a USB 2 one not a USB 3 one for compatibility. YMMV).

* Download an appropriate image from the [Gallium Download Page](https://galliumos.org/download) - For the Dell 11 3120's it's the "Bay Trail" image. They are named after Intel processor architectures.
* Boot this live image. You'll end up on a desktop.
* Configure your wifi. In the task bar near the clock there is a double arrow icon. Click that to select your network.
* Click the "Install GalliumOS 3.0" icon on the desktop
* Select English US - don't get fancy!
* Allow it to download updates while installing GalliumOS (the default) and also click the "Install third-party software for graphics and WiFi hardware and additional media formats" - we want all the things.
* For the Installation Type select the automagic "Erase disk and install GalliumOS". Don't enable encryption or LVM. 
    At this point for me it wanted to write the changes to `MMC/SD card #1 (mmcblk0)` and to make two partions on it. Let it.
* You're in Los Angeles (sets the timezone)

On the "Who are you?" screen, here's the recommendation:

*Your name* - leave it blank
*Your computer's name* - be sassy. The first 3 machines are Sally, Becky, and Jennifer. Jennifer doesn't boot right.
*Username* - baaahs
*Password* - notsecure <- or whatever. This is on you. We'll deal with standardizing it later. I didn't use this password.

After this the installation should proceed normally. At the end, reboot. You shouldn't pull out the installation USB drive until it tells you to at the end of the shutdown of the live system.

Upon reboot you should be able to enter your username and password and be back at the Gallium desktop. You will need to configure your wifi again as before. 

While you'll almost certainly be prompted about software updated, don't bother because the ansible scripts should handle those.

*Ok!* We're almost done. Because we want to do the rest of the things remotely we need sshd running so open a shell and:

    sudo apt-get install openssh-server

At this point you should be able to ssh to the laptop via something like the following from your controller machine (i.e. not the laptop you are setting up)

    ssh baaahs@sally.local        # Using whatever name you gave your machine of course

Assuming that works, and that you are a sane ssh user with a private key already setup locally, then it's time to make this new laptop trust you. Again - use whatever name you're working with. Something along the lines of

    ssh-copy-id -i ~/.ssh/id_rsa.pub baaahs@sally.local

might work, or you might need the more old school approach

    ssh baaahs@sally.local mkdir -m 700 .ssh
    scp ~/.ssh/id_rsa.pub baaahs@sally.local:.ssh/authorized_keys

**Now!** Finally. We are ready to deploy the Ansible!

Local Virtualization
====================

Lets say you want to work on this configuration project upstairs instead of in front of a laptop. Well you would want to get the [xubuntu ISO file](https://xubuntu.org/download) and run it in [VirtualBox](https://www.virtualbox.org). That  way you could hack at this config and not have to have the laptop in front of you.

Follow the same setup instructions from above and you can get an image up and running. You'll almost certainly want to install the kernel extensions that make the desktop part all happy and good though, and that requires work because xubuntu isn't setup to build kernel extensions out of the box. So we get to do some of this stuff inside the VM as root.

    #apt-get install gcc make perl

And then run the `autorun.sh` script to install the tools into the instance and reboot it.

I also had to change the default networking from NAT to Bridged in order to get the machine on the same network as the host so that ansible can talk to it. Changing this in virtualbox required bouncing the interface in xubuntu.


Ansible
=======

Right now - we're working on it. Or I'm working on it. Someone might be. Basically, this is the real meat and potatoes of where changes are happening.

The basic idea is that you will be running ansible on a controller machine which will then remotely connect to the laptop and will twiddle things around via ssh. So you need ansible installed locally on the controller machine. For Mac Os this is probably

    brew install ansible

Read the output, deal with any local system issues related to python environments and the like. It might just work for you and if it doesn't you probably caused the issue that is breaking it yourself and are qualified to resolve it. I'm just a text file. I can't fix all the things.

The magic "just go do it already" command is as follows, but see below for the "only do one machine" version.

    ansible-playbook -K chat_laptop_playbook.yml 

The `-K` says ask for the sudo password - which would be the password you used during installation.

At this point, there should hopefully be error messages about things that don't work. As mentioned, this doesn't currently result in a complete chat laptop, but eventually it will.

Hey I finally made progress on sanity. There are now playbooks that target a single machine, the name of which can be passed as a variable. So during development quick turns you might want to do something more like this:

    ansible-playbook -K -e "the_one=sally" update_one.yml
    
There is also a `setup_one.yml` file that can be used instead of the `chat_laptop_playbook.yml` file from above when you are setting up one machine at a time (which is probably a good idea to be doing).

THEN I discovered a new thing - the install process doesn't actually create a unique machine id the way it really really should. Come on guys. Anyway, there is a one off playbook that is also a "only target one host" playbook you can run to fix this. It's the `new_machine_id_one.yml` and it wants a `the_one` variable defined as above. You really should only run this once per machine and shouldn't need to ever again. The only reason this came up is that I'm using the machine id in the `bchat-tty` program so that a machine which goes offline can be re-identified. But to be real big boy linux computers this id really should be getting set (much like windows hell, but whatever. It's not 100% stupid)


Notes on Inventory
==================

The way ansible knows which systems to talk to is the result of both the `ansible.cfg` file in this directory which references `inventory.yml`. 

Things Left To Do
=================

These should all be done via ansible

* Set a good display mode/size for the tty


Fix for q character on laptops
==============================

In 2023 right before the playa we discovered that the a "q" character would be displayed in the input field after all whitespace. So if you are typing a sentence like "Hello out there world" it would display in the input field as "Helloqoutqthereqworldq" or some such nonsense. The issue was that we had configure the tty to use the term type `linux` because, well, that is what the other terminals were configured for and it makes sense right? Well wrong, fuck you, ncurses doesn't work right when you do that.

Thus the fix is to configure the terminal as `ansi` which seems to still preserve the color capabilities without fucking up other necessary things. The ansible task has been updated to do this correctly now, but here is the step by step manual fix that is probably easier on the playa (at least until we update the bchat code???)

  * Get a root console, either via ssh or using one of the other virtual consoles (press CTRL-ALT-F2 etc.)
  * As root, edit `/etc/systemd/system/getty@tty1.service`
  * Search for the `ExecStart` line which specifies the `-/sbin/agetty`. Change the last parameter of this command from `linux` to `ansi`. This is the terminal type.
  * Run `systemctl daemon-reload` to pick up the change to the service file

After this you should be able to switch back to console 1 and hit `^C` to exit the application, which should then get restarted with the right terminal type. You _might_ need to restart the laptop, but probably not (although not a bad idea).

