package dbresolver

import (
	"fmt"
	"log"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive
	. "github.com/onsi/gomega"    //nolint:revive
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID     uint
	Name   string
	Orders []Order
}

type Product struct {
	ID   uint
	Name string
}

type Order struct {
	ID      uint
	OrderNo string
	UserID  uint
}

func TestDBRESOLVER(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DBRESOLVER Suite")
}

var (
	pool      *dockertest.Pool
	resources = map[string]*dockertest.Resource{}

	dsn13306 = DSN{
		Path:     "localhost",
		Port:     13306,
		Dbname:   "dbresolver",
		Username: "dbresolver",
		Password: "dbresolver",
	}
	dsn13307 = DSN{
		Path:     "127.0.0.1",
		Port:     13307,
		Dbname:   "dbresolver",
		Username: "dbresolver",
		Password: "dbresolver",
	}
	dsn13308 = DSN{
		Path:     "localhost",
		Port:     13308,
		Dbname:   "dbresolver",
		Username: "dbresolver",
		Password: "dbresolver",
	}
	dsn13309 = DSN{
		Path:     "127.0.0.1",
		Port:     13309,
		Dbname:   "dbresolver",
		Username: "dbresolver",
		Password: "dbresolver",
		Config:   "charset=utf8&parseTime=True&loc=Local",
	}
)

func runMySQL(port string) {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "8.0",
		Env: []string{
			"MYSQL_DATABASE=dbresolver",
			"MYSQL_USER=dbresolver",
			"MYSQL_PASSWORD=dbresolver",
			`MYSQL_RANDOM_ROOT_PASSWORD="yes"`,
		},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"3306/tcp": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = false
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err := resource.Expire(600); err != nil { // Tell docker to hard kill the container in 120 seconds
		log.Fatalf("Could not expire resource: %s", err)
	}

	waitForStart(port)
	resources[port] = resource
}

