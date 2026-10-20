package main

import (
	"fmt"
	"log"
)

func main() {
	lights, err := LoadLights()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(len(lights), "lights")
	fmt.Printf("%+v\n", lights[0])

	upperRight := SelectorState{
		Selector: MakeSelector(Label, "Right Lamp Upper"),
		LightState: LightState{
			Power:    "on",
			Duration: 0.5,
			Fast:     true,
			Colour: Colour{
				Hue:        60,
				Saturation: 0.5,
				Brightness: 1.0,
				Kelvin:     3000,
			},
		},
	}
	upperLeft := SelectorState{
		Selector: MakeSelector(Label, "Right Lamp Lower"),
		LightState: LightState{
			Power:    "on",
			Duration: 10,
			Fast:     true,
			Colour: Colour{
				Hue:        90,
				Saturation: 0.5,
				Brightness: 1.0,
				Kelvin:     3000,
			},
		},
	}

	err = SetMultipleLights(upperRight, upperLeft)
	if err != nil {
		log.Fatal(err)
	}
}
