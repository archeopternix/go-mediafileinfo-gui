// Media File Info GUI
//
// This program provides a graphical user interface (GUI) for viewing metadata
// of media files such as MP4, AVI, MOV, and more. Users can select one
// video file, and the application will display detailed information including
// file size, duration, format, codec details, resolution, frame rate, aspect ratio,
// bitrate, and audio channels. The GUI is built using go-fltk, and metadata is displayed
// in a table format using gofltk-keyvalue.
//
// Author: archeopternix
// Repository: https://github.com/archeopternix/go-mediafileinfo-gui

package main

import (
	"fmt"
	"log/slog"
	"math"

	fltk "github.com/archeopternix/go-fltk"
	"github.com/archeopternix/go-mediafileinfo"
	"github.com/archeopternix/gofltk-keyvalue"
)

// currFile holds the currently selected filename (default: example.mp4)
var (
	// kvgrid is the global reference to the KeyValueGrid for displaying metadata
	kvgrid *keyvalue.KeyValueGrid
)

// main is the entry point of the application
func main() {
	fltk.InitStyles() // Initialize custom FLTK styles

	// Create main window
	win := fltk.NewWindow(580, 620)
	win.SetLabel("Media File Info GUI")
	win.Resizable(win)
	win.Begin()

	setupOpenFileButton(win) // Add "Open File" button
	setupKeyValueGrid(win)   // Add KeyValueGrid for metadata display

	win.End()
	win.Show()
	fltk.Run() // Start the FLTK event loop
}

// setupOpenFileButton creates and configures the "Open File" button
func setupOpenFileButton(win *fltk.Window) {
	openFileBtn := fltk.NewButton(10, 30, 80, 70, "Open File")
	openFileBtn.SetTooltip("Open File")
	openFileBtn.SetAlign(fltk.ALIGN_IMAGE_OVER_TEXT)
	imgFile, err := fltk.NewPngImageLoad("img/document-open.png")
	if err != nil {
		slog.Error("button open image", "image:", err)
	}
	openFileBtn.SetImage(imgFile)
	openFileBtn.SetCallback(openFile) // Assign openFile callback
}

// setupKeyValueGrid creates the KeyValueGrid widget
func setupKeyValueGrid(win *fltk.Window) {
	grid := keyvalue.NewKeyValueGrid(win, 100, 10, 460, 600)
	kvgrid = grid
}

// openFile prompts the user to select one or more video files and loads metadata for the first selection
func openFile() {
	chooser := fltk.NewFileChooser(
		".",                                // Default directory
		"*.{mp4,mpeg,avi,vob,mpg,mov,m2t}", // Video file filter
		fltk.FileChooser_MULTI,             // Allow multiple selection
		"Select File",                      // Dialog title
	)
	chooser.Show()

	// Wait for user selection
	for chooser.Shown() {
		fltk.Wait()
	}

	list := chooser.Selection()
	if len(list) == 0 {
		slog.Info("open files", "no video files selected")
		return
	}

	addMediaInfo(kvgrid, list[0]) // Load metadata for the first file
}

// addMediaInfo gets media info for the file and displays it in the grid
func addMediaInfo(grid *keyvalue.KeyValueGrid, filename string) {
	grid.ClearAll() // Clear previous entries

	info, err := mediafileinfo.GetMediaInfo(filename)
	if err != nil {
		slog.Error("Failed to get media info", "err", err)
		return
	}

	// Add general file info
	grid.Add("File", "Name", info.Filename)
	grid.Add("File", "Size", info.FileSizeText)
	grid.Add("File", "Duration", info.DurationText)
	grid.Add("File", "Format Detail", info.FormatLongName)
	grid.Add("File", "Format", info.FormatName)

	// Add stream-specific info
	for _, stream := range info.Streams {
		if stream.CodecParameters.CodecTypeText == "UNKNOWN" {
			continue // Skip unknown streams
		}

		stext := fmt.Sprintf("Stream %d (%s)", stream.Index, stream.CodecParameters.CodecTypeText)
		grid.Add(stext, "Codec ID", stream.CodecParameters.CodecIDText)

		// If video stream, show resolution, FPS, aspect ratio
		if stream.CodecParameters.Width > 0 && stream.CodecParameters.Height > 0 {
			grid.Add(stext, "Resolution", fmt.Sprintf("%d : %d", stream.CodecParameters.Width, stream.CodecParameters.Height))

			fps := calcFPS(stream.AverageFrameRate.Num, stream.AverageFrameRate.Den)
			fpsType := "progressive"
			if stream.CodecParameters.FieldOrder > 1 {
				fpsType = "interlaced"
			}
			grid.Add(stext, "FPS", fmt.Sprintf("%.2f (%s)", fps, fpsType))

			grid.Add(stext, "Aspect Ratio", formatAspectRatio(stream.CodecParameters.AspectRatio.Num, stream.CodecParameters.AspectRatio.Den))
		}

		// Bitrate for all streams
		grid.Add(stext, "Bitrate", formatBitsPerSecond(stream.CodecParameters.BitRate))

		// If audio stream, show channel count
		if stream.CodecParameters.Channels > 0 {
			grid.Add(stext, "Channels", fmt.Sprintf("%d", stream.CodecParameters.Channels))
		}
	}

	// Print AV context info to the log (for debugging)
	if err := mediafileinfo.PrintAVContextJSON(info); err != nil {
		slog.Error("Failed to print media info", "err", err)
	}
}

// calcFPS calculates frames-per-second from numerator and denominator
func calcFPS(num, den int) float32 {
	if den == 0 {
		return 0
	}
	return float32(num) / float32(den)
}

// formatAspectRatio returns human-readable aspect ratio string
func formatAspectRatio(num, den int) string {
	if den == 0 {
		return "N/A"
	}
	if num == 0 {
		return fmt.Sprintf("1:%d", den)
	}
	return fmt.Sprintf("1:%.2f", float32(num)/float32(den))
}

// formatBitsPerSecond converts bits per second into human-readable units
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
