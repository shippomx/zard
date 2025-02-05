package dbresolver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/conc"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	SelectVersionComment = "SELECT @@version_comment LIMIT 1"
)

func HostResolve(host string) ([]net.IP, error) {
	cname, err := net.LookupCNAME(host)
	if err != nil {
		return nil, err
	}

	return net.LookupIP(cname)
}

type DataSourceType string

const (
	DataSourceSource  = "source"
	DataSourceReplica = "replica"
)

type Addr string // String return database instance identity, format is ipv4:port, eg. "127.0.0.1:3306".

// func (a Addr) String() string {
// 	return fmt.Sprintf("%s:%d", a.ipv4, a.port)
// }

type ServerManager interface {
	RunAndWait()

	CronDBStatusLog()
	CronHealthCheck()
	CronResolve(dsn DSN)

	SetPrimary(isPrimary bool)
	SetPeerManager(peerManager ServerManager)
	NotifyPeer(addr Addr)
	Remove(addr Addr)

	GetInstances() (list []Instance)
	Close() error
}

var (
	_ ServerManager = &EmptyServerManager{}
	_ ServerManager = &ServerManagerImpl{}
)

type EmptyServerManager struct{}

func (EmptyServerManager) RunAndWait()                     {}
func (EmptyServerManager) CronDBStatusLog()                {}
func (EmptyServerManager) CronHealthCheck()                {}
func (EmptyServerManager) CronResolve(_ DSN)               {}
func (EmptyServerManager) SetPrimary(_ bool)               {}
func (EmptyServerManager) SetPeerManager(_ ServerManager)  {}
func (EmptyServerManager) NotifyPeer(_ Addr)               {}
func (EmptyServerManager) Remove(_ Addr)                   {}
func (EmptyServerManager) GetInstances() (list []Instance) { return }
func (EmptyServerManager) Close() error                    { return nil }

type ServerManagerImpl struct {
	ctx    context.Context
	logger LoggerWithDebug

	done      chan bool
	once      sync.Once
	closeOnce sync.Once

	isPrimary   bool
	peerManager ServerManager

	conf          Config
	hostDSNs      []DSN
	ipMap         map[Addr]DSN
	instanceStore sync.Map
}

func NewServerManager(ctx context.Context, logger LoggerWithDebug, conf Config, dsns []DSN, dsType DataSourceType) ServerManager {
	m := &ServerManagerImpl{
		ctx:    ctx,
		done:   make(chan bool),
		logger: logger,

		hostDSNs: []DSN{},
		ipMap:    map[Addr]DSN{}, // 兼容数据源配置中使用 ip 格式

		conf: conf,

		instanceStore: sync.Map{},
	}
	for _, item := range dsns {
		m.initDSN(item, dsType) // nolint: contextcheck
	}
	return m
}

// databaseRole 返回数据库配置的角色.
func (m *ServerManagerImpl) databaseRole() string {
	return m.conf.Role
}

func (m *ServerManagerImpl) SetPeerManager(peerManager ServerManager) { m.peerManager = peerManager }
func (m *ServerManagerImpl) SetPrimary(isPrimary bool)                { m.isPrimary = isPrimary }

// initDSN 以地址的类型分离 DSN，ip 类型存在 m.ipMap，域名类型存在 m.hostDSNs .
func (m *ServerManagerImpl) initDSN(dsn DSN, dsType DataSourceType) {
	dsn.Type = dsType
	if net.ParseIP(dsn.Path) != nil {
		// ip: no need resolve
		dsn.IPv4 = dsn.Path
		m.ipMap[dsn.Addr()] = dsn
		m.createInstanceIfNotExists(dsn)
	} else {
		m.hostDSNs = append(m.hostDSNs, dsn)
	}
}

func (m *ServerManagerImpl) RunAndWait() {
	m.once.Do(m.run)
	stop := make(chan bool)
	go func() {
		for {
			<-time.After(10 * time.Millisecond)
			if len(m.GetInstances()) > 0 {
				stop <- true
				return
			}
		}
	}()

	select {
	case <-stop:
		close(stop)
		return
	case <-time.After(1 * time.Minute):
		panic("timeout for database connection, please check the database configuration")
	}
}

