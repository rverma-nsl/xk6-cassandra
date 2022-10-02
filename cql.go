package cql

import (
	"fmt"
	"log"
	"strings"

	"github.com/gocql/gocql"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
)

// init is called by the Go runtime at application startup.
func init() {
	modules.Register("k6/x/cassandra", new(RootModule))
}

type (
	// RootModule is the global module instance that will create module instances for each VU
	RootModule struct{}

	// CQL represents an instance of the JS module
	CQL struct {
		// vu provides methods for accessing internal k6 objects for a VU
		vu modules.VU
	}
)

var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &CQL{}
)

// NewModuleInstance implements the modules.Module interface returning a new instance for each VU.
func (m *RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &CQL{vu: vu}
}

func (cql *CQL) Exports() modules.Exports {
	return modules.Exports{Default: cql}
}

type Config struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Timeout  string `json:"timeout"`
	DC       string `json:"dc"`
}

func (cql *CQL) Connect(config Config) (*gocql.Session, error) {
	rt := cql.vu.Runtime()
	if config.URL == "" {
		common.Throw(rt, fmt.Errorf("url is required"))
	}
	if config.Timeout == "" {
		config.Timeout = "10s"
	}
	hosts := strings.Split(config.URL, ",")
	cluster := gocql.NewCluster(hosts...)
	cluster.Authenticator = gocql.PasswordAuthenticator{Username: config.Username, Password: config.Password}
	if config.DC != "" {
		cluster.PoolConfig.HostSelectionPolicy = gocql.DCAwareRoundRobinPolicy(config.DC)
	}
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (*CQL) Exec(session *gocql.Session, stmt string) error {
	if err := session.Query(stmt).Exec(); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (*CQL) CheckTable(session *gocql.Session, keyspace string, table string) bool {
	ks, err := session.KeyspaceMetadata(keyspace)
	if err != nil {
		log.Fatal("can't get session metadata")
	}
	if _, ok := ks.Tables[table]; ok {
		log.Println("Table already exists, skipping table creation")
		return ok
	}
	return false
}

func (*CQL) Insert(session *gocql.Session, keyspace string, table string, col []string, vals []string) error {
	stmt := fmt.Sprintf("INSERT INTO %s.%s ( %s ) VALUES ( %s );", keyspace, table, strings.Join(col, ","), strings.Join(vals, ","))
	if err := session.Query(stmt).Exec(); err != nil {
		log.Fatal(err)
	}
	return nil
}

/*func (*CQL) Query(session *gocql.Session, table string) error {
	var id gocql.UUID
	var text string

	if err := session.Query(`SELECT id, text FROM ?  WHERE timeline = ? LIMIT 1`, table, "me").Consistency(gocql.One).Scan(&id, &text); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Tweet:", id, text)
	return nil
}
func (*CQL) Scan(session *gocql.Session, table string) error {
	var id gocql.UUID
	var text string

	scanner := session.Query(`SELECT id, text FROM tweet WHERE timeline = ?`,
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
}*/
