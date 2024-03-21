package workers

import (
	"github.com/customerio/gospec"
	. "github.com/customerio/gospec"
)

func ConfigSpec(c gospec.Context) {
	var recoverOnPanic = func(f func()) (err interface{}) {
		defer func() {
			if cause := recover(); cause != nil {
				err = cause
			}
		}()

		f()

		return
	}

	c.Specify("sets redis pool size which defaults to 1", func() {
		c.Expect(Config.Pool.MaxIdle, Equals, 1)

		Configure(Options{
			Address:   "localhost:6379",
			ProcessID: "1",
			PoolSize:  20,
		})

		c.Expect(Config.Pool.MaxIdle, Equals, 20)
	})

	c.Specify("can specify custom process", func() {
		c.Expect(Config.processId, Equals, "1")

		Configure(Options{
			Address:   "localhost:6379",
			ProcessID: "2",
		})

		c.Expect(Config.processId, Equals, "2")
	})

	c.Specify("requires a server parameter", func() {
		err := recoverOnPanic(func() {
			Configure(Options{ProcessID: "2"})
		})

		c.Expect(err, Equals, "Configure requires a 'Address' option, which identifies a Redis instance")
	})

	c.Specify("requires a process parameter", func() {
		err := recoverOnPanic(func() {
			Configure(Options{Address: "localhost:6379"})
		})

		c.Expect(err, Equals, "Configure requires a 'ProcessID' option, which uniquely identifies this instance")
	})

	c.Specify("adds ':' to the end of the namespace", func() {
		c.Expect(Config.Namespace, Equals, "")

		Configure(Options{
			Address:   "localhost:6379",
			ProcessID: "1",
			Namespace: "prod",
		})

		c.Expect(Config.Namespace, Equals, "prod:")
	})

	c.Specify("defaults poll interval to 15 seconds", func() {
		Configure(Options{
			Address:   "localhost:6379",
			ProcessID: "1",
		})

		c.Expect(Config.PoolInterval, Equals, 15)
	})

	c.Specify("allows customization of poll interval", func() {
		Configure(Options{
			Address:      "localhost:6379",
			ProcessID:    "1",
			PoolInterval: 1,
		})

		c.Expect(Config.PoolInterval, Equals, 1)
	})
}
