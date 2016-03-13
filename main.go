package main

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
)

var threads = 8
var Sheets = 20

func main() {
	var Items []*ItemListItem
	d, err := ioutil.ReadFile("oldschool/db/itemlist/names.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(d, &Items)
	if err != nil {
		panic(err)
	}

	Normal, out0 := SheetBuilder(0, 32*48, false, nil)
	Slider, out1 := SheetBuilder(1, 372, false, nil)
	outB := make([]chan OutItemKP, Sheets)
	outB[0] = out0
	outB[1] = out1
	last := Normal
	last, outB[2] = SheetBuilder(2, 32*32, true, last) //Notes get a taller image
	for i := 3; i < Sheets; i++ {
		last, outB[i] = SheetBuilder(i, 8*32, true, last)
	}
	out := make(chan OutItemKP, 10)
	go Mux(out, outB)

	go func() {
		for _, Item := range Items {
			it := &LoadedItem{
				Image: LoadAndCrop(fmt.Sprintf("iconsmelter/icons/%v.png", Item.ID)),
				ID:    fmt.Sprint(Item.ID),
			}
			if Item.Name == "Sliding piece" || Item.Name == "Sliding button" {
				Slider <- it
			} else {
				last <- it
			}
		}
		close(last)
		close(Slider)
	}()

	outMap := map[string]OutItem{}

	for v := range out {
		outMap[v.ID] = v.OutItem
	}

	d, err = json.Marshal(outMap)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("oldschool/db/itemlist/coords.json", d, 0777)
}

type LoadedItem struct {
	Image *image.RGBA
	ID    string
}

type OutItem struct {
	X, Y, W, H int    `json:",omitempty"`
	Sheet      int    `json:",omitempty"`
	Item       string `json:",omitempty"`
}

type Icon struct {
	Image *image.RGBA
	IDs   []string
	Pos   image.Rectangle
}

type OutItemKP struct {
	ID      string
	OutItem OutItem
}

type ItemListItem struct {
	ID      int
	Name    string
	Members bool
	Noted   bool
}
