package tests

import (
	"context"
	"log"
	"os"
	"testing"

	usersvcv1 "github.com/mlukasik-dev/faceit-usersvc/gen/go/faceit/usersvc/v1"
	"github.com/mlukasik-dev/faceit-usersvc/internal/appconfig"
	"github.com/mlukasik-dev/faceit-usersvc/internal/controller"
	"github.com/mlukasik-dev/faceit-usersvc/internal/events"
	"github.com/mlukasik-dev/faceit-usersvc/internal/store"
	"go.uber.org/zap"
)

var (
	s        *store.Store
	ctr      usersvcv1.ServiceServer
	testData = struct {
		users []*store.User
	}{
		[]*store.User{
			{FirstName: "John", LastName: "Doe", Email: "john.doe@gmail.com", Country: "UK"},
			{FirstName: "Jane", LastName: "Doe", Email: "jane.doe@gmail.com", Country: "UK"},
			{FirstName: "Jan", LastName: "Kowalski", Email: "jan.kowalski@gmail.com", Country: "PL"},
		},
	}
)

var testConfig = `
mongodb:
  uri: ${MONGODB_URI:?uri was not provided}
`

func TestMain(m *testing.M) {
	if err := appconfig.Init([]byte(testConfig)); err != nil {
		log.Fatal(err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	client, err := store.Connect(appconfig.AppConfig.Mongodb.URI)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	s = store.New(client)
	if err := s.CreateIndexes(context.Background()); err != nil {
		log.Fatal(err)
	}

	if err := s.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	if err := seedDB(); err != nil {
		log.Fatal(err)
	}

	e := events.New()

	ctr = controller.New(s, logger, e)

	code := m.Run()

	unseedDB()

	os.Exit(code)
}

func seedDB() error {
	var users []*store.User
	for _, u := range testData.users {
		user, err := s.CreateUser(context.Background(), u, "123456")
		if err != nil {
			return err
		}
		users = append(users, user)
	}
	testData.users = users
	return nil
}

func unseedDB() error {
	for _, u := range testData.users {
		err := s.DeleteUser(context.Background(), u.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
