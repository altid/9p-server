# Example 9p Client

This client is a very simple client which writes the body of `document`, scrolls `feed`, or writes the URL of `stream` to stdout, and reads from stdin for command/input messages

## Usage

```
mkfifo somefifo

example <ip address> [<ip address> ...] < somefifo

echo /ctrl message > somefifo
echo some input > somefifo
```
