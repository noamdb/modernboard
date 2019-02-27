package main

import (
	"flag"
	"fmt"
	
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	log "github.com/sirupsen/logrus"


	"gitlab.com/noamdb/modernboard/config"
	"gitlab.com/noamdb/modernboard/repository"
	"gitlab.com/noamdb/modernboard/tasks"
)

var initialize bool

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	config.InitConfig()

	flag.BoolVar(&initialize, "init", false, "Initilize app for the first time")
	flag.Parse()

	repo := &repository.Repository{}
	repo.Connect(viper.GetString("database_url"))
	t := tasks.Tasks{repo}
	t.Run()

	if initialize {
		initialRun(repo)
		return
	}

	StartServer(repo)
}

func initialRun(repo *repository.Repository) {
	fmt.Println("Initializng the app for the first run")
	addAdmin(repo)
}

func addAdmin(repo *repository.Repository) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(viper.GetString("admin_password")), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	admin := &repository.UserCreate{
		Name:     viper.GetString("admin_name"),
		Password: string(hashedPassword),
		Role:     "admin"}

	repo.CreateUser(*admin)
}