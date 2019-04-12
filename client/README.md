# Example 9p Client

This client is a very simple client which writes the body of `document` or writes the URL of `stream` to stdout, and reads from stdin for command/input messages
`feed` elements are not currently supported, to keep things simple.

## Usage

`example <ip address> [<ip address> ...]`

## Commands

Any input with a leading slash will be read as a command, and sent through to the underlying service, except for the following

 - `/tabs` will list all current buffers
 - `/sidebar` will list the items in the sidebar
 - `/status` will print the current buffer status
 - `/title` will print the current buffer title
 - `/quit` will end the program

## Input

Any text without a leading slash will be written to the `input` file of the given buffer, if it exists. If input is disabled for the buffer, an error will be returned.

Full markdown support is implemented, and for example the following can be used for richer input:

 - `This is a %[red text](red)` 
 - `This is a %[red text with blue background](red, blue)`
 - `This is **bold text**`
 - `This is *emphasized text*`

