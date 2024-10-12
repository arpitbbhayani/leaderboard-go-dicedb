package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/dicedb/dicedb-go"
	"github.com/gorilla/websocket"
)

var (
	client   *dicedb.Client
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	watchTopics map[string]string = map[string]string{}

	watchConn      *dicedb.WatchConn
	watchCh        <-chan *dicedb.WatchResult
	connectedUsers []*websocket.Conn
)

type Score struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

func main() {
	client = dicedb.NewClient(&dicedb.Options{
		Addr: "localhost:7379",
	})
	connectedUsers = []*websocket.Conn{}
	go initWatch()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/update", handleUpdate)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func initWatch() {
	ctx := context.Background()
	watchConn = client.WatchConn(ctx)
	if watchConn == nil {
		log.Fatal("Failed to create watch")
		return
	}
	watchCh = watchConn.Channel()

	res, err := watchConn.ZRangeWatch(ctx, "leaderboard", "0", "5", "REV", "WITHSCORES")
	if err != nil {
		log.Println("Failed to start watch:", err)
		return
	}
	watchTopics[res.Fingerprint] = "global_leaderboard"

	for {
		select {
		case msg := <-watchCh:
			switch watchTopics[msg.Fingerprint] {
			case "global_leaderboard":
				var scores []Score
				for _, z := range msg.Data.([]dicedb.Z) {
					scores = append(scores, Score{
						Name:  z.Member.(string),
						Score: int(z.Score),
					})
				}

				for _, conn := range connectedUsers {
					if err := conn.WriteJSON(scores); err != nil {
						log.Println("WebSocket write error:", err)
						return
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	connectedUsers = append(connectedUsers, conn)
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	var score Score
	if err := json.NewDecoder(r.Body).Decode(&score); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := client.ZAdd(r.Context(), "leaderboard", dicedb.Z{
		Score:  float64(score.Score),
		Member: score.Name,
	}).Err()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
