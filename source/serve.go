package main

// import "battlemap/structs"
import "battlemap/server"
import "encoding/json"
import "fmt"
import "log"
import "net/http"
import "os"

func main() {

	cache := server.NewCache()
	fsys := os.DirFS("public")
	fsrv := http.FileServer(http.FS(fsys))

	// data_fsys := os.DirFS("../data")
	// system_entries := data_fsys.Readdir("/systems")

	// fmt.Println(system_entries)

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
