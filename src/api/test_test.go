package api

import (
	"dpv/dpv/src/repository/dpv"
	"dpv/dpv/src/repository/graph"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	Cleanup()
	os.Exit(exitCode)
}

func Cleanup() {
	var err error
	config, err := dpv.NewConfig("../../config.yml")
	if err != nil {
		log.Fatalf("could not initialise config instance: %s", err)
	}
	c, err := graph.Connect(config, true)
	if err != nil {
		log.Fatalf("could not connect to database: %s", err)
	}
	err = graph.DropTestDatabases(c)
}
