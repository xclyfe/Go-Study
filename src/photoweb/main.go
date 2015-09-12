package main

import (
	"photoweb/microService"
)

func main() {
	photoService := microService.NewPhotoService()
	photoService.Start()
}
