package main

import (
	"CHainGate/backend/database"
	"CHainGate/backend/routes"
	"CHainGate/backend/utils"
)

func main() {
	utils.NewOpts() // create utils.Opts (env variables)
	database.Connect()
	routes.Setup()
}
