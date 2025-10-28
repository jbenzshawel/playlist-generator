# Playlist Generator
This is a simple tool that pulls a list of songs from a configured source
and syncs them to a playlist in spotify. I found that my Spotify Discover Weekly
playlist was no longer interesting and occasionally included AI music. 

## Overview

### Example
The tool can be run with the following command:
```
./playlist-generator -date=2025-10-28
```

### Authentication 
Spotify requires using the OAuth authentication code grant type when accessing
any information specific to a user, such as a playlist. On startup a link will display
that can be used to authenticate with your Spotify account.  

### Sources
Sources represent a source that provides a list of songs played. 

The only supported source is currently IPR Studio One, however I  plan on adding 
support for MPR The Current if I can find a source to download a daily list of songs
played on that station. 

### Playlists
Playlists represent where the playlist is created. Spotify is currently the only supported 
streaming provider, however I may add support for Apple Music or Tidal if they have free
APIs available. 



