package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type SelectorType int

const (
	Label SelectorType = iota
	ID
	GroupID
	Group
	LocationID
	Location
	SceneID
	All
)

type Light struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	GroupID      string     `json:"group_id"`
	GroupName    string     `json:"group_name"`
	LocationID   string     `json:"location_id"`
	LocationName string     `json:"location_name"`
	MaxKelvin    int        `json:"max_kelvin"`
	MinKelvin    int        `json:"min_kelvin"`
	State        LightState `json:"light_state"`
}

type Colour struct {
	// The hue of the colour, from 0 to 360.
	Hue float64 `json:"hue,omitempty"`
	// The saturation of the colour, from 0.0 to 1.0.
	Saturation float64 `json:"saturation,omitempty"`
	// The brightness of the colour, from 0.0 to 1.0.
	Brightness float64 `json:"brightness,omitempty"`
	// The colour temperature in Kelvin.
	Kelvin int `json:"kelvin,omitempty"`
}

type LightState struct {
	// The power state you want to set on the selector. on or off
	Power string `json:"power,omitempty"`
	// The colour to set the light to.
	Colour Colour `json:"color"`
	// The brightness level from 0.0 to 1.0. Overrides any brightness set in color (if any).
	Brightness float64 `json:"brightness,omitempty"`
	// How long in seconds you want the power action to take. Range: 0.0 – 3155760000.0 (100 years)
	Duration float64 `json:"duration,omitempty"`
	// The maximum brightness of the infrared channel from 0.0 to 1.0.
	Infrared float64 `json:"infrared,omitempty"`
	// Execute the query fast, without initial state checks and wait for no results.
	Fast bool `json:"fast,omitempty"`
}

func toLight(m map[string]any) Light {
	product, _ := m["product"].(map[string]any)
	caps, _ := product["capabilities"].(map[string]any)

	id, _ := m["id"].(string)
	label, _ := m["label"].(string)
	power, _ := m["power"].(string)
	brightness, _ := m["brightness"].(float64)
	location, _ := m["location"].(map[string]any)
	locationID, _ := location["id"].(string)
	locationName, _ := location["name"].(string)
	maxKelvin, _ := caps["max_kelvin"].(float64)
	minKelvin, _ := caps["min_kelvin"].(float64)

	group, _ := m["group"].(map[string]any)
	groupID, _ := group["id"].(string)
	groupName, _ := group["name"].(string)

	color, _ := m["color"].(map[string]any)
	hue, _ := color["hue"].(float64)
	saturation, _ := color["saturation"].(float64)
	kelvin, _ := color["kelvin"].(float64)

	return Light{
		ID:           id,
		Name:         label,
		GroupID:      groupID,
		GroupName:    groupName,
		LocationID:   locationID,
		LocationName: locationName,
		MaxKelvin:    int(maxKelvin),
		MinKelvin:    int(minKelvin),
		State: LightState{
			Power:      power,
			Brightness: brightness,
			Colour: Colour{
				Hue:        hue,
				Saturation: saturation,
				Kelvin:     int(kelvin),
			},
		},
	}
}

func loadLights() ([]Light, error) {
	url := "https://api.lifx.com/v1/lights"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("LIFX_TOKEN"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", resp.Status, body)
	}

	var maps []map[string]any
	if err := json.Unmarshal(body, &maps); err != nil {
		return nil, err
	}

	lights := make([]Light, 0, len(maps))
	for _, m := range maps {
		lights = append(lights, toLight(m))
	}
	return lights, nil
}

func setLight(selector string, state LightState) (*http.Response, error) {
	url := "https://api.lifx.com/v1/lights/" + selector + "/state"

	body, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("LIFX_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	return http.DefaultClient.Do(req)
}

func makeSelector(selectorType SelectorType, light Light) string {
	switch selectorType {
	case All:
		return "all"
	case GroupID:
		return "group_id:" + light.GroupID
	case Group:
		return "group:" + light.GroupName
	case LocationID:
		return "location_id:" + light.LocationID
	case Location:
		return "location:" + light.LocationName
	case ID:
		return "id:" + light.ID
	case SceneID:
		// TODO: Implement scene ID selector
		return "scene_id:"
	default:
		return ""
	}
}

func main() {
	lights, err := loadLights()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(len(lights), "lights")
	fmt.Printf("%+v\n", lights[0])

	resp, err := setLight(makeSelector(GroupID, lights[0]), LightState{
		Power: "on",
		Colour: Colour{
			Hue:        100,
			Saturation: 1.0,
			Brightness: 1.0,
			Kelvin:     5000,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}
