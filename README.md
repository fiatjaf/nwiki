nwiki
=====

Command line client for the [Nostr](https://github.com/fiatjaf/nostr) wiki.

## Run in container, without installation

If you don't have `go` installed or don't like installing stuff just to
quickly try it out, you can easily run this in a `docker` or `podman`
container:

```
podman run --rm -it golang bash
mkdir -p /root/.config/nostr/
echo '{"relays": {"wss://nostr-pub.wellorder.net": {"read": true,"write": true}},"privatekey": "d2c8bb39f07285067b6d027b3f3a82a07febef57fd9a3c94ed5abde11e29804c"}' > /root/.config/nostr/config.json
apt update; apt install -y vim
go install github.com/fiatjaf/nwiki@latest
nwiki bitcoin
```

## Installation

Compile with `go install github.com/fiatjaf/nwiki@latest` or [download a binary](https://github.com/fiatjaf/nwiki/releases).

## Usage

First this requires a config file at `~/.config/nostr/config.json` with private key and relays configured in it, such as the file that is created and managed by [noscl](https://github.com/fiatjaf/noscl).

Or you can edit it manually into something that looks like this:

```json
{
  "relays": {
    "wss://expensive-relay.fiatjaf.com": {
      "read": true,
      "write": true
    }
  },
  "privatekey": "d2c8bb39f07285067b6d027b3f3a82a07febef57fd9a3c94ed5abde11e29804c"
}
```

The call it with `nwiki '<article>'` (in which `<article>` is the name of the article you want to read, create or edit).

You'll be shown with all the article versions from people on your configured relays -- if any. Pressing `Enter` will enter the edit screen, and exiting that will publish it (unless you save an empty file or an unchanged file).

## Video Demo

[![asciicast](https://asciinema.org/a/9VxZEWV0MDUUsBAkKqgYaQH0P.svg)](https://asciinema.org/a/9VxZEWV0MDUUsBAkKqgYaQH0P)
