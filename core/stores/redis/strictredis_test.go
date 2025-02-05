package redis

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet_SuccessfulUnmarshal(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		strictRedis := &StrictRedis{Redis: client}
		var v struct {
			Name string
			Age  int
		}
		v.Name = "Alice"
		v.Age = 30
		//set
		err := strictRedis.Set("valid_key", v)
		assert.Nil(t, err)
		//get
	})
}

func TestCmds(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		runOnRedis(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			var v struct {
				Name string
				Age  int
			}
			v.Name = "Alice"
			v.Age = 30
			//set
			err := strictRedis.Set("valid_key", v)
			assert.Nil(t, err)
		})
	})

	t.Run("get", func(t *testing.T) {
		runOnRedis(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			var v struct {
				Name string
				Age  int
			}
			v.Name = "Alice"
			v.Age = 30
			//set
			err := strictRedis.Set("valid_key", v)
			assert.Nil(t, err)
			//get
			err = strictRedis.Get("valid_key").To(&v)
			assert.Nil(t, err)

			err = strictRedis.Set("invalid_key", "123")
			assert.Nil(t, err)
			sv := ""
			err = strictRedis.Get("invalid_key").To(&sv)
			assert.Nil(t, err)
			assert.Equal(t, "123", sv)

			err = strictRedis.Set("invalid_key", 1)
			assert.Nil(t, err)
			iv := 0
			err = strictRedis.Get("invalid_key").To(&iv)
			assert.Nil(t, err)
			assert.Equal(t, 1, iv)
		})
	})

	t.Run("mset", func(t *testing.T) {
		runOnRedis(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			var v struct {
				Name string
				Age  int
			}
			v.Name = "Alice"
			v.Age = 30
			//set
			err := strictRedis.MSet(map[string]any{"valid_key": v, "valid_key2": v})
			assert.Nil(t, err)
		})
	})

	t.Run("mget", func(t *testing.T) {
		runOnRedis(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			v := TestStruct{
				Name: "Alice",
				Age:  30,
			}
			v2 := TestStruct{
				Name: "Bob",
				Age:  40,
			}
			//set
			err := strictRedis.MSet(map[string]any{"valid_key": v, "valid_key2": v2})
			assert.Nil(t, err)
			//get
			vs := []TestStruct{}
			err = strictRedis.Mget("valid_key", "valid_key2").To(&vs)
			assert.Nil(t, err)
			assert.Equal(t, 2, len(vs))
			assert.Equal(t, "Alice", vs[0].Name)
			assert.Equal(t, "Bob", vs[1].Name)

			//sets
			vv := []TestStruct{v, v2}
			err = strictRedis.MSet(map[string]any{"valid_key": vv, "valid_key2": vv})
			assert.Nil(t, err)
			vv2 := [][]TestStruct{}
			err = strictRedis.Mget("valid_key", "valid_key2").To(&vv2)
			assert.Nil(t, err)
			fmt.Println(vv2)

			assert.Equal(t, 2, len(vv2[0]))
			assert.Equal(t, 2, len(vv2[1]))
			assert.Equal(t, "Alice", vv2[0][0].Name)
			assert.Equal(t, "Bob", vv2[1][1].Name)

		})

		runOnRedisWithError(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			vs := []TestStruct{}
			err := strictRedis.Mget("valid_key", "valid_key2").To(&vs)
			assert.NotNil(t, err)
		})

	})

	t.Run("setex", func(t *testing.T) {
		runOnRedis(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			var v struct {
				Name string
				Age  int
			}
			v.Name = "Alice"
			v.Age = 30
			//set
			err := strictRedis.Setex("valid_key", v, 10)
			assert.Nil(t, err)
		})
	})

	t.Run("setnx", func(t *testing.T) {
		runOnRedis(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			var v struct {
				Name string
				Age  int
			}
			v.Name = "Alice"
			v.Age = 30
			//set
			bool, err := strictRedis.Setnx("valid_key", v)
			assert.Nil(t, err)
			assert.True(t, bool)
		})
	})

	t.Run("setnxex", func(t *testing.T) {
		runOnRedis(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			var v struct {
				Name string
				Age  int
			}
			v.Name = "Alice"
			v.Age = 30
			//set
			bool, err := strictRedis.SetnxEx("valid_key", v, 10)
			assert.Nil(t, err)
			assert.True(t, bool)
		})
	})

	t.Run("hset", func(t *testing.T) {
		runOnRedis(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			var v struct {
				Name string
				Age  int
			}
			v.Name = "Alice"
			v.Age = 30
			//set
			err := strictRedis.Hset("valid_key", "field	", v)
			assert.Nil(t, err)
		})
	})

	t.Run("hget", func(t *testing.T) {
		runOnRedis(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			var v struct {
				Name string
				Age  int
			}
			v.Name = "Alice"
			v.Age = 30
			//set
			err := strictRedis.Hset("valid_key", "field", v)
			assert.Nil(t, err)
			//get
			err = strictRedis.Hget("valid_key", "field").To(&v)
			assert.Nil(t, err)
			assert.Equal(t, "Alice", v.Name)
			assert.Equal(t, 30, v.Age)
		})
	})

	t.Run("hmget", func(t *testing.T) {
		runOnRedis(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			v := TestStruct{
				Name: "Alice",
				Age:  30,
			}
			//set
			err := strictRedis.Hmset("valid_key", map[string]any{"field": v})
			assert.Nil(t, err)
			//get
			vs := []TestStruct{}
			err = strictRedis.Hmget("valid_key", "field").To(&vs)
			assert.Nil(t, err)
			assert.Equal(t, "Alice", v.Name)
			assert.Equal(t, 30, v.Age)
		})
	})

	t.Run("hmset", func(t *testing.T) {
		runOnRedis(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			v := TestStruct{
				Name: "Alice",
				Age:  30,
			}
			//set
			err := strictRedis.Hmset("valid_key", map[string]any{"field": v})
			assert.Nil(t, err)
		})
	})

	t.Run("hgetall", func(t *testing.T) {
		runOnRedis(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			v := TestStruct{
				Name: "Alice",
				Age:  30,
			}
			err := strictRedis.Hmset("valid_key", map[string]any{"field": v, "field2": v, "field3": v})
			assert.Nil(t, err)
			//get
			vs := map[string]TestStruct{}
			err = strictRedis.Hgetall("valid_key").To(&vs)
			assert.Nil(t, err)
			assert.Equal(t, "Alice", vs["field"].Name)
			assert.Equal(t, 30, vs["field"].Age)
			assert.Equal(t, "Alice", vs["field2"].Name)
			assert.Equal(t, 30, vs["field2"].Age)
		})
	})

	t.Run("hdel", func(t *testing.T) {
		runOnRedis(t, func(client *Redis) {
			strictRedis := &StrictRedis{Redis: client}
			ok, err := strictRedis.Hdel("valid_key", "field")
			assert.Nil(t, err)
			assert.False(t, ok)
		})
	})

}

type TestStruct struct {
	Name string
	Age  int
}

func TestRedisValue_To_NilPointerDereference(t *testing.T) {
	r := &RedisValue{}
	err := r.To(nil)
	if err == nil || err.Error() != "value is nil" {
		t.Errorf("Expected error 'value is nil', got '%v'", err)
	}
}

func TestRedisValue_To_NonNilPointer(t *testing.T) {
	r := &RedisValue{}
	err := r.To(1)
	if err == nil || err != NilValueErr {
		t.Errorf("Expected error 'v must be a non-nil pointer', got '%v'", err)
	}
}

func TestRedisValue_To_UnhandledType(t *testing.T) {
	r := &RedisValue{result: make(chan int)}
	v := make([]int, 2)
	err := r.To(&v)
	if err == nil || err.Error() != "unhandled type" {
		t.Errorf("Expected error 'unhandled type', got '%v'", err)
	}
}

func TestRedisValue_To_DifferentLengths(t *testing.T) {
	r := &RedisValue{result: []string{"1", "2", "3"}}
	v := make([]int, 2)
	err := r.To(&v)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, v)
}
