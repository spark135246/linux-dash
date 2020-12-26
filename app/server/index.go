package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/rakyll/statik/fs"
	"io/ioutil"
	_ "linux-dash/statik"
	"log"
	"net/http"
	"os"
	"os/exec"
)

var (
	listenAddress = flag.String("listen", "0.0.0.0:8080", "Where the server listens for connections. [interface]:port")
)

func init() {
	flag.Parse()
}

func main() {

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	// 写入文件
	r, err := statikFS.Open("/server/linux_json_api.sh")
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile(
		"linux_json_api.sh",
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	if err != nil {
		panic(err)
	}
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	_, err = file.Write(contents)
	if err != nil {
		panic(err)
	}
	_ = file.Close()
	_ = r.Close()

	http.Handle("/", http.FileServer(statikFS))
	http.HandleFunc("/server/", func(w http.ResponseWriter, r *http.Request) {
		module := r.URL.Query().Get("module")
		if module == "" {
			http.Error(w, "No module specified, or requested module doesn't exist.", 406)
			return
		}

		// Execute the command
		cmd := exec.Command("./linux_json_api.sh", module)
		var output bytes.Buffer
		cmd.Stdout = &output
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error executing '%s': %s\n\tScript output: %s\n", module, err.Error(), output.String())
			http.Error(w, "Unable to execute module.", http.StatusInternalServerError)
			return
		}

		w.Write(output.Bytes())
	})

	fmt.Println("Starting http server at:", *listenAddress)
	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		fmt.Println("Error starting http server:", err)
		os.Exit(1)
	}
}