func (m *ServerManagerImpl) run() {
	go m.CronHealthCheck()

	for _, dsn := range m.hostDSNs {
		go m.CronResolve(dsn)
	}

	go m.CronDBStatusLog()
}

func (m *ServerManagerImpl) CronDBStatusLog() {
	job := func() {
		msg := []string{}
		m.instanceStore.Range(func(k, v any) bool {
			vv := v.(*Instance)
			msg = append(msg, fmt.Sprintf("{DSN={%s}, healthy=%t, retry=%d}", vv.dsn.DEBUG(), vv.GetHealthy(), vv.GetRetry()))
			return true
		})
		m.logger.Info(m.ctx, "[SUMMARY] cached instances, len=%d\tdetails=%v\trole=%s", len(msg), msg, m.databaseRole())
	}
	StartCronJob(m.done, true, m.conf.DBStatusLogInterval, "DBStatusLog: ", job)
}

// Clear close all instances and remove from the store.
func (m *ServerManagerImpl) Clear() (err error) {
	m.instanceStore.Range(func(k, inst any) bool {
		if inst != nil {
			err = errors.Join(err, inst.(*Instance).Close())
		}
		m.instanceStore.Delete(k)
		return true
	})
	return
}

// Remove and close instance of the given address.
func (m *ServerManagerImpl) Remove(addr Addr) {
	value, exist := m.instanceStore.LoadAndDelete(addr)
	if exist && value != nil {
		inst := value.(*Instance)
		defer m.logger.Info(m.ctx, "success to remove instance, DSN={%s}\trole=%s", inst.dsn.DEBUG(), m.databaseRole())
		inst.Close()
	}
}

// Close stops the manager.
func (m *ServerManagerImpl) Close() (err error) {
	m.closeOnce.Do(func() {
		err = m.close()
	})
	return
}

func (m *ServerManagerImpl) close() error {
	// Note(roby): 停止定时任务，避免在关闭连接期间创建新的连接.
	close(m.done)
	return m.Clear()
}

// NotifyPeer tell peer manager to remove Some(addr) instance.
func (m *ServerManagerImpl) NotifyPeer(addr Addr) {
	if m.peerManager != nil {
		m.peerManager.Remove(addr)
	}
}

// CronResolve start a cron job of a host resolving.
func (m *ServerManagerImpl) CronResolve(dsn DSN) {
	host := dsn.Path
	job := func() {
		ips, err := HostResolve(host)
		if err != nil {
			m.logger.Error(m.ctx, "fail to resolve host, err=%s\tDSN={%s}\trole=%s", err, dsn.DEBUG(), m.databaseRole())
		}
		m.logger.Debug(m.ctx, "found ips, ips=%s\tDSN={%s}\trole=%s", ips, dsn.DEBUG(), m.databaseRole())
		for _, item := range ips {
			if item.To4() == nil {
				continue
			}
			cp := dsn
			cp.IPv4 = item.String()
			if m.createInstanceIfNotExists(cp) {
				m.NotifyPeer(cp.Addr())
			}
		}
	}
	StartCronJob(m.done, true, m.conf.HostResolveInterval, "Resolve: "+host, job)
}

// CronHealthCheck start a cron job of instances health checking.
func (m *ServerManagerImpl) CronHealthCheck() {
	duration := m.conf.HealthCheckInterval
	job := func() {
		var removals []Addr
		var wg conc.WaitGroup
		m.instanceStore.Range(func(_, value any) bool {
			wg.Go(func() {
				m.healthCheck(value.(*Instance))
			})
			return true
		})
		wg.Wait()

		m.instanceStore.Range(func(addr, value any) bool {
			_addr, _inst := addr.(Addr), value.(*Instance)
			if _, ok := m.ipMap[_addr]; ok { // ip 配置的不移除，避免恢复不了
				return true
			}
			if _inst.GetHealthy() {
				return true
			}
			if _inst.GetRetry() >= m.conf.MaxHealthCheckRetry {
				removals = append(removals, _addr)
			}
			return true
		})

		for _, addr := range removals {
			m.Remove(addr)
		}
	}
	StartCronJob(m.done, true, duration, "HealthCheck", job)
}

