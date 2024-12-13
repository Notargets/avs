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
	chart.Screen.LoadFont("assets/fonts/Noto-Sans/static/NotoSans-Regular.ttf", 32)
	chart.Screen.AddString(screen.NEW, "Hello World", -0, 0, [3]float32{1, 1, 1}, 0.5)
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
