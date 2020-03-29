[![Build Status](https://github.com/antonbaumann/spotify-jukebox/workflows/build/badge.svg)](https://github.com/antonbaumann/spotify-jukebox/actions?workflow=build)
[![codecov](https://codecov.io/gh/antonbaumann/spotify-jukebox/branch/master/graph/badge.svg?token=juTAuitYfJ)](https://codecov.io/gh/antonbaumann/spotify-jukebox)
# Spotify Jukebox

## Quick start
#### Build and run
```sh
docker-compose up -d

# on first time building docker image
# initialize mongo db replication
mongo admin -u root -p root
rs.initiate()

go build .
./spotify-jukebox
```

## REST Api
#### User related
##### join: 
- `POST /users/join/{username}`
- response: `{"username": <username>, "secret": <secret>, "is_admin": false, "score": <score>}`
##### list:
- `GET /users/{username}/list`
- headers: `{"Authorization": <secret>}`
- response: `[{"username": <username>, "is_admin": false, "score": <score>}]`
##### suggest song
- `POST /users/{username}/suggest/{song_id}`
- headers: `{"Authorization": <secret>}`
- response: `{
                 "duration_ms" : <duration>,
                 "preview_url" : <url>,
                 "cover_url": <url>,
                 "album_name": <name>,
                 "score" : <score>,
                 "id" : <id>,
                 "time_added" : <time string>,
                 "suggested_by" : <user>,
                 "name" : <name>,
                 "artists" : []
               }`
##### vote up/down
- `POST /users/{username}/vote/{song_id}/up`
- `POST /users/{username}/vote/{song_id}/down`
- headers: `{"Authorization": <secret>}`
- response: `[SongInfo]`
##### list songs
- `GET /users/{username}/listSongs`
- headers: `{"Authorization": <secret>}`
- response: `[{
                 "duration_ms" : <duration>,
                 "preview_url" : <url>,
                 "cover_url": <url>,
                 "album_name": <name>,
                 "score" : <score>,
                 "id" : <id>,
                 "time_added" : <time string>,
                 "suggested_by" : <user>,
                 "name" : <name>,
                 "artists" : []
               }]`
#### Admin related
##### login: 
- `POST /admin/login` 
- request: `{"username": <username>, "password": <password>}`
- response: `{"username": <username>, "secret": <secret>, "is_admin": true, "score": <score>}`
##### remove song: 
- `DELETE /users/{username}/removeSong/{song_id}`
- headers: `{"Authorization": <secret>}`
- response: `[{
                   "duration_ms" : <duration>,
                   "preview_url" : <url>,
                   "cover_url": <url>,
                   "album_name": <name>,
                   "score" : <score>,
                   "id" : <id>,
                   "time_added" : <time string>,
                   "suggested_by" : <user>,
                   "name" : <name>,
                   "artists" : []
                 }]`
                 
#### events
- `GET /events`

#### Server related
##### ping:
- `GET /ping`