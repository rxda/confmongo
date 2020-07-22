package confmongo

import (
	"context"
	"fmt"
	"github.com/go-courier/envconf"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

type Mongo struct {
	Host     string `env:",upstream"` // mongodb://simpleUser:simplePass@your.db.ip.address:27017/foo
	Port     int
	User     string           `env:""`
	Password envconf.Password `env:""`
	DB       string
	*options.ClientOptions
}

func (m *Mongo) SetDefaults() {
	if m.Host == "" {
		m.Host = "127.0.0.1"
	}

	if m.Port == 0 {
		m.Port = 27017
	}
}

func (m *Mongo) url() string {
	userAndPassword := ""
	if m.Password != "" && m.User != "" {
		userAndPassword = m.User + ":" + m.Password.String()
	}
	return fmt.Sprintf("mongodb://%s@%s:%d/%s", userAndPassword, m.Host, m.Port, m.DB)
}

func (m *Mongo) conn() *options.ClientOptions {
	clientOption:= options.Client().ApplyURI(m.url()).SetMaxPoolSize(10)
	return clientOption
}

func (m *Mongo) Init() {
	m.SetDefaults()
	r := Retry{Repeats: 5, Interval: envconf.Duration(1 * time.Second)}

	err := r.Do(func() error {
		clientOptions := m.conn()
		m.ClientOptions = clientOptions
		return nil
	})

	if err != nil {
		panic(err)
	}
}

func (m *Mongo) LivenessCheck() map[string]string {
	res := map[string]string{}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, m.ClientOptions)
	if err != nil {
		res[m.Host] = err.Error()
		return res
	}
	defer c.Disconnect(ctx)
	err = c.Ping(ctx, readpref.Primary())
	if err != nil {
		res[m.Host] = err.Error()
	} else {
		res[m.Host] = "ok"
	}
	return res
}

func (m *Mongo) Get(ctx context.Context) (*mongo.Client,error) {
	return mongo.Connect(ctx, m.ClientOptions)
}