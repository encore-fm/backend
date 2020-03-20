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
- response: `{"username": "<username>", "secret": "<secret>", "is_admin": false, "score": <score>}`
#### Admin related
##### login: 
- `POST /admin/login` 
- request: `{"username": "<username>", "password": "<password>"}`
- response: `{"username": "<username>", "secret": "<secret>", "is_admin": true, "score": <score>}`
#### Server related
- ping: `GET /ping`