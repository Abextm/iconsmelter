package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func SheetBuilder(sheetID int, colHeight int, removeBG bool, fallback chan *LoadedItem) (in chan *LoadedItem, out chan OutItemKP) {
	in = make(chan *LoadedItem, 10)
	out = make(chan OutItemKP, 10)
	go func() {
		sheetName := fmt.Sprintf("%v.png", sheetID)
		var BGImg *image.RGBA
		if removeBG {
			BGImg = LoadAndCrop("static/os/ico/bg" + sheetName)
		}

		fmt.Printf("%v - Loading\n", sheetID)

		//Call ProcessItem in parallel
		outs := make([]chan *LoadedItem, threads)
		for i := range outs {
			outs[i] = make(chan *LoadedItem, 3)
			go func(channel chan *LoadedItem) {
				for item := range in {
					if fallback == nil || BGImg == nil || HasBG(item.Image, BGImg) {
						ProcessItem(item, BGImg)
						channel <- item
					} else {
						fallback <- item
					}
				}
				close(channel)
			}(outs[i])
		}
		newItems := make(chan *LoadedItem, 10)
		go Mux(newItems, outs)

		//categorize them
		//map[width]map[height]Icon
		icons := map[int]map[int][]*Icon{}
		for item := range newItems {
			width := item.Image.Bounds().Dx()
			height := item.Image.Bounds().Dy()
			//make the map (if needed)
			_, ok := icons[width]
			if !ok {
				icons[width] = make(map[int][]*Icon)
			}
			_, ok = icons[width][height]
			if !ok {
				icons[width][height] = []*Icon{}
			}

			found := false
			for _, icon := range icons[width][height] {
				if CmpU8Arr(icon.Image.Pix, item.Image.Pix) { //dupelicate icon
					found = true
					icon.IDs = append(icon.IDs, item.ID)
					out <- OutItemKP{
						ID: item.ID,
						OutItem: OutItem{
							Item: icon.IDs[0],
						},
					}
					break
				}
			}
			if !found { //unique item
				icons[width][height] = append(icons[width][height], &Icon{
					Image: item.Image,
					IDs:   []string{item.ID},
				})
			}
		}
		if fallback != nil {
			close(fallback)
		}

		fmt.Printf("%v - Mapping\n", sheetID)

		//Map out the sheet
		currentPos := image.Point{}
		currentWidth := 0
		for width, iconH := range icons {
			currentWidth = max(currentWidth, width)
			for _, iconS := range iconH {
				for _, icon := range iconS {
					dbound := image.Rectangle{Min: image.ZP, Max: icon.Image.Bounds().Size()}.Add(currentPos)
					if dbound.Max.Y > colHeight {
						currentPos.Y = 0
						currentPos.X += currentWidth
						currentWidth = width
						dbound = image.Rectangle{Min: image.ZP, Max: icon.Image.Bounds().Size()}.Add(currentPos)
					}
					icon.Pos = dbound
					currentPos.Y = dbound.Max.Y
					out <- OutItemKP{
						ID: icon.IDs[0],
						OutItem: OutItem{
							X:     dbound.Min.X,
							Y:     dbound.Min.Y,
							W:     dbound.Dx(),
							H:     dbound.Dy(),
							Sheet: sheetID,
						},
					}
				}
			}
		}

		fmt.Printf("%v - Drawing\n", sheetID)

		//draw the sheet
		outImg := image.NewRGBA(image.Rect(0, 0, currentPos.X+currentWidth, colHeight))
		done := make(chan struct{})
		inImg := make(chan *Icon, 10)
		for i := 0; i < threads; i++ {
			go func() {
				for img := range inImg {
					draw.Draw(outImg, img.Pos, img.Image, image.ZP, draw.Src)
				}
				done <- struct{}{}
			}()
		}
		for _, iconH := range icons {
			for _, iconS := range iconH {
				for _, icon := range iconS {
					inImg <- icon
				}
			}
		}
		close(inImg)
		for i := 0; i < threads; i++ {
			<-done
		}

		fmt.Printf("%v - Saving\n", sheetID)

		tmpfi := "iconsmelter/uncrushed" + sheetName
		//save the sheet
		func() {
			fi, err := os.OpenFile(tmpfi, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
			if err != nil {
				panic(err)
			}
			defer fi.Close()
			err = (&png.Encoder{
				png.BestSpeed,
			}).Encode(fi, outImg)
			if err != nil {
				panic(err)
			}
		}()

		fmt.Printf("%v - Crushing\n", sheetID)

		crush(tmpfi, "static/os/ico/"+sheetName)

		fmt.Printf("%v - Done\n", sheetID)
		close(out)
	}()
	return
}

//Remove duplicate pixels
func ProcessItem(item *LoadedItem, bg *image.RGBA) {
	img := item.Image
	if bg != nil {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
				ar, ag, ab, aa := img.At(x, y).RGBA()
				br, bg, bb, ba := bg.At(x, y).RGBA()
				if inthresh(aa, ba) && inthresh(ar, br) && inthresh(ag, bg) && inthresh(ab, bb) {
					img.Set(x, y, color.RGBA{0, 0, 0, 0})
				}
			}
		}
		item.Image = crop(img, true)
	}
}

func inthresh(a, b uint32) bool {
	t := int(a) - int(b)
	return t < 6 && t > -6
}
