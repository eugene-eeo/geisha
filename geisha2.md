## geisha 2.0

 - workflow:
   - `geishad` simple queue server
   - `geishac` talk to geishad (scripting)
   - `geisha-tagged` manage tags db, load songs etc

 - simpler queue (test impl!) and flags system
   - flag for play-once

       | id | song-ref | flags |

   - enqueue()
   - next() => check flags and act accordingly
   - nextForce() => ignore flags
   - shuffle()
   - setRepeatMode(REPEAT-ONE | REPEAT-PLAYLIST)
   - remove(id)

 - song tags (playlist-esque)
   - loadTag(t)
   - unloadTag(t)
   - mergeTag(t, u)
   - deleteTag(t)
   - listTags()
