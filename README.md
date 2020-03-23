[![Build Status](https://github.com/antonbaumann/spotify-jukebox/workflows/build/badge.svg)](https://github.com/antonbaumann/spotify-jukebox/actions?workflow=build)
[![codecov](https://codecov.io/gh/antonbaumann/spotify-jukebox/branch/master/graph/badge.svg)](https://codecov.io/gh/antonbaumann/spotify-jukebox)

# Spotify Jukebox

## Quick start
#### Build and run
```sh
go build .
./spotify-jukebox
```

## REST Api
#### User related
##### join: 
- `GET /users/join/{username}`
- response: `{"username": <username>, "secret": <secret>, "is_admin": false, "score": <score>}`
##### list:
- `GET /users/{username}/list`
- headers: `{"Authorization": <secret>}`
- response: `[{"username": <username>, "is_admin": false, "score": <score>}]`
##### suggest song
- `GET /users/{username}/suggest/{song_id}`
- headers: `{"Authorization": <secret>}`
- response: `{
                 "duration_ms" : <duration>,
                 "preview_url" : <url>,
                 "score" : <score>,
                 "id" : <id>,
                 "time_added" : <time string>,
                 "suggested_by" : <user>,
                 "name" : <name>,
                 "artists" : []
               }`
##### list songs
- `GET /users/{username}/listSongs`
- headers: `{"Authorization": <secret>}`
- response: `[{
                 "duration_ms" : <duration>,
                 "preview_url" : <url>,
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
#### Server related
- ping: `GET /ping`