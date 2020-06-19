# bipper
Bipper (beep is bip in French) is a generic beeper application. Beep cycles are configured
using simple yaml files.

Bipper is suitable for any routing training (tabata for instance) or any activities that require
simple or advanced beepers (i.e: cooking).

## YAML format
Please have a look at the file `example.yaml`. It provides a simple example on how to use the app.

For a complete description of the file format, please read on:
``` yaml
---
# loop is true if the following actions must be repeated when finished
loop: true
sections:
  - name: Making coffee
    duration: 1m

  - name: Drink up
    duration: 30s
```

The `golang` time format is used to describe time durations - `1d2h3m34s`.

## Platforms

* Windows
* macOS
* Linux
* FreeBSD
* OpenBSD
* Android
* iOS
* Web browsers ([GopherJS](https://github.com/gopherjs/gopherjs) and WebAssembly)

## Prerequisite

Bipper uses `oto` under the hood, so the following dependencies are required.

### macOS

Oto requies `AudioToolbox.framework`, but this is automatically linked.

### iOS

Oto requies these frameworks:

* `AVFoundation.framework`
* `AudioToolbox.framework`

Add them to "Linked Frameworks and Libraries" on your Xcode project.

### Linux

libasound2-dev is required. On Ubuntu or Debian, run this command:

```sh
apt install libasound2-dev
```

In most cases this command must be run by root user or through `sudo` command.

### FreeBSD

OpenAL is required. Install openal-soft:

```sh
pkg install openal-soft
```

### OpenBSD

OpenAL is required. Install openal:

```sh
pkg_add -r openal
```