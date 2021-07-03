package double

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewDocker() *mongo.Database {
	return clientFromContainer().Database(uuid.NewString())
}

func Purge() {
	m.Lock()
	defer m.Unlock()

	if purge != nil {
		purge()
	}
}

var (
	m      sync.Mutex
	client *mongo.Client
	purge  func()
)

func clientFromContainer() *mongo.Client {
	m.Lock()
	defer m.Unlock()

	if client != nil {
		return client
	}

	ctx := context.Background()

	keyFile := tempKeyfile()
	defer os.Remove(keyFile)

	pool, err := dockertest.NewPool(os.Getenv("DOCKER_HOST"))
	if err != nil {
		panic(err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "4.4",
		Mounts:     []string{keyFile + ":/etc/keyfile"},
		Cmd:        []string{"mongod", "--replSet", "rs", "--keyFile", "/etc/keyfile"},
		Env: []string{
			"MONGO_INITDB_ROOT_USERNAME=mongo",
			"MONGO_INITDB_ROOT_PASSWORD=mongo",
		},
	})
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := recover(); err != nil {
			_ = pool.Purge(resource)
			panic(err)
		}
	}()

	if err := resource.Expire(60); err != nil {
		panic(err)
	}

	err = pool.Retry(func() error {
		port := resource.GetPort("27017/tcp")
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://mongo:mongo@localhost:%s", port)).SetDirect(true))
		if err != nil {
			return err
		}
		res := client.Database("admin").RunCommand(ctx, bson.M{"replSetInitiate": bson.M{
			"_id": "rs",
			"members": []bson.M{
				{"_id": 0, "host": "localhost:27017"},
			},
		}})
		if err := res.Err(); err != nil {
			return err
		}
		res = client.Database("admin").RunCommand(ctx, bson.M{"replSetGetStatus": 1})
		if err := res.Err(); err != nil {
			return err
		}
		var repl struct {
			OK int `bson:"ok"`
		}
		_ = res.Decode(&repl)
		if repl.OK != 1 {
			return errors.New("replset not ready")
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	purge = func() {
		_ = client.Disconnect(ctx)
		_ = pool.Purge(resource)
	}

	return client
}

func tempKeyfile() string {
	file, err := ioutil.TempFile("", "key")
	if err != nil {
		panic(err)
	}
	_, _ = file.WriteString("testkeyfile")
	if err := file.Chmod(0600); err != nil { // nolint
		_ = os.Remove(file.Name())
		panic(err)
	}
	return file.Name()
}