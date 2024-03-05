package main

import (
	"false.kr/WebChecker-Node/files"
	"false.kr/WebChecker-Node/routes"
)

func main() {
	app := routes.Router()
	files.Init()
	app.Listen(":" + files.Config.Port)
}
