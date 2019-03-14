# Example 9p Client

This client is a very simple client which writes the body of `document` or writes the URL of `stream` to stdout, and reads from stdin for command/input messages
`feed` elements are not currently supported, to keep things simple.

## Usage

`example <ip address> [<ip address> ...]`

Anything typed in the terminal will be read as stdin.
To quit, type `/exit`

