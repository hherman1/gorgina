package main

import (
	"context"
	"database/sql"
	"embed"
	_ "embed"
	"encoding/csv"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hherman1/gorgina/db/persist"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	dbUrlKey = "DATABASE_URL"
	portKey  = "PORT"
)

// A series of schema setup queries to be run during startup.
//go:embed db/schema.sql
var initDB string

// All of our HTML pages
//go:embed web
var web embed.FS

// The database, useful for initiating transactions
var db *sql.DB

func main() {
	if err := run(); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

// to test locally:
// 'docker run --name gorgina -p 5432:5432 -e POSTGRES_USER=gorgina -e POSTGRES_PASSWORD=gorgina -e POSTGRES_DB=gorgina -d postgres'
// 'PORT=8081 DATABASE_URL=postgresql://gorgina:gorgina@localhost:5432/gorgina go run .'

var queries *persist.Queries

func run() error {
	ctx := context.Background()

	// Parse args
	port, err := strconv.Atoi(os.Getenv(portKey))
	if err != nil {
		return fmt.Errorf("parse port: %w", err)
	}

	// setup DB
	db, err = sql.Open("pgx", os.Getenv(dbUrlKey))
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	_, err = db.ExecContext(ctx, initDB)
	if err != nil {
		return fmt.Errorf("initialize DB tables: %w", err)
	}
	queries = persist.New(db)

	// Setup routes
	http.DefaultServeMux.Handle("/component/putCatalog", HandlerFuncE(handlePutComponent))
	http.DefaultServeMux.Handle("/api/put", HandlerFuncE(handlePut))
	http.DefaultServeMux.Handle("/api/use", HandlerFuncE(handleUse))
	http.DefaultServeMux.Handle("/component/list", HandlerFuncE(handleList))

	http.DefaultServeMux.Handle("/data/catalog.csv", HandlerFuncE(handleCatalog))
	http.DefaultServeMux.Handle("/data/activity.csv", HandlerFuncE(handleActivity))

	contents, err := fs.Sub(web, "web")
	if err != nil {
		return fmt.Errorf("chrooting web dir: %w", err)
	}
	http.DefaultServeMux.Handle("/", http.FileServer(http.FS(contents)))

	// Start server
	return http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}

// Component for creating new catalog entries or editing existing ones
func handlePutComponent(response http.ResponseWriter, req *http.Request) error {
	id := strings.Join(req.URL.Query()["id"], "")
	var c persist.Catalog
	if id != "" {
		// load catalog
		var err error
		c, err = queries.GetCatalog(req.Context(), id)
		if err != nil {
			return fmt.Errorf("loading catalog entry: %w", err)
		}
	}
	_, err := response.Write([]byte(putForm(c)))
	if err != nil {
		return fmt.Errorf("writing response: %w", err)
	}
	return nil
}

// They added a new item
func handlePut(response http.ResponseWriter, req *http.Request) error {
	err := req.ParseForm()
	if err != nil {
		return fmt.Errorf("parse form data: %w", err)
	}
	id := strings.Join(req.Form["id"], "")
	if id == "" {
		id = uuid.NewString()
	}
	title := req.Form["title"]
	description := req.Form["description"]
	category := req.Form["category"]
	brand := req.Form["brand"]
	color := req.Form["color"]
	pattern := req.Form["pattern"]
	priceRaw := strings.Join(req.Form["price"], "")
	price := sql.NullFloat64{}
	if priceRaw != "" {
		d, err := strconv.ParseFloat(priceRaw, 32)
		if err != nil {
			return fmt.Errorf("parse price: %w", err)
		}
		price.Valid = true
		price.Float64 = d
	}
	_, err = queries.PutItem(req.Context(), persist.PutItemParams{
		ID:          id,
		Category:    ns(category),
		Brand:       ns(brand),
		Color:       ns(color),
		Pattern:     ns(pattern),
		Title:       ns(title),
		Description: ns(description),
		Price:       price,
	})
	fmt.Println("<", id, ">")
	if err != nil {
		return fmt.Errorf("saving result: %w", err)
	}

	err = handleList(response, req)
	if err != nil {
		return fmt.Errorf("rendering list view: %w", err)
	}
	return nil
}

// Load the list of all results.
func handleList(response http.ResponseWriter, req *http.Request) error {
	search := strings.Join(req.URL.Query()["search"], " ")
	var cs []persist.Catalog
	var err error
	if search == "" {
		cs, err = queries.ListCatalog(req.Context())
		if err != nil {
			return fmt.Errorf("list catalog: %w", err)
		}
	} else {
		cs, err = queries.SearchCatalog(req.Context(), search)
		if err != nil {
			return fmt.Errorf("search catalog(%v): %w", search, err)
		}
	}
	for _, c := range cs {
		fmt.Println("<", c.ID, ">")
	}
	r, err := listCatalog(cs)
	if err != nil {
		return fmt.Errorf("render catalog: %w", err)
	}
	_, _ = response.Write([]byte(r))
	return nil
}

// Mark an item as used, and render an updated view of it.
func handleUse(response http.ResponseWriter, req *http.Request) error {
	cid := strings.Join(req.URL.Query()["id"], "")
	tx, err := db.BeginTx(req.Context(), nil)
	if err != nil {
		return fmt.Errorf("open transaction: %w", err)
	}
	defer tx.Rollback()

	queries := queries.WithTx(tx)
	t := time.Now().UTC()
	_, err = queries.LogUsage(req.Context(), persist.LogUsageParams{
		ID:  uuid.NewString(),
		CID: cid,
		Ts:  t,
	})
	if err != nil {
		return fmt.Errorf("log usage: %w", err)
	}
	_, err = queries.UpdateLastUsed(req.Context(), persist.UpdateLastUsedParams{
		LastActivity: sql.NullTime{Valid: true, Time: t},
		ID:           cid,
	})
	if err != nil {
		return fmt.Errorf("update last used: %w", err)
	}
	c, err := queries.GetCatalog(req.Context(), cid)
	if err != nil {
		return fmt.Errorf("fetch catalog entry: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Render result
	r, err := loggedCatalogItem(c)
	if err != nil {
		return fmt.Errorf("render item: %w", err)
	}
	_, err = response.Write([]byte(r))
	if err != nil {
		return fmt.Errorf("write result: %w", err)
	}
	return nil
}

// Renders all catalog data as a csv
func handleCatalog(response http.ResponseWriter, req *http.Request) error {
	cs, err := queries.ListCatalog(req.Context())
	if err != nil {
		return fmt.Errorf("list catalog: %w", err)
	}
	w := csv.NewWriter(response)
	err = w.Write([]string{"id", "category", "brand", "color", "pattern", "title", "description", "price", "last_activity"})
	if err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	for _, c := range cs {
		err = w.Write([]string{c.ID,
			strings.TrimSpace(c.Category.String),
			strings.TrimSpace(c.Brand.String),
			strings.TrimSpace(c.Color.String),
			strings.TrimSpace(c.Pattern.String),
			strings.TrimSpace(c.Title.String),
			c.Description.String,
			fmt.Sprintf("%.2f", c.Price.Float64),
			strconv.Itoa(int(c.LastActivity.Time.UnixMilli()))})
		if err != nil {
			return fmt.Errorf("write row (%v): %w", c, err)
		}
	}
	w.Flush()
	if w.Error() != nil {
		return fmt.Errorf("flush csv: %w", w.Error())
	}
	return nil
}

// Renders all activity data as a csv
func handleActivity(response http.ResponseWriter, req *http.Request) error {
	as, err := queries.ListUsage(req.Context())
	if err != nil {
		return fmt.Errorf("list activity: %w", err)
	}
	w := csv.NewWriter(response)
	err = w.Write([]string{"id", "cid", "time"})
	if err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	for _, a := range as {
		err = w.Write([]string{a.ID,
			a.CID,
			strconv.Itoa(int(a.Ts.UnixMilli()))})
		if err != nil {
			return fmt.Errorf("write row (%v): %w", a, err)
		}
	}
	w.Flush()
	if w.Error() != nil {
		return fmt.Errorf("flush csv: %w", w.Error())
	}
	return nil
}

func ns(ss []string) sql.NullString {
	if len(ss) == 0 {
		return sql.NullString{Valid: false}
	}
	s := strings.Join(ss, " ")
	if s == "" {
		return sql.NullString{Valid: false}
	} else {
		return sql.NullString{Valid: true, String: s}
	}
}

// Wraps an HTMX component as a handler... maybe dumb
func component(contents string) http.HandlerFunc {
	return http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {
		_, _ = response.Write([]byte(contents))
	})
}

// Convenience for error handling, for now.
func HandlerFuncE(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {
		err := f(response, req)
		if err != nil {
			response.WriteHeader(502)
			response.Write([]byte("Error: " + err.Error()))
		}
	})
}
