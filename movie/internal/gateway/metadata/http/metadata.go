package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mkvy/movies-app/metadata/pkg/model"
	"github.com/mkvy/movies-app/movie/internal/gateway"
	"github.com/mkvy/movies-app/pkg/discovery"
	"log"
	"math/rand"
	"net/http"
)

// Gateway defines an HTTP gateway for a movie metadata service.
type Gateway struct {
	registry discovery.Registry
}

// New creates a new HTTP gateway for a movie metadata service.
func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

// Get gets movie metadata by a movie id.
func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	url, err := getUrl(ctx, g.registry)
	if err != nil {
		return nil, err
	}
	log.Printf("Calling metadata service. Request GET " + url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	values := req.URL.Query()
	values.Add("id", id)
	req.URL.RawQuery = values.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, gateway.ErrNotFound
	} else if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("non-2xx response: %v", resp)
	}
	var v *model.Metadata
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	return v, nil
}

// getUrl returns random instance url from service registry.
func getUrl(ctx context.Context, registry discovery.Registry) (string, error) {
	addrs, err := registry.ServiceAddresses(ctx, "metadata")
	log.Println(err)
	if err != nil {
		return "", err
	}
	url := "http://" + addrs[rand.Intn(len(addrs))] + "/metadata"
	return url, nil
}
