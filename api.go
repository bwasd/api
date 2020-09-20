// Package api contains routines that expose a RESTful API via HTTP
package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/bwasd/api/storage"
)

var (
	logger            *log.Logger
	apiHealth         uint32
	ProductRepository *API
)

var (
	ErrAttribute   = errors.New("Attribute ill-formed")
	ErrSKUName     = errors.New("SKU name ill-formed")
	ErrSKUConflict = errors.New("SKU name conflict")
)

func checkContentType(expect string, r *http.Request) error {
	ct := r.Header.Get("Content-Type")
	// Not an error: Content-Type was not specified with an empty request body
	if ct == "" {
		if r.Body == http.NoBody {
			return nil
		}
	}
	// Otherwise, check that Content-Type is a parseable media type value
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return err
	}
	// And matches the expected content-type
	if mt != expect {
		return fmt.Errorf("bad Content-Type: expected %s, got %s", expect, mt)
	}

	return nil
}

// ProductOps provides external operations for persisting products to underlying
// storage.
type ProductOps interface {
	// Lookup looks up a product for the given SKU
	Lookup(ctx context.Context, sku string) (*Product, error)

	// List returns a list of Product records in the range lo, hi
	List(ctx context.Context, lo, hi int) ([]*Product, error)

	// Store writes or updates a product to storage
	Store(ctx context.Context, sku string, prod *Product) error

	// Remove removes the Product identified by SKU from storage
	// TODO: implement this
	// Remove(ctx context.Context, sku string) error
}

type API struct {
	db *sql.DB
}

// NewRepository returns a new Repository object using the given operations
func NewRepository() (*API, error) {
	db, err := storage.Open()
	if err != nil {
		return nil, err
	}
	return &API{db: db}, nil
}

func handleListProducts(w http.ResponseWriter, req *http.Request) {
	qs := func(q string) int { v, _ := strconv.Atoi(req.URL.Query().Get(q)); return v }
	lo := qs("lo")
	hi := qs("hi")

	prod, err := ProductRepository.List(req.Context(), lo, hi)
	if err != nil {
		panic(err)
	}
	v, err := json.Marshal(prod)
	if err != nil {
		panic(err)
	}
	if _, err = w.Write(v); err != nil {
		panic(err)
	}
}

func handleCreateProduct(w http.ResponseWriter, req *http.Request) {
	if !(req.Method == "POST" || req.Method == "PUT") {
		e := &apiError{
			Code:    http.StatusMethodNotAllowed,
			Message: "Supported methods: POST, PUT",
		}
		e.Write(w)
		return
	}

	var prod Product
	if err := json.NewDecoder(req.Body).Decode(&prod); err != nil {
		e := &apiError{
			Code:    http.StatusUnprocessableEntity,
			Message: http.StatusText(http.StatusUnprocessableEntity),
		}
		e.Write(w)
		return
	}

	err := ProductRepository.Store(req.Context(), &prod)
	if err != nil {
		panic(err)
	}

	// Return a URI with the create/updated resource's location
	loc := fmt.Sprintf("/products/%s/", prod.SKU)
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(`{"product":"` + loc + `"}`)); err != nil {
		panic(err)
	}
}

func (r *API) Lookup(ctx context.Context, sku string) (*Product, error) {
	var prod Product
	if err := r.db.
		QueryRowContext(ctx, `
			SELECT
				p.sku,
				p.attrs
			FROM test.product AS p
			WHERE p.sku=$1
			`,
			sku).Scan(&prod.SKU, &prod.Attrs); err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, err
		}
		return nil, err
	}
	return &prod, nil
}

func (r *API) Store(ctx context.Context, prod *Product) error {
	tx, err := r.db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO test.product(
			sku,
			attrs
		) VALUES( $1, $2 )
		ON CONFLICT (sku) DO UPDATE SET attrs=$2
		`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	// TODO: this re-marshals attributes from map[string]string -> []bytes but
	// it might be better to let the database do this processing with
	// json_build_object()
	v, err := json.Marshal(prod.Attrs)
	if err != nil {
		panic(err)
	}

	if _, err := stmt.ExecContext(ctx, prod.SKU, v); err != nil {
		panic(err)
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}
	return nil
}

func (r *API) List(ctx context.Context, lo, hi int) ([]*Product, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			p.sku,
			p.attrs
		FROM test.product AS p
		WHERE p.id >=$1 LIMIT $2`,
		lo, hi)
	if err != nil {
		return nil, err
	}

	prods := make([]*Product, 0)
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.SKU, &p.Attrs); err != nil {
			panic(err)
		}
		prods = append(prods, &p)
	}
	return prods, nil
}

