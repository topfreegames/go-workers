package workers

import (
	"reflect"

	"github.com/customerio/gospec"
	. "github.com/customerio/gospec"
)

var called chan bool

func myJob(_ *Msg) {
	called <- true
}

func WorkersSpec(c gospec.Context) {
	const queueName = "queue-workers"

	c.Specify("Workers", func() {
		c.Specify("allows running in tests", func() {
			called = make(chan bool)

			Process(queueName, myJob, 10)

			Start()

			_, err := Enqueue(queueName, "Add", []int{1, 2})
			if err != nil {
				panic(err)
			}
			<-called

			Quit()
		})

		// TODO make this test more deterministic, randomly locks up in travis.
		//c.Specify("allows starting and stopping multiple times", func() {
		//	called = make(chan bool)

		//	Process(queueName, myJob, 10)

		//	Start()
		//	Quit()

		//	Start()

		//	Enqueue(queueName, "Add", []int{1, 2})
		//	<-called

		//	Quit()
		//})

		c.Specify("runs beforeStart hooks", func() {
			hooks := []string{}

			BeforeStart(func() {
				hooks = append(hooks, "1")
			})
			BeforeStart(func() {
				hooks = append(hooks, "2")
			})
			BeforeStart(func() {
				hooks = append(hooks, "3")
			})

			Start()

			c.Expect(reflect.DeepEqual(hooks, []string{"1", "2", "3"}), IsTrue)

			Quit()

			// Clear out global hooks variable
			beforeStart = nil
		})

		c.Specify("runs beforeStart hooks", func() {
			hooks := []string{}

			DuringDrain(func() {
				hooks = append(hooks, "1")
			})
			DuringDrain(func() {
				hooks = append(hooks, "2")
			})
			DuringDrain(func() {
				hooks = append(hooks, "3")
			})

			Start()

			c.Expect(reflect.DeepEqual(hooks, []string{}), IsTrue)

			Quit()

			c.Expect(reflect.DeepEqual(hooks, []string{"1", "2", "3"}), IsTrue)

			// Clear out global hooks variable
			duringDrain = nil
		})
	})
}
