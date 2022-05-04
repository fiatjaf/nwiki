nwiki
=====

Command line client for the [Nostr](https://github.com/fiatjaf/nostr) wiki.

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

[![asciicast](https://asciinema.org/a/dtrzdbg7BnMq0hUMzDE3F6yDe.svg)](https://asciinema.org/a/dtrzdbg7BnMq0hUMzDE3F6yDe)