func handleGetProduct(w http.ResponseWriter, req *http.Request) {
	sku := req.URL.Query().Get("sku")
	prod, err := ProductRepository.Lookup(req.Context(), sku)
	if errors.Is(err, sql.ErrNoRows) {
		e := &apiError{
			Code:    http.StatusNotFound,
			Message: "Product does not exist",
		}
		if err := e.Write(w); err != nil {
			panic(err)
		}
		return
	}

	v, err := json.Marshal(prod)
	if err != nil {
		panic(err)
	}
	if _, err = w.Write(v); err != nil {
		panic(err)
	}
}

func index() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// The "/" pattern matches everything.
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

// logHandler is a simple route logging middleware that captures basic request
// info
func logHandler(l *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				ip, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					l.Printf("ip: %q", r.RemoteAddr)
				}
				l.Println(r.Method, r.URL.Path, ip)
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type uaKey struct{}

// handle wraps most API endpoints, adding request context and header fields
// that apply to all requests
func handle(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), uaKey{}, r.Header.Get("User-Agent"))
		r = r.WithContext(ctx)
		if err := checkContentType("application/json", r); err != nil {
			e := &apiError{
				Code:    http.StatusUnsupportedMediaType,
				Message: http.StatusText(http.StatusUnsupportedMediaType),
			}
			e.Write(w)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Inhibit processing the request if the API is in an unhealthy state.
		// In future, the server could also send a Retry-After header field for
		// clients to wait and retry (with exponential backoff).
		if atomic.LoadUint32(&apiHealth) == 0 {
			e := &apiError{
				Code: http.StatusServiceUnavailable,
				Message: http.StatusText(http.
					StatusServiceUnavailable),
			}
			e.Write(w)
			return
		}
		fn(w, r)
	}
}

// Server creates an http.Server capable of handling requests against the
// Product REST API.
func Server(addr string) *http.Server {
	logger = log.New(os.Stdout, "api: ", log.Lmicroseconds|log.LUTC)
	logger.Println("Server started")

	mux := http.NewServeMux()
	// The hacky workaround below registers two route handlers for each endpoint
	// to make endpoints trailing-slash invariant.
	mux.Handle("/product/get/", handle(handleGetProduct))
	mux.Handle("/product/get", handle(handleGetProduct))
	mux.Handle("/product/create/", handle(handleCreateProduct))
	mux.Handle("/product/create", handle(handleCreateProduct))
	mux.Handle("/product/list/", handle(handleListProducts))
	mux.Handle("/product/list", handle(handleListProducts))
	mux.Handle("/healthz", http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			if h := atomic.LoadUint32(&apiHealth); h == 0 {
				w.WriteHeader(http.StatusServiceUnavailable)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
	mux.Handle("/", index())

	srv := &http.Server{
		Addr:         addr,
		Handler:      logHandler(logger)(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return srv
}

func Main(addr string) {
	srv := Server(addr)
	var err error
	ProductRepository, err = NewRepository()
	if err != nil {
		logger.Printf("%v: db cluster unavailable!", err)
		atomic.StoreUint32(&apiHealth, 0)
	} else {
		atomic.StoreUint32(&apiHealth, 1)
	}

	// Start a go routine to receive interrupt signals; the server will attempt
	// to shutdown gracefully, without interrupting active connections.
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		ch := <-sigint
		logger.Printf("Signal %s received; server shutting down\n", ch)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Printf("Server failed to shutdown gracefully: %v", err)
			os.Exit(2)
		}
	}()

	logger.Printf("Server listening on address: %s\n", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("%v", err)
	}

	logger.Println("Server stopped")
	os.Exit(0)
}
