package game

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand/v2"
	"net/http"
	"rso-game/circuitbreaker"
	"rso-game/config"
	"rso-game/grpc"
	___ "rso-game/grpc/lobby"
	"rso-game/nats"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/sony/gobreaker/v2"
)

var conf config.Config
var PLAYER_SPEED = 500
var NUM_FOOD = 150
var TICK_RATE = 30 * time.Millisecond

func SetGlobalConfig(c config.Config) {
	conf = c
}

type Game struct {
	ID string

	humanPlayers  atomic.Int32
	botPlayers    atomic.Int32
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
	minPlayers   int

	terminate bool

	gameLoopTicker     *time.Ticker
	isPaused           bool
	gameTerminateTimer *time.Timer
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
var runningGamesLock sync.RWMutex

func (g *Game) manageBots() {
	if circuitbreaker.BotsBreaker.State() == gobreaker.StateOpen {
		return
	}

	botsNeeded := g.minPlayers - len(g.players)
	for i := 0; i < botsNeeded; i++ {
		token := uuid.New().String()
		botName := "bot-" + token[:4]

		log.Info("Requesting a new bot ", botName, " for game ", g.ID)
		err := grpc.CreateBot(g.ID, botName, token, "medium")
		if err != nil {
			log.WithError(err).Warn("Failed to create bot, skipping bot management")
			return // Skip further bot creation attempts if we encounter an error
		}
		g.botTokens[token] = botName
	}

	if botsNeeded < 0 {
		log.Info("Disconnecting ", -botsNeeded, " bots from game ", g.ID)
		go g.disconnectBots(-botsNeeded)
	}
}

func (g *Game) informLobby() {
	if circuitbreaker.LobbyBreaker.State() == gobreaker.StateOpen {
		return
	}

	players := make([]*___.Player, 0, len(g.players))
	for _, player := range g.players {
		players = append(players, &___.Player{
			Username: player.PlayerName,
			Size:     player.Circle.Radius,
		})
	}

	err := grpc.NotifyLiveData(g.ID, players)
	if err != nil {
		log.WithError(err).Warn("Failed to notify lobby about game state")
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

	grpc.NotifyGameDeleted(g.ID, ___.EndGameReason_INACTIVITY)

	deleteRunningGame(g.ID)
	log.Info("Game ", g.ID, " stopped")

	runtime.Goexit()
}

func (g *Game) Run() {
	g.gameLoopTicker = time.NewTicker(TICK_RATE)
	defer g.gameLoopTicker.Stop()

	g.gameTerminateTimer = time.NewTimer(time.Hour)
	g.gameTerminateTimer.Stop()

	backupTicker := time.NewTicker(5 * time.Second)
	defer backupTicker.Stop()

	maintenanceTicker := time.NewTicker(2 * time.Second)
	defer maintenanceTicker.Stop()
	replayTicker := time.NewTicker(500 * time.Millisecond)
	defer replayTicker.Stop()

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
					g.humanPlayers.Add(-1)
					log.Info("There are now ", g.humanPlayers.Load(), " human players in game ", g.ID)
				} else {
					g.botPlayers.Add(-1)
				}

				g.broadcast(playerLeftMsg)
			}
		case message := <-g.messages_in:
			g.handleMessage(message)
		case time := <-g.gameLoopTicker.C:
			g.loop(time)
			if g.terminate {
				g.Terminate()
			}
		case <-maintenanceTicker.C:
			if !g.isPaused {
				g.manageBots()
				g.informLobby()
			}
		case <-backupTicker.C:
			if !g.isPaused {
				state := g.GetGameState()
				SaveToBackup(g.ID, state)
			}
		case <-replayTicker.C:
			if !g.isPaused {
				state := g.GetGameState()
				SendToReplays(g.ID, state)
			}
		case <-g.gameTerminateTimer.C:
			log.WithField("id", g.ID).Info("Terminating game due to inactivity")
			g.Terminate()
		}

	}
}

func SendToReplays(key string, data GameState) {
	state := GameStateMessage{
		Type: "gameState",
		Data: data,
	}
	stateBytes, err := json.Marshal(state)
	if err != nil {
		log.WithError(err).Error("Failed to marshal game state")
		return
	}

	nats_channel := "game_state." + key
	nats.Publish(nats_channel, stateBytes)
}

