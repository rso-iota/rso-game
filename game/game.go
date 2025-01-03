package game

import (
	"encoding/json"
	"math"
	"math/rand/v2"
	"net/http"
	"time"
	"rso-game/config"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var PLAYER_SPEED = 150
var NUM_FOOD = 150
var gameConfig *config.Config

// SetConfig sets the global game configuration
func SetConfig(cfg *config.Config) {
	gameConfig = cfg
}

type Game struct {
	ID string

	players map[*Player]*PlayerData
	food    []Food

	connect    chan *Player
	disconnect chan *Player

	messages_in  chan PlayerMessage
	moveMessages map[*Player]MoveMessage

	previousTime time.Time
	botClient    *BotClient
	minPlayers   int
}

type PlayerMessage struct {
	Player  *Player
	Message []byte
}

type PlayerData struct {
	PlayerName string `json:"playerName"`
	Alive      bool   `json:"alive"`
	Circle     Circle `json:"circle"`
}

type Food struct {
	Index  int    `json:"index"`
	Circle Circle `json:"circle"`
}

type Circle struct {
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Radius float32 `json:"radius"`
}

type GameState struct {
	Players []PlayerData `json:"players"`
	Food    []Food      `json:"food"`
}

func (l Circle) overlap(other Circle) bool {
	sqrDist := (l.X-other.X)*(l.X-other.X) + (l.Y-other.Y)*(l.Y-other.Y)
	sqrRadSum := (l.Radius + other.Radius) * (l.Radius + other.Radius)
	return sqrDist < sqrRadSum
}

func (l *Circle) addArea(radius float32) {
	l.Radius = float32(math.Sqrt(float64(l.Radius*l.Radius + radius*radius)))
}

var runningGames = make(map[string]*Game)

func (g *Game) manageBots() {
	if g.botClient == nil {
		return
	}

	// Only try to add bots if we have a minimum player requirement
	if g.minPlayers <= 0 {
		return
	}

	if len(g.players) < g.minPlayers {
		botsNeeded := g.minPlayers - len(g.players)
		for i := 0; i < botsNeeded; i++ {
			_, err := g.botClient.CreateBot(g.ID, "medium")
			if err != nil {
				log.WithError(err).Warn("Failed to create bot, skipping bot management")
				return // Skip further bot creation attempts if we encounter an error
			}
		}
	}
}

func (g *Game) Run() {
	ticker := time.NewTicker(30 * time.Millisecond)
	defer ticker.Stop()

	botCheckTicker := time.NewTicker(5 * time.Second)
	defer botCheckTicker.Stop()

	for {
		select {
		case player := <-g.connect:
			log.Printf("Player connected to game %s", g.ID)
			g.players[player] = &PlayerData{}
		case player := <-g.disconnect:
			if p, ok := g.players[player]; ok {
				log.Printf("Player disconnected from game %s", g.ID)
				delete(g.players, player)
				close(player.send)

				playerLeftMsg := PlayerLeftMessage{
					Type: "playerLeft",
					Data: *p,
				}

				g.broadcast(playerLeftMsg)
			}
		case message := <-g.messages_in:
			g.handleMessage(message)
		case <-botCheckTicker.C:
			g.manageBots()
		case time := <-ticker.C:
			g.loop(time)
		}
	}
}

func (g *Game) loop(time time.Time) {
	delta := time.Sub(g.previousTime).Seconds()
	g.previousTime = time

	updatedPlayers := make([]PlayerData, 0, len(g.players))
	for player, move := range g.moveMessages {
		if !g.players[player].Alive {
			continue
		}
		playerData := g.players[player]
		magnitude := math.Sqrt(float64(move.X*move.X + move.Y*move.Y))

		// Skip movement update if magnitude is 0
		if magnitude == 0 {
			continue
		}

		playerData.Circle.X += move.X * float32(delta) * float32(PLAYER_SPEED) / float32(magnitude)
		playerData.Circle.Y += move.Y * float32(delta) * float32(PLAYER_SPEED) / float32(magnitude)

		updatedPlayers = append(updatedPlayers, *playerData)
	}
	clear(g.moveMessages)

	// Check for collisions with food
	foodsToChange := make([]int, 0)
	for i, food := range g.food {
		for _, player := range g.players {
			if food.Circle.overlap(player.Circle) {
				player.Circle.addArea(food.Circle.Radius)
				foodsToChange = append(foodsToChange, i)
				break
			}
		}
	}

	updatedFood := make([]Food, len(foodsToChange))
	for i, foodInd := range foodsToChange {
		food := Food{
			Index: foodInd,
			Circle: Circle{
				X:      rand.Float32() * 800,
				Y:      rand.Float32() * 600,
				Radius: 5,
			}}
		g.food[foodInd] = food
		updatedFood[i] = food
	}

	// Check for collisions with other players
	for player, playerData := range g.players {
		for otherPlayer, otherPlayerData := range g.players {
			if player == otherPlayer || !playerData.Alive || !otherPlayerData.Alive {
				continue
			}

			if playerData.Circle.overlap(otherPlayerData.Circle) {
				if playerData.Circle.Radius > otherPlayerData.Circle.Radius {
					playerData.Circle.addArea(otherPlayerData.Circle.Radius)
					otherPlayerData.Alive = false
				} else {
					otherPlayerData.Circle.addArea(playerData.Circle.Radius)
					playerData.Alive = false
				}
				updatedPlayers = append(updatedPlayers, *playerData)
				updatedPlayers = append(updatedPlayers, *otherPlayerData)
			}
		}
	}

	if len(updatedPlayers) > 0 {
		state := GameStateMessage{
			Type: "update",
			Data: GameState{
				Players: updatedPlayers,
				Food:    updatedFood,
			},
		}

		g.broadcast(state)
	}
}

func (g *Game) broadcast(message interface{}) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Error("Error marshalling message")
		return
	}

	for player := range g.players {
		select {
		case player.send <- bytes:
			// Message sent successfully
		default:
			// Channel is full, skip this message
			log.WithField("player", g.players[player].PlayerName).Warn("Skipping message - send channel full")
		}
	}
}

