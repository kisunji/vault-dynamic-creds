package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	flag "github.com/spf13/pflag"

	vault "github.com/hashicorp/vault/api"
)

var (
	service    = flag.StringP("service", "s", "", "(required) Service name")
	role       = flag.StringP("role", "r", "", "(required) Role name (ro|rw)")
	vaultToken = flag.StringP("token", "t", os.Getenv("VAULT_TOKEN"),
		"Vault token to use for authentication. Defaults to env var VAULT_TOKEN.")
	vaultAddr = flag.StringP("addr", "a", os.Getenv("VAULT_ADDR"),
		"Vault address to use for API calls. Defaults to env var VAULT_ADDR.")
)

func main() {
	flag.Parse()
	if *vaultToken == "" {
		log.Fatal("VAULT_TOKEN is required")
	}
	if *vaultAddr == "" {
		log.Fatal("VAULT_ADDR is required")
	}
	if *service == "" {
		log.Fatal("service is required")
	}
	if *role == "" {
		log.Fatal("role is required")
	}
	config := vault.DefaultConfig()
	config.Address = *vaultAddr
	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Fetching dynamic credentials")
	secret, err := client.Logical().Read(fmt.Sprintf("%s/database/postgres/creds/%s_role", *service, *role))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Username: ", secret.Data["username"])
	log.Println("Password: ", secret.Data["password"])

	input := &vault.LifetimeWatcherInput{
		Secret:        secret,
		RenewBehavior: vault.RenewBehaviorErrorOnErrors,
	}

	watcher, err := client.NewLifetimeWatcher(input)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting renewal routine. This will block your terminal until the program is terminated.")
	go watcher.Start()
	defer watcher.Stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-watcher.DoneCh():
			if err != nil {
				log.Fatal(err)
			}
		case <-sigs:
			log.Fatal("Received signal, shutting down")
		case <-watcher.RenewCh():
			log.Println("Successfully renewed")
		}
	}
}
