[![Build Status](https://github.com/antonbaumann/spotify-jukebox/workflows/build/badge.svg)](https://github.com/antonbaumann/spotify-jukebox/actions?workflow=build)
[![Go Report Card](https://goreportcard.com/badge/github.com/antonbaumann/spotify-jukebox)](https://goreportcard.com/report/github.com/antonbaumann/spotify-jukebox)
[![codecov](https://codecov.io/gh/antonbaumann/spotify-jukebox/branch/master/graph/badge.svg?token=juTAuitYfJ)](https://codecov.io/gh/antonbaumann/spotify-jukebox)
# Spotify Jukebox

## Quick start
#### Build and run
```sh
docker-compose up -d
go build . && ./spotify-jukebox
```
#### Run Component Tests
```sh
docker-compose up -d

# start backend with test configuration
go test ./... -tags ci
```
#### Run System Tests
```sh
# start mongodb container containing test db
docker-compose -f systest/docker-compose.yml up -d

# start backend with test configuration
go build && ./spotify-jukebox -config <test_config_path>
go test ./systest/... -config <test_config_path>
```
## Models
#### Song
```js
Song = {
  "id": "7unF2ARDGldwWxZWCmlwDM",
  "name": "A Love Supreme, Pt. II - Resolution",
  "artists": ["list", "of", "artists"],
  "duration_ms": 1337,
  "cover_url": "https://i.scdn.co/image/ab67616d0000b2737fe4eca2f931b806a9c9a9dc",
  "album_name": "A Love Supreme",
  "preview_url": "url to 30 second song preview",
  "suggested_by": "anton",
  "score": 3,
  "time_added": "time string",
  "upvoters": ["omar", "cybotter", "anton"],
  "downvoters": []
}
```
#### User 
```js
User = {
  "id": "omar@sessionID",
  "username": "omar",
  "secret": "secret",
  "session_id": "128 character random alphanumerical string",
  "is_admin": nope, //bool
  "score": 9001,
  "spotify_authorized": true,
}
```

#### User List Element
```js
UserListElement = {
  "username": "omar", 
  "is_admin": false,
  "score": 9001
}
```

## REST Api
#### User related
##### join: 
- `POST /users/join/{username}/session/{sessionID}`
- response: `{"user_info": User, "auth_url": "spotify authorization url"}`
- errors: `[SessionNotFoundError, UserConflictError, InternalServerError]`
##### ping: 
- `POST /users/{username}/ping`
- headers: `{"Authorization": <secret>, "Session": <sessionID>}`
- response: `{"message": "pong"}`
- errors: `[InternalServerError]`
##### list:
- `GET /users/{username}/list`
- headers: `{"Authorization": <secret>, "Session": <sessionID>}`
- response: `[UserListElement]`
- errors: `[InternalServerError]`
##### suggest song
- `POST /users/{username}/suggest/{song_id}`
- headers: `{"Authorization": <secret>, "Session": <sessionID>}`
- response: `Song`
- errors: `[InternalServerError]`
##### vote up/down
- `POST /users/{username}/vote/{song_id}/up`
- `POST /users/{username}/vote/{song_id}/down`
- headers: `{"Authorization": <secret>, "Session": <sessionID>}`
- response: `[Song]`
- errors: `[BadVoteError, InternalServerError]`
##### list songs
- `GET /users/{username}/listSongs`
- headers: `{"Authorization": <secret>, "Session": <sessionID>}`
- response: `[Song]`
- errors: `[InternalServerError]`
##### client token
- `GET /users/{username}/clientToken`
- headers: `{"Authorization": <secret>, "Session": <sessionID>}`
- response `{"access_token": "...", "token_type": "...", "expiry": Time`}
- errors: `[InternalServerError]`
#### auth token
- `GET /users/{username}/authToken`
- headers: `{"Authorization": <secret>, "Session": <sessionID>}`
- response `{"access_token": "...", "token_type": "...", "expiry": Time`}
- errors: `[RequestNotAuthorized, SpotifyNotAuthenticated, InternalServerError]`

#### Admin related
##### Create Session: 
- `POST /admin/{username}/createSession` 
- response: `{"user_info": User, "auth_url": "spotify authorization url"}`
- errors: `[SessionConflictError, UserConflictError, InternalServerError]`
##### remove song: 
- `DELETE /users/{username}/removeSong/{song_id}`
- headers: `{"Authorization": <secret>, "Session": <sessionID>}`
- response: `[Song]`
- errors: `[SessionConflictError, SongNotFoundError, InternalServerError]`
       
#### events
- `GET /events/{username}/{session_id}`
- response: `event stream`

#### Server related
##### ping:
- `GET /ping`