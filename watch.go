package main

/*
Registering a client in a multiple-server paradigm:
SIGUSR won't work with multiple servers, especially if arbitrarily named
FIFO won't work, if we have multiple servers digesting them
Inotify, recursive would be fine likely, but webfs and such will grow well beyond the watch limits

Inotify on inpath, add watch to folder until we see `event`, then tail event
fs's will append to events - `printf '%s\n' "title" >> event
If event is deleted, add back to watch
We end up with the following structure:

inpath/
    ircfs/
        event
        ctl
        irc.freenode.net/
        ...
    webfs/
        event
        ctl
        https/
        ...
    ...

We let the os handle write contentions on our behalf, and multiple servers can register to listen to these directories (9p, http, circle (from tickit)?)
File servers should periodically flush their event file as well, to keep the size minimal
*/

