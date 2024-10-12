module github.com/arpitbbhayani/leaderboard-go-dicedb

go 1.22.1

require (
	github.com/dicedb/dicedb-go v0.0.0
	github.com/gorilla/websocket v1.5.3
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)

replace github.com/dicedb/dicedb-go => ../dicedb-go
