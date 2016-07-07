package blameplugin_test

import (
	"bytes"
	"io/ioutil"
	"strings"

	"github.com/cloudfoundry/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/diego-blame/blameplugin"
)

var _ = Describe("DiegoBlame", func() {
	Describe("Given a DiegoBlame object", func() {
		Context("when calling Run w/ the wrong args", func() {
			var plgn *DiegoBlame
			BeforeEach(func() {
				plgn = new(DiegoBlame)
			})
			It("then it should panic and exit", func() {
				Ω(func() {
					cli := new(pluginfakes.FakeCliConnection)
					plgn.Run(cli, []string{})
				}).Should(Panic())
			})
		})

		Context("when calling Run w/ the proper args", func() {
			var plgn *DiegoBlame
			var b *bytes.Buffer
			BeforeEach(func() {
				b = new(bytes.Buffer)
				plgn = &DiegoBlame{
					Writer: b,
				}
			})
			It("then it should print a table with the correct columns", func() {
				cli := new(pluginfakes.FakeCliConnection)
				plgn.Run(cli, []string{"hi", "there"})
				for _, col := range []string{"app name/instance", "State", "Host:Port", "Org", "Space", "Disk-Usage", "Disk-Quota", "Mem-Usage", "Mem-Quota", "CPU-Usage", "Uptime", "URIs"} {
					Ω(strings.ToLower(b.String())).Should(ContainSubstring(strings.ToLower(col)))
				}
			})
			It("then it should print something", func() {
				cli := new(pluginfakes.FakeCliConnection)
				plgn.Run(cli, []string{"hi", "there"})
				Ω(b.String()).ShouldNot(BeEmpty())
			})
		})
	})

	Describe("Given CallAppsApi", func() {
		Context("when called with a valid endpoint and cli", func() {
			var guidArray []string
			BeforeEach(func() {
				cli := new(pluginfakes.FakeCliConnection)
				b, _ := ioutil.ReadFile("fixtures/app.json")
				cli.CliCommandWithoutTerminalOutputReturns([]string{string(b)}, nil)
				guidArray = CallAppsAPI("/v2/apps", cli)
			})

			It("should return a valid list of app guids", func() {
				Ω(len(guidArray)).Should(Equal(1))
				Ω(guidArray).Should(ConsistOf("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"))
			})
		})
	})

	Describe("Given CallStatsApi", func() {
		Context("when called with a valid guid and cli and selector and a host the app is running on", func() {
			var stats []AppStat
			BeforeEach(func() {
				cli := new(pluginfakes.FakeCliConnection)
				b, _ := ioutil.ReadFile("fixtures/stats.json")
				cli.CliCommandWithoutTerminalOutputReturns([]string{string(b)}, nil)
				stats = CallStatsAPI("xxx-xxx-xxxxx-xxxxx", cli, "192.xxx.x.255")
			})
			It("then it should return the stats for that app instance", func() {
				Ω(len(stats)).Should(Equal(1))
				Ω(stats[0].Stats.Name).Should(Equal("cool-server"))
			})
		})

		Context("when called with a valid guid and cli and selector and a host the app is NOT running on", func() {
			var stats []AppStat
			BeforeEach(func() {
				cli := new(pluginfakes.FakeCliConnection)
				b, _ := ioutil.ReadFile("fixtures/stats.json")
				cli.CliCommandWithoutTerminalOutputReturns([]string{string(b)}, nil)
				stats = CallStatsAPI("xxx-xxx-xxxxx-xxxxx", cli, "192.not.x.111")
			})
			It("then it should return no stats", func() {
				Ω(len(stats)).Should(Equal(0))
			})
		})
	})
})
