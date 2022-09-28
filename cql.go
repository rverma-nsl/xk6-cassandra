package xk6_mongo

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/gocql/gocql"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
	"log"
	"strings"
)

func init() {
	modules.Register("k6/x/cassandra", new(cassandraModule))
}

// CassandraModule is the k6 extension for interacting Cassandra endpoints.
type CassandraModule struct {
	vu modules.VU
}

type cassandraModule struct{}

var _ modules.Module = &cassandraModule{}

func (m *cassandraModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &CassandraModule{vu: vu}
}

func (r *CassandraModule) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"Client": r.xclient,
		},
	}
}

// Client is the k6 construct for interacting with cassandra.
type Client struct {
	session *gocql.Session
	cfg     *Config
	vu      modules.VU
}

type Config struct {
	Url        string `json:"url"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Timeout    string `json:"timeout"`
	TenantName string `json:"tenant_name"`
}

// xclient represents
func (r *CassandraModule) xclient(c goja.ConstructorCall) *goja.Object {
	var config Config
	rt := r.vu.Runtime()
	err := rt.ExportTo(c.Argument(0), &config)
	if err != nil {
		common.Throw(rt, fmt.Errorf("Client constructor expects first argument to be Config"))
	}
	if config.Url == "" {
		log.Fatal(fmt.Errorf("url is required"))
	}
	if config.Timeout == "" {
		config.Timeout = "10s"
	}
	if config.TenantName == "" {
		config.TenantName = "tn"
	}
	hosts := strings.Split(config.Url, ",")
	var cluster = gocql.NewCluster(hosts...)
	cluster.Authenticator = gocql.PasswordAuthenticator{Username: config.Username, Password: config.Password}
	cluster.PoolConfig.HostSelectionPolicy = gocql.DCAwareRoundRobinPolicy("AWS_AP_SOUTH_1")

	session, err := cluster.CreateSession()
	if err != nil {
		panic("Failed to connect to cluster")
	}
	return rt.ToValue(&Client{
		session: session,
		cfg:     &config,
		vu:      r.vu,
	}).ToObject(rt)
}

func (c *Client) Insert() error {
	if err := c.session.Query(`INSERT INTO tweet (timeline, id, text) VALUES (?, ?, ?)`,
		"me", gocql.TimeUUID(), "hello world").Exec(); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (c *Client) Select() error {
	var id gocql.UUID
	var text string

	if err := c.session.Query(`SELECT id, text FROM tweet WHERE timeline = ? LIMIT 1`,
		"me").Consistency(gocql.One).Scan(&id, &text); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Tweet:", id, text)
	return nil
}

func (c *Client) Scan() error {
	var id gocql.UUID
	var text string

	scanner := c.session.Query(`SELECT id, text FROM tweet WHERE timeline = ?`,
		"me").Iter().Scanner()
	for scanner.Next() {
		err := scanner.Scan(&id, &text)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Tweet:", id, text)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return nil
}
