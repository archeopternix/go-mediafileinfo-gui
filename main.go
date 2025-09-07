package main

import (
	"fmt"
	"log/slog"
	"math"

	fltk "github.com/archeopternix/go-fltk"
	"github.com/archeopternix/go-mediafileinfo"
	"github.com/archeopternix/gofltk-keyvalue"
)

var currFile = "example.mp4"
var kvgrid *keyvalue.KeyValueGrid

func main() {
	fltk.InitStyles()

	win := fltk.NewWindow(580, 620)
	win.SetLabel("Media File Info GUI")
	win.Resizable(win)
	win.Begin()

	// Add a button that calls open() when clicked
	openFileBtn := fltk.NewButton(10, 30, 80, 70, "Open File")
	openFileBtn.SetTooltip("Open File")
	openFileBtn.SetAlign(fltk.ALIGN_IMAGE_OVER_TEXT)
	imgFile, err := fltk.NewPngImageLoad("img/document-open.png")
	if err != nil {
		slog.Error("button open image", "image:", err)
	}
	openFileBtn.SetImage(imgFile)
	openFileBtn.SetCallback(func() {
		openFile()
	})

	// Create the KeyValueGrid (replace NewKeyValueGrid with correct import if needed)
	grid := keyvalue.NewKeyValueGrid(win, 100, 10, 460, 600)
	kvgrid = grid

	win.End()
	win.Show()
	fltk.Run()
}

// openFile prompts the user to select one or more video files, processes them,
// and adds them to the scrollable list.
func openFile() {
	// Create a new file chooser dialog for video files
	chooser := fltk.NewFileChooser(
		".",                                // Default directory
		"*.{mp4,mpeg,avi,vob,mpg,mov,m2t}", // Video file filter
		fltk.FileChooser_MULTI,             // Mode: Select multiple files
		"Select File",                      // Dialog title
	)
	chooser.Show()

	// Wait for the user to make a selection
	for chooser.Shown() {
		fltk.Wait()
	}

	// Handle case where no files are selected
	if len(chooser.Selection()) == 0 {
		slog.Info("open directory", "no files selected")
		return
	}

	list := chooser.Selection() // Process selected files

	// If no valid video files found, log and return
	if len(list) == 0 {
		slog.Info("open files", "no video files selected")
		return
	}
	// currFile = list[0]
	addMediaInfo(kvgrid, list[0])
}

func addMediaInfo(grid *keyvalue.KeyValueGrid, filename string) {
	var fps float32

	grid.ClearAll()

	info, err := mediafileinfo.GetMediaInfo(filename)
	if err != nil {
		slog.Error("Failed to get media info: %v", err)
	}

	grid.Add("File", "Name", info.Filename)
	grid.Add("File", "Size", info.FileSizeText)
	grid.Add("File", "Duration", info.DurationText)
	grid.Add("File", "Format Detail", info.FormatLongName)
	grid.Add("File", "Format", info.FormatName)

	for _, stream := range info.Streams {
		if stream.CodecParameters.CodecTypeText != "UNKNOWN" {
			stext := fmt.Sprintf("Stream %d (%s)", stream.Index, stream.CodecParameters.CodecTypeText)
			grid.Add(stext, "Codec ID", stream.CodecParameters.CodecIDText)

			// video stream?
			if (stream.CodecParameters.Width > 0) && (stream.CodecParameters.Height > 0) {
				grid.Add(stext, "Resolution", fmt.Sprintf("%d : %d", stream.CodecParameters.Width, stream.CodecParameters.Height))
				fps = (float32)(stream.AverageFrameRate.Num) / (float32)(stream.AverageFrameRate.Den)
				if stream.CodecParameters.FieldOrder > 1 {
					grid.Add(stext, "FPS", fmt.Sprintf("%.2f (interlaced)", fps))
				} else {
					grid.Add(stext, "FPS", fmt.Sprintf("%.2f (progressive)", fps))
				}
				if stream.CodecParameters.AspectRatio.Num == 0 {
					grid.Add(stext, "Aspect Ratio", fmt.Sprintf("1:%d", stream.CodecParameters.AspectRatio.Den))

				} else {
					var ar float32
					ar = float32(stream.CodecParameters.AspectRatio.Num) / float32(stream.CodecParameters.AspectRatio.Den)
					grid.Add(stext, "Aspect Ratio", fmt.Sprintf("1:%.2f", ar))

				}

			}
			grid.Add(stext, "Bitrate", formatBitsPerSecond(stream.CodecParameters.BitRate))

			//audio stream?
			if stream.CodecParameters.Channels > 0 {
				grid.Add(stext, "Channels", fmt.Sprintf("%d", stream.CodecParameters.Channels))
			}
		}
	}

	err = mediafileinfo.PrintAVContextJSON(info)
	if err != nil {
		slog.Error("Failed to print media info: %v", err)
	}
}

// FormatBitsPerSecond converts an int64 value of bytes into a human-readable string using b/s, kb/s, mb/s, or gb/s (1024 basis).
// The result is rounded to 2 decimals (e.g., 1536 -> 1.50 KB).
func formatBitsPerSecond(bits int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	b := float64(bits)
	round2 := func(val float64) float64 {
		return math.Round(val*100) / 100
	}

	switch {
	case b >= TB:
		return fmt.Sprintf("%.2f tb/s", round2(b/TB))
	case b >= GB:
		return fmt.Sprintf("%.2f gb/s", round2(b/GB))
	case b >= MB:
		return fmt.Sprintf("%.2f mb/s", round2(b/MB))
	case b >= KB:
		return fmt.Sprintf("%.2f kb/s", round2(b/KB))
	default:
		return fmt.Sprintf("%d b/s", bits)
	}
}
