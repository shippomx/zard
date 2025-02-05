package dbresolver

import (
	"gorm.io/gorm"
)

type resolver struct {
	policy            Policy
	dbResolver        *DBResolver
	traceResolverMode bool

	primaryManager  ServerManager
	replicasManager ServerManager
}

func (r *resolver) sources() []Instance {
	return r.primaryManager.GetInstances()
}

func (r *resolver) replicas() []Instance {
	return r.replicasManager.GetInstances()
}

func (r *resolver) resolve(stmt *gorm.Statement, op Operation) gorm.ConnPool {
	replicas := r.replicas()
	sources := r.sources()
	var inst Instance
	if op == Read {
		if len(replicas) == 1 {
			inst = replicas[0]
		} else {
			if len(replicas) == 0 {
				return &EmptyConnPool{}
			}
			inst = r.policy.Resolve(replicas)
		}
		if r.traceResolverMode {
			markStmtResolverMode(stmt, ResolverModeReplica)
		}
	} else if len(sources) == 1 {
		inst = sources[0]
		if r.traceResolverMode {
			markStmtResolverMode(stmt, ResolverModeSource)
		}
	} else {
		if len(sources) == 0 {
			return &EmptyConnPool{}
		}
		inst = r.policy.Resolve(sources)
		if r.traceResolverMode {
			markStmtResolverMode(stmt, ResolverModeSource)
		}
	}

	connPool := inst.GetConnPool()
	prepareStmt := inst.GetPrepareStmt()
	if stmt.DB.PrepareStmt {
		if prepareStmt != nil {
			return &gorm.PreparedStmtDB{
				ConnPool: connPool,
				Mux:      prepareStmt.Mux,
				Stmts:    prepareStmt.Stmts,
			}
		}
	}

	return connPool
}
