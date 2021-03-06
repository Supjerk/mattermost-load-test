package ltparse

import (
	"encoding/json"
	"io"
	"sort"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-load-test/loadtest"
)

type ResultsConfig struct {
	Input         io.Reader
	BaselineInput io.Reader
	Output        io.Writer
	Display       string
	Aggregate     bool
	Verbose       bool
}

type templateData struct {
	Actual   *loadtest.RouteStats
	Baseline *loadtest.RouteStats
	Verbose  bool
}

func sortedRoutes(routesMap map[string]*loadtest.RouteStats) []*loadtest.RouteStats {
	routeNames := make([]string, 0, len(routesMap))
	for routeName := range routesMap {
		routeNames = append(routeNames, routeName)
	}
	sort.Strings(routeNames)

	routes := make([]*loadtest.RouteStats, 0, len(routesMap))
	for _, routeName := range routeNames {
		routes = append(routes, routesMap[routeName])
	}

	return routes
}

func parseTimings(input io.Reader) ([]*loadtest.ClientTimingStats, error) {
	allTimings := make(map[string]*loadtest.ClientTimingStats)
	decoder := json.NewDecoder(input)
	foundStructuredLogs := false
	for decoder.More() {
		log := map[string]interface{}{}
		if err := decoder.Decode(&log); err != nil {
			return nil, errors.Wrap(err, "failed to decode")
		}
		foundStructuredLogs = true

		// Look for result logs
		if log["tag"] == "timings" {
			timings := &loadtest.ClientTimingStats{}
			if err := mapstructure.Decode(log["timings"], timings); err != nil {
				continue
			}

			var instanceId string
			if instanceIdValue := log["instance_id"]; instanceIdValue != nil {
				instanceId = instanceIdValue.(string)
			}
			if instanceId == "" {
				instanceId = "default"
			}

			allTimings[instanceId] = allTimings[instanceId].Merge(timings)
		}
	}

	if !foundStructuredLogs {
		return nil, errors.New("failed to find structured logs")
	}
	if len(allTimings) == 0 {
		return nil, errors.New("failed to find results")
	}

	allTimingsList := make([]*loadtest.ClientTimingStats, 0, len(allTimings))
	for _, timings := range allTimings {
		allTimingsList = append(allTimingsList, timings)
	}

	return allTimingsList, nil
}

func ParseResults(config *ResultsConfig) error {
	allTimings, err := parseTimings(config.Input)
	if err != nil {
		return err
	}

	allBaselineTimings := []*loadtest.ClientTimingStats{}
	if config.BaselineInput != nil {
		allBaselineTimings, err = parseTimings(config.BaselineInput)
		if err != nil {
			return err
		}
	}

	var timings *loadtest.ClientTimingStats
	if !config.Aggregate {
		timings = allTimings[len(allTimings)-1]
	} else {
		for _, t := range allTimings {
			timings = timings.Merge(t)
		}
	}

	var baselineTimings *loadtest.ClientTimingStats
	if len(allBaselineTimings) > 0 {
		if !config.Aggregate {
			baselineTimings = allBaselineTimings[len(allBaselineTimings)-1]
		} else {
			for _, t := range allBaselineTimings {
				baselineTimings = timings.Merge(t)
			}
		}
	}

	timings.CalcResults()
	if baselineTimings != nil {
		baselineTimings.CalcResults()
	}

	switch config.Display {
	case "markdown":
		if err := dumpTimingsMarkdown(timings, baselineTimings, config.Output, config.Verbose); err != nil {
			return errors.Wrap(err, "failed to dump timings")
		}
	case "text":
		if len(allBaselineTimings) > 0 {
			return errors.New("cannot compare to baseline using text display")
		}
		fallthrough
	default:
		if err := dumpTimingsText(timings, config.Output, config.Verbose); err != nil {
			return errors.Wrap(err, "failed to dump timings")
		}
	}

	return nil
}
