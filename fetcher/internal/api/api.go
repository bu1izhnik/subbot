package api

import (
	"github.com/BulizhnikGames/subbot/fetcher/internal/fetcher"
	"net/http"
)

type Api struct {
	f *fetcher.Fetcher
}

func Init(f *fetcher.Fetcher) *Api {
	return &Api{f: f}
}

func (api *Api) HandleSubscribe(w http.ResponseWriter, r *http.Request) {

}
