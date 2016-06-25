package socket

import (
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/ppincak/gse/store"
	"net/http"
	"errors"
)

type Server struct {
	// root namespace
	*Namespace
	// map of all namespaces
	namespaces		map[string]*Namespace
	// gorilla websocket upgrader
	upgrader   		*websocket.Upgrader
	// server configuration
	conf         	*ServerConf
	//  store factory
	storeFactory 	socket.StoreFactory
	// flag indicating that the server is running
	isRunning		bool
}

func NewServer(storeFactory socket.StoreFactory, config *ServerConf) *Server {
	if storeFactory == nil {
		storeFactory = socket.NewLocalStore
	}
	if config == nil {
		config = DefaultConf()
	}

	upgrader := &websocket.Upgrader{
		ReadBufferSize: 	config.ReadBufferSize,
		WriteBufferSize: 	config.WriteBufferSize,
	}

	server := &Server {
		upgrader: 		upgrader,
		namespaces:     make(map[string] *Namespace),
		storeFactory: 	storeFactory,
		conf: 			config,
	}
	server.Namespace = rootNamespace(server)
	return server
}

// warning: Thread unsafe
func (server *Server) Run() {
	go server.Namespace.Run()
	server.isRunning = true
}

// warning: Thread unsafe
func (server *Server) Stop() {
	server.Namespace.Stop()
	server.isRunning = false
}

func (server *Server) AddNamespace(namespaceName string) (*Namespace, error) {
	if server.isRunning {
		return nil, errors.New("Server is already running")
	}
	logrus.Infof("Registering namespace: %s ", namespaceName)
	namespace := newNamespace(namespaceName, server)
	server.namespaces[namespaceName] = namespace
	go namespace.Run()
	return namespace, nil
}

func (server *Server) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error(err)
		return
	}

	client := NewClient(server, ws, server.storeFactory())
	server.addClient(client)
	logrus.Infof("Client connection established, sessionId: %s", client.GetSessionId())

	go client.readPump()
	go client.writePump()
}

func (server *Server) GetAllNamespaces() []*Namespace {
	namespaces := make([]*Namespace, len(server.namespaces))

	i := 0
	for _, namespace := range server.namespaces {
		namespaces[i] = namespace
		i++
	}
	return namespaces
}

func (server *Server) addClient(client *Client) {
	server.mtx.Lock()
	server.clients[client.uuid] = client
	server.mtx.Unlock()
	client.addNamespace(server.Namespace)
}

func (server *Server) addNamespaceClient(client *Client, namespaceName string) error {
	server.mtx.Lock()
	defer server.mtx.Unlock()
	namespace, ok := server.namespaces[namespaceName]
	if !ok {
		return errors.New("Namespace doesn't exist")
	}
	namespace.addClient(client)
	return nil
}

func (server *Server) removeClient(client *Client) {
	server.mtx.Lock()
	delete(server.Namespace.clients, client.uuid)
	client.destroy()
	server.mtx.Unlock()
}

func (server *Server) removeNamespaceClient(client *Client, namespaceName string) error {
	server.mtx.Lock()
	defer server.mtx.Unlock()
	namespace, ok := server.namespaces[namespaceName]
	if !ok {
		return errors.New("Namespace doesn't exist")
	}
	namespace.removeClient(client)
	return nil
}