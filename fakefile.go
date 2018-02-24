package main

import (
//	"log"
//	"github.com/mortdeus/go9p/srv"
)
// open foo site | LIST FILES | ADD FILES (use other style of 9p server)
// Then, add inotifywatch | add/delete files
// buffer foo site | list files | add files
// close - delete inotifywatch
// Amount of watches increases with clients - stress this some
// We need to figure out if this is tunable per-client.
// We could have a seperate root per client connected

// TODO: Research ordering of connOpened and attach requests from go9p
/* TODO: When we encounter one of these files, we need to synthesize the response; else we pass the regular requested file along, if it exists 
Normalfiles (
	ctl
	status
	input
	feed
	doc
	title
	tabs
	etc
)
	// Readdirnames(n int) (names []string, err error)
	// add list of files in the buffer we're watching
	// add list of Normalfiles we want to check for
	// - if file already exists in our dir object, don't add. (map by string)
*/
