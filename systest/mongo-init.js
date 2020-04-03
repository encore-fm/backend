db.createUser({
    user: "root",
    pwd: "root",
    roles: [ { role: "readWrite", db: "users" } ]
});

db.users.insert({
    _id: "baumanto@1",
    username: "baumanto",
    secret: "secret",
    session_id: "1",
    is_admin: true,
    score: 1,
    spotify_authorized: false,
    auth_token: null,
    auth_state: "4321"
});

db.sessions.insert({
    _id: "1",
    song_list: []
});
