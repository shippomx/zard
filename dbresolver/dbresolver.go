package dbresolver

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	Write Operation = "write"
	Read  Operation = "read"
)

type DBResolver struct {
	*gorm.DB
	loggerWithDebug  LoggerWithDebug
	configs          []Config
	resolversMap     map[string]*resolver
	resolversList    []*resolver
	global           *resolver
	prepareStmtStore map[gorm.ConnPool]*gorm.PreparedStmtDB
}

type Config struct {
	Sources             []DSN
	Replicas            []DSN
	MaxIdleConns        int
	MaxOpenConns        int
	ConnMaxIdleTime     time.Duration
	ConnMaxLifetime     time.Duration
	HostResolveInterval time.Duration
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
	DBStatusLogInterval time.Duration
	MaxHealthCheckRetry uint32
	Policy              Policy
	TraceResolverMode   bool
	datas               []interface{}
	Role                string
}

func (c *Config) Validate() error {
	if len(c.Sources) == 0 && len(c.Replicas) == 0 {
		return errors.New("no source or replica")
	}

	if len(c.Sources) > 1 {
		return errors.New("only one source is allowed")
	}

	for _, src := range c.Sources {
		for _, replica := range c.Replicas {
			if src.EqualAddrTo(replica) {
				if len(c.Sources) == 1 && len(c.Replicas) == 1 {
					return errors.New("source and replica can't be the same")
				}
				return errors.New("one mysql backend server conflict in source and replica")
			}
		}
	}

	return nil
}

// resolvers store all registered dbresolver instance
// We use it to implement `Close` function.
var drResolvers []*DBResolver

// Close 用来主动关闭 dbresolver 管理的所有数据库连接.
func Close() (e error) {
	for _, dr := range drResolvers {
		for _, r := range dr.resolversList {
			if r.primaryManager != nil {
				e = errors.Join(e, r.primaryManager.Close())
			}
			if r.replicasManager != nil {
				e = errors.Join(e, r.replicasManager.Close())
			}
		}
	}
	drResolvers = nil
	return
}

func Register(config Config, datas ...interface{}) *DBResolver {
	r := (&DBResolver{}).Register(config, datas...)
	drResolvers = append(drResolvers, r)
	return r
}

func (dr *DBResolver) Register(config Config, datas ...interface{}) *DBResolver {
	if dr.prepareStmtStore == nil {
		dr.prepareStmtStore = map[gorm.ConnPool]*gorm.PreparedStmtDB{}
	}

	if dr.resolversMap == nil {
		dr.resolversMap = map[string]*resolver{}
	}

	if config.Policy == nil {
		config.Policy = &RandomPolicy{}
	}

	config.datas = datas
	dr.configs = append(dr.configs, config)
	if dr.DB != nil {
		dr.compileConfig(config)
	}
	return dr
}

func (dr *DBResolver) Name() string {
	return "gorm:db_resolver"
}

func (dr *DBResolver) Initialize(db *gorm.DB) error {
	docLink := "https://gtglobal.jp.larksuite.com/wiki/NNqawMHhqiawlWkGF6rjkD6xpMd"
	for _, config := range dr.configs {
		if err := config.Validate(); err != nil {
			db.Logger.Error(context.Background(), "config err: %s, please read doc %s to get help", err, docLink)
			return err
		}
	}
	dr.DB = db
	dr.loggerWithDebug = TransformLogger(dr.Logger)
	dr.registerCallbacks(db)
	return dr.compile()
}

func (dr *DBResolver) compile() error {
	for _, config := range dr.configs {
		if err := dr.compileConfig(config); err != nil {
			return err
		}
	}
	return nil
}

func (dr *DBResolver) compileConfig(config Config) (err error) {
	r := resolver{
		dbResolver:        dr,
		policy:            config.Policy,
		traceResolverMode: config.TraceResolverMode,
	}
	dr.resolversList = append(dr.resolversList, &r)

	if len(config.Sources) == 0 {
		r.primaryManager = &EmptyServerManager{}
	} else {
		r.primaryManager = dr.initServerManager(config.Sources, config, DataSourceSource)
		r.primaryManager.SetPrimary(true)
	}

	if len(config.Replicas) == 0 {
		r.replicasManager = r.primaryManager
	} else {
		r.replicasManager = dr.initServerManager(config.Replicas, config, DataSourceReplica)
		r.replicasManager.SetPeerManager(r.primaryManager)
		r.primaryManager.SetPeerManager(r.replicasManager)
	}

	r.primaryManager.RunAndWait()
	r.replicasManager.RunAndWait()

	if len(config.datas) > 0 {
		for _, data := range config.datas {
			if t, ok := data.(string); ok {
				dr.resolversMap[t] = &r
			} else {
				stmt := &gorm.Statement{DB: dr.DB}
				if err := stmt.Parse(data); err == nil {
					dr.resolversMap[stmt.Table] = &r
				} else {
					return err
				}
			}
		}
	} else if dr.global == nil {
		dr.global = &r
	} else {
		return errors.New("conflicted global resolver")
	}

	if config.TraceResolverMode {
		dr.Logger = NewResolverModeLogger(dr.Logger)
	}

	return nil
}

func (dr *DBResolver) initServerManager(dsns []DSN, conf Config, dsType DataSourceType) ServerManager {
	return NewServerManager(context.Background(), dr.loggerWithDebug, conf, dsns, dsType)
}

func (dr *DBResolver) resolve(stmt *gorm.Statement, op Operation) gorm.ConnPool {
	if r := dr.getResolver(stmt); r != nil {
		return r.resolve(stmt, op)
	}
	return stmt.ConnPool
}

func (dr *DBResolver) getResolver(stmt *gorm.Statement) *resolver {
	if len(dr.resolversMap) > 0 {
		if u, ok := stmt.Clauses[usingName].Expression.(using); ok && u.Use != "" {
			if r, ok := dr.resolversMap[u.Use]; ok {
				return r
			}
		}

		if stmt.Table != "" {
			if r, ok := dr.resolversMap[stmt.Table]; ok {
				return r
			}
		}

		if stmt.Model != nil {
			if err := stmt.Parse(stmt.Model); err == nil {
				if r, ok := dr.resolversMap[stmt.Table]; ok {
					return r
				}
			}
		}

		if stmt.Schema != nil {
			if r, ok := dr.resolversMap[stmt.Schema.Table]; ok {
				return r
			}
		}

		if rawSQL := stmt.SQL.String(); rawSQL != "" {
			if r, ok := dr.resolversMap[getTableFromRawSQL(rawSQL)]; ok {
				return r
			}
		}
	}

	return dr.global
}
