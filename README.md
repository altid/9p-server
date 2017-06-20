# Server
============

This serves up a multiplexed directory to clients, each one getting a unique view. For example, Client A could be viewing an IRC channel as its' buffer, while Client B is viewing a game stream. There's no reason why Client A and Client B cannot both view the same channel, though.

## Usage

### Setup

Ensure you have one or more fileservers set up to lay out the directories that ubqt-server expects. They are in the following form:

```
ircfs
	ctl
	foo/
		feed
		title
		status	
		sidebar
		input
	bar/
		...

docfs
	ctl
	status
	foo/
		document
		title
		sidebar
		<images...>
	bar/
		...
```

As can be seen above, the file "status", as well as title, sidebar, input, and ctl can exist in both the base of the directory, and in an individual subdirectory. If a file exists in a subdirectory, it will be served with precedence over the base level! This is important, as status will have to be removed from a specific buffer directory if you want to show errors in commands with the base level status file. (Note, you do not need to implement all files, if they don't make sense; nor do they have to exist at all times. Simply put, if they exist at a given time, they will be handed to the client. If they're only relevent 1/10 times, only show it that frequently.)

### Running

Simply call ubqt-server, with the path of the directories you've set up in the last step.

`ubqt-server -p <port> ~/my/path/to/ircfs ~/my/path/to/docfs` 

It's also possible to call it as follows: `ubqt-server ~/my/path/to`, so you can actively add or remove services while running. 
By default, ubqt-server will listen on port 564, or whatever is passed in with -p

## Notes

This is currently a work in progress, but please do not hesitate to add your advices, opinions, or actual code to the project. A major goal in writing this is to have it approachable for entry level coders and seasoned veterans alike. 
Thanks!
