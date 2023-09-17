package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var tmpStagingDir string = "/tmp/scan-to-mayan"
var mayanWatchDir string
var found bool
var pathList = []string{}
var boundList binding.StringList

func init() {
	os.MkdirAll(tmpStagingDir, 0755)
	mayanWatchDir, found = os.LookupEnv("MAYAN_WATCH_DIR")

	if !found {
		log.Fatal("Environment variable MAYAN_WATCH_DIR not set")
	}

	matches, err := filepath.Glob(filepath.Join(tmpStagingDir, "*"))
	if err != nil {
		log.Fatalf("Failed to glob: %v", err)
	} else {
		log.Printf("Found existing files in tmpStagingDir=%v so adding those to pathList\n", tmpStagingDir)
	}
	pathList = matches

	boundList = binding.BindStringList(&pathList)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Scanner App")
	cmdOutput := binding.NewString()
	cmdOutputLabel := widget.NewLabelWithData(cmdOutput)

	list := widget.NewListWithData(boundList,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})

	var submit *widget.Button

	submit = widget.NewButton("Submit", func() {
		println("submit")
		println("Out:", combineScans())
		if len(pathList) == 0 { // all docs were correctly combined and pathList was emptied
			submit.Disable()
		}
	})

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()
	var add *widget.Button
	add = widget.NewButton("Add Page", func() {
		add.Disable()
		submit.Disable()
		fname := fmt.Sprintf("filescan-%d.tiff", len(pathList))
		progressBar.Show()
		val := scanToFile(fname)
		progressBar.Hide()
		cmdOutput.Set(val)
		boundList.Append(fname)
		add.Enable()
		submit.Enable()
	})

	myWindow.SetContent(container.NewBorder(cmdOutputLabel, container.NewVBox(progressBar, container.NewHBox(layout.NewSpacer(), submit, add, layout.NewSpacer())), nil, nil, list))
	myWindow.ShowAndRun()
}

func combineScans() string {
	// TODO: this should be the mayan staging dir to output to
	args := append(pathList, filepath.Join(tmpStagingDir, "output.pdf"))
	cmd := exec.Command("/usr/bin/convert", args...)
	println(cmd.String())

	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}
	return string(out)
}

func scanToFile(fname string) string {
	cmdPath := "/usr/bin/scanimage"
	outputPath := filepath.Join(tmpStagingDir, fname)
	log.Printf("Running scan command: %s -o %s", cmdPath, outputPath)

	cmd := exec.Command(cmdPath, "-o", outputPath)
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("Command failed: %v", err)
	}
	return string(out)
}
