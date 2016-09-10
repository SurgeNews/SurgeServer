package main

import (
	controller "github.com/SurgeNews/SurgeServer/controller"
)

func main() {
	controller.Server.Run(":8080")
}
