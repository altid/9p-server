# Example 9p Client

This client is a very simple client which writes the body of `document` or writes the URL of `stream` to stdout, and reads from stdin for command/input messages
`feed` elements are not currently supported, to keep things simple.

## Usage

```
# tail -f is available everywhere I use this

touch fifo
tail -f fifo | example <ip address> [<ip address> ...]

# In another terminal
echo /ctrl message >> fifo
echo some input >> fifo
```
