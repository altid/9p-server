# Server
============

This serves up a multiplexed directory to clients, each one getting a unique view. For example, Client A could be viewing an IRC channel as its' buffer, while Client B is viewing a game stream. There's no reason why Client A and Client B cannot both view the same channel, though.

## Usage

### Setup

Ensure you have one or more fileservers set up to lay out the directories that ubqt-server expects. They are in the following form:

```
ircfs
	ctl
	event
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
	event
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

9p-server [-d directory] [-p port]
By default, ubqt-server will listen on port 4567, and watch ~/ubqt
Any directories created within your directory will be added to an Inotify watch, until such a time as a named file `event` is created. Then the server will tail the event file, receiving updates and synthesizing required files based on that data. If a fileserver is closed, the event file will be removed and that directory will be added back on to the Inotify watch.

## Notes

This is currently a work in progress, but please do not hesitate to add your advices, opinions, or actual code to the project. A major goal in writing this is to have it approachable for entry level coders and seasoned veterans alike. 
Thanks!