func (g *Game) sendTo(player *Player, message interface{}) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Error("Error marshalling message")
		return
	}

	select {
	case player.send <- bytes:
		// Message sent successfully
	default:
		// Channel is full, skip this message
		log.WithField("player", g.players[player].PlayerName).Warn("Skipping message - send channel full")
	}
}

func (g *Game) broadcastExcept(message interface{}, except *Player) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Error("Error marshalling message")
		return
	}

	for player := range g.players {
		if player == except {
			continue
		}
		select {
		case player.send <- bytes:
			// Message sent successfully
		default:
			// Channel is full, skip this message
			log.WithField("player", g.players[player].PlayerName).Warn("Skipping message - send channel full")
		}
	}
}

func (g *Game) handleMessage(playerMessage PlayerMessage) {
	var msg Message
	err := json.Unmarshal(playerMessage.Message, &msg)
	if err != nil {
		log.WithError(err).Error("Error unmarshalling message")
		return
	}

	if msg.Type == "join" {
		var join JoinMessage
		err := json.Unmarshal(msg.Data, &join)
		if err != nil {
			log.WithError(err).Error("Error unmarshalling join message")
			return
		}

		g.players[playerMessage.Player] = &PlayerData{
			PlayerName: join.PlayerName,
			Alive:      true,
			Circle: Circle{
				X:      rand.Float32() * 800,
				Y:      rand.Float32() * 600,
				Radius: 10,
			},
		}

		state := GameStateMessage{
			Type: "gameState",
			Data: GameState{
				Players: g.onlinePlayers(),
				Food:    g.food,
			},
		}

		g.sendTo(playerMessage.Player, state)

		spawnMsg := SpawnMessage{
			Type: "spawn",
			Data: *g.players[playerMessage.Player],
		}
		g.broadcastExcept(spawnMsg, playerMessage.Player)
	} else if msg.Type == "move" {
		player, ok := g.players[playerMessage.Player]
		if !ok {
			return
		}
		if !player.Alive {
			return
		}
		var move MoveMessage
		err := json.Unmarshal(msg.Data, &move)
		if err != nil {
			log.WithError(err).Error("Error unmarshalling move message")
			return
		}
		g.moveMessages[playerMessage.Player] = move
	}
}

func (g *Game) onlinePlayers() []PlayerData {
	players := make([]PlayerData, len(g.players))
	i := 0
	for _, player := range g.players {
		players[i] = *player
		i++
	}

	return players
}

func CreateGame() string {
	if gameConfig == nil {
		log.Error("Game configuration not set. Call SetConfig() before creating games.")
		return ""
	}

	id := uuid.New().String()
	food := make([]Food, NUM_FOOD)
	for i := 0; i < NUM_FOOD; i++ {
		food[i] = Food{
			Index: i,
			Circle: Circle{
				X:      rand.Float32() * 800,
				Y:      rand.Float32() * 600,
				Radius: 5,
			},
		}
	}

	var botClient *BotClient
	var minPlayers int

	// Only try to set up bot client if bot service URL is configured
	if gameConfig.BotServiceURL != "" {
		var err error
		botClient, err = NewBotClient(gameConfig.BotServiceURL)
		if err != nil {
			log.WithError(err).Info("Bot service unavailable, game will run without bots")
		} else {
			// Only set minPlayers if we successfully connected to the bot service
			minPlayers = gameConfig.MinPlayers
		}
	}

	game := &Game{
		ID:           id,
		players:      make(map[*Player]*PlayerData),
		food:         food,
		connect:      make(chan *Player),
		disconnect:   make(chan *Player),
		messages_in:  make(chan PlayerMessage),
		moveMessages: make(map[*Player]MoveMessage),
		previousTime: time.Now(),
		botClient:    botClient,
		minPlayers:   minPlayers,
	}
	runningGames[id] = game

	go game.Run()
	log.Printf("Creating new game %s. There are now %d running games", id, len(runningGames))

	return id
}

func RunningGameIDs() []string {
	ids := make([]string, len(runningGames))

	i := 0
	for id := range runningGames {
		ids[i] = id
		i++
	}

	return ids
}

func HandleNewConnection(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("gameID")
	game, ok := runningGames[gameID]
	if !ok {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	// TODO: check if the player can join (he went through the lobby service)
	serveWebSocket(game, w, r)
}