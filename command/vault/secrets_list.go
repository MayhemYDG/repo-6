package vault

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/seatgeek/hashi-helper/command/vault/helper"
	"github.com/seatgeek/hashi-helper/config"
	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"
)

// SecretsList ...
func SecretsList(c *cli.Context) error {
	if c.GlobalBool("remote") {
		return secretListRemote(c)
	}

	return secretListLocal(c)
}

func secretListLocal(c *cli.Context) error {
	config, err := config.NewConfigFromCLI(c)
	if err != nil {
		return err
	}

	for _, secret := range config.VaultSecrets {
		logger := log.WithFields(log.Fields{
			"env":    secret.Environment.Name,
			"app":    secret.Application.Name,
			"secret": secret.Key,
		})

		for k, v := range secret.VaultSecret.Data {
			logger.Printf("%s = %s", k, v)
		}

		log.Println()
	}

	spew.Dump(config)

	return nil
}

// secretListRemote ...
func secretListRemote(c *cli.Context) error {
	secrets := helper.IndexRemoteSecrets(c.GlobalString("environment"), c.GlobalInt("concurrency"))

	if c.Bool("detailed") {
		printDetailedSecrets(secrets, c.GlobalInt("concurrency"))
		return nil
	}

	log.Println()
	for _, secret := range secrets {
		log.Infof("%s @ %s: %s", secret.Application.Name, secret.Environment.Name, secret.Path)
	}

	return nil
}

func printDetailedSecrets(paths config.VaultSecrets, concurrency int) {
	secrets, err := helper.ReadRemoteSecrets(paths, concurrency)
	if err != nil {
		log.Fatal(err)
	}

	for _, secret := range secrets {
		log.Println()
		log.Infof("%s @ %s: %s", secret.Application.Name, secret.Environment.Name, secret.Path)

		for k, v := range secret.VaultSecret.Data {
			switch vv := v.(type) {
			case string:
				log.Info("  ⇛ ", k, " = ", vv)
			case int:
				log.Println("  ⇛ ", k, " = ", vv)
			default:
				log.Panic("  ⇛ ", k, "is of a type I don't know how to handle")
			}
		}
	}
}
