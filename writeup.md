# Vector controlling my lights

## Background/why

Developers have gotten Vector to do smart home stuff via wire-pod, though this usually involves an interaction like the following:

```
User: "Hey Vector, turn the lights on."
<lights on>
Vector (TTS): "Turned the lights on."
```

I prefer Vector to speak as little as possible. It just feels like cheating to divert things to the TTS, and it breaks character. I wanted to implement it like this:

```
User: "Hey Vector, turn the lights on."
Vector: shows animation, lights turn on in sync with the animation
```

**This is not possible on a production robot. If your robot is unlocked, this won't work for you.**

## Demo

[Demo Video](images/vlightexample.mp4)

## Challenges/how

1. [Digital Dream Labs open-sourced the vector-animations-raw git](https://github.com/digital-dream-labs/vector-animations-raw). This gives us access to the Anki's VictorAnim and Vector rig, which allows us to create and export animations in Maya.

-   I created animations so you don't have to. Though, for anyone looking to install Maya and the plugin, here's how I did it:

    -   Obtain Maya 2020 (versions 2018-2020 work, 2020 is what I used) from Autodesk. I did this by getting a student license.
    -   [Use this to install on Arch](https://github.com/MyHCel/Maya-For-Arch)
    -   Clone the vector-animations-raw repo.
        -   This will not work as DDL has run out of LFS tokens and it doesn't seem like they will be paying for more anytime soon.
        -   I forked and archived the repo with "Include Git LFS in archive" enabled, which allowed the LFS objects to be downloaded.
            -   So, instead of cloning it, you should download this and unzip it to your home directory then rename it to `vector-animations-raw` [https://github.com/kercre123/vector-animations-raw/archive/refs/heads/main.zip](https://github.com/kercre123/vector-animations-raw/archive/refs/heads/main.zip).
    -   Follow the instructions in the repo. When you get to the step where you copy the Maya.env file, make sure the destination directory matches with the version of Maya you got.
    -   Here is a description of the VictorAnim shelf items: [https://github.com/kercre123/vector-animations-raw/blob/main/documentation/VictorAnim%20Shelf%20in%20Maya.md](https://github.com/kercre123/vector-animations-raw/blob/main/documentation/VictorAnim%20Shelf%20in%20Maya.md)

2. We cannot implement new noises into Vector and we cannot put sounds into new animations. This is due to proprietary code which could not be released to the public and the lack of full engine source code. I got around this by taking advantage of Vector's APIs. Dev robots host a webserver for debugging. This gives us lots of control. We can directly set behaviors, play sound events, play animations, etc all via POST requests/websocket messages. In the final program I created, the webserver is used to play audio events which already exist but couldn't be implemented directly into the animation. As for the "thock" sound, I just used an AudioStream via the SDK.

3. We can't add an intent via normal methods due to lack of access to the engine source code. I got around this by creating a custom program which runs on the bot and looks at the log for responses from the cloud, then takes control of the bot when it sees a certain phrase.

## What exactly is implemented

-   "Hey vector, turn the lights on"
-   "Hey Vector, turn the lights off"

That's it, for now.

## How to install my implementation

1. Make sure you have an OSKR/dev-unlocked bot which is connected to a wire-pod instance. You will also need a Linux machine for compilation.

2. Install docker on your machine:

```
# Ubuntu
sudo snap install docker
# Debian
sudo apt install docker.io
# Arch
sudo pacman -S docker
```

3. [Install Go](https://go.dev/doc/install) (should already be done if you have wire-pod installed on the same Linux machine)

4. Make sure you have your bot's SSH key. If you don't have it, [follow these instructions](https://web.archive.org/web/20230401010147/https://oskr.ddl.io/article/451-oskr-detailed-unlock-steps)

5. Do an SSH test. Run the following command:

```
ssh -i <path/to/sshkey> root@<vectorip> "uname -a"
# replace <path/to/sshkey> with the path to the ssh key, <vectorip> with vector's ip address
```
You should see something like the following:

```
Linux Vector-H3W7 3.18.66 #1 SMP PREEMPT Thu Jul 2 16:34:20 PDT 2020 armv7l GNU/Linux
```

If you see an error saying `no mutual signature accepted` (or something like that), run:
(**this should only be run once**)
```
sudo -s
echo "PubkeyAcceptedKeyTypes +ssh-rsa" >>/etc/ssh/ssh_config
exit
```

6. Run:

```
cd ~
git clone https://github.com/kercre123/vlight.git
cd vlight
```

7. Open main.go in a text editor. Replace the URLs near the top of the file, most importantly the podURL (should correspond to the IP address of your wire-pod instance). You'll have to modify the Go code to get it to communicate with your smarthome solution.

8. Run:

```
sudo make docker-builder
ssh-add <path/to/sshkey>
./send.sh <vector_ip>
```

9. The commands should now be implemented in Vector. Test it by saying "Hey vector, turn the lights on".

10. The commands should work, but Vector will play the "IDK" noise before playing the animation and will run the "UnclaimedIntent" behavior, which is out of place. To get rid of these:

    1. Create a custom intent in WirePod with the settings in [this screenshot](images/vlightintent.png). This will get rid of the "IDK" noise before the lights on/lights off animations.
    2. SSH into Vector (ssh -i <path/to/sshkey> root@<vector_ip>). You should be at a prompt which says "root@Vector-####:~# ". Run the following (one at a time. anki-robot.target should be given a few seconds to fully stop):

    ```
    systemctl stop anki-robot.target
    sed -i '/"ReactToUnclaimedIntent",/d' /anki/data/assets/cozmo_resources/config/engine/behaviorComponent/behaviors/victorBehaviorTree/globalInterruptions.json
    systemctl start anki-robot.target
    ```
    
    A 923 error is normal.

    This will get rid of the UnclaimedIntent animation after the lights on/lights off animations.

11. This feature should now be fully implemented!

## How I made my lights smart

Servo, ESP8266, [3d-printed thing](https://www.thingiverse.com/thing:2146951/files) (not the exact one I am using, but close enough)

Right now, Vector just directly tells the ESP8266 what to do. Ideally this should be done through an MQTT server of some sort.
