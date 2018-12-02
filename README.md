# geisha

A simple MP3 playing daemon for Linux or OSX.

    $ geishad &
    $ geishac play song.mp3   # play song immediately
    $ geishac queue song2.mp3 # play after song.mp3
    $ geishac pause
    $ geishac sub             # get events from geisha server
    $ geishac get_state       # get server state
    $ geishac shutdown        # shutdown server

Similar architecture to `herbstluftwm` - there is a server process
running and clients use `geishac` or the `geisha` Go library to
communicate with the server.

### TODO

 - [ ] `geishac`
 - [x] implement `MethodEnqueue`
 - [x] `MethodSort`
 - [x] `MethodShuffle`
 - [x] playback modes:
   - [x] `repeat`
   - [x] `loop`
 - [ ] clean up `geishad` code
 - [ ] clean up `geishac` code (PLEASE)
 - [ ] check for memory leaks
 - [ ] tests (lol)
