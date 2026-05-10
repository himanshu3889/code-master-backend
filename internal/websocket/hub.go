package appWebsocket

import (
	"github.com/himanshu3889/code-master-backend/internal/store"
	"github.com/jmoiron/sqlx"
	"sync"
)

var (
	websocketStore     *store.Store // to call the store methods
	websocketStoreOnce sync.Once
	userClient         *Client // let's fixed the single user client
	clientMu           sync.Mutex
)

func InitializeWebsocketStore(db *sqlx.DB) {
	websocketStoreOnce.Do(func() {
		websocketStore = store.New(db)
	})
}
