package game

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand/v2"
	"net/http"
	"rso-game/config"
	"rso-game/nats"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var conf config.Config
var PLAYER_SPEED = 150
var NUM_FOOD = 150

func SetGlobalConfig(c config.Config) {
	conf = c
}

type Game struct {
	ID string

	humanPlayers  int
	isPaused      bool
	players       map[*Player]*PlayerData
	playersToDel  []*Player
	loadedPlayers []PlayerData
	food          []Food

	botTokens map[string]string

	connect    chan *Player
	disconnect chan *Player

	messages_in  chan PlayerMessage
	moveMessages map[*Player]MoveMessage

	previousTime time.Time
	botClient    *BotClient
	minPlayers   int

	delta     float64
	terminate bool
}

type PlayerMessage struct {
	Player  *Player
	Message []byte
}

type PlayerData struct {
	PlayerId   string
	IsBot      bool   `json:"isBot"`
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
	Food    []Food       `json:"food"`
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
			token := uuid.New().String()
			botName := "bot-" + token[:4]

			log.Info("Requesting a new bot ", botName, " for game ", g.ID)
			_, err := g.botClient.CreateBot(g.ID, botName, token, "medium")
			if err != nil {
				log.WithError(err).Warn("Failed to create bot, skipping bot management")
				return // Skip further bot creation attempts if we encounter an error
			}
			g.botTokens[token] = botName
		}
	}
}

func (g *Game) Stop() {
	g.terminate = true
}

func (g *Game) Terminate() {
	closeMessage := CloseMessage{
		Type:   "close",
		Reason: "Game stopped",
	}

	g.broadcast(closeMessage)
	DeleteBackup(g.ID)

	log.Info("Game ", g.ID, " stopped")
}

func (g *Game) Run() {
	g.manageBots()
	gameLoopTicker := time.NewTicker(50 * time.Millisecond)
	defer gameLoopTicker.Stop()

	botCheckTicker := time.NewTicker(5 * time.Second)
	defer botCheckTicker.Stop()

	backupTicker := time.NewTicker(5 * time.Second)
	defer backupTicker.Stop()

	for {
		select {
		case player := <-g.connect:
			g.players[player] = &PlayerData{
				PlayerName: "TEMP",
				PlayerId:   "TEMP",
				Alive:      false,
				Circle: Circle{
					X:      -10000,
					Y:      -10000,
					Radius: 0,
				},
			}

		case player := <-g.disconnect:
			if p, ok := g.players[player]; ok {
				log.Info("Player disconnected from lobby", g.ID)
				g.playersToDel = append(g.playersToDel, player)

				playerLeftMsg := PlayerLeftMessage{
					Type: "playerLeft",
					Data: *p,
				}

				if !p.IsBot {
					g.humanPlayers--
					log.Info("There are now ", g.humanPlayers, " human players in game ", g.ID)
				}

				if g.humanPlayers <= 0 {
					log.Info("No human players left in game ", g.ID, ", pausing game")
				}

				g.broadcast(playerLeftMsg)
			}
		case message := <-g.messages_in:
			g.handleMessage(message)
		case <-botCheckTicker.C:
			g.manageBots()
		case time := <-gameLoopTicker.C:
			g.delta = time.Sub(g.previousTime).Seconds()
			g.previousTime = time
			if !g.isPaused {
				g.loop()
			}
			if g.terminate {
				g.Terminate()
				return
			}
		case <-backupTicker.C:
			state := g.GetGameState()
			SaveToBackup(g.ID, state)
		}
	}
}

func (g *Game) GetGameState() GameState {
	return GameState{
		Players: g.onlinePlayers(),
		Food:    g.food,
	}
}

