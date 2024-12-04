const buttons = document.getElementById("buttons");
const canvas = document.getElementById("gameCanvas");
const ctx = canvas.getContext("2d");

const playerName = "player" + Math.floor(Math.random() * 1000);
const otherPlayers = {};
const food = {}
let player;

let ws;

function updatePlayers(playerData) {
    for (const p of playerData) {
        if (p.playerName === playerName) {
            player = { x: p.circle.x, y: p.circle.y, r: p.circle.radius, alive: p.alive };
        } else {
            otherPlayers[p.playerName] = { x: p.circle.x, y: p.circle.y, r: p.circle.radius, alive: p.alive };
        }
    }
}

function updateFood(foodData) {
    for (const f of foodData) {
        food[f.index] = { x: f.circle.x, y: f.circle.y, radius: f.circle.radius };
    }
}

const handleMessage = (msg) => {
    const message = JSON.parse(msg.data);

    switch (message.type) {
        case "spawn":
            otherPlayers[message.data.playerName] = {
                x: message.data.circle.x,
                y: message.data.circle.y,
                r: message.data.circle.radius,
                alive: message.data.alive
            };

            break;
        case "gameState":
        case "update":
            updatePlayers(message.data.players);
            updateFood(message.data.food);
            break;
        case "playerLeft":
            console.log("Player left: ", message.data.playerName);
            delete otherPlayers[message.data.playerName];
            break;
    }
};

fetch("http://localhost:8080/list").then(
    async (res) => {
        const games = await res.json();
        for (const game of games) {
            const button = document.createElement("button");
            button.innerText = game;
            button.addEventListener("click", async () => {
                if (ws !== undefined) {
                    ws.close();
                }

                ws = new WebSocket("ws://localhost:8080/connect/" + game);
                ws.onopen = () => {
                    ws.send(JSON.stringify({ type: "join", data: { playerName } }));
                };
                ws.onmessage = handleMessage;
            });
            buttons.appendChild(button);
        }
    }
);



const keyPressed = {};

window.addEventListener("keydown", (e) => {
    keyPressed[e.key] = true;
});

window.addEventListener("keyup", (e) => {
    keyPressed[e.key] = false;
});

let start;
function step(timestamp) {
    if (start === undefined) {
        start = timestamp;
    }
    const elapsed = timestamp - start;

    ctx.clearRect(0, 0, canvas.width, canvas.height);

    let dx = 0;
    let dy = 0;
    if (keyPressed["w"]) {
        dy -= 1;
    }
    if (keyPressed["s"]) {
        dy += 1;
    }
    if (keyPressed["a"]) {
        dx -= 1;
    }
    if (keyPressed["d"]) {
        dx += 1;
    }

    if (dx !== 0 || dy !== 0) {
        ws.send(JSON.stringify({ type: "move", data: { x: dx, y: dy } }));
    }

    if (player?.alive) {
        ctx.fillStyle = "green";
        ctx.beginPath();
        ctx.arc(player.x, player.y, player.r, 0, 2 * Math.PI);
        ctx.fill();

    }
    for (const playerName in otherPlayers) {
        if (!otherPlayers[playerName].alive) {
            continue;
        }
        ctx.fillStyle = "red";

        ctx.beginPath();
        ctx.arc(otherPlayers[playerName].x, otherPlayers[playerName].y, otherPlayers[playerName].r, 0, 2 * Math.PI);
        ctx.fill();
    }

    for (const f in food) {
        ctx.fillStyle = "blue";
        ctx.beginPath();
        ctx.arc(food[f].x, food[f].y, food[f].radius, 0, 2 * Math.PI);
        ctx.fill();
    }
    requestAnimationFrame(step);
}

requestAnimationFrame(step);



