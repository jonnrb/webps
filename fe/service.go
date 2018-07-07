package fe

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elazarl/go-bindata-assetfs"
	"go.jonnrb.io/webps/fe/assets"
	"go.jonnrb.io/webps/pb"
)

var staticAssets = http.FileServer(&assetfs.AssetFS{
	Asset:     assets.Asset,
	AssetDir:  assets.AssetDir,
	AssetInfo: assets.AssetInfo,
})

func New(be []webpspb.WebPsBackendClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static") {
			staticAssets.ServeHTTP(w, r)
			return
		}

		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if p, ok := w.(http.Pusher); ok {
			if err := p.Push("/static/bootstrap.min.css", nil); err != nil {
				log.Println("Error pushing bootstrap.min.css: %v", err)
			} else if err := p.Push("/static/page.css", nil); err != nil {
				log.Println("Error pushing page.css: %v", err)
			}
		}

		var list webpspb.ListResponse
		for _, cli := range be {
			sub, err := cli.List(ctx, &webpspb.ListRequest{})
			if err != nil {
				http.Error(w, fmt.Sprintf("Error listing containers: %v", err), http.StatusInternalServerError)
				return
			}
			list.Container = append(list.Container, sub.Container...)
		}

		var buf bytes.Buffer

		if err := PageFromList(&list).Render(&buf); err != nil {
			http.Error(w, fmt.Sprintf("Error rendering page: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		if _, err := io.Copy(w, &buf); err != nil {
			log.Printf("Error writing rendered page to response: %v", err)
		}
	})
}
