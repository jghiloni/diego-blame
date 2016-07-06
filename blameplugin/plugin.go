package blameplugin

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/krujos/cfcurl"
	"github.com/olekukonko/tablewriter"
	"github.com/xchapter7x/lo"
)

const name = "diego-blame"

type DiegoBlame struct {
	Version string
}

func (c *DiegoBlame) Run(cli plugin.CliConnection, args []string) {
	var appStatusArray []AppStat
	var hostSelector string

	if len(args) == 2 {
		hostSelector = args[1]

	} else {
		fmt.Println("invalid argument set: ", args)
		lo.G.Panic("invalid argument set: ", args)
	}

	guidArray := callAppsAPI("/v2/apps", cli)

	for i, guid := range guidArray {
		if i > 10 {
			break
		}
		appStatusArray = append(appStatusArray, callStatsAPI(guid, cli, hostSelector)...)
	}

	prettyColumnPrint(appStatusArray)
}

func prettyColumnPrint(appStatusArray []AppStat) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"AppName", "State", "Host:Port", "Disk-Usage", "Disk-Quota", "Mem-Usage", "Mem-Quota", "CPU-Usage", "Uptime", "URIs"})

	for _, status := range appStatusArray {
		row := []string{
			status.Stats.Name,
			status.State,
			fmt.Sprintf("%v:%v", status.Stats.Host, status.Stats.Port),
			fmt.Sprintf("%v", status.Stats.Usage.Disk),
			fmt.Sprintf("%v", status.Stats.DiskQuota),
			fmt.Sprintf("%v", status.Stats.Usage.Mem),
			fmt.Sprintf("%v", status.Stats.MemQuota),
			fmt.Sprintf("%v", status.Stats.Usage.CPU),
			fmt.Sprintf("%v", status.Stats.Uptime),
			fmt.Sprintf("%v", status.Stats.URIs),
		}
		table.Append(row)
	}
	table.Render()
}

func getStatsURL(guid string) (statsURL string) {
	return fmt.Sprintf("/v2/apps/%s/stats", guid)
}

func callStatsAPI(guid string, cli plugin.CliConnection, hostSelector string) (res []AppStat) {
	endpoint := getStatsURL(guid)

	if appStatsRaw, err := cfcurl.Curl(cli, endpoint); err == nil {
		appsStatsJson, _ := json.Marshal(appStatsRaw)
		var appStats = make(map[string]AppStat)
		json.Unmarshal(appsStatsJson, &appStats)

		for _, appStatObj := range appStats {

			if appStatObj.Stats.Host == hostSelector {
				res = append(res, appStatObj)
			}
		}
	} else {
		fmt.Println("err:", err)
	}
	return
}

func callAppsAPI(endpoint string, cli plugin.CliConnection) (res []string) {
	lo.G.Debugf("calling: %s", endpoint)
	var apps = new(Apps)
	if appsRaw, err := cfcurl.Curl(cli, endpoint); err == nil {
		lo.G.Debug("appsraw: ", appsRaw)

		if appsJson, err := json.Marshal(appsRaw); err == nil {
			lo.G.Debug("appsJson: ", appsJson)
			json.Unmarshal(appsJson, apps)
			lo.G.Debug("apps: ", apps)

			for _, resource := range apps.Resources {
				res = append(res, resource.Metadata.GUID)
			}

			if apps.NextURL != "" {
				res = append(res, callAppsAPI(apps.NextURL, cli)...)
			}
		} else {

			lo.G.Errorf("error in json marshal call to %s: %s", endpoint, err.Error())
		}
	} else {
		lo.G.Errorf("error in curl call to %s: %s", endpoint, err.Error())
	}
	return
}

type AppStat struct {
	State string `json:"state"`
	Stats Stats  `json:"stats"`
}
type Stats struct {
	Usage     Usage    `json:"usage"`
	Name      string   `json:"name"`
	URIs      []string `json:"uris"`
	Host      string   `json:"host"`
	Port      float64  `json:"port"`
	Uptime    float64  `json:"uptime"`
	MemQuota  float64  `json:"mem_quota"`
	DiskQuota float64  `json:"disk_quota"`
	FDSQuota  float64  `json:"fds_quota"`
}

type Usage struct {
	Disk float64   `json:"disk"`
	Mem  float64   `json:"mem"`
	CPU  float64   `json:"cpu"`
	Time time.Time `json:"time"`
}

type Apps struct {
	Total      float64    `json:"total_results"`
	TotalPages float64    `json:"total_pages"`
	NextURL    string     `json:"next_url"`
	Resources  []Resource `json:"resources"`
}

type Resource struct {
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	GUID string `json:"guid"`
}

func (c *DiegoBlame) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:    name,
		Version: c.GetVersionType(),
		Commands: []plugin.Command{
			plugin.Command{
				Name:     name,
				HelpText: "Run a scan to find all apps on a given diego cell in order to identify utilization spike causes",
				UsageDetails: plugin.Usage{
					Usage: fmt.Sprintf("cf %s 1.2.3.4", name),
				},
			},
		},
	}
}

func (c *DiegoBlame) GetVersionType() plugin.VersionType {
	versionArray := strings.Split(strings.TrimPrefix(c.Version, "v"), ".")
	major, _ := strconv.Atoi(versionArray[0])
	minor, _ := strconv.Atoi(versionArray[1])
	build, _ := strconv.Atoi(versionArray[2])
	return plugin.VersionType{
		Major: major,
		Minor: minor,
		Build: build,
	}
}