func (m *ServerManagerImpl) healthCheck(inst *Instance) {
	if inst.GetConnPool() == nil {
		// InCase: 服务启动时，ip 配置的实例初始化失败.
		n, err := m.newInstance(inst.dsn)
		if err != nil {
			inst.IncreaseRetry()
			return
		}
		inst.SetConnPoolAndPrepareStmt(n.GetConnPool(), n.GetPrepareStmt())
	}

	ctx, cancel := context.WithTimeout(m.ctx, m.conf.HealthCheckTimeout)
	defer cancel()
	rows, err := inst.GetConnPool().QueryContext(ctx, SelectVersionComment)
	if rows != nil {
		defer rows.Close()
	}
	if (rows != nil && rows.Err() != nil) || err != nil {
		m.logger.Warn(m.ctx, "instance health check failed, err=%s\tDSN={%s}\trole=%s", err.Error(), inst.dsn.DEBUG(), m.databaseRole())
		inst.SetHealthy(false)
		inst.IncreaseRetry()
	} else {
		m.logger.Debug(m.ctx, "instance health check success, DSN={%s}\trole=%s", inst.dsn.DEBUG(), m.databaseRole())
		inst.SetHealthy(true)
		inst.ResetRetry()
	}
}

// GetInstances return ordered healthy instances.
func (m *ServerManagerImpl) GetInstances() (list []Instance) {
	m.instanceStore.Range(func(_, inst any) bool {
		_inst := inst.(*Instance)
		if _inst.GetHealthy() {
			list = append(list, *_inst)
		}
		return true
	})
	sort.Slice(list, func(i, j int) bool {
		return list[i].Addr() < list[j].Addr()
	})
	return list
}

func (m *ServerManagerImpl) createInstanceIfNotExists(dsn DSN) (created bool) {
	addr := dsn.Addr()
	_, exist := m.instanceStore.Load(addr)
	if exist {
		return
	}

	inst, err := m.newInstance(dsn)
	if err != nil {
		m.logger.Error(m.ctx, "failed to create instance, err=%s\tDSN={%s}\trole=%s", err.Error(), inst.dsn.DEBUG(), m.databaseRole())
	}
	m.logger.Info(m.ctx, "success to create instance, DSN={%s}\trole=%s", inst.dsn.DEBUG(), m.databaseRole())
	if m.isPrimary { // only one instance exsists in primary store
		_ = m.Clear()
	}

	actual, exists := m.instanceStore.LoadOrStore(addr, inst)
	if exists && actual != nil {
		actual.(*Instance).Close()
	}

	created = true
	return
}