func (g *Game) GetGameState() GameState {
	return GameState{
		Players: g.onlinePlayers(),
		Food:    g.food,
	}
}

func (g *Game) loop(t time.Time) {
	delta := t.Sub(g.previousTime).Seconds()
	g.previousTime = t

	if g.humanPlayers.Load() <= 0 && g.botPlayers.Load() > 0 {
		go g.disconnectBots(int(g.botPlayers.Load()))
	}

	if len(g.players) == 0 {
		log.WithField("id", g.ID).Info("No players left in game, terminating game in ", conf.TerminateMinutes, " minutes")
		g.isPaused = true
		g.gameLoopTicker.Stop()
		g.gameTerminateTimer.Reset(time.Duration(conf.TerminateMinutes) * time.Minute)
	}

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

		slow := float32(2.0 / math.Sqrt(float64(playerData.Circle.Radius)))

		playerData.Circle.X += move.X * float32(delta) * float32(PLAYER_SPEED) / float32(magnitude) * slow
		playerData.Circle.Y += move.Y * float32(delta) * float32(PLAYER_SPEED) / float32(magnitude) * slow

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

func (g *Game) disconnectBots(n int) {
	i := 0
	for player, playerData := range g.players {
		if i >= n {
			break
		}

		if playerData.IsBot {
			g.disconnect <- player
			i++
		}
	}
}

func (g *Game) broadcast(message interface{}) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Error("Error marshalling message")
		return
	}

	for player := range g.players {
		player.send <- bytes
	}
}

func (g *Game) sendTo(player *Player, message interface{}) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Error("Error marshalling message")
		return
	}

	player.send <- bytes
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
		player.send <- bytes
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
			if g.isPaused {
				log.Info("Resuming game ", g.ID, ", stopping termination timer")
				g.gameLoopTicker.Reset(TICK_RATE)
				g.gameTerminateTimer.Stop()
				g.isPaused = false
			}

			log.Info("There are now ", g.humanPlayers.Load(), " human players in game ", g.ID)
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

func addRunningGame(game *Game) {
	runningGamesLock.Lock()
	runningGames[game.ID] = game
	runningGamesLock.Unlock()
}

func deleteRunningGame(id string) {
	runningGamesLock.Lock()
	delete(runningGames, id)
	runningGamesLock.Unlock()
}

func getGame(id string) (*Game, error) {
	runningGamesLock.RLock()
	defer runningGamesLock.RUnlock()

	game, ok := runningGames[id]
	if !ok {
		return nil, errors.New("Game not found")
	}

	return game, nil
}

func CreateGameStruct(id string, players []PlayerData, food []Food) *Game {
	game := &Game{
		ID:            id,
		players:       make(map[*Player]*PlayerData),
		loadedPlayers: players,
		food:          food,
		connect:       make(chan *Player),
		disconnect:    make(chan *Player),
		messages_in:   make(chan PlayerMessage),
		moveMessages:  make(map[*Player]MoveMessage),
		previousTime:  time.Now(),
		minPlayers:    conf.MinPlayers,
		botTokens:     make(map[string]string),
	}

	return game
}

func RestoreFromBackup() {
	games := LoadBackup()

	for id, state := range games {
		game := CreateGameStruct(id, state.Players, state.Food)
		go game.Run()

		addRunningGame(game)
		log.WithField("id", id).Info("Restored game from backup")
	}
}

func DeleteGame(id string) error {
	game, err := getGame(id)
	if err != nil {
		return err
	}

	game.Stop()
	deleteRunningGame(id)
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

	game := CreateGameStruct(id, []PlayerData{}, food)

	go game.Run()
	addRunningGame(game)
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
	game, err := getGame(gameID)
	if err != nil {
		log.Error("Game not found")
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	token := r.URL.Query().Get("token")

	bot, botInfo := IsBot(game, token)
	if bot {
		game.botPlayers.Add(1)
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
	game.humanPlayers.Add(1)
	log.Info("Player connected to game", playerInfo)

	serveWebSocket(playerInfo, game, w, r)
}
