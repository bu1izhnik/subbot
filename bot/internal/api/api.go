package api

import (
	"encoding/json"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type Api struct {
	server *http.Server
	db     *orm.Queries
}

func Init(db *orm.Queries, port string) *Api {
	api := &Api{db: db}
	router := chi.NewRouter()
	router.Post("/{fetcherID}", api.registerFetcher)
	api.server = &http.Server{
		Handler: router,
		Addr:    ":" + port,
	}
	return api
}

func (api *Api) Run() error {
	return api.server.ListenAndServe()
}

func (api *Api) registerFetcher(w http.ResponseWriter, r *http.Request) {
	fetcherIDStr := chi.URLParam(r, "fetcherID")
	fetcherID, err := strconv.ParseInt(fetcherIDStr, 10, 64)
	if err != nil {
		responseWithError(w, http.StatusBadRequest, err.Error())
	}
	type fetcherParams struct {
		Phone string `json:"phone"`
		IP    string `json:"ip"`
		Port  string `json:"port"`
	}
	decoder := json.NewDecoder(r.Body)
	fetcher := &fetcherParams{}
	err = decoder.Decode(fetcher)
	if err != nil {
		responseWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := api.db.AddFetcher(r.Context(), orm.AddFetcherParams{
		ID:    fetcherID,
		Phone: fetcher.Phone,
		Ip:    fetcher.IP,
		Port:  fetcher.Port,
	})
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responseWithJSON(w, http.StatusOK, res)
}
