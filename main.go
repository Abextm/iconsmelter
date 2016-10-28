package main

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"os"
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
	BaseImage(2, 4208)
	BaseImage(4, 229)
	BaseImage(5, 11732)
	BaseImage(6, 13167)
	BaseImage(7, 1925)
	BaseImage(8, 8007)
	BaseImage(9, 11477)
	BaseImage(10, 5376)
	BaseImage(11, 1931)
	BaseImage(13, 1978)
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
				Image: LoadAndCrop(ItemPath(Item.ID)),
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

func ItemPath(id int) string {
	return fmt.Sprintf("iconsmelter/icons/%v.png", id)
}

func BaseImage(bgi, id int) {
	err := cp(fmt.Sprintf("static/os/ico/bg%v.png", bgi), ItemPath(id))
	if err != nil {
		panic(err)
	}
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
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Members bool   `json:"members"`
	Noted   bool   `json:"noted"`
}

func cp(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()
	d, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}
