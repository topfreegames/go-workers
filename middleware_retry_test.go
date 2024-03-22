package workers

import (
	"time"

	"github.com/customerio/gospec"
	. "github.com/customerio/gospec"
	"github.com/gomodule/redigo/redis"
)

func MiddlewareRetrySpec(c gospec.Context) {
	var panicingJob = (func(message *Msg) {
		panic("AHHHH")
	})

	var wares = NewMiddleware(
		&MiddlewareRetry{},
	)

	layout := "2006-01-02 15:04:05 MST"
	manager := newManager("myqueue", panicingJob, 1)
	worker := newWorker(manager)

	was := Config.Namespace
	Config.Namespace = "prod:"

	c.Specify("puts messages in retry queue when they fail", func() {
		message, _ := NewMsg("{\"jid\":\"2\",\"retry\":true}")

		wares.call("myqueue", message, func() {
			worker.process(message)
		})

		conn := Config.Pool.Get()
		defer conn.Close()

		retries, _ := redis.Strings(conn.Do("zrange", "prod:"+RETRY_KEY, 0, 1))
		c.Expect(retries[0], Equals, message.ToJson())
	})

	c.Specify("allows disabling retries", func() {
		message, _ := NewMsg("{\"jid\":\"2\",\"retry\":false}")

		wares.call("myqueue", message, func() {
			worker.process(message)
		})

		conn := Config.Pool.Get()
		defer conn.Close()

		count, _ := redis.Int(conn.Do("zcard", "prod:"+RETRY_KEY))
		c.Expect(count, Equals, 0)
	})

	c.Specify("doesn't retry by default", func() {
		message, _ := NewMsg("{\"jid\":\"2\"}")

		wares.call("myqueue", message, func() {
			worker.process(message)
		})

		conn := Config.Pool.Get()
		defer conn.Close()

		count, _ := redis.Int(conn.Do("zcard", "prod:"+RETRY_KEY))
		c.Expect(count, Equals, 0)
	})

	c.Specify("allows numeric retries", func() {
		message, _ := NewMsg("{\"jid\":\"2\",\"retry\":5}")

		wares.call("myqueue", message, func() {
			worker.process(message)
		})

		conn := Config.Pool.Get()
		defer conn.Close()

		retries, _ := redis.Strings(conn.Do("zrange", "prod:"+RETRY_KEY, 0, 1))
		c.Expect(retries[0], Equals, message.ToJson())
	})

	c.Specify("handles new failed message", func() {
		message, _ := NewMsg("{\"jid\":\"2\",\"retry\":true}")

		wares.call("myqueue", message, func() {
			worker.process(message)
		})

		conn := Config.Pool.Get()
		defer conn.Close()

		retries, _ := redis.Strings(conn.Do("zrange", "prod:"+RETRY_KEY, 0, 1))
		message, _ = NewMsg(retries[0])

		queue, _ := message.Get("queue").String()
		error_message, _ := message.Get("error_message").String()
		error_class, _ := message.Get("error_class").String()
		retry_count, _ := message.Get("retry_count").Int()
		error_backtrace, _ := message.Get("error_backtrace").String()
		failed_at, _ := message.Get("failed_at").String()

		c.Expect(queue, Equals, "myqueue")
		c.Expect(error_message, Equals, "AHHHH")
		c.Expect(error_class, Equals, "")
		c.Expect(retry_count, Equals, 0)
		c.Expect(error_backtrace, Equals, "")
		c.Expect(failed_at, Equals, time.Now().UTC().Format(layout))
	})

	c.Specify("handles recurring failed message", func() {
		message, _ := NewMsg("{\"jid\":\"2\",\"retry\":true,\"queue\":\"default\",\"error_message\":\"bam\",\"failed_at\":\"2013-07-20 14:03:42 UTC\",\"retry_count\":10}")

		wares.call("myqueue", message, func() {
			worker.process(message)
		})

		conn := Config.Pool.Get()
		defer conn.Close()

		retries, _ := redis.Strings(conn.Do("zrange", "prod:"+RETRY_KEY, 0, 1))
		message, _ = NewMsg(retries[0])

		queue, _ := message.Get("queue").String()
		error_message, _ := message.Get("error_message").String()
		retry_count, _ := message.Get("retry_count").Int()
		failed_at, _ := message.Get("failed_at").String()
		retried_at, _ := message.Get("retried_at").String()

		c.Expect(queue, Equals, "myqueue")
		c.Expect(error_message, Equals, "AHHHH")
		c.Expect(retry_count, Equals, 11)
		c.Expect(failed_at, Equals, "2013-07-20 14:03:42 UTC")
		c.Expect(retried_at, Equals, time.Now().UTC().Format(layout))
	})

	c.Specify("handles recurring failed message with customized max", func() {
		message, _ := NewMsg("{\"jid\":\"2\",\"retry\":true,\"queue\":\"default\",\"error_message\":\"bam\",\"failed_at\":\"2013-07-20 14:03:42 UTC\",\"retry_count\":8,\"retry_max\":10}")

		wares.call("myqueue", message, func() {
			worker.process(message)
		})

		conn := Config.Pool.Get()
		defer conn.Close()

		retries, _ := redis.Strings(conn.Do("zrange", "prod:"+RETRY_KEY, 0, 1))
		message, _ = NewMsg(retries[0])

		queue, _ := message.Get("queue").String()
		error_message, _ := message.Get("error_message").String()
		retry_count, _ := message.Get("retry_count").Int()
		failed_at, _ := message.Get("failed_at").String()
		retried_at, _ := message.Get("retried_at").String()

		c.Expect(queue, Equals, "myqueue")
		c.Expect(error_message, Equals, "AHHHH")
		c.Expect(retry_count, Equals, 9)
		c.Expect(failed_at, Equals, "2013-07-20 14:03:42 UTC")
		c.Expect(retried_at, Equals, time.Now().UTC().Format(layout))
	})

	c.Specify("doesn't retry after default number of retries", func() {
		message, _ := NewMsg("{\"jid\":\"2\",\"retry\":true,\"retry_count\":25}")

		wares.call("myqueue", message, func() {
			worker.process(message)
		})

		conn := Config.Pool.Get()
		defer conn.Close()

		count, _ := redis.Int(conn.Do("zcard", "prod:"+RETRY_KEY))
		c.Expect(count, Equals, 0)
	})

	c.Specify("doesn't retry after customized number of retries", func() {
		message, _ := NewMsg("{\"jid\":\"2\",\"retry\":true,\"retry_max\":3,\"retry_count\":3}")

		wares.call("myqueue", message, func() {
			worker.process(message)
		})

		conn := Config.Pool.Get()
		defer conn.Close()

		count, _ := redis.Int(conn.Do("zcard", "prod:"+RETRY_KEY))
		c.Expect(count, Equals, 0)
	})

	c.Specify("use retry_options when provided - min_delay", func() {
		message, _ := NewMsg("{\"jid\":\"2\",\"retry\":true,\"retry_options\":{\"exp\":2,\"min_delay\":1200,\"max_rand\":0}}")
		var now int
		wares.call("myqueue", message, func() {
			worker.process(message)
			now = int(time.Now().Unix())
		})

		conn := Config.Pool.Get()
		defer conn.Close()

		count, _ := redis.Int(conn.Do("zcard", "prod:"+RETRY_KEY))
		c.Expect(count, Equals, 1)
		values, _ := redis.Values(conn.Do("ZRANGE", "prod:"+RETRY_KEY, 0, -1, "WITHSCORES"))

		msg := ""
		nextAt := -1.0
		redis.Scan(values, &msg, &nextAt)
		c.Expect(int(nextAt), Equals, now+1200)
	})

	c.Specify("use retry_options when provided - exp", func() {
		message, _ := NewMsg("{\"jid\":\"2\",\"retry\":true,\"retry_count\":2,\"retry_options\":{\"exp\":2,\"min_delay\":0,\"max_rand\":0}}")
		var now int
		wares.call("myqueue", message, func() {
			worker.process(message)
			now = int(time.Now().Unix())
		})

		conn := Config.Pool.Get()
		defer conn.Close()

		count, _ := redis.Int(conn.Do("zcard", "prod:"+RETRY_KEY))
		c.Expect(count, Equals, 1)
		values, _ := redis.Values(conn.Do("ZRANGE", "prod:"+RETRY_KEY, 0, -1, "WITHSCORES"))

		msg := ""
		nextAt := -1.0
		redis.Scan(values, &msg, &nextAt)
		c.Expect(int(nextAt), Equals, now+9)
	})

	c.Specify("use retry_options when provided - max_delay", func() {
		message, _ := NewMsg("{\"jid\":\"2\",\"retry\":true,\"retry_options\":{\"exp\":20,\"min_delay\":600,\"max_delay\":10,\"max_rand\":10000}}")
		var now int
		wares.call("myqueue", message, func() {
			worker.process(message)
			now = int(time.Now().Unix())
		})

		conn := Config.Pool.Get()
		defer conn.Close()

		count, _ := redis.Int(conn.Do("zcard", "prod:"+RETRY_KEY))
		c.Expect(count, Equals, 1)
		values, _ := redis.Values(conn.Do("ZRANGE", "prod:"+RETRY_KEY, 0, -1, "WITHSCORES"))

		msg := ""
		nextAt := -1.0
		redis.Scan(values, &msg, &nextAt)
		c.Expect(int(nextAt), Equals, now+10)
	})

	c.Specify("use retry_options when provided - max_rand", func() {
		message, _ := NewMsg("{\"jid\":\"2\",\"retry\":true,\"retry_options\":{\"exp\":2,\"min_delay\":0,\"max_rand\":100}}")
		var now int
		wares.call("myqueue", message, func() {
			worker.process(message)
			now = int(time.Now().Unix())
		})

		conn := Config.Pool.Get()
		defer conn.Close()

		count, _ := redis.Int(conn.Do("zcard", "prod:"+RETRY_KEY))
		c.Expect(count, Equals, 1)
		values, _ := redis.Values(conn.Do("ZRANGE", "prod:"+RETRY_KEY, 0, -1, "WITHSCORES"))

		msg := ""
		nextAt := -1.0
		redis.Scan(values, &msg, &nextAt)
		c.Expect(nextAt, Satisfies, nextAt <= float64(now+100))
		c.Expect(nextAt, Satisfies, nextAt >= float64(now+0))
	})

	Config.Namespace = was
}
