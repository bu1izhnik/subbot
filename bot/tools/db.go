package tools

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
)

// Trying to get fetcher with providing func, if fails gets random fetcher, if it also fails returns error
func GetFetcher(ctx context.Context, db *orm.Queries, get GetFetcherRequest) (*FetcherParams, error) {
	fetcher, err := get(ctx, db)
	if err != nil {
		randomFetcher, err := db.GetRandomFetcher(ctx)
		if err != nil {
			return nil, err
		}
		return &FetcherParams{
			ID:   randomFetcher.ID,
			Ip:   randomFetcher.Ip,
			Port: randomFetcher.Port,
		}, nil
	}
	return &FetcherParams{
		ID:   fetcher.ID,
		Ip:   fetcher.Ip,
		Port: fetcher.Port,
	}, nil
}

func LeastFull(ctx context.Context, db *orm.Queries) (*FetcherParams, error) {
	f, err := db.GetLeastFullFetcher(ctx)
	if err != nil {
		return nil, err
	}
	return &FetcherParams{
		ID:   f.ID,
		Ip:   f.Ip,
		Port: f.Port,
	}, nil
}

func MostFull(ctx context.Context, db *orm.Queries) (*FetcherParams, error) {
	f, err := db.GetMostFullFetcher(ctx)
	if err != nil {
		return nil, err
	}
	return &FetcherParams{
		ID:   f.ID,
		Ip:   f.Ip,
		Port: f.Port,
	}, nil
}
