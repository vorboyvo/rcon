rcon is a command-line application that allows you to issue commands remotely
to servers running Team Fortress 2, Counter-Strike: Global Offensive, ~~Minecraft~~,
and other games which support the [Source RCON Protocol](https://developer.valvesoftware.com/wiki/Source_RCON_Protocol).

It is currently in alpha. Functionality has not been fully tested; expect bugs and instability.

# Features

* Connect to SRCDS servers and run commands remotely
* Send commands in either the command body or in standard input
* Save server information in config and reuse it to avoid having to retype hostname, port, password

## Planned Features

* For certain games (especially Team Fortress 2), cache server information and allow for tab-completion
* Minecraft support

# Installation

To be finished

# Usage

To be finished. Currently, you can use `rcon -h` or `rcon --help` to get a usage message.

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