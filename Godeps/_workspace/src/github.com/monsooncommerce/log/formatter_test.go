package log

import (
	"fmt"
	"os"
	"testing"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestFormatter(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Default Formatter", func() {
		g.It("should return the default format", func() {
			h, _ := os.Hostname()
			t := os.Args[0]
			p := os.Getpid()
			m := "test debug message"

			f := &DefaultFormat{
				hostname: h,
				pid:      p,
				tag:      t,
			}

			line := f.Format(Debug, m)
			Expect(line).To(ContainSubstring(fmt.Sprintf("%s %s[%d]: %s [%v]", h, t, p, Debug, m)))
		})
	})
}
