package main

import (
	"fmt"
	log "github.com/inconshreveable/log15"
	"io"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

var (
	DefaultUpgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	DefaultDialer = websocket.DefaultDialer
)

type WebsocketProxy struct {
	Director func(incoming *http.Request, out http.Header)

	Backend func(*http.Request) *url.URL

	Upgrader *websocket.Upgrader

	Dialer *websocket.Dialer
}

func WsProxyHandler(target []*url.URL) http.Handler {
	return WsProxy(target)
}

func WsProxy(target []*url.URL) *WebsocketProxy {
	backend := func(r *http.Request) *url.URL {
		max := len(backend.Websocket) - 1
		pointer.Websocket = point(pointer.Websocket, max)
		u := *backend.Websocket[pointer.Websocket]
		u.Fragment = r.URL.Fragment
		u.Path = r.URL.Path
		u.RawQuery = r.URL.RawQuery
		log.Info("Websocket endpoint", "url", u.String(), "index", pointer.Websocket)
		return &u
	}
	return &WebsocketProxy{Backend: backend}
}

// ServeHTTP implements the http.Handler that proxies WebSocket connections.
func (w *WebsocketProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if w.Backend == nil {
		log.Error("websocketproxy: backend function is not defined")
		http.Error(rw, "internal server error (code: 1)", http.StatusInternalServerError)
		return
	}

	backendURL := w.Backend(req)
	if backendURL == nil {
		log.Error("websocketproxy: backend URL is nil")
		http.Error(rw, "internal server error (code: 2)", http.StatusInternalServerError)
		return
	}

	dialer := w.Dialer
	if w.Dialer == nil {
		dialer = DefaultDialer
	}

	connBackend, resp, err := dialer.Dial(backendURL.String(), nil)
	if err != nil {
		log.Error("websocketproxy: couldn't dial to remote backend url", "error", err)
		if resp != nil {
			if err := copyResponse(rw, resp); err != nil {
				log.Error("websocketproxy: couldn't write response after failed remote backend handshake", "error", err)
			}
		} else {
			http.Error(rw, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		}
		return
	}
	defer connBackend.Close()

	upgrader := w.Upgrader
	if w.Upgrader == nil {
		upgrader = DefaultUpgrader
	}

	connPub, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Error("websocketproxy: couldn't upgrade", "error", err)
		return
	}
	defer connPub.Close()

	errClient := make(chan error, 1)
	errBackend := make(chan error, 1)
	replicateWebsocketConn := func(dst, src *websocket.Conn, errc chan error) {
		for {
			msgType, msg, err := src.ReadMessage()
			if err != nil {
				m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
				if e, ok := err.(*websocket.CloseError); ok {
					if e.Code != websocket.CloseNoStatusReceived {
						m = websocket.FormatCloseMessage(e.Code, e.Text)
					}
				}
				errc <- err
				dst.WriteMessage(websocket.CloseMessage, m)
				break
			}
			err = dst.WriteMessage(msgType, msg)
			if err != nil {
				errc <- err
				break
			}
		}
	}

	go replicateWebsocketConn(connPub, connBackend, errClient)
	go replicateWebsocketConn(connBackend, connPub, errBackend)

	var message string
	select {
	case err = <-errClient:
		message = "websocketproxy: Error when copying from backend to client: %v"
	case err = <-errBackend:
		message = "websocketproxy: Error when copying from client to backend: %v"

	}
	if e, ok := err.(*websocket.CloseError); !ok || e.Code == websocket.CloseAbnormalClosure {
		log.Error("Websocket error", "msg", message, "error", err)
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func copyResponse(rw http.ResponseWriter, resp *http.Response) error {
	copyHeader(rw.Header(), resp.Header)
	rw.WriteHeader(resp.StatusCode)
	defer resp.Body.Close()

	_, err := io.Copy(rw, resp.Body)
	return err
}
