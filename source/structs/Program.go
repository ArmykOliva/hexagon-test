package structs

import "tholian-endpoint/console"
import "tholian-endpoint/matchers"
import utils_path "tholian-endpoint/utils/path"
import utils_encoding "tholian-endpoint/utils/encoding"
import "strings"

type Program struct {
	PID          uint                  `json:"pid"`         // 1337
	Name         string                `json:"name"`        // stealth.mjs
	Command      string                `json:"command"`     // /usr/bin/node
	Arguments    []string              `json:"arguments"`   // [ "stealth/stealth.mjs", "serve" ]
	Folder       string                `json:"folder"`      // /home/cookiengineer/Software/tholian-network/stealth
	Environment  map[string]string     `json:"environment"` // { LD_PRELOAD: "/tmp/evil.so" }
	User         matchers.User         `json:"user"`
	Manager      matchers.Manager      `json:"manager"`
	Filesystem   []string              `json:"filesystem"`
	Connections  []matchers.Connection `json:"connections"`
	Dependencies []matchers.Package    `json:"dependencies"`
	Packages     []Package             `json:"packages"`
}

func NewProgram(pid uint, name string) Program {

	var program Program

	program.PID = pid
	program.SetName(name)
	program.Arguments = make([]string, 0)
	program.Environment = make(map[string]string)
	program.Filesystem = make([]string, 0)
	program.Connections = make([]matchers.Connection, 0)
	program.Dependencies = make([]matchers.Package, 0)
	program.Packages = make([]Package, 0)

	return program

}

func (program *Program) Debug() {

	if program.Name != "" {

		if program.PID == 0 {
			console.Error("programs/" + program.Name + ": Invalid PID")
		}

		if strings.HasPrefix(program.Command, "/") == false {
			console.Error("programs/" + program.Name + ": Invalid Command \"" + program.Command + "\"")
		}

		if strings.HasPrefix(program.Folder, "/") == false {
			console.Error("programs/" + program.Name + ": Invalid Folder \"" + program.Folder + "\"")
		}

		if len(program.Environment) > 0 {

			for key, val := range program.Environment {

				if len(val) == 0 {
					console.Error("programs/" + program.Name + ": Invalid Environment key value for \"" + key + "\"")
				}

			}

		}

		if program.User.IsValid() == false {
			console.Error("programs/" + program.Name + ": Invalid User")
			console.Error(utils_encoding.ToJSON(program.User))
		}

		if program.Manager.IsValid() == false {
			console.Error("programs/" + program.Name + ": Invalid Manager")
			console.Error(utils_encoding.ToJSON(program.Manager))
		}

		if len(program.Filesystem) > 0 {

			for f := 0; f < len(program.Filesystem); f++ {

				file := program.Filesystem[f]

				if utils_path.IsWatchedFile(file) == false {
					console.Error("programs/" + program.Name + ": Invalid File \"" + file + "\"")
				}

			}

		}

		for c := 0; c < len(program.Connections); c++ {

			connection := program.Connections[c]

			if connection.IsValid() == false {
				console.Error("programs/" + program.Name + ": Invalid Connection")
				console.Error(utils_encoding.ToJSON(connection))
			}

		}

		if len(program.Dependencies) > 0 {

			for d := 0; d < len(program.Dependencies); d++ {

				dependency := program.Dependencies[d]

				if dependency.IsValid() == false {
					console.Error("programs/" + program.Name + ": Invalid Dependency \"" + dependency.Name + "\"")
					console.Error(utils_encoding.ToJSON(dependency))
				}

			}

		}

		if len(program.Packages) > 0 {

			for p := 0; p < len(program.Packages); p++ {

				pkg := program.Packages[p]

				if pkg.IsValid() == false {
					console.Error("programs/" + program.Name + ": Invalid Package \"" + pkg.Name + "\"")
					console.Error(utils_encoding.ToJSON(pkg))
				}

			}

		}

	}

}

func (program *Program) IsProgram() bool {

	if program.Name != "" {

		if program.Command != "" && program.Folder != "" {

			var is_client bool = true

			for c := 0; c < len(program.Connections); c++ {

				if program.Connections[c].Type == "server" {
					is_client = false
					break
				}

			}

			return is_client

		}

	}

	return false

}

func (program *Program) IsService() bool {

	if program.Name != "" {

		if program.Command != "" && program.Folder != "" {

			var is_server bool = false

			for c := 0; c < len(program.Connections); c++ {

				if program.Connections[c].Type == "server" {
					is_server = true
					break
				}

			}

			return is_server

		}

	}

	return false

}

func (program *Program) IsValid() bool {

	if program.PID != 0 && program.Name != "" {
		return true
	}

	return false

}

func (program *Program) SetName(value string) {
	program.Name = strings.TrimSpace(value)
}

func (program *Program) SetArguments(value []string) {
	program.Arguments = value
}

func (program *Program) SetCommand(value string) {
	program.Command = strings.TrimSpace(value)
}

func (program *Program) AddConnection(value matchers.Connection) {

	if value.IsValid() {

		var found bool = false

		for c := 0; c < len(program.Connections); c++ {

			if program.Connections[c].IsIdentical(value) {
				found = true
				break
			}

		}

		if found == false {
			program.Connections = append(program.Connections, value)
		}

	}

}

func (program *Program) RemoveConnection(value matchers.Connection) {

	var index int = -1

	for c := 0; c < len(program.Connections); c++ {

		if program.Connections[c].IsIdentical(value) {
			index = c
			break
		}

	}

	if index != -1 {
		program.Connections = append(program.Connections[:index], program.Connections[index+1:]...)
	}

}

