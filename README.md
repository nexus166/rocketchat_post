# deployment

## webhook
Visit hxxps://api.slack.com/apps/YOURAPP/incoming-webhooks, create/obtain webhook for the channel/user/group you want the messages to be posted to


## go POST executable
### Build the go program for the target(s):
```sh
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -v -ldflags="-s -w -X main.defaultWebHook=https://hooks.slack.com/services/abc/def/xyz"
```
The` main.defaultWebHook` flag is optional. Assumes you are fine with including the (secret) webhook path in the resulting binary. It is possible to pass `-W` parameter to specify that in a more secure way, if you have reason to do so.

### Installation
1. Copy the resulting binary in `/usr/local/bin/hermes` on all the hosts. Ensure it is executable.


# usage
A simple message can be sent with
```sh
hermes "test webhook message"
```
In general, all text that is piped to the binary will be wrapped in the POST and sent to the webhook.
```sh
ip a | hermes
```
Based on this, there are a couple of usage examples:
- [PAM sshd hook](./etc/pam.scripts/hermes.sh): A [pam_exec](https://www.freebsd.org/cgi/man.cgi?query=pam_exec&) _optional_ hook to a [script](./etc/pam.scripts/hermes.sh) that will alert of all SSHd login/out activities on the host.
- [reboot_required](./etc/cron.daily/reboot_required): A simple script to test for the /var/run/reboot_required* files (where this is supported), which will report the need to reboot the host and why you should do that.
