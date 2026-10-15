package main

import (
	"context"
	"fmt"
	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
	"log"
	"os"
)

func main() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	folder, err := client.Assets.Folders.Create(context.Background(), bird.AssetsFoldersCreateParams{Name: "Campaign images"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(folder.Id)
	for asset, err := range client.Assets.List(context.Background(), bird.AssetsListParams{Kind: "image"}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(asset.Id, *asset.Url)
	}
}
