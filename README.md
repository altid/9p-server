# 9p-Server

9p-server is a ubqt server, used to connect to clients over the 9p protocol.

`go install github.com/ubqt-systems/9p-server

## Usage

`9p-server [-t] [-d <dir>] [-c certfile] [-k keyfile] [-u username]`

 - `-t` enables TLS use
 - `-c <certfile>` certificate file for use with TLS connections (Default /etc/ssl/certs/ubqt.pem)
 - `-k <keyfile>` key file for use with TLS connections (Not required for systems with factotum, default /etc/ssl/private/ubqt.pem)
 - `-d <dir>` directory to watch (Default /tmp/ubqt)
 - `-u <username>` Run as user (Default is current user)

## Configuration

```
# ubqt.cfg - place this in your operating system's default config directory

service=foo
	#listen_address=192.168.1.144:12345
```
 - listen_address is a more advanced topic, explained here: [Using listen_address](https://ubqt-systems.github.io/using-listen-address.html)

## Plan9

 - On Plan9, the default certfile is set to $home/lib/ubqt.pem
 - You must run all services in the same namespace that 9p-server is running