func (program *Program) SetConnections(value []matchers.Connection) {

	var filtered []matchers.Connection

	for v := 0; v < len(value); v++ {

		if value[v].IsValid() {
			filtered = append(filtered, value[v])
		}

	}

	program.Connections = filtered

}

func (program *Program) AddDependency(value matchers.Package) {

	if value.IsValid() {

		var found bool = false

		for d := 0; d < len(program.Dependencies); d++ {

			if program.Dependencies[d].IsIdentical(value) {
				found = true
				break
			}

		}

		if found == false {
			program.Dependencies = append(program.Dependencies, value)
		}

	}

}

func (program *Program) HasDependency(dependency matchers.Package) bool {

	var result bool = false

	for d := 0; d < len(program.Dependencies); d++ {

		other := program.Dependencies[d]

		if dependency.Name == other.Name &&
			dependency.Version == other.Version {
			result = true
			break
		}

	}

	return result

}

func (program *Program) RemoveDependency(value matchers.Package) {

	var index int = -1

	for d := 0; d < len(program.Dependencies); d++ {

		if program.Dependencies[d].IsIdentical(value) {
			index = d
			break
		}

	}

	if index != -1 {
		program.Dependencies = append(program.Dependencies[:index], program.Dependencies[index+1:]...)
	}

}

func (program *Program) ResolveDependencies(packages []Package) {

	if len(program.Filesystem) > 0 && len(packages) > 0 {

		for f := 0; f < len(program.Filesystem); f++ {

			file := program.Filesystem[f]

			var resolved matchers.Package

			for p := 0; p < len(packages); p++ {

				other := packages[p]

				if other.HasFilesystem(file) {

					resolved.Name = other.Name
					resolved.Version = other.Version.String()

					if other.Manager != "" {
						resolved.Manager = other.Manager.String()
					} else {
						resolved.Manager = "any"
					}

					if other.Vendor != "" {
						resolved.Vendor = other.Vendor
					} else {
						resolved.Vendor = "any"
					}

				}

				if resolved.Name != "" {
					break
				}

			}

			if resolved.Name != "" {

				if !program.HasDependency(resolved) {
					program.AddDependency(resolved)
				}

			}

		}

	}

}

func (program *Program) SetDependencies(value []matchers.Package) {

	var filtered []matchers.Package

	for v := 0; v < len(value); v++ {

		if value[v].IsValid() {
			filtered = append(filtered, value[v])
		}

	}

	program.Dependencies = filtered

}

func (program *Program) AddEnvironment(key string, val string) {

	var found bool = false

	for other_key, other_val := range program.Environment {

		if other_key == key && other_val == val {
			found = true
			break
		}

	}

	if found == false {
		program.Environment[key] = val
	}

}

func (program *Program) RemoveEnvironment(key string) {

	_, ok := program.Environment[key]

	if ok == true {
		delete(program.Environment, key)
	}

}

func (program *Program) SetEnvironment(key string, val string) {

	if len(key) > 0 && len(val) > 0 {
		program.Environment[key] = val
	}

}

func (program *Program) AddFilesystem(value string) {

	if utils_path.IsWatchedFile(value) {

		var found bool = false

		for f := 0; f < len(program.Filesystem); f++ {

			file := program.Filesystem[f]

			if file == value {
				found = true
				break
			}

		}

		if found == false {
			program.Filesystem = append(program.Filesystem, value)
		}

	}

}

func (program *Program) RemoveFilesystem(value string) {

	var index int = -1

	for f := 0; f < len(program.Filesystem); f++ {

		if program.Filesystem[f] == value {
			index = f
			break
		}

	}

	if index != -1 {
		program.Filesystem = append(program.Filesystem[:index], program.Filesystem[index+1:]...)
	}

}

func (program *Program) SetFilesystem(value []string) {

	var filtered []string

	for v := 0; v < len(value); v++ {

		file := value[v]

		if utils_path.IsWatchedFile(file) {

			var found bool = false

			for f := 0; f < len(filtered); f++ {

				if filtered[f] == file {
					found = true
					break
				}

			}

			if found == false {
				filtered = append(filtered, file)
			}

		}

	}

	program.Filesystem = filtered

}

func (program *Program) SetFolder(value string) {

	if strings.HasPrefix(value, "/") {

		if len(value) > 1 && strings.HasSuffix(value, "/") {
			value = value[0 : len(value)-1]
		}

		program.Folder = value

	}

}

func (program *Program) SetManager(value matchers.Manager) {

	if value.IsValid() {
		program.Manager = value
	}

}

func (program *Program) AddPackage(value Package) {

	if value.IsValid() {

		var found bool = false

		for p := 0; p < len(program.Packages); p++ {

			if program.Packages[p].IsIdentical(value) {
				found = true
				break
			}

		}

		if found == false {
			program.Packages = append(program.Packages, value)
		}

	}

}

func (program *Program) RemovePackage(value Package) {

	var index int = -1

	for p := 0; p < len(program.Packages); p++ {

		if program.Packages[p].IsIdentical(value) {
			index = p
			break
		}

		if index != -1 {
			program.Packages = append(program.Packages[:index], program.Packages[index+1:]...)
		}

	}

}

func (program *Program) AddPackages(packages []Package) {

	for p := 0; p < len(packages); p++ {

		if packages[p].IsValid() {
			program.Packages = append(program.Packages, packages[p])
		}

	}

}

func (program *Program) SetPackages(packages []Package) {

	var filtered []Package

	for p := 0; p < len(packages); p++ {

		if packages[p].IsValid() {
			filtered = append(filtered, packages[p])
		}

	}

	program.Packages = filtered

}

func (program *Program) SetUser(value matchers.User) {

	if value.IsValid() {
		program.User = value
	}

}
