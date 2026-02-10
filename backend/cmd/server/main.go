package main

import (
	"fmt"
	"net/http"
)

/*
Features to implement
Music
- Filter music by
	- artist
	- name
	- popularity
	- likes
	- what I am following
Music needs to show
	Basic Information / Tags / Artist / ArtistMember

- Artists
	- Music they produced
	- Albums they produced
	- Follow status (you vs total)

- Follow / Likes
	- add/remove follow
	- get current state

- Playlist
	- Add to playlist
	- List playlists
	- List music within playlist
	- Remove from playlist
	- Reorder in playlist

- Album
	- Add to album
	- List albums by artist
	- Remove from album
	- Reorder in album
	- List albums
	- List music within album

- ListeningHistory
	- Append to history
	- Get history
*/

func main() {
	mux := http.NewServeMux()

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println(err.Error())
	}
}
