package main

import (
    "fmt"
    "log"
    "net/http"
    "bytes"
    "io/ioutil"
    "github.com/gorilla/mux"
    "github.com/santhosh-tekuri/jsonschema"
)

func EventsPostHandler(w http.ResponseWriter, r *http.Request) {

    // Would probably be much nicer to have a validation package - could reuse in the importer
    schema, err := jsonschema.Compile("./mock_store/eventSchema.json")
    if err != nil {
        panic(err)
    }

    body, _ := ioutil.ReadAll(r.Body)
    fmt.Printf("%s\n", body)

    if err = schema.Validate(bytes.NewReader(body)); err != nil {

        w.WriteHeader(http.StatusBadRequest)

        err := fmt.Sprintf("%s\n", err)
        w.Write([]byte(err))

    } else {

        /*
            Aggregate + pipe over to some persistent store here..
         */

        message := "Danke für die Veranstaltungen..."
        w.Write([]byte(message))

    }
}


func main() {

    r := mux.NewRouter()
    r.HandleFunc("/events", EventsPostHandler).Methods("POST")

    log.Println("Mock store API running on :8001")
    if err := http.ListenAndServe(":8001", r); err != nil {
        panic(err)
    }

}