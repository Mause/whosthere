# whosthere

Knock knock.. Who's there? 🚪

Whosthere is a TUI network scanner written in Go that quickly discovers devices and services on your local network.
It's built for fast, concurrent scans and designed to run without `sudo` by using normal user‑space network requests (
e.g., TCP/UDP connects and multicast queries).
Pluggable scanners run concurrently and feed a central engine that merges results and enriches device manufacturer
information via an [**OUI registry**](https://standards-oui.ieee.org/).

I am by no means a network expert, and mainly built this application to learn more about networking and concurrency in
Go.
Feel free to raise feature requests, open issues, or contribute code! I will try my best to review and merge
contributions.

## Installation

You can download precompiled binaries for Linux and macOS from the [releases page](tbd), or simply install via your OS
package manager if available.

TODO(ramon): start with install script instead, later on introduce all package managers

### Homebrew (macOS/Linux)

```bash
brew install whosthere
```

### apt (Debian/Ubuntu)

```bash
apt install whosthere
```

### yum (Fedora/CentOS)

```bash
yum install whosthere
```

### pacman (Arch Linux)

```bash
pacman -S whosthere
```

### From source

Make sure you have [Go](https://golang.org/dl/) installed.

```bash
go install github.com/yourusername/whosthere@latest
```

## Configuration

A lot of behavior within whosthere can be configured to your liking. By default, whosthere will try to look for a
configuration
file at `$XDG_CONFIG_HOME/whosthere/config.yaml`, or `~/.config/whosthere/config.yaml` if the [**XDG Base Directory
**](https://specifications.freedesktop.org/basedir/latest/#basics)
environment variables are not set. If no configuration file is found, whosthere will create one with default values on
first run.
The location of the configuration file can be overridden by providing the `--config` (`-c`) flag when starting
whosthere,
or the `WHOSTHERE_CONFIG` environment variable.

Here is an example configuration file with all available options and their default values:

```yaml
scan_interval: 20s # how often a scan gets triggered
scan_duration: 10s # maximum duration of a scan
splash:
  enabled: true # show splash screen on startup
  delay: 1s     # delay for the splash screen
theme:
  # pick a built-in theme by name; unknown names fall back to "default"
  name: default
  # default palette (uncomment and set name: custom to tweak)
  # primitive_background_color: "#000a1a"
  # contrast_background_color: "#001a33"
  # more_contrast_background_color: "#003366"
  # border_color: "#0088ff"
  # title_color: "#00ffff"
  # graphics_color: "#00ffaa"
  # primary_text_color: "#cceeff"
  # secondary_text_color: "#6699ff"
  # tertiary_text_color: "#ffaa00"
  # inverse_text_color: "#000a1a"
  # contrast_secondary_text_color: "#88ddff"
scanners:
  mdns:
    enabled: true
  ssdp:
    enabled: true
  arp:
    enabled: true
```

## Logging

Whosthere supports logging to a file for debugging and monitoring purposes. By default, logs are written to
`$XDG_STATE_HOME/whosthere/whosthere.log`, or `~/.local/state/whosthere/whosthere.log` if the [**XDG Base Directory
**](https://specifications.freedesktop.org/basedir/latest/#basics)
environment variables are not set. The log level is set to `info` by default, but can be changed via the `WHOSTHERE_LOG`
environment variable.

For example, to set the log level to `debug`, you can start whosthere with the following command:

```bash
WHOSTHERE_LOG=debug whosthere
```

## Platforms

This application has been tested on Linux and macOS. Windows support is not currently available, but contributions
to add Windows compatibility are welcome!

## Engine

### Scanners

...

### OUI Table

https://standards-oui.ieee.org/oui/oui.csv