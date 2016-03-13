package main

import "os/exec"

func crush(tmpfi, outfi string) {
	//pngcrush -rem alla -nofilecheck -reduce -m 7 "iconsmelter/uncrushed"+sheetName "static/os/ico/"+sheetName
	pngcrush, err := exec.LookPath("pngcrush")
	if err != nil {
		pngcrush = "iconsmelter/pngcrush.exe"
	}
	cmd := exec.Command(pngcrush, "-rem", "alla", "-nofilecheck", "-m", "7", tmpfi, outfi)
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	_ = out
	//fmt.Println((string)(out))
}
