# Ymail-smtp

A direct-delivery SMTP server—it accepts emails from clients and forwards them to recipients directly via **DNS MX Lookup** (without an intermediate relay).

Built using the following libraries:
- [Cobra](https://github.com/spf13/cobra/)
- [go-smtp](https://github.com/emersion/go-smtp)
- Other standard Go libraries

## Features
- Email reception
- Direct delivery to the recipient
- STARTTLS during delivery, if supported by the recipient
- Does not store emails—transit forwarding only

## Running

```shell
git clone https://github.com/ngo-yellowhat/ymail-smtp
cd ymail-smtp
make build run
```
This starts the SMTP server on `localhost:2525`.
See [Makefile](Makefile).

### Docker

```shell
git clone https://github.com/ngo-yellowhat/ymail-smtp
cd ymail-smtp
make build-docker run-docker
```
The container runs in `--network=host` mode, using the host's network stack directly (without port mapping). This means the port the server listens on inside the container is available on the host on the same port number without additional configuration.
See [Dockerfile](Dockerfile).

## Testing
```shell
telnet localhost 2525
```
Or use the specific port you configured.
Testing via `netcat` (nc) may cause issues: by default, `nc` uses `\n` instead of the protocol-required `\r\n`, which can prevent the server from recognizing the `DATA` termination sequence. `telnet` or a real SMTP client (e.g., Python with `smtplib`) is recommended for reliable testing. ***

## Minor caveats
1. No authentication between servers
2. Port 25 is blocked by most residential ISPs (as a security measure against spam bots)
3. The server processes each recipient individually (one connection per client)

## LICENSE
GNU General Public License v3
