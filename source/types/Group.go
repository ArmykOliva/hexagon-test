package types

import "tholian-endpoint/console"
import "strconv"
import "strings"

type Group struct {
	ID       uint16 `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Type     string `json:"type"`
}

func NewGroup() Group {

	var group Group

	group.SetType("user")

	return group

}

func ToGroup(name string) Group {

	var group Group

	group.SetName(name)
	group.SetType("user")

	return group

}

func (group *Group) Debug() {

	if group.Name != "" {

		if group.ID == 0 && group.Name != "root" {
			console.Warn("groups/" + group.Name + ": Invalid ID " + strconv.FormatUint(uint64(group.ID), 10))
		}

	}

}

func (group *Group) IsValid() bool {

	if group.Name != "" {

		var result bool = true

		if group.ID == 0 && group.Name != "root" {
			result = false
		}

		return result

	}

	return false

}

func (group *Group) SetID(value uint16) {

	// User "nobody" is last user id
	if value >= 0 && value <= 65534 {
		group.ID = value
	}

}

func (group *Group) SetName(value string) {
	group.Name = strings.TrimSpace(value)
}

func (group *Group) SetPassword(value string) {
	group.Password = value
}

func (group *Group) SetType(value string) {

	if value == "user" {
		group.Type = "user"
	} else if value == "program" {
		group.Type = "program"
	}

}
