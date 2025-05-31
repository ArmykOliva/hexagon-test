package main

import "battlemap/structs"
import "battlemap/server"
import "encoding/json"
import "fmt"
import "io/fs"
import "log"
import "net/http"
import "os"

func main() {

	cache := server.NewCache()
	fsys := os.DirFS("public")
	fsrv := http.FileServer(http.FS(fsys))

	data_fsys := os.DirFS("../data")

	files1, err1 := fs.ReadDir(data_fsys, "systems")

	if err1 == nil {

		for _, file := range files1 {

			buffer, err11 := fs.ReadFile(data_fsys, "systems/" + file.Name())

			if err11 == nil {

				var system structs.System

				err12 := json.Unmarshal(buffer, &system)

				if err12 == nil {
					cache.SetSystem(system)
				}

			}

		}

	}

	files2, err2 := fs.ReadDir(data_fsys, "vulnerabilities")

	if err2 == nil {

		for _, file := range files2 {

			buffer, err21 := fs.ReadFile(data_fsys, "vulnerabilities/" + file.Name())

			if err21 == nil {

				var vulnerability structs.Vulnerability

				err22 := json.Unmarshal(buffer, &vulnerability)

				if err22 == nil {
					cache.SetVulnerability(vulnerability)
				}

			}

		}

	}

	http.Handle("/", fsrv)

	http.HandleFunc("/api/systems", func(response http.ResponseWriter, request *http.Request) {

		if request.Method == http.MethodGet {

			payload, err := json.MarshalIndent(cache.Systems, "", "\t")

			if err == nil {

				fmt.Println("> GET /api/systems: ok")

				response.Header().Set("Content-Type", "application/json")
				response.WriteHeader(http.StatusOK)
				response.Write(payload)

			} else {

				fmt.Println("> GET /api/systems: error")

				response.Header().Set("Content-Type", "application/json")
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte("[]"))

			}

		} else {

			response.Header().Set("Content-Type", "application/json")
			response.WriteHeader(http.StatusMethodNotAllowed)
			response.Write([]byte("[]"))

		}

	})

	fmt.Println("Listening on http://localhost:3000")

	err := http.ListenAndServe(":3000", nil)

	if err != nil {
		log.Fatal(err)
	}

}
