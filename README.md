# geisha (芸者)

A simple MP3 playing daemon for Linux or OSX.

    $ geishad &
    $ geishac play song.mp3     # play song immediately
    $ geishac enqueue song2.mp3 # play after song.mp3
    $ geishac pause
    $ geishac sub               # get events from geisha server
    $ geishac get_state         # get server state
    $ geishac shutdown          # shutdown server

Similar architecture to `herbstluftwm` - there is a server process
running and clients use `geishac` or the `geisha` Go library to
communicate with the server. geisha is meant to be used alongside
other tools. For instance, you can use `fzf` to add a song to the
queue:

    $ geishac enqueue $(find ~/music | fzf)

Long term plan would be to have a suite of tools including:

 - [ ] `geisha-controls` termbox play/pause/mute/etc.
 - [ ] `geisha-albumart` extract album art from currently playing tracks


### Commands

 - `play <id>`
 - `next <song ...>`
 - `enqueue <song ...>`
 - `remove <id ...>`
 - `pause|resume|toggle|skip|prev|toggle|fwd|bwd`
 - `shuffle|sort|loop`
 - `get_queue`
 - `get_state`
 - `sub`


### TODO

 - [x] `geishac`
 - [x] implement `MethodEnqueue`
 - [x] `MethodSort`
 - [x] `MethodShuffle`
 - [x] playback modes:
   - [x] `repeat`
   - [x] `loop`
 - [x] stable queue shuffle + sort
 - [x] `MethodRemove`
 - [ ] clean up `geishad` code
 - [ ] clean up `geishac` code (PLEASE)
 - [ ] check for memory leaks
 - [ ] tests
   - [ ] hairy queue logic
 - [ ] support playback for other filetypes
