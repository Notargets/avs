package main

import (
	"os"
	"path/filepath"

	"github.com/notargets/avs/screen"

	"github.com/notargets/avs/chart2d"
)

func main() {
	chart := chart2d.NewChart2D(-10, 10, -20, 20, 1000, 1000)
	chart.AddAxis(chart2d.Color{1., 1., 1.})
	chart.Screen.LoadFont("assets/fonts/Noto-Sans/static/NotoSans-Regular.ttf", 64)
	//chart.Screen.LoadFont("assets/fonts/snob.org/sans-serif.fnt", 64)
	chart.Screen.AddString(screen.NEW, "Hello World - There once was a man in natucket. His head could fill a big bucket. He tried one on, then banged into his lawn. His friends were told to buck it.", 0.5, 0.5, [3]float32{1, 1, 1}, 1.0)
	select {}
}

func getFontConfigPath() (path string) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic("Failed to find user config directory: " + err.Error())
	}
	path = filepath.Join(configDir, "avs", "fonts")
	err = os.MkdirAll(path, 0755)
	if err != nil {
		panic("Failed to create font config directory: " + err.Error())
	}
	return path
}
