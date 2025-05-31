package structs

import "tholian-endpoint/console"
import "tholian-endpoint/types"
import utils_encoding "tholian-endpoint/utils/encoding"
import "net/url"
import "strings"

type Update struct {
	Name         string             `json:"name"`
	Version      types.Version      `json:"version"`
	Architecture types.Architecture `json:"architecture"`
	Manager      types.Manager      `json:"manager"`
	URL          string             `json:"url"`
}

func NewUpdate(manager string) Update {

	var update Update

	update.SetArchitecture("any")
	update.SetManager(manager)

	return update

}

func (update *Update) Debug() {

	if update.Name != "" {

		if update.Version.IsValid() == false {
			console.Error("updates/" + update.Name + ": Invalid Version")
			console.Error(utils_encoding.ToJSON(update.Version))
		}

		if update.Architecture.IsValid() == false {
			console.Error("updates/" + update.Name + ": Invalid Architecture")
			console.Error(utils_encoding.ToJSON(update.Architecture))
		}

		if update.Manager.IsValid() == false {
			console.Error("updates/" + update.Name + ": Invalid Manager")
			console.Error(utils_encoding.ToJSON(update.Manager))
		}

		if update.URL == "" {
			console.Error("updates/" + update.Name + ": Invalid URL")
		}

	}

}

func (update *Update) IsValid() bool {

	if update.Name != "" {

		var result bool = true

		if update.Version.IsValid() == false {
			result = false
		}

		if update.Architecture.IsValid() == false {
			result = false
		}

		if update.Manager.IsValid() == false {
			result = false
		}

		return result

	}

	return false

}

func (update *Update) SetArchitecture(value string) {

	architecture := types.ParseArchitecture(value)

	if architecture != nil {
		update.Architecture = *architecture
	}

}

func (update *Update) SetManager(value string) {

	manager := types.ParseManager(value)

	if manager != nil {
		update.Manager = *manager
	}

}

func (update *Update) SetName(value string) {
	update.Name = strings.TrimSpace(value)
}

func (update *Update) SetURL(value string) {

	tmp, err := url.ParseRequestURI(value)

	if err == nil {

		if tmp.Scheme == "https" || tmp.Scheme == "http" {
			update.URL = tmp.String()
		}

	}

}

func (update *Update) SetVersion(value string) {

	version := types.ToVersion(value)

	if version.IsValid() {
		update.Version = version
	}

}