func (g *Game) loop() {
	g.isPaused = g.humanPlayers <= 0

	for _, player := range g.playersToDel {
		delete(g.players, player)
		close(player.send)
	}
	g.playersToDel = nil

	updatedPlayers := make([]PlayerData, 0, len(g.players))
	for player, move := range g.moveMessages {
		if p, ok := g.players[player]; !ok || !p.Alive {
			continue
		}
		playerData := g.players[player]
		magnitude := math.Sqrt(float64(move.X*move.X + move.Y*move.Y))

		// Skip movement update if magnitude is 0
		if magnitude == 0 {
			continue
		}
		playerData.Circle.X += move.X * float32(g.delta) * float32(PLAYER_SPEED) / float32(magnitude)
		playerData.Circle.Y += move.Y * float32(g.delta) * float32(PLAYER_SPEED) / float32(magnitude)

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

				if player.IsBot {
					nats.Publish("bot_food", []byte(""))
				} else {
					nats.Publish("food", []byte(player.PlayerId))
				}

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
				var smaller *PlayerData
				var bigger *PlayerData

				if playerData.Circle.Radius > otherPlayerData.Circle.Radius {
					bigger = playerData
					smaller = otherPlayerData
				} else {
					bigger = otherPlayerData
					smaller = playerData
				}

				bigger.Circle.addArea(smaller.Circle.Radius)
				smaller.Alive = false

				if smaller.IsBot {
					nats.Publish("bot_died", []byte(""))
				} else {
					nats.Publish("died", []byte(smaller.PlayerId))
				}

				if bigger.IsBot {
					nats.Publish("bot_kill", []byte(""))
				} else {
					nats.Publish("kill", []byte(bigger.PlayerId))
				}

				updatedPlayers = append(updatedPlayers, *smaller)
				updatedPlayers = append(updatedPlayers, *bigger)
			}
		}
	}

	if len(updatedPlayers) > 0 {
		state := GameStateMessage{
			Type: "update",
			Data: g.GetGameState(),
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

		// Backward compatibility
		if playerMessage.Player.info == (PlayerInfo{}) {
			playerMessage.Player.info = PlayerInfo{
				Username: join.PlayerName,
				Id:       join.PlayerName,
			}
		}

		var data *PlayerData = nil
		for _, player := range g.loadedPlayers {
			if player.PlayerId == playerMessage.Player.info.Id {
				data = &player
				break
			}
		}

		if data == nil {
			data = &PlayerData{
				PlayerName: playerMessage.Player.info.Username,
				PlayerId:   playerMessage.Player.info.Id,
				Alive:      true,
				Circle: Circle{
					X:      rand.Float32() * 800,
					Y:      rand.Float32() * 600,
					Radius: 10,
				},
				IsBot: playerMessage.Player.info.IsBot,
			}
		}

		if !data.IsBot {
			if g.humanPlayers == 0 {
				log.Info("Human player joined game ", g.ID, ", resuming game")
				g.isPaused = false
			}

			g.humanPlayers++
			log.Info("There are now ", g.humanPlayers, " human players in game ", g.ID)
		}

		g.players[playerMessage.Player] = data

		state := GameStateMessage{
			Type: "gameState",
			Data: g.GetGameState(),
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

func CreateGameStruct(id string, players []PlayerData, food []Food, botClient *BotClient, minPlayers int) Game {
	game := Game{
		ID:            id,
		humanPlayers:  0,
		players:       make(map[*Player]*PlayerData),
		loadedPlayers: players,
		food:          food,
		connect:       make(chan *Player),
		disconnect:    make(chan *Player),
		messages_in:   make(chan PlayerMessage),
		moveMessages:  make(map[*Player]MoveMessage),
		previousTime:  time.Now(),
		botClient:     botClient,
		minPlayers:    minPlayers,
		botTokens:     make(map[string]string),
	}

	return game
}

func RestoreFromBackup() {
	games := LoadBackup()

	var botClient *BotClient
	var minPlayers int

	for id, state := range games {
		// Only try to set up bot client if bot service URL is configured
		if conf.BotServiceURL != "" {
			var err error
			botClient, err = NewBotClient(conf.BotServiceURL)
			if err != nil {
				log.WithError(err).Info("Bot service unavailable, game will run without bots")
			} else {
				// Only set minPlayers if we successfully connected to the bot service
				minPlayers = conf.MinPlayers
				log.Info("Bot service available, will run with bots")
			}
		}

		game := CreateGameStruct(id, state.Players, state.Food, botClient, minPlayers)
		runningGames[id] = &game
		go game.Run()
		log.WithField("id", id).Info("Restored game from backup")
	}
}

func DeleteGame(id string) error {
	game, ok := runningGames[id]
	if !ok {
		return errors.New("Game not found")
	}

	game.Stop()
	delete(runningGames, id)
	log.Printf("Deleted game %s. There are now %d running games", id, len(runningGames))

	return nil
}

// func GetPlayerSizes(id string) (*pb.GameData, error) {
// 	game, ok := runningGames[id]
// 	if !ok {
// 		return nil, errors.New("Game not found")
// 	}

// 	data := &pb.GameData{
// 		Players: make([]*pb.Player, 0, len(game.players)),
// 	}

// 	for _, player := range game.players {
// 		data.Players = append(data.Players, &pb.Player{
// 			Username: player.PlayerName,
// 			Size:     player.Circle.Radius,
// 		})
// 	}
// 	return data, nil
// }

func CreateGame() string {
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
	if conf.BotServiceURL != "" {
		var err error
		botClient, err = NewBotClient(conf.BotServiceURL)
		if err != nil {
			log.WithError(err).Info("Bot service unavailable, game will run without bots")
		} else {
			// Only set minPlayers if we successfully connected to the bot service
			minPlayers = conf.MinPlayers
			log.Info("Bot service available, will run with bots")
		}
	}

	game := CreateGameStruct(id, []PlayerData{}, food, botClient, minPlayers)
	runningGames[id] = &game

	go game.Run()
	log.Printf("Creating new game %s. There are now %d running games", id, len(runningGames))

	return id
}

func EnsureGames(num int) {
	for len(runningGames) < num {
		CreateGame()
	}
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

type AuthError struct {
	Message string
	Code    int
}

func (e AuthError) Error() string {
	return e.Message
}

func Authorize(token string) (PlayerInfo, AuthError) {
	if token == "" {
		return PlayerInfo{}, AuthError{"No token provided", http.StatusUnauthorized}
	}

	req, err := http.NewRequest("GET", conf.AuthEndpoint, nil)
	if err != nil {
		return PlayerInfo{}, AuthError{"Failed to create auth request", http.StatusInternalServerError}
	}

	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return PlayerInfo{}, AuthError{"Failed to authorize", http.StatusInternalServerError}
	}

	if res.StatusCode != http.StatusOK {
		return PlayerInfo{}, AuthError{"Failed to authorize", res.StatusCode}
	}

	var playerInfo PlayerInfo
	err = json.NewDecoder(res.Body).Decode(&playerInfo)
	if err != nil {
		return PlayerInfo{}, AuthError{"Failed to decode auth response", http.StatusInternalServerError}
	}
	playerInfo.IsBot = false

	return playerInfo, AuthError{}
}

func IsBot(game *Game, token string) (bool, PlayerInfo) {
	if name, ok := game.botTokens[token]; ok {
		delete(game.botTokens, token)

		return true, PlayerInfo{
			Username: name,
			Id:       name,
			IsBot:    true,
		}
	}

	return false, PlayerInfo{}
}

func HandleNewConnection(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("gameID")
	game, ok := runningGames[gameID]
	if !ok {
		log.Error("Game not found")
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	token := r.URL.Query().Get("token")

	bot, botInfo := IsBot(game, token)
	if bot {
		log.Info("Bot ", botInfo.Username, " connected to game ", gameID)
		serveWebSocket(botInfo, game, w, r)
		return
	}

	playerInfo := PlayerInfo{}
	if conf.RequireAuth {
		data, err := Authorize(token)
		if err != (AuthError{}) {
			http.Error(w, err.Message, err.Code)
			log.WithError(err).Warn("Failed to authorize player")
			return
		}
		playerInfo = data

	}
	log.Info("Player connected to game", playerInfo)

	serveWebSocket(playerInfo, game, w, r)
}
