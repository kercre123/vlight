# vlight

Code which allows Vector to control my lights.

This includes animations, a sound, and a fully functional implementation.

You'll have to change the following in main.go:

```
var lightsOnEndpoint string = "http://192.168.1.75:8080/lights_on"
var lightsOffEndpoint string = "http://192.168.1.75:8080/lights_off"
```

Only works on unlocked robots.

A gist is coming soon.

You will need to install Docker.

To install onto your unlocked Vector:


```
git clone https://github.com/kercre123/vlight
cd vlight
sudo make docker-builder
ssh-add <path/to/ssh_key>
./send.sh <bot_ip_address>
```



