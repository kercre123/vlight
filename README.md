# vlight

Code which allows Vector to control my lights.

This includes animations, a sound, and a fully functional implementation.

You'll have to change the following in main.go:

```
var podURL string = "192.168.1.222:8080"
var lightsOnEndpoint string = "http://192.168.1.75:8080/lights_on"
var lightsOffEndpoint string = "http://192.168.1.75:8080/lights_off"
```

Only works on unlocked robots.

A gist is coming soon.

To install onto your unlocked Vector:


```
git clone https://github.com/kercre123/vlight
cd vlight
ssh-add <path/to/ssh_key>
./send.sh <bot_ip_address>
```



