package resources

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"github.com/9spokes/go/db"
	"github.com/9spokes/go/network"
	event "github.com/9spokes/go/services/events"
	"github.com/9spokes/go/types"
	"github.com/go-redis/redis"
	"github.com/gorilla/sessions"
)

// New returns a *types.Resources object with all the downstream resources pre-initialised
func New(opt types.Document) (*types.Resources, error) {

	res := types.Resources{}

	if store, err := initSessionCookie(opt["SessionCookieKey"]); err == nil {
		res.Store = store
	} else {
		return nil, err
	}

	if mongo, err := initMongoDB(opt["DbURL"]); err == nil {
		res.Mongo = mongo
	} else {
		return nil, err
	}

	if redis, err := initRedisDB(opt["CacheHost"]); err == nil {
		res.RedisDB = redis
	} else {
		return nil, err
	}

	if creds, err := initCreds(opt["Creds"]); err == nil {
		res.Creds = creds
	} else {
		return nil, err
	}

	if clients, err := initClients(opt["Clients"]); err == nil {
		res.Clients = clients
	} else {
		return nil, err
	}

	if ev, err := initEvents(opt["EventURL"]); err == nil {
		res.Events = ev
	} else {
		return nil, err
	}

	if keystore, err := initX509Keystore(opt["Keystore"]); err == nil {
		res.Keystore = keystore
	} else {
		return nil, err
	}

	return &res, nil
}

func initSessionCookie(key interface{}) (*sessions.CookieStore, error) {

	if _, ok := key.(string); !ok {
		return nil, fmt.Errorf("The BOARD_SESSION_COOKIE_KEY environment variable is missing")
	}
	return sessions.NewCookieStore([]byte(key.(string))), nil
}

func initMongoDB(url interface{}) (db.MongoDB, error) {

	if url == nil {
		return db.MongoDB{}, nil
	}

	// Connect to MongoDB
	re := regexp.MustCompile("^mongodb://([^:]+:[^@]+@)?([^:]+:[0-9]+)/?")

	if _, ok := url.(string); !ok {
		return db.MongoDB{}, fmt.Errorf("Invalid DB URL")
	}

	if !re.MatchString(url.(string)) {
		return db.MongoDB{}, fmt.Errorf("Invalid MongoDB URL %s", url)
	}
	host := re.FindStringSubmatch(url.(string))[2]
	network.Dial("tcp", host, 120)

	mongo, err := db.Connect(url.(string))
	if err != nil {
		return db.MongoDB{}, fmt.Errorf("Failed to connect to database: %s", err.Error())
	}

	return mongo, nil
}

func initRedisDB(url interface{}) (*redis.Client, error) {

	if url == nil {
		return nil, nil
	}

	if _, ok := url.(string); !ok {
		return nil, fmt.Errorf("Invalid Redis URL")
	}

	redisOpts, err := redis.ParseURL(url.(string))
	if err != nil {
		return nil, fmt.Errorf("Failed to parse Redis URL: %s", err.Error())
	}

	redisdb := redis.NewClient(&redis.Options{
		Addr:         redisOpts.Addr,
		Password:     redisOpts.Password,
		DB:           redisOpts.DB,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})
	_, err = redisdb.Ping().Result()
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Redis: %s", err.Error())
	}

	return redisdb, nil
}

func initCreds(file interface{}) (types.Credentials, error) {

	if file == nil {
		return types.Credentials{}, nil
	}

	var clients types.Credentials

	if _, ok := file.(string); !ok {
		return types.Credentials{}, fmt.Errorf("Invalid file path")
	}

	credsFile, err := os.Open(file.(string))
	if err != nil {
		return types.Credentials{}, fmt.Errorf("Failed to load '%s': %s", file, err.Error())
	}
	defer credsFile.Close()

	byteValue, _ := ioutil.ReadAll(credsFile)
	err = json.Unmarshal(byteValue, &clients)
	if err != nil {
		return types.Credentials{}, fmt.Errorf("Failed to read '%s': %s", file, err.Error())
	}

	return clients, nil

}

func initClients(file interface{}) (map[string]string, error) {

	if file == nil {
		return nil, nil
	}

	clients := make(map[string]string)

	if _, ok := file.(string); !ok {
		return nil, fmt.Errorf("Invalid file path")
	}

	clientsFile, err := os.Open(file.(string))
	if err != nil {
		return nil, fmt.Errorf("Failed to load '%s': %s", file, err.Error())
	}
	defer clientsFile.Close()

	byteValue, _ := ioutil.ReadAll(clientsFile)
	err = json.Unmarshal(byteValue, &clients)
	if err != nil {
		return nil, fmt.Errorf("Failed to read '%s': %s", file, err.Error())
	}

	return clients, nil
}

func initEvents(url interface{}) (*event.Context, error) {

	if url == nil {
		return nil, nil
	}

	if _, ok := url.(string); !ok {
		return nil, fmt.Errorf("Invalid Events URL")
	}

	e, err := event.New(url.(string))
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the Event service: %s", err.Error())
	}

	return e, nil
}

func initX509Keystore(path interface{}) ([]x509.Certificate, error) {

	if _, ok := path.(string); !ok || path.(string) == "" {
		return nil, nil
	}

	pool := make([]x509.Certificate, 0)

	raw, err := ioutil.ReadFile(path.(string))
	if err != nil {
		return nil, fmt.Errorf("while reading keystore %s: %s", path, err.Error())
	}

	for i := 0; ; i++ {
		block, rest := pem.Decode(raw)
		if block == nil {
			break
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			//logger.Warningf("Failed to load certificate #%d", i+1)
		} else {
			pool = append(pool, *cert)
		}
		raw = rest
	}

	return pool, nil
}