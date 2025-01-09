# Game service

Main service for handling game logic and state.

## Configuration

See `defaults.env` for default configuration values.
Environment variables will override the defaults.

Explanation of configuration values:
- `INTERNAL_HTTP_PORT`: Port on which the internal HTTP server will listen
- `EXTERNAL_HTTP_PORT`: Port on which the external HTTP server will listen
- `TEST_SERVER`: If set to `true`, the service will serve a testing webpage on port `EXTERNAL_HTTP_PORT`
- `NUM_TEST_GAMES`: Number of test games to create on startup
- `LOG_JSON`: If set to `true`, logs will be in JSON format
- `NATS_URL`: URL of the NATS server
- `AUTH_EP`: URL of the authentication service
- `REQUIRE_AUTH`: If set to `true`, the players will have to authenticate before joining a game
- `BACKUP_REDIS_URL`: URL of the Redis server used for backup
- `BOT_SERVICE_URL`: URL of the bot service
- `MIN_PLAYERS`: Minimum number of players, if there are less players than this number, the service will ask for bots to join
- `TERMINATE_MINUTES`: Number of minutes after which the game will be terminated if there are no players

## API

To connect to WebSocket, use the `:EXTERNAL_HTTP_PORT/connect/{gameID}` endpoint.
When `TEST_SERVER` is set to `true`, the service will serve a testing webpage on port `EXTERNAL_HTTP_PORT`.

There is a single endpoint on the internal HTTP server: `:INTERNAL_HTTP_PORT/game`
This endpoint has 3 methods: `GET`, `POST`, and `DELETE`.
- `GET` will return a list of all games (no parameters needed)
- `POST` will create a new game (no parameters needed)
- `DELETE` will delete a game (query parameter `id` is required)


## Connecting to a game

Create a new WebSocket connection to `/connect/{gameID}` where `{gameID}` is the ID of the game you want to connect to (all IDs are listed on the `/list` endpoint).

On connection, send the following JSON object:

```json
{
  "type":"join",
  "data":{
    "playerName":"PLAYER_NAME"
  }
}
```

In response, you will receive a JSON object with the current game state:

```json
{
  "type":"gameState",
  "data":{
    "players":[
      {
        "nameName":"PLAYER_NAME",
        "alive":true,
        "circle":{
          "x":100,
          "y":100,
          "radius":10
        }
      }
    ],
    "food":[
      {
        "index":0,
        "circle":{
          "x":200,
          "y":200,
          "radius":5
        }
      }
    ]
  }
}
```

All other players will receive a message that a new player has joined:

```json
{
  "type":"spawn",
  "data":{
    "nameName":"PLAYER_NAME",
    "alive":true,
    "circle":{
      "x":100,
      "y":100,
      "radius":10
    }
  }
}
```

When the state is changed (eg. a player moves), all players will receive a message with all updates:

```json
{
  "type":"update",
  "data":{
    "players":[
      {
        "playerName":"PLAYER_NAME",
        "alive":true,
        "circle":{
          "x":400,
          "y":400,
          "radius":20
        }
      }
    ],
    "food":[
      {
        "index":65,
        "circle":{
          "x":75,
          "y":52,
          "radius":5
        }
      }
    ]
  }
}
```

In the above example, the player with the name `PLAYER_NAME` is now at position `(400, 400)` and has a radius of `20`. If the player died, the value `alive` will be set to `false`. The piece of food with index 65 got eaten and a new piece of food was spawned at `(75, 52)`, on the same index.

### Moving

To move, send the following JSON object:

```json
{
  "type":"move",
  "data":{
    "x":"dx",
    "y":"dy"
  }
}
```