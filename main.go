package main

import (
	"log"

	"github.com/archeopternix/go-mediafileinfo"
)

func main() {
	info, err := mediafileinfo.GetMediaInfo("example.mp4")
	if err != nil {
		log.Fatalf("Failed to get media info: %v", err)
	}
	err = mediafileinfo.PrintAVContextJSON(info)
	if err != nil {
		log.Fatalf("Failed to print media info: %v", err)
	}
}
