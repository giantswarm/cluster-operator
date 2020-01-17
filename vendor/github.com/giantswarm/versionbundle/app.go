package versionbundle

import "encoding/json"

type App struct {
	App              string `yaml:"app"`
	ComponentVersion string `yaml:"componentVersion"`
	Version          string `yaml:"version"`
}

func CopyApps(apps []App) []App {
	raw, err := json.Marshal(apps)
	if err != nil {
		panic(err)
	}

	var appList []App
	err = json.Unmarshal(raw, &appList)
	if err != nil {
		panic(err)
	}

	return appList
}
