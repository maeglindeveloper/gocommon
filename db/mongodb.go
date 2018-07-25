package db

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

// Define the interface for database management
type DataBaseManager interface {
	Init() error
	Connect() error
	FindOne(collectionName string, filter *bson.Document) (*bson.Document, error)
}

// Define the database configuration structure
type Config struct {
	Addr string
	Name string
}

// Define the Mongo database manager structure
type MongoDataBaseManager struct {
	Client *mongo.Client
	Config Config
	IsInit bool
}

var (
	ErrConfigCorrupted = errors.New("The database configuration is corrupted")
	ErrUnableToConnect = errors.New("")
)

// Init the mongo db database manager
func (m *MongoDataBaseManager) Init() error {

	flag.StringVar(&m.Config.Addr, "mongo.db.addr", "mongodb://localhost:27017", "MongoDB address, including port number")
	flag.StringVar(&m.Config.Addr, "mongo.db.database", "", "MongoDB database name")

	// Use environment variables, if set. Flags have priority over Env vars.
	if addr := os.Getenv("MONGO_DB_ADDR"); addr != "" {
		m.Config.Addr = addr
	}
	if name := os.Getenv("MONGO_DB_DATABASE"); name != "" {
		m.Config.Name = name
	}

	// Check consistency
	if m.Config.Addr == "" || m.Config.Name == "" {
		return ErrConfigCorrupted
	}

	return nil
}

// Define the Mongo database manager structure
func (m *MongoDataBaseManager) Connect() error {

	var err error
	m.Client, err = mongo.NewClient(m.Config.Addr)

	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		return err
	}

	err = m.Client.Connect(context.TODO())
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		return err
	}

	return nil
}

func (m *MongoDataBaseManager) FindOne(collectionName string, filter *bson.Document) (*bson.Document, error) {

	var err error

	// Check if is init
	if !m.IsInit {
		return nil, err
	}

	collection := m.Client.Database(m.Config.Name).Collection(collectionName)
	var result = bson.NewDocument()
	err = collection.FindOne(context.TODO(), filter).Decode(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetMongoDataBaseManager() (*MongoDataBaseManager, error) {
	var db *MongoDataBaseManager = new(MongoDataBaseManager)
	var err error

	if err = db.Init(); err != nil {
		return nil, err
	}
	if err = db.Connect(); err != nil {
		return nil, err
	}

	return db, nil
}
