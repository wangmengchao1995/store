package main

import (
	_ "bj2qFresh/routers"
	"github.com/astaxie/beego"
	_ "bj2qFresh/models"
)

func main() {
	beego.Run()
}

