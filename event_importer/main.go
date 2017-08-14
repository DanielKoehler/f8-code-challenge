package main

import (
	"fmt"
	"net/http"
	"os"
	"encoding/json"
	"strconv"
	"io/ioutil"
	"strings"
	"bytes"
    "github.com/santhosh-tekuri/jsonschema"
)

const feedsAPIBase = "http://localhost:8000"

type Option struct {
    ID string `json:"id"`
    Name string `json:"name"`
    Numerator  int `json:"num"`
    Denominator  int `json:"den"`
}

type Market struct {
    ID string `json:"id"`
    Type string `json:"type"`
    Options []Option `json:"options"`
}

func (m *Market) AddOption(o Option) []Option {
    m.Options = append(m.Options, o)
    return m.Options
}

type Event struct {
    ID string `json:"id"`
    Name string `json:"name"`
    Time string `json:"time"` // timestamp (RFC 3339) 
    Markets []Market `json:"markets"`
}

func (e *Event) AddMarket(m Market) []Market {
    e.Markets = append(e.Markets, m)
    return e.Markets
}

type MarketResult struct {
	Error bool
    Market Market
}

type EventResult struct {
	Error bool
    Event Event
}


// Operator specific function - would be nice to simplify some of the non implementation specific logic.
func GetMarket(marketId string, eventMarketsChannel chan MarketResult) {
	
    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Error in Market{%s}: %s - giving up on market.\n", marketId, r)
            eventMarketsChannel <- MarketResult{Error: true}
        }
    }()

    resp, err := http.Get(feedsAPIBase + "/football/markets/" + marketId)


	if err != nil {
        panic(err)
    }

    // Ignore non-200 OK responses.
    if resp.StatusCode != 200 {
    	panic(fmt.Sprintf("Feeds server status %d", resp.StatusCode))
    }
  
    bytes, _ := ioutil.ReadAll(resp.Body)    

    var genericMarket interface{}
    json.Unmarshal(bytes, &genericMarket)

    market := genericMarket.(map[string]interface{})

    newMarket := Market{
    	ID: PluckString(market, "id"),
    	Type: PluckString(market, "type"),
        Options: []Option{}}

    // Parse Options List.
    optionsArray := EnsureArray(market["options"])

    for _, optionInterface := range optionsArray {

		option := optionInterface.(map[string]interface {})

		newOption := Option{
	    	ID: PluckString(option, "id"),
	    	Name: PluckString(option, "name"),
	    }

	    // Split num / den - assign to struct.
	    odds := strings.Split(PluckString(option, "odds"), "/")
	    newOption.Numerator, _ = strconv.Atoi(odds[0])
	    newOption.Denominator, _ = strconv.Atoi(odds[1])

	    newMarket.AddOption(newOption)
	
    }

	eventMarketsChannel <- MarketResult{Error: false, Market: newMarket}
}


// Again very operator specific - would be nice to repeat stuff like unmarshalling to a generic interface 
func GetEvent(eventId string, eventsChannel chan EventResult) {
	
	defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Error in Event{id:%s}: %s - giving up on event.\n", eventId, r)
        	eventsChannel <- EventResult{Error: true}
        }
    }()

	resp, err := http.Get(feedsAPIBase + "/football/events/" + eventId)
	if err != nil {
        panic(err)
    }

    // Ignore non-200 OK responses.
    if resp.StatusCode != 200 {
    	panic(fmt.Sprintf("Feeds server status %d", resp.StatusCode))
    }

    bytes, _ := ioutil.ReadAll(resp.Body)

    var genericEvent interface{}
    json.Unmarshal(bytes, &genericEvent)

    event := genericEvent.(map[string]interface{})
  
    newEvent := Event{
    	ID: PluckString(event, "id"),
    	Time: PluckRFC3339(event, "time"), 
        // These kind of plucking / casting helpers don't seem super nice. 
        // I'd also probably like to register CustomTimeParseFormats here, 
        // rather than setting up operator specific formats inside what should be generic helpers.
    	Name: PluckString(event, "name"),
        Markets: []Market{}}

    eventMarketsChannel := make(chan MarketResult)
    
    // Lets Go get the markets.
    marketArray := EnsureArray(event["markets"])

    for _, marketIdInterface := range marketArray {

		// This will currently panic if this isn't an int array.
		marketId := EnsureString(marketIdInterface)
	
		go GetMarket( marketId , eventMarketsChannel )
	
    }

    // Receive market results
	for range marketArray {
	    
	    marketResult := <- eventMarketsChannel

	    if marketResult.Error == false {
	    	newEvent.AddMarket(marketResult.Market)
	    }

	}

	eventsChannel <- EventResult{Error: false, Event: newEvent}
}

// Would have liked to have moved the event type definitions, schema, and validation out to another package.
func IsValidEvent(eventJSON []byte) (bool) {

    schema, err := jsonschema.Compile("./event_importer/eventSchema.json")
    if err != nil {
        panic(err)
    }

    if err = schema.Validate(bytes.NewReader(eventJSON)); err != nil {
        return false
    }

    return true

}


func PostToStore(event Event) {

	eventData, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}

    if !IsValidEvent(eventData) {
        panic( "Invalid event" )
    }

	endpoint := fmt.Sprintf("http://%s/events", StoreAdress)

    req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(eventData))
    req.Header.Set("User-Agent", "SomeKindaEventsScraper 0.0.1")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    fmt.Println("response Status:", resp.Status)
    fmt.Println("response Headers:", resp.Header)
    body, _ := ioutil.ReadAll(resp.Body)
    fmt.Println("response Body:", string(body))


}
func GetEventsAndPostToStore() {

	
	resp, err := http.Get(feedsAPIBase + "/football/events")
	if err != nil {
        panic(err)
    }

    var eventIds []int
    err = json.NewDecoder( resp.Body ).Decode( &eventIds )
    if err != nil {
        panic(err)
    }

    // Setup a channel to receive the event payloads
    eventsChannel := make(chan EventResult)

    for _, eventId := range eventIds {

	    eventId := strconv.Itoa(eventId)
	    go GetEvent(eventId, eventsChannel)
	
	}

    successfulFetches := 0.0

	// Recieve events - build stats or whatever else, then post to store.
    // Obviously relying on a result being dumped onto a channel here isn't super nice. 
    // Posting also becomes syncronous in this implementation - should probably be sorted in some future iteration.
	for range eventIds {
	    
	    eventResult := <- eventsChannel

	    if eventResult.Error == false {

            successfulFetches++
	    	PostToStore(eventResult.Event)

	    }

	}
	   
    if len(eventIds) > 0 {

        fmt.Printf("\nSuccess ratio: %f%%\n", successfulFetches / float64(len(eventIds)) * 100)
    
    }


}

var StoreAdress = "localhost:8001"

func main() {
	

	if remote := os.Getenv("STORE_ADDR"); remote != "" {
	    StoreAdress = remote
	}

	GetEventsAndPostToStore()

}













