package main

import (
	"bufio"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/pprof"

	_ "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/sashankg/hold/dao"
	"github.com/sashankg/hold/graphql"
	"github.com/sashankg/hold/handlers"
	"github.com/sashankg/hold/util"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	goose.SetBaseFS(migrations)

	schemaDb, err := NewSchemaDb()
	if err != nil {
		panic(err)
	}

	recordDb, err := NewRecordDb()
	if err != nil {
		panic(err)
	}

	daoObj := dao.NewDao(schemaDb, recordDb)

	validator := graphql.NewValidator(daoObj)
	resolver := graphql.NewResolver(daoObj)

	privKey, err := util.LoadIdentity("server.key")
	if err != nil {
		panic(err)
	}

	relayAddrInfo, err := peer.AddrInfoFromString(
		"/ip4/127.0.0.1/tcp/4002/ws/p2p/QmNpBvAKWrjigDHP4Mn3LpqCmin5F2K9TiVFoFGTC6ayV3",
	)
	if err != nil {
		panic(err)
	}
	println("Relay ID", relayAddrInfo.ID.String())
	println("Relay Addr", relayAddrInfo.Addrs[0].String())

	host, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.EnableAutoRelayWithStaticRelays(
			[]peer.AddrInfo{
				*relayAddrInfo,
			},
		),
	)
	if err != nil {
		panic(err)
	}

	println("Host ID", host.ID().String())
	for _, addr := range host.Network().ListenAddresses() {
		println("Listening on", addr.String())
	}

	reservation, err := client.Reserve(context.Background(), host, *relayAddrInfo)
	if err != nil {
		panic(err)
	}
	// println("Reservation", reservation.Addrs[0].String())
	println("Reservation", reservation.LimitData)

	mux := http.NewServeMux()
	mux.Handle("/graph", handlers.NewGraphqlHandler(validator, resolver))
	mux.Handle("/upload", handlers.NewUploadHandler(&handlers.FnvHasher{}))

	server := NewServer(mux)

	listener, err := gostream.Listen(host, "/http/1.1")
	if err != nil {
		panic(err)
	}

	go func() {
		err := server.Serve(listener)
		if err != nil {
			println(err.Error())
		}
	}()

	util.WaitForInterrupt()

	err = server.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}

	host.Close()
}

func NewServer(serveMux *http.ServeMux) *http.Server {
	serveMux.Handle("/debug/pprof/", pprof.Handler("heap"))
	return &http.Server{
		Handler:  serveMux,
		ErrorLog: log.Default(),
	}
}

func NewSchemaDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "schema.db")
	if err != nil {
		return nil, err
	}
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return nil, err
	}
	return db, err
}

func NewRecordDb() (*sql.DB, error) {
	return sql.Open("sqlite3", "record.db")
}

func NewHttpProxy(h http.Handler) network.StreamHandler {
	return func(s network.Stream) {
		defer s.Close()
		// Create a new buffered reader, as ReadRequest needs one.
		// The buffered reader reads from our stream, on which we
		// have sent the HTTP request
		buf := bufio.NewReader(s)
		b, err := io.ReadAll(buf)
		fmt.Println("%w", b)

		// Read the HTTP request from the buffer
		req, err := http.ReadRequest(buf)
		if err != nil {
			s.Reset()
			fmt.Println(err)
			return
		}
		defer req.Body.Close()

		h.ServeHTTP(&ResponseWriter{
			Writer: s,
		}, req)
	}
}

type ResponseWriter struct {
	header      http.Header
	wroteHeader bool
	Writer      io.Writer
}

var _ http.ResponseWriter = &ResponseWriter{}

// Header implements http.ResponseWriter.
func (r *ResponseWriter) Header() http.Header {
	return r.header
}

// WriteHeader implements http.ResponseWriter.
func (r *ResponseWriter) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}
	r.Writer.Write(
		[]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, http.StatusText(statusCode))),
	)
	for k, v := range r.header {
		for _, vv := range v {
			r.Writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, vv)))
		}
	}
	r.Writer.Write([]byte("\r\n"))
}

// Write implements http.ResponseWriter.
func (r *ResponseWriter) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.Writer.Write(b)
}
