package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type StrictRedis struct {
	*Redis
}

// NewStrictRedis returns a new StrictRedis based on the given RedisConf and options.
//
// NewStrictRedis creates a new Redis and sets it on the StrictRedis.
// If there is an error creating the Redis, NewStrictRedis returns the error.
//
// The StrictRedis is returned regardless of whether there is an error or not.
func NewStrictRedis(conf RedisConf, opts ...Option) (*StrictRedis, error) {
	r, err := NewRedis(conf, opts...)
	if err != nil {
		return nil, err
	}
	return NewStrictRedisFromRedis(r), nil
}

func NewStrictRedisFromRedis(r *Redis) *StrictRedis {
	return &StrictRedis{r}
}

type RedisValue struct {
	result any
	err    StrictRedisError
}

type StrictRedisError interface {
	error
}

var NilValueErr error = StrictRedisError(errors.New("value is nil"))
var ErrTypeMismatch = StrictRedisError(errors.New("unhandled type"))

func (r RedisValue) To(v any) error {
	if r.err != nil {
		return r.err
	}
	if r.result == nil {
		return NilValueErr
	}
	res := r.result // the value stored in the RedisValue

	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Ptr || vv.IsNil() {
		return errors.New("v must be a non-nil pointer")
	}

	var err error
	switch res := res.(type) {
	case string: // res is a string
		err = json.Unmarshal([]byte(res), vv.Interface())
		if err != nil {
			return StrictRedisError(err)
		}
	case []string: // res is a []string
		if vv.Elem().Kind() != reflect.Slice {
			return errors.New("v must be a slice pointer")
		}
		vv.Elem().Set(reflect.MakeSlice(vv.Elem().Type(), len(res), len(res)))
		for i := range res {
			err = json.Unmarshal([]byte(res[i]), vv.Elem().Index(i).Addr().Interface())
			if err != nil {
				return StrictRedisError(fmt.Errorf("at index %d: %w", i, err))
			}
		}
	case map[string]string: // res is a map[string]string
		if vv.Elem().Kind() != reflect.Map {
			return errors.New("v must be a map pointer")
		}
		for k, v := range res {
			vvv := reflect.New(vv.Elem().Type().Elem()).Elem()
			if vvv.Kind() != reflect.Ptr {
				vvv = vvv.Addr()
			}
			err = json.Unmarshal([]byte(v), vvv.Interface())
			if err != nil {
				return StrictRedisError(fmt.Errorf("at key %s: %w", k, err))
			}
			vv.Elem().SetMapIndex(reflect.ValueOf(k), vvv.Elem())
		}
	default: // unhandled type
		return ErrTypeMismatch
	}

	return nil
}

func (r *StrictRedis) Get(key string) RedisValue {
	res, err := r.Redis.Get(key)
	if err != nil {
		return RedisValue{result: nil, err: err}
	}
	return RedisValue{result: res, err: err}
}

func (r *StrictRedis) Set(key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.Redis.Set(key, string(b))
}

func (r *StrictRedis) Mget(keys ...string) RedisValue {
	res, err := r.Redis.Mget(keys...)
	if err != nil {
		return RedisValue{result: nil, err: err}
	}
	return RedisValue{result: res, err: err}
}

// MSet executes a Redis MSET command on all nodes in the cluster
// with the provided key-value pairs.
//
// The keys and values in the kvs map are first marshaled to JSON strings,
// and then the key-value pairs are appended to a slice of interface{} values.
// The slice is then passed to Redis.RawExec(), which will execute the MSET
// command on each node in the cluster.
//
// The context argument is not used, but is provided by Redis.RawExec().
//
// The MSET command returns the number of key-value pairs set, so the returned
// error is the error from the last node in the cluster's MSET command.
//
// If an error occurs marshaling any key or value to JSON, the entire MSET
// operation is aborted and the error is returned.
func (r *StrictRedis) MSet(kvs map[string]any) error {
	msetfn := func(ctx context.Context, rds RedisNode) error {
		var pairs []interface{}
		for k, v := range kvs {
			b, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("error marshaling value for key '%s': %w", k, err)
			}
			pairs = append(pairs, k, string(b))
		}
		_, err := rds.MSet(ctx, pairs...).Result()
		return err
	}
	return r.Redis.RawExec(context.Background(), msetfn)
}

func (r *StrictRedis) Setnx(key string, v any) (bool, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return false, err
	}
	return r.Redis.Setnx(key, string(b))
}

func (r *StrictRedis) SetnxEx(key string, v any, seconds int) (bool, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return false, err
	}
	return r.Redis.SetnxEx(key, string(b), seconds)
}

func (r *StrictRedis) Setex(key string, v any, seconds int) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.Redis.Setex(key, string(b), seconds)
}

func (r *StrictRedis) Hset(key, field string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.Redis.Hset(key, field, string(b))
}

func (r *StrictRedis) Hget(key, field string) RedisValue {
	res, err := r.Redis.Hget(key, field)
	if err != nil {
		return RedisValue{result: nil, err: err}
	}
	return RedisValue{result: res, err: err}
}

func (r *StrictRedis) Hmset(key string, kvs map[string]any) error {
	values := make(map[string]string)
	for k, v := range kvs {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		values[k] = string(b)
	}

	return r.Redis.Hmset(key, values)
}

func (r *StrictRedis) Hmget(key string, fields ...string) RedisValue {
	res, err := r.Redis.Hmget(key, fields...)
	if err != nil {
		return RedisValue{result: nil, err: err}
	}
	return RedisValue{result: res, err: err}
}

func (r *StrictRedis) Hgetall(key string) RedisValue {
	res, err := r.Redis.Hgetall(key)
	if err != nil {
		return RedisValue{result: nil, err: err}
	}
	return RedisValue{result: res, err: err}
}
