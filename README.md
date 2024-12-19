# Game service

Main service for handling game logic and state.

## Configuration

See `defaults.env` for default configuration values.
Environment variables will override the defaults.

## API

When `GAME_TEST_SERVER` is set to `true`, the service will serve a testing webpage and also have endpoints `/new` and `/list` for creating and listing games. Both endpoins are GET and take no parameters.

Regardless of the value of `GAME_TEST_SERVER`, the service always has one HTTP endpoint, `/game/{gameID}`, which is a WebSocket endpoint.


### Connecting to a game

Create a new WebSocket connection to `/game/{gameID}` where `{gameID}` is the ID of the game you want to connect to (all IDs are listed on the `/list` endpoint).

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
        "name":"PLAYER_NAME",
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
    "name":"PLAYER_NAME",
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

## gRPC

The service also has a gRPC endpoint for creating and listing games. These endpoints are always available, regardless of the value of `GAME_TEST_SERVER`. See the [proto file](rso-comms/game.proto) for more information.