#!/bin/bash

set -x
set -e

sudo make vic-custom
chmod 600 resources/ssh_root_key
ssh-add resources/ssh_root_key
ssh root@$1 "mount -o rw,remount /"
if [[ ! $2 == "jc" ]]; then
	ssh root@$1 "systemctl stop anki-robot.target"
	ssh root@$1 "mount -o rw,remount /"
	scp -O build/vic-custom root@$1:/anki/bin/vic-custom
	scp -O resources/lights_on.json root@$1:/anki/data/assets/cozmo_resources/assets/animations/
	scp -O resources/lights_off.json root@$1:/anki/data/assets/cozmo_resources/assets/animations/
	scp -O resources/light.pcm root@$1:/data/
	scp -O resources/vic-custom.service root@$1:/lib/systemd/system/
	scp -O resources/vic-cloud.service root@$1:/lib/systemd/system/
	scp -O resources/cloud-verbose root@$1:/anki/bin/
        ssh root@$1 "chmod +rwx /anki/bin/cloud-verbose /anki/bin/vic-custom"
	ssh root@$1 "systemctl daemon-reload"
	ssh root@$1 "systemctl enable vic-custom"
	ssh root@$1 "systemctl start anki-robot.target"
else
	ssh root@$1 "systemctl stop vic-custom"
	scp -O build/vic-custom root@$1:/anki/bin/vic-custom
	ssh root@$1 "systemctl start vic-custom"
fi
