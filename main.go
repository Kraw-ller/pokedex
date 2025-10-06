package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"pokedex/internal/pokecache"
	"strings"
	"time"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

type config struct {
	Next     string
	Previous string
	arg1     string
}

type LocationAreas struct {
	Count    int     `json:"count"`
	Next     string  `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type LocationAreaData struct {
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	GameIndex int `json:"game_index"`
	ID        int `json:"id"`
	Location  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Name  string `json:"name"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int   `json:"chance"`
				ConditionValues []any `json:"condition_values"`
				MaxLevel        int   `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

var commands map[string]cliCommand
var cache pokecache.Cache

func init() {

	commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays the names of 20 location areas in the Pokemon world",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "It's similar to the map command, however, instead of displaying the next 20 locations, it displays the previous 20 locations",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Lists the Pokemons in a location area",
			callback:    commandExplore,
		},
	}

	// create new cache
	cache = *pokecache.NewCache(5 * time.Second)
}

func cleanInput(text string) []string {
	text = strings.TrimSpace(text) // remove leading/trailing spaces
	text = strings.ToLower(text)   // make lowercase
	words := strings.Fields(text)  // split by any whitespace
	return words
}

func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config) error {
	fmt.Print("Welcome to the Pokedex!\nUsage:\n\n")
	for _, cmd := range commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func commandMap(cfg *config) error {
	url := ""
	if cfg.Next == "" {
		url = "https://pokeapi.co/api/v2/location-area/"
	} else {
		url = cfg.Next
	}

	// check if we have the data in cache first
	entry, exists := cache.Get(url)
	if !exists {
		// make a get request
		res, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
			return err
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal(err)
			return err
		}

		// body is ready
		// add the data to the cache
		cache.Add(url, body)

		// unmarshal the json into a go struct
		locationarea := LocationAreas{}
		err = json.Unmarshal(body, &locationarea)
		if err != nil {
			fmt.Println(err)
			return err
		}

		cfg.Next = locationarea.Next
		if locationarea.Previous != nil {
			cfg.Previous = *locationarea.Previous
		} else {
			cfg.Previous = ""
		}

		for _, loc := range locationarea.Results {
			fmt.Println(loc.Name)
		}
		return nil
	} else {
		// the data is already in cache
		fmt.Println("getting data from cache")
		// unmarshal the json into a go struct
		locationarea := LocationAreas{}
		err := json.Unmarshal(entry, &locationarea)
		if err != nil {
			fmt.Println(err)
			return err
		}

		cfg.Next = locationarea.Next
		if locationarea.Previous != nil {
			cfg.Previous = *locationarea.Previous
		} else {
			cfg.Previous = ""
		}

		for _, loc := range locationarea.Results {
			fmt.Println(loc.Name)
		}
		return nil
	}

}

func commandMapb(cfg *config) error {
	if cfg.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	}

	// check if we have the data in cache first
	entry, exists := cache.Get(cfg.Previous)
	if !exists {
		fmt.Println("getting data from the API")
		res, err := http.Get(cfg.Previous)
		if err != nil {
			log.Fatal(err)
			return err
		}

		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal(err)
			return err
		}

		cache.Add(cfg.Previous, body)

		locationarea := LocationAreas{}
		err = json.Unmarshal(body, &locationarea)
		if err != nil {
			fmt.Println(err)
			return err
		}

		cfg.Next = locationarea.Next

		if locationarea.Previous != nil {
			cfg.Previous = *locationarea.Previous
		} else {
			cfg.Previous = ""
		}

		for _, loc := range locationarea.Results {
			fmt.Println(loc.Name)
		}
		return nil
	} else {
		// the data is already in cache
		fmt.Println("getting data from cache")
		// unmarshal the json into a go struct
		locationarea := LocationAreas{}
		err := json.Unmarshal(entry, &locationarea)
		if err != nil {
			fmt.Println(err)
			return err
		}

		cfg.Next = locationarea.Next
		if locationarea.Previous != nil {
			cfg.Previous = *locationarea.Previous
		} else {
			cfg.Previous = ""
		}

		for _, loc := range locationarea.Results {
			fmt.Println(loc.Name)
		}
		return nil
	}
}

func commandExplore(cfg *config) error {
	if cfg.arg1 == "" {
		fmt.Println("Missing argument for command explore")
		return nil
	}
	url := "https://pokeapi.co/api/v2/location-area/" + cfg.arg1

	// check if we have the data in cache first
	entry, exists := cache.Get(url)
	if !exists {
		fmt.Println("getting data from the API")
		res, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
			return err
		}

		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal(err)
			return err
		}

		// add data to the cache
		cache.Add(url, body)

		locationareadata := LocationAreaData{}
		err = json.Unmarshal(body, &locationareadata)
		if err != nil {
			fmt.Println(err)
			return err
		}

		// get the names of the pokemons
		for _, encounter := range locationareadata.PokemonEncounters {
			fmt.Println(encounter.Pokemon.Name)
		}

	} else {
		// use cache
		fmt.Println("getting data from cache")
		locationareadata := LocationAreaData{}
		err := json.Unmarshal(entry, &locationareadata)
		if err != nil {
			fmt.Println(err)
			return err
		}

		// get the names of the pokemons
		for _, encounter := range locationareadata.PokemonEncounters {
			fmt.Println(encounter.Pokemon.Name)
		}
	}

	return nil
}

func main() {

	cfg := &config{} // shared config instance

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Pokedex > ")
	for scanner.Scan() {
		user_input := scanner.Text()
		words := cleanInput(user_input)
		// fmt.Println("Your command was:", words[0])
		command := words[0]
		if len(words) > 1 {
			cfg.arg1 = words[1]
		} else {
			cfg.arg1 = ""
		}

		cmd, exists := commands[command]
		if !exists {
			fmt.Println("Unknown command:", command)
		} else {
			err := cmd.callback(cfg)
			if err != nil {
				fmt.Println("Error:", err)
			}
		}

		fmt.Print("Pokedex > ")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