func waitForStart(port string) {
	if err := pool.Retry(func() error {
		_, err := getDB(port)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
}

func getDB(port string) (*gorm.DB, error) {
	DB, err := gorm.Open(mysql.Open(fmt.Sprintf("dbresolver:dbresolver@tcp(localhost:%s)/dbresolver?charset=utf8&parseTime=True&loc=Local", port)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Default().Println("failed to open connection to mysql, err:", err)
		return nil, err
	}
	return DB, nil
}

var _ = BeforeSuite(func() {
	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 180 * time.Second

	for _, port := range []string{"13306", "13307", "13308", "13309"} {
		runMySQL(port)
		db, err := getDB(port)
		Expect(err).To(BeNil())
		_ = db.AutoMigrate(&User{}, &Order{}, &Product{})
		user := User{Name: fmt.Sprintf("%v", port)}
		db.Create(&user)
		db.Create(&Product{Name: fmt.Sprintf("%v", port)})
		db.Create(&Order{OrderNo: fmt.Sprintf("%v", port), UserID: user.ID})
	}
})

var _ = Describe("The DBRESOLVER repository", Serial, func() {
	Context("config validate", Serial, Ordered, func() {
		AfterEach(func() {
			Expect(Close()).Should(Succeed())
		})
		It("no source or replica", func() {
			db, err := getDB("13306")
			Expect(err).To(BeNil())
			err = db.Use(
				Register(Config{
					Sources:             []DSN{},
					Replicas:            []DSN{},
					HostResolveInterval: 2 * time.Second,
					HealthCheckInterval: 2 * time.Second,
					HealthCheckTimeout:  2 * time.Second,
					DBStatusLogInterval: 10 * time.Second,
					MaxOpenConns:        10,
					MaxIdleConns:        10,
					Policy:              PolicyFrom("roundRobin"),
					TraceResolverMode:   true,
				}))
			Expect(err.Error()).To(ContainSubstring("no source or replica"))
		})
		It("only one source is allowed", func() {
			db, err := getDB("13306")
			Expect(err).To(BeNil())
			err = db.Use(
				Register(Config{
					Sources:             []DSN{dsn13306, dsn13307},
					Replicas:            []DSN{},
					HostResolveInterval: 2 * time.Second,
					HealthCheckInterval: 2 * time.Second,
					HealthCheckTimeout:  2 * time.Second,
					DBStatusLogInterval: 10 * time.Second,
					MaxOpenConns:        10,
					MaxIdleConns:        10,
					Policy:              PolicyFrom("roundRobin"),
					TraceResolverMode:   true,
				}))
			Expect(err.Error()).To(ContainSubstring("only one source is allowed"))
		})
		It("source and replica can't be the same", func() {
			db, err := getDB("13306")
			Expect(err).To(BeNil())
			err = db.Use(
				Register(Config{
					Sources:             []DSN{dsn13306},
					Replicas:            []DSN{dsn13306},
					HostResolveInterval: 2 * time.Second,
					HealthCheckInterval: 2 * time.Second,
					HealthCheckTimeout:  2 * time.Second,
					DBStatusLogInterval: 10 * time.Second,
					MaxOpenConns:        10,
					MaxIdleConns:        10,
					Policy:              PolicyFrom("roundRobin"),
					TraceResolverMode:   true,
				}))
			Expect(err.Error()).To(ContainSubstring("source and replica can't be the same"))
		})
		It("one mysql backend server conflict in source and replica", func() {
			db, err := getDB("13306")
			Expect(err).To(BeNil())
			err = db.Use(
				Register(Config{
					Sources:             []DSN{dsn13306},
					Replicas:            []DSN{dsn13306, dsn13307, dsn13308},
					HostResolveInterval: 2 * time.Second,
					HealthCheckInterval: 2 * time.Second,
					HealthCheckTimeout:  2 * time.Second,
					DBStatusLogInterval: 10 * time.Second,
					MaxOpenConns:        10,
					MaxIdleConns:        10,
					Policy:              PolicyFrom("roundRobin"),
					TraceResolverMode:   true,
				}))
			Expect(err.Error()).To(ContainSubstring("one mysql backend server conflict in source and replica"))
		})
	})
	Context("multiple registers", Serial, Ordered, func() {
		var db *gorm.DB
		BeforeAll(func() {
			// init db and use dbresolver plugin
			fmt.Println("init db and use dbresolver plugin")
			_db, err := getDB("13306")
			db = _db
			Expect(err).To(BeNil())
			err = db.Use(
				Register(Config{
					Sources:             []DSN{dsn13306},
					Replicas:            []DSN{dsn13307, dsn13308, dsn13309},
					HostResolveInterval: 2 * time.Second,
					HealthCheckInterval: 2 * time.Second,
					HealthCheckTimeout:  2 * time.Second,
					DBStatusLogInterval: 10 * time.Second,
					MaxOpenConns:        10,
					MaxIdleConns:        10,
					Policy:              PolicyFrom("roundRobin"),
					TraceResolverMode:   true,
				}).Register(Config{
					Sources:             []DSN{dsn13308},
					Replicas:            []DSN{dsn13309},
					HostResolveInterval: 2 * time.Second,
					HealthCheckInterval: 2 * time.Second,
					HealthCheckTimeout:  2 * time.Second,
					DBStatusLogInterval: 10 * time.Second,
					MaxOpenConns:        10,
					MaxIdleConns:        10,
					Policy:              PolicyFrom("roundRobin"),
					TraceResolverMode:   true,
				}, "users"))
			Expect(err).To(BeNil())
			time.Sleep(4 * time.Second)
		})
		It("data specific backend db instance", func() {
			fmt.Println("==> data specific backend db instance")
			var user User
			db.First(&user)
			Expect(user.Name).To(BeEquivalentTo("13309"))
			db.Clauses(Read).First(&user)
			Expect(user.Name).To(BeEquivalentTo("13309"))
			db.Clauses(Write).First(&user)
			Expect(user.Name).To(BeEquivalentTo("13308"))

			var order Order
			tx := db.Begin()
			tx.Find(&order)
		})
		It("test transaction", func() {
			fmt.Println("==> test transaction")
			var order Order

			tx := db.Begin()
			tx.Find(&order)
			Expect(order.OrderNo).To(BeEquivalentTo("13306"))
			tx.Rollback()

			tx = db.Clauses(Read).Begin()
			tx.Find(&order)
			Expect(order.OrderNo).To(BeEquivalentTo("13306"))
			tx.Rollback()

			tx = db.Clauses(Write).Begin()
			tx.Find(&order)
			Expect(order.OrderNo).To(BeEquivalentTo("13306"))
			tx.Rollback()

			// test commit
			var prods []Product
			tx = db.Clauses(Write).Begin()
			newProd := &Product{Name: "test-rollback"}
			Expect(tx.Create(newProd).Error).To(BeNil())
			Expect(tx.Commit().Error).To(BeNil())
			Expect(db.Clauses(Write).Where("name = ?", "test-rollback").Find(&prods).Error).To(BeNil())
			Expect(len(prods)).To(BeEquivalentTo(1))
			Expect(db.Delete(newProd).Error).To(BeNil())
			Expect(db.Clauses(Write).Where("name = ?", "test-rollback").Find(&prods).Error).To(BeNil())
			Expect(len(prods)).To(BeEquivalentTo(0))

			// test rollback
			tx = db.Clauses(Write).Begin()
			tx.Create(newProd)
			tx.Rollback()
			Expect(db.Clauses(Write).Where("name = ?", "test-rollback").Find(&prods).Error).To(BeNil())
			Expect(len(prods)).To(BeEquivalentTo(0))
		})
		It("test crud", func() {
			fmt.Println("==> test crud")
			var name string
			// test create
			db13308, err := getDB("13308")
			Expect(err).To(BeNil())
			db.Create(&User{Name: "create"})
			Expect(db.First(&User{}, "name = ?", "create").Error).NotTo(BeNil())
			Expect(db.Clauses(Write).First(&User{}, "name = ?", "create").Error).To(BeNil())
			Expect(db13308.First(&User{}, "name = ?", "create").Error).To(BeNil())

			// test update
			Expect(db.Model(&User{}).Where("name = ?", "create").Update("name", "update").Error).To(BeNil())
			Expect(db13308.First(&User{}, "name = ?", "update").Error).To(BeNil())
			Expect(db.Clauses(Write).Raw("select name from users where name = ?", "update").Row().Scan(&name)).To(BeNil())
			Expect(name).To(BeEquivalentTo("update"))

			// test raw
			Expect(db.Raw("select name from users where name = ?", "update").Row().Scan(&name)).NotTo(BeNil())
			Expect(db.Clauses(Read).Raw("select name from users where name = ?", "update").Row().Scan(&name)).NotTo(BeNil())
			Expect(db.Clauses(Write).Raw("select name from users where name = ?", "update").Row().Scan(&name)).To(BeNil())
			Expect(db.Clauses(Write).Raw("select name from users where name = ? FOR UPDATE", "update").Row().Scan(&name)).To(BeNil())

			// test delete
			Expect(db.Where("name = ?", "update").Delete(&User{}).Error).To(BeNil())
			Expect(db13308.First(&User{}, "name = ?", "update").Error).NotTo(BeNil())
		})
		AfterAll(func() {
			Expect(Close()).Should(Succeed())
		})
	})
	Context("only 1 master configured", Serial, Ordered, func() {
		var db *gorm.DB
		BeforeAll(func() {
			// init db and use dbresolver plugin
			fmt.Println("init db and use dbresolver plugin")
			_db, err := getDB("13306")
			db = _db
			Expect(err).To(BeNil())
			err = db.Use(Register(Config{
				Sources:             []DSN{dsn13306},
				Replicas:            []DSN{},
				HostResolveInterval: 2 * time.Second,
				HealthCheckInterval: 2 * time.Second,
				HealthCheckTimeout:  2 * time.Second,
				DBStatusLogInterval: 10 * time.Second,
				MaxOpenConns:        10,
				MaxIdleConns:        10,
				Policy:              PolicyFrom("roundRobin"),
			}))
			Expect(err).To(BeNil())
			time.Sleep(4 * time.Second)
		})
		It("all query use primary", func() {
			fmt.Println("==> all query use primary")
			var user User
			db.First(&user)
			Expect(user.Name).To(BeEquivalentTo("13306"))
			db.Clauses(Read).First(&user)
			Expect(user.Name).To(BeEquivalentTo("13306"))
			db.Clauses(Write).First(&user)
			Expect(user.Name).To(BeEquivalentTo("13306"))
		})
		AfterAll(func() {
			Expect(Close()).Should(Succeed())
		})
	})
	Context("with 0 master and 3 replicas in random policy", Serial, Ordered, func() {
		var db *gorm.DB
		BeforeAll(func() {
			// init db and use dbresolver plugin
			_db, err := getDB("13306")
			db = _db
			Expect(err).To(BeNil())
			err = db.Use(Register(Config{
				Sources:             []DSN{},
				Replicas:            []DSN{dsn13307, dsn13308, dsn13309},
				HostResolveInterval: 2 * time.Second,
				HealthCheckInterval: 2 * time.Second,
				HealthCheckTimeout:  2 * time.Second,
				DBStatusLogInterval: 10 * time.Second,
				MaxOpenConns:        10,
				MaxIdleConns:        10,
				Policy:              PolicyFrom("random"),
			}))
			Expect(err).To(BeNil())
			time.Sleep(4 * time.Second)
		})
		It("load balance in random policy", func() {
			fmt.Println("==> load balance in random policy")
			var user User
			r := map[string]int{
				"13307": 0,
				"13308": 0,
				"13309": 0,
			}
			for i := 0; i < 3; i++ {
				db.First(&user)
				r[user.Name]++
			}
		})
		AfterAll(func() {
			Expect(Close()).Should(Succeed())
		})
	})
	Context("with 0 master and 3 replicas in roundRobin policy", Serial, Ordered, func() {
		var db *gorm.DB
		BeforeAll(func() {
			// init db and use dbresolver plugin
			_db, err := getDB("13306")
			db = _db
			Expect(err).To(BeNil())
			err = db.Use(Register(Config{
				Sources:             []DSN{},
				Replicas:            []DSN{dsn13307, dsn13308, dsn13309},
				HostResolveInterval: 2 * time.Second,
				HealthCheckInterval: 2 * time.Second,
				HealthCheckTimeout:  2 * time.Second,
				DBStatusLogInterval: 10 * time.Second,
				MaxOpenConns:        10,
				MaxIdleConns:        10,
				Policy:              PolicyFrom("roundRobin"),
			}))
			Expect(err).To(BeNil())
			time.Sleep(4 * time.Second)
		})
		It("load balance in randRobin policy", func() {
			fmt.Println("==> load balance in randRobin policy")
			var user User
			r := map[string]int{
				"13307": 0,
				"13308": 0,
				"13309": 0,
			}
			for i := 0; i < 3; i++ {
				db.First(&user)
				r[user.Name]++
			}
			Expect(r["13307"]).To(BeEquivalentTo(1))
			Expect(r["13308"]).To(BeEquivalentTo(1))
			Expect(r["13309"]).To(BeEquivalentTo(1))
		})
		It("primary query raise error", func() {
			var user User
			r := db.Clauses(Write).First(&user)
			Expect(r.Error).To(BeEquivalentTo(ErrEmptyConnPool))
		})
		AfterAll(func() {
			Expect(Close()).Should(Succeed())
		})
	})
	Context("with 1 master and 3 replicas", Serial, Ordered, func() {
		var db *gorm.DB
		config := Config{
			Sources:             []DSN{dsn13306},
			Replicas:            []DSN{dsn13307, dsn13308, dsn13309},
			HostResolveInterval: 2 * time.Second,
			HealthCheckInterval: 2 * time.Second,
			HealthCheckTimeout:  2 * time.Second,
			DBStatusLogInterval: 10 * time.Second,
			MaxHealthCheckRetry: 2,
			MaxOpenConns:        10,
			MaxIdleConns:        10,
			Policy:              PolicyFrom("roundRobin"),
		}
		BeforeAll(func() {
			// init db and use dbresolver plugin
			fmt.Println("init db and use dbresolver plugin")
			_db, err := getDB("13306")
			db = _db
			Expect(err).To(BeNil())
			err = db.Use(Register(config))
			Expect(err).To(BeNil())
			time.Sleep(4 * time.Second)
		})
		When("every db instance is running healthy", func() {
			It("read write splitting", func() {
				fmt.Println("==> read write splitting")
				var user User
				db.First(&user)
				Expect(user.Name).To(BeElementOf([]string{"13307", "13308", "13309"}))
				db.Clauses(Read).First(&user)
				Expect(user.Name).To(BeElementOf([]string{"13307", "13308", "13309"}))
				db.Clauses(Write).First(&user)
				Expect(user.Name).To(BeEquivalentTo("13306"))
			})
			It("load balance in randRobin policy", func() {
				fmt.Println("==> load balance in randRobin policy")
				var user User
				r := map[string]int{
					"13307": 0,
					"13308": 0,
					"13309": 0,
				}
				for i := 0; i < 3; i++ {
					db.First(&user)
					r[user.Name]++
				}
				Expect(r["13307"]).To(BeEquivalentTo(1))
				Expect(r["13308"]).To(BeEquivalentTo(1))
				Expect(r["13309"]).To(BeEquivalentTo(1))
			})
			It("load balance between replicas", func() {
				fmt.Println("==> load balance between replicas")
				var user User
				r := map[string]int{
					"13307": 0,
					"13308": 0,
					"13309": 0,
				}
				for i := 0; i < 3*10; i++ {
					db.First(&user)
					r[user.Name]++
				}
				Expect(r["13307"] / r["13308"]).To(BeEquivalentTo(1))
				Expect(r["13308"] / r["13309"]).To(BeEquivalentTo(1))
			})
		})
		When("one replica is down", Ordered, func() {
			BeforeAll(func() {
				err := pool.Client.StopContainer(resources["13307"].Container.ID, 0)
				Expect(err).To(BeNil())
				time.Sleep(time.Second * 2)
			})
			It("read from master and replicas", func() {
				fmt.Println("==> read from master and replicas")
				var user User
				r := map[string]int{
					"13307": 0,
					"13308": 0,
					"13309": 0,
				}
				db.Clauses(Write).First(&user)
				Expect(user.Name).To(BeEquivalentTo("13306"))
				for i := 0; i < 3*2; i++ {
					db.First(&user)
					r[user.Name]++
				}
				Expect(r["13307"]).To(BeEquivalentTo(0))
				Expect(r["13308"]).To(BeEquivalentTo(3))
				Expect(r["13309"]).To(BeEquivalentTo(3))
			})
			It("recover after ip instance back online", func() {
				time.Sleep(config.HealthCheckInterval*time.Duration(config.MaxHealthCheckRetry) + 2*time.Second)
				err := pool.Client.StartContainer(resources["13307"].Container.ID, nil)
				Expect(err).To(BeNil())
				waitForStart("13307")
				time.Sleep(config.HealthCheckInterval + time.Second*1)
				var user User
				r := map[string]int{
					"13307": 0,
					"13308": 0,
					"13309": 0,
				}
				for i := 0; i < 3*2; i++ {
					db.First(&user)
					r[user.Name]++
				}
				Expect(r["13307"]).To(BeEquivalentTo(2))
				Expect(r["13308"]).To(BeEquivalentTo(2))
				Expect(r["13309"]).To(BeEquivalentTo(2))
			})
		})
		AfterAll(func() {
			Expect(Close()).Should(Succeed())
			Expect(db.Clauses(Write).Exec("SELECT VERSION()").Error).ShouldNot(BeNil())
			Expect(db.Exec("SELECT VERSION()").Error).ShouldNot(BeNil())
		})
	})
})

var _ = AfterSuite(func() {
	// Purge function destroys container
	for _, resource := range resources {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("purge pool err: %s", err)
		}
	}
})
