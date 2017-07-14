package main

/* For most every file, we won't have to do much; but tabs, events, ctl, input will have to be synthesized
input, ctl: will listen for server-specific tokens, otherwise pass through to underlying server
tabs: will use a map of buffers to synthesize the contents (per client)
events: will aggregate all events from underlying servers as part of the update mechanism, and will forward anything pertinent to client updates
all other files are 1:1 passed though
*/
