package main

import "main/initializers"

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
}
func main() {

}
