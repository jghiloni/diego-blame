package blameplugin

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
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
	Writer  io.Writer
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

	guidArray := CallAppsAPI("/v2/apps", cli)

	for _, guid := range guidArray {
		appStatusArray = append(appStatusArray, CallStatsAPI(guid, cli, hostSelector)...)
	}

	prettyColumnPrint(appStatusArray, cli, c.Writer)
}

func prettyColumnPrint(appStatusArray []AppStat, cli plugin.CliConnection, writer io.Writer) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"App Name/Instance", "State", "Host:Port", "Org", "Space", "Disk-Usage", "Disk-Quota", "Mem-Usage", "Mem-Quota", "CPU-Usage", "Uptime", "URIs"})

	for _, status := range appStatusArray {
		location := GetLocation(status.AppGUID, cli)
		row := []string{
			fmt.Sprintf("%v/%v", status.Stats.Name, status.AppInstance),
			status.State,
			fmt.Sprintf("%v:%v", status.Stats.Host, status.Stats.Port),
			location.Org,
			location.Space,
			formatBytes(status.Stats.Usage.Disk),
			formatBytes(status.Stats.DiskQuota),
			formatBytes(status.Stats.Usage.Mem),
			formatBytes(status.Stats.MemQuota),
			formatPercent(status.Stats.Usage.CPU),
			formatTime(status.Stats.Uptime),
			fmt.Sprintf("%v", status.Stats.URIs),
		}
		table.Append(row)
	}
	table.Render()
}

func getStatsURL(guid string) (statsURL string) {
	return fmt.Sprintf("/v2/apps/%s/stats", guid)
}

func getAppUrl(guid string) (appUrl string) {
	return fmt.Sprintf("/v2/apps/%s", guid)
}

func GetLocation(guid string, cli plugin.CliConnection) (loc AppLocation) {
	endpoint := getAppUrl(guid)

	if appRaw, err := cfcurl.Curl(cli, endpoint); err == nil {
		appJson, _ := json.Marshal(appRaw)
		var app = &AppResource{}
		json.Unmarshal(appJson, app)

		spaceName, orgUrl := getSpaceInfo(app.Entity.SpaceUrl, cli)
		orgName := getOrgName(orgUrl, cli)

		loc = AppLocation{
			Org:   orgName,
			Space: spaceName,
		}
	} else {
		fmt.Println("err:", err)
	}
	return
}

func getSpaceInfo(spaceUrl string, cli plugin.CliConnection) (name string, orgUrl string) {
	if spaceRaw, err := cfcurl.Curl(cli, spaceUrl); err == nil {
		spaceJson, _ := json.Marshal(spaceRaw)

		var space = &SpaceResource{}
		json.Unmarshal(spaceJson, space)

		name = space.Entity.Name
		orgUrl = space.Entity.OrganizationUrl

	} else {
		fmt.Println("err:", err)
	}
	return
}

func getOrgName(orgUrl string, cli plugin.CliConnection) (name string) {
	if orgRaw, err := cfcurl.Curl(cli, orgUrl); err == nil {
		orgJson, _ := json.Marshal(orgRaw)

		var org = &Resource{}
		json.Unmarshal(orgJson, org)

		name = org.Entity.Name
	} else {
		fmt.Println("err:", err)
	}
	return
}

func CallStatsAPI(guid string, cli plugin.CliConnection, hostSelector string) (res []AppStat) {
	endpoint := getStatsURL(guid)

	if appStatsRaw, err := cfcurl.Curl(cli, endpoint); err == nil {
		appsStatsJson, _ := json.Marshal(appStatsRaw)
		var appStats = make(map[string]AppStat)
		json.Unmarshal(appsStatsJson, &appStats)

		for instanceIndex, appStatObj := range appStats {

			if appStatObj.Stats.Host == hostSelector {
				appStatObj.AppGUID = guid
				appStatObj.AppInstance = instanceIndex
				res = append(res, appStatObj)
			}
		}
	} else {
		fmt.Println("err:", err)
	}
	return
}

func CallAppsAPI(endpoint string, cli plugin.CliConnection) (res []string) {
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
				res = append(res, CallAppsAPI(apps.NextURL, cli)...)
			}
		} else {

			lo.G.Errorf("error in json marshal call to %s: %s", endpoint, err.Error())
		}
	} else {
		lo.G.Errorf("error in curl call to %s: %s", endpoint, err.Error())
	}
	return
}

func formatBytes(numBytes float64) string {
	suffixes := []string{"B", "KB", "MB", "GB", "TB"}

	suffixIndex := math.Floor(math.Log(numBytes) / math.Log(1024))
	divisor := math.Pow(1024, suffixIndex)

	return fmt.Sprintf("%v %v", round(numBytes/divisor, 3), suffixes[int(suffixIndex)])
}

func round(num float64, decPlaces float64) float64 {
	mult := math.Pow(10, decPlaces)

	return math.Floor(num*mult) / mult
}

func formatPercent(absolute float64) string {
	return fmt.Sprintf("%v%%", round(absolute*100, 2))
}

func formatTime(seconds float64) string {
	days := int(seconds / 86400)
	seconds = math.Mod(seconds, 86400)

	hours := int(seconds / 3600)
	seconds = math.Mod(seconds, 3600)

	minutes := int(seconds / 60)
	seconds = math.Mod(seconds, 60)

	return fmt.Sprintf("%d:%02d:%02d:%02d", days, hours, minutes, int(seconds))
}

type AppRecord struct {
	Location     AppLocation
	AppInstances []AppStat
}

type AppLocation struct {
	Org   string
	Space string
}

type AppStat struct {
	AppGUID     string
	AppInstance string
	State       string `json:"state"`
	Stats       Stats  `json:"stats"`
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
	Total      float64       `json:"total_results"`
	TotalPages float64       `json:"total_pages"`
	NextURL    string        `json:"next_url"`
	Resources  []AppResource `json:"resources"`
}

type Resource struct {
	Metadata Metadata `json:"metadata"`
	Entity   Entity   `json:"entity"`
}

type AppResource struct {
	Metadata Metadata  `json:"metadata"`
	Entity   AppEntity `json:"entity"`
}

type SpaceResource struct {
	Metadata Metadata    `json:"metadata"`
	Entity   SpaceEntity `json:"entity"`
}

type Entity struct {
	Name string `json:"name"`
}

type AppEntity struct {
	Name     string `json:"name"`
	SpaceUrl string `json:"space_url"`
}

type SpaceEntity struct {
	Name            string `json:"name"`
	OrganizationUrl string `json:"organization_url"`
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
