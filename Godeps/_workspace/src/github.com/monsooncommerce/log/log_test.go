package log

import (
	"fmt"
	"os"
	"testing"

	. "github.com/franela/goblin"
	"github.com/monsooncommerce/mockwriter"
	. "github.com/onsi/gomega"
)

func TestLogging(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Logging levels", func() {
		var h, t string
		var p int
		var m *mockwriter.MockWriter
		var logger *Log

		g.BeforeEach(func() {
			h, _ = os.Hostname()
			t = os.Args[0]
			p = os.Getpid()
			m = &mockwriter.MockWriter{}
			logger = New(m, Debug)
		})

		g.It("should write a debug message", func() {
			logger.Debug("debug message")
			Expect(m.Written).To(ContainSubstring("DEBUG [[[debug message]]]"))
		})

		g.It("should write a formatted debug message", func() {
			logger.Debugf("additional: %s", "debug message")
			Expect(m.Written).To(ContainSubstring("additional: [debug message]"))
			Expect(m.Written).To(ContainSubstring("DEBUG"))
		})

		g.It("should write an info message", func() {
			logger.Info("info message")
			Expect(m.Written).To(ContainSubstring("INFO [[[info message]]]"))
		})

		g.It("should write a formatted info message", func() {
			logger.Infof("additional: %s", "info message")
			Expect(m.Written).To(ContainSubstring("additional: [info message]"))
			Expect(m.Written).To(ContainSubstring("INFO"))
		})

		g.It("should write a notice message", func() {
			logger.Notice("notice message")
			Expect(m.Written).To(ContainSubstring("NOTICE [[[notice message]]]"))
		})

		g.It("should write a formatted notice message", func() {
			logger.Noticef("additional: %s", "notice message")
			Expect(m.Written).To(ContainSubstring("additional: [notice message]"))
			Expect(m.Written).To(ContainSubstring("NOTICE"))
		})

		g.It("should write a warning message", func() {
			logger.Warning("warning message")
			Expect(m.Written).To(ContainSubstring("WARNING [[[warning message]]]"))
		})

		g.It("should write a formatted warning message", func() {
			logger.Warningf("additional: %s", "warning message")
			Expect(m.Written).To(ContainSubstring("additional: [warning message]"))
			Expect(m.Written).To(ContainSubstring("WARNING"))
		})

		g.It("should write a error message", func() {
			logger.Error("error message")
			Expect(m.Written).To(ContainSubstring("ERROR [[[error message]]]"))
		})

		g.It("should write a formatted error message", func() {
			logger.Errorf("additional: %s", "error message")
			Expect(m.Written).To(ContainSubstring("additional: [error message]"))
			Expect(m.Written).To(ContainSubstring("ERROR"))
		})
	})

	g.Describe("Thresholds", func() {
		g.It("should write debug and higher severity", func() {
			m := &mockwriter.MockWriter{}
			logger := New(m, Debug)
			logger.Debug("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("DEBUG"))

			logger.Info("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("INFO"))

			logger.Notice("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("NOTICE"))

			logger.Warning("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("WARNING"))

			logger.Error("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("ERROR"))
		})

		g.It("should write info and higher severity", func() {
			m := &mockwriter.MockWriter{}
			logger := New(m, Info)
			logger.Debug("test message")
			Expect(m.Written).To(BeNil())

			logger.Info("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("INFO"))

			logger.Notice("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("NOTICE"))

			logger.Warning("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("WARNING"))

			logger.Error("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("ERROR"))
		})

		g.It("should write notice and higher severity", func() {
			m := &mockwriter.MockWriter{}
			logger := New(m, Notice)
			logger.Debug("test message")
			Expect(m.Written).To(BeNil())

			logger.Info("test message")
			Expect(m.Written).To(BeNil())

			logger.Notice("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("NOTICE"))

			logger.Warning("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("WARNING"))

			logger.Error("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("ERROR"))
		})

		g.It("should write warning and higher severity", func() {
			m := &mockwriter.MockWriter{}
			logger := New(m, Warning)
			logger.Debug("test message")
			Expect(m.Written).To(BeNil())

			logger.Info("test message")
			Expect(m.Written).To(BeNil())

			logger.Notice("test message")
			Expect(m.Written).To(BeNil())

			logger.Warning("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("WARNING"))

			logger.Error("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("ERROR"))
		})

		g.It("should write error and higher severity", func() {
			m := &mockwriter.MockWriter{}
			logger := New(m, Error)
			logger.Debug("test message")
			Expect(m.Written).To(BeNil())

			logger.Info("test message")
			Expect(m.Written).To(BeNil())

			logger.Notice("test message")
			Expect(m.Written).To(BeNil())

			logger.Warning("test message")
			Expect(m.Written).To(BeNil())

			logger.Error("test message")
			Expect(m.Written).To(ContainSubstring("test message"))
			Expect(m.Written).To(ContainSubstring("ERROR"))
		})
	})

	g.Describe("Custom Formatter", func() {
		g.It("should allow setting a custom formatter", func() {
			m := &mockwriter.MockWriter{}
			f := &CustomFormat{}
			logger := New(m, Debug)
			logger.SetFormatter(f)

			logger.Debug("test message")
			Expect(f.Formatted).To(Equal(1))
			Expect(m.Written).To(ContainSubstring("Custom: [DEBUG] -- [[[test message]]]"))

			logger.Error("test message")
			Expect(f.Formatted).To(Equal(2))
			Expect(m.Written).To(ContainSubstring("Custom: [ERROR] -- [[[test message]]]"))
		})
	})
}

type CustomFormat struct {
	Formatted int
}

func (c *CustomFormat) Format(l Level, args ...interface{}) string {
	c.Formatted++
	return fmt.Sprintf("Custom: [%s] -- %s", l, args)
}
