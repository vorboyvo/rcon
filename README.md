rcon is a command-line application that allows you to issue commands remotely
to servers running Team Fortress 2, Counter-Strike: Global Offensive, ~~Minecraft~~,
and other games which support the [Source RCON Protocol](https://developer.valvesoftware.com/wiki/Source_RCON_Protocol).

It is currently in beta. Functionality has not been fully tested; expect bugs and instability, but nothing too breaking.

# Features

* Connect to SRCDS servers and run commands remotely
* Send commands in either the command body or in standard input
* Save server information in config and reuse it to avoid having to retype hostname, port, password

## Planned Features

* Validate for strings being ASCII
* Dynamic window title/server/connection status
* Localization
* For certain games (especially Team Fortress 2), cache server information and allow for tab-completion
* Minecraft support

# Installation

## Build

Prerequisites:
* git
* go 1.19
  * github.com/BurntSushi/toml v1.2.1
  * github.com/cheynewallace/tabby v1.1.1
  * github.com/spf13/pflag v1.0.5

Clone the code into a local directory:

```$ git clone https://github.com/vibeisveryo/rcon.git```

```$ go build```

Copy the output file (`rcon.exe` on Windows, `rcon` on Mac/Linux) to your path.

## Release

### Windows

Download and run the installer `rcon_install.exe`.

To uninstall, select "rcon" in the Apps and Features sections of Settings, and uninstall using the given installer.

### Mac

See build instructions; there is no package for macOS.

### Linux

Download and extract the .tar.gz archive `rcon.tar.gz`. In the extracted directory, run with root permissions:

```# ./rcon_install.sh```

To uninstall, run with root permissions:
```# rm /usr/bin/rcon```

# Usage

You can use `rcon -h` or `rcon --help` to get a usage message.

There are two ways to use rcon: to issue a single command, or to take commands interactively (take over the shell window).

To issue a single command, run `rcon [options] [your command here]`. To take commands interactively, run `rcon [options]`, and then issue commands into the terminal. In any event, options must include either a hostname and an RCON password and, if different from the default 27015, port, or a server from the configuration file.

To exit out of interactive mode, send an end-of-file signal to the terminal. This can be done on Linux or Mac by pressing Ctrl+D, or on Windows by pressing Ctrl+Z then Enter.

## Configuration file

The configuration file is located in the "rcon" subdirectory in the user config directory; by default:
* on Windows: C:\Users\YourUsernameHere\AppData\Roaming\rcon\config.toml
* on macOS: ~/Library/Application Support/rcon/config.toml
* on Linux: ~/.config/rcon/config.toml

In it, you can store a list of servers, such as ones you use frequently, for easy use, using a TOML-based format as follows:

```
[someservername1]
hostname = "172.0.0.1"
port = 27015
password = "somepassword"

[someservername2]
hostname = "172.0.0.2"
port = 27035
password = "differentpassword"
```

Then, you can call rcon as follows:

```rcon -s someservername1```

## Examples

```$ rcon -H example.com -p 27035 -P myPassword status```

```$ rcon -s exampleServer sv_password hello```

## Security

Note that the RCON protocol sends passwords in unsecured plain text over the internet; this is universal to RCON, not specific to this program. If this is a concern to you, you should consider running this program through an SSH tunnel.

# License

Copyright 2023 vorboyvo.

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public
License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later
version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see
https://www.gnu.org/licenses/.

# See also

* This project was inspired by n0la's [rcon](https://github.com/n0la/rcon).