func (m *ServerManagerImpl) newInstance(dsn DSN, opts ...gorm.Option) (*Instance, error) {
	db, err := gorm.Open(mysql.Open(dsn.Str()), opts...)
	if err != nil {
		return &Instance{
			logger:  m.logger,
			mutex:   &sync.RWMutex{},
			dsn:     dsn,
			retry:   new(atomic.Uint32),
			healthy: NewAtomicBool(false),
		}, err
	}
	connPool := db.Config.ConnPool
	if preparedStmtDB, ok := connPool.(*gorm.PreparedStmtDB); ok {
		connPool = preparedStmtDB.ConnPool
	}

	if conn, ok := connPool.(interface{ SetMaxOpenConns(int) }); ok {
		conn.SetMaxOpenConns(m.conf.MaxOpenConns)
	} else {
		db.Logger.Error(context.Background(), "SetMaxOpenConns not implemented for %#v", conn)
	}

	if conn, ok := connPool.(interface{ SetMaxIdleConns(int) }); ok {
		conn.SetMaxIdleConns(m.conf.MaxIdleConns)
	} else {
		db.Logger.Error(context.Background(), "SetMaxIdleConns not implemented for %#v", conn)
	}

	// If d <= 0, connections are not closed due to a connection's age.
	if conn, ok := connPool.(interface{ SetConnMaxLifetime(d time.Duration) }); ok {
		conn.SetConnMaxLifetime(m.conf.ConnMaxLifetime)
	} else {
		db.Logger.Error(context.Background(), "SetConnMaxLifetime not implemented for %#v", conn)
	}

	// If d <= 0, connections are not closed due to a connection's idle time.
	if conn, ok := connPool.(interface{ SetConnMaxIdleTime(d time.Duration) }); ok {
		conn.SetConnMaxIdleTime(m.conf.ConnMaxIdleTime)
	} else {
		db.Logger.Error(context.Background(), "SetConnMaxIdleTime not implemented for %#v", conn)
	}

	inst := &Instance{
		logger:   m.logger,
		mutex:    &sync.RWMutex{},
		db:       db,
		dsn:      dsn,
		healthy:  NewAtomicBool(true),
		retry:    new(atomic.Uint32),
		connPool: connPool,
		prepareStmt: &gorm.PreparedStmtDB{
			ConnPool: db.Config.ConnPool,
			Stmts:    map[string]*gorm.Stmt{},
			Mux:      &sync.RWMutex{},
		},
	}
	GlobalInstances.Insert(inst)
	return inst, nil
}

var GlobalInstances = &InstanceStore{}

type InstanceStore sync.Map

func (s *InstanceStore) Insert(inst *Instance) {
	(*sync.Map)(s).Store(inst.Addr(), inst)
}

func (s *InstanceStore) List() []*Instance {
	var list []*Instance
	(*sync.Map)(s).Range(func(_, v any) bool {
		inst := v.(*Instance)
		if inst != nil {
			list = append(list, inst)
		}
		return true
	})
	return list
}

func (s *InstanceStore) Delete(inst *Instance) {
	(*sync.Map)(s).Delete(inst.Addr())
}

type Instance struct {
	logger logger.Interface

	mutex *sync.RWMutex

	db          *gorm.DB
	dsn         DSN
	healthy     *atomic.Bool
	retry       *atomic.Uint32
	connPool    gorm.ConnPool
	prepareStmt *gorm.PreparedStmtDB
}

func (i *Instance) Close() error {
	prepareStmt := i.GetPrepareStmt()
	if prepareStmt == nil {
		return nil
	}
	db, err := prepareStmt.GetDBConn()
	if err != nil {
		i.logger.Warn(context.TODO(), "get db conn err: %s", err)
		return err
	}
	if err := db.Close(); err != nil {
		i.logger.Warn(context.TODO(), "close db conn err: %s", err)
		return err
	}
	GlobalInstances.Delete(i)
	return nil
}

func (i *Instance) GetConnPool() gorm.ConnPool {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.connPool
}

func (i *Instance) GetPrepareStmt() *gorm.PreparedStmtDB {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.prepareStmt
}

func (i *Instance) SetConnPoolAndPrepareStmt(connPool gorm.ConnPool, prepareStmt *gorm.PreparedStmtDB) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.connPool = connPool
	i.prepareStmt = prepareStmt
}

func (i *Instance) SetHealthy(healthy bool) {
	i.healthy.Store(healthy)
}

func (i *Instance) GetHealthy() bool {
	return i.healthy.Load()
}

func (i *Instance) IncreaseRetry() {
	i.retry.Add(1)
}

func (i *Instance) ResetRetry() {
	i.retry.Store(0)
}

func (i *Instance) GetRetry() uint32 {
	return i.retry.Load()
}

func (i *Instance) SQLDB() (*sql.DB, error) { return i.db.DB() }

func (i *Instance) Addr() Addr { return i.dsn.Addr() }

func (i *Instance) GetAddress() string { return string(i.Addr()) }
