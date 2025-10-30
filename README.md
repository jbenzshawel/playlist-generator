# Playlist Generator
This is a simple tool that pulls a list of songs from a configured source
and syncs them to a playlist in spotify. I found that my Spotify Discover Weekly
playlist was no longer interesting and occasionally included AI music. 

## Overview
The tool currently scopes all playlists to songs played over a month at a given source. For example, if
the tool runs for multiple days in a month all songs collected from a source will be in a playlist called
"Source Name YYYY-MM".

In the future additional playlist time range scopes may be added. Another idea I'm considering is generating a
new playlist with a random number of songs played from a specific source. 

### Options
The playlist generator currently supports two modes `single` and `recurring`. 

The `single` mode in combination with the `date` option can be used to update a playlist with songs 
played on a specific date. The `single` mode in combination with the `month` flag will create a playlist with
alls songs played from a source in the given month. 

The `recurring` flag can be used to run the tool in the background to add songs played from a source in real time. For 
example, if the interval is set to `5` minutes the sources playlist for the current month will be updated every five 
minutes with the most recent song(s) added. 

| Flag name  | Default      | Description                                                                                                                                                           |
|------------|--------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `mode`     | single       | The generator mode (single or recurring). In single mode the generator runs once based on parameters. In recurring mode the generator runs at the configured interval | 
| `date`     | current date | The date to download songs for in YYYY-MM-DD. This option is only used in single mode.                                                                                |
| `month`    |              | The month to download songs for in YYYY-MM. This option is only used in single mode.                                                                                  |
| `interval` | 60           | The interval, in minutes, between updating the playlist. This option is only used in recurring mode                                                                   |
| `verbose`  | false        | Whether to include detailed logs                                                                                                                                      | 

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



