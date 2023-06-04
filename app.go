package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	avatar "github.com/holys/initials-avatar"
	"github.com/mitchellh/go-homedir"
	"github.com/o1egl/govatar"
	"github.com/skip2/go-qrcode"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

type Placeholder struct {
	ID          int     `json:"id"`
	StartX      float64 `json:"startX"`
	StartY      float64 `json:"startY"`
	W           float64 `json:"w"`
	H           float64 `json:"h"`
	CsvKey      string  `json:"csv_key"`
	Color       string  `json:"color"`
	Font        string  `json:"font"`
	FontContent string  `json:"fontContent"`
	TextAlign   string  `json:"textAlign"`
	FontSize    float64 `json:"fontSize"`
	Transform   string  `json:"transform"`
}

func (a *App) Proceed(b64image string, placehldr string, csvData string) string {
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64image))
	m, _, err := image.Decode(reader)
	if err != nil {
		log.Println(err)
		return err.Error()
	}

	// Load placeholders
	var placeholders []Placeholder
	json.Unmarshal([]byte(placehldr), &placeholders)

	// Load CSV
	r := csv.NewReader(bytes.NewReader([]byte(csvData)))
	headers, _ := r.Read()
	rowIndex := map[string]int{}
	for i, h := range headers {
		rowIndex[h] = i
	}

	// Save each image
	home, err := homedir.Dir()
	if err != nil {
		log.Println(err)
		return err.Error()
	}
	savePath := home + fmt.Sprintf("%sDocuments%sBatch Image Generator", string(os.PathSeparator), string(os.PathSeparator))
	err = os.MkdirAll(savePath, os.ModePerm)
	if err != nil {
		log.Println(err)
		return err.Error()
	}

	// counter for naming output images
	counter := 0

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}

		// Create a copy of the image for each row
		imgCopy := image.NewRGBA(m.Bounds())
		draw.Draw(imgCopy, imgCopy.Bounds(), m, image.Point{0, 0}, draw.Src)

		dc := gg.NewContextForImage(imgCopy)

		// Draw each placeholder
		for _, ph := range placeholders {
			if ph.CsvKey != "" {
				dc.Push()

				// Load font
				data, err := base64.StdEncoding.DecodeString(ph.FontContent)
				if err == nil {
					if font, err := truetype.Parse(data); err == nil {
						face := truetype.NewFace(font, &truetype.Options{Size: ph.FontSize})
						dc.SetFontFace(face)
					} else {
						log.Println(err)
					}
				} else {
					fmt.Println(err)
				}

				// Set color
				colorHex := strings.TrimPrefix(ph.Color, "#")
				colorRgb, _ := hex.DecodeString(colorHex)
				r := float64(colorRgb[0]) / 255.0
				g := float64(colorRgb[1]) / 255.0
				b := float64(colorRgb[2]) / 255.0
				dc.SetRGB(r, g, b)

				// Text alignment
				w, h := dc.MeasureString(row[rowIndex[ph.CsvKey]])
				startX := ph.StartX
				startY := ph.StartY + ph.H - h*0.8 // adjust the baseline factor if needed

				switch ph.TextAlign {
				case "center":
					startX = ph.StartX + (ph.W-w)/2
				case "right":
					startX = ph.StartX + ph.W - w
				}

				// Transformation
				content := row[rowIndex[ph.CsvKey]]
				switch ph.Transform {
				case "uppercase":
					content = strings.ToUpper(content)
				case "qrcode":
					qrCode, err := qrcode.New(content, qrcode.Medium)
					if err != nil {
						log.Println(err)
					} else {
						qrImage := qrCode.Image(int(ph.W))
						dc.DrawImage(qrImage, int(startX), int(startY))
						dc.Pop()
						continue
					}
				case "avatar":
					// current gender are still static, this need to be fixed
					if avtr, err := govatar.Generate(govatar.MALE); err == nil {
						// Resize the image
						resizedImg := imaging.Resize(avtr, int(ph.W), int(ph.W), imaging.Lanczos)

						dc.DrawImage(resizedImg, int(startX), int(startY))
						dc.Pop()
						continue
					}
				case "initial-avatar":
					a := avatar.New("")
					if raw, err := a.DrawToBytes(content, int(ph.W), "png"); err == nil {
						if img, perr := png.Decode(bytes.NewReader(raw)); perr == nil {
							dc.DrawImage(img, int(startX), int(startY))
							dc.Pop()
							continue
						}
					}
				}

				// Draw the string
				dc.DrawStringAnchored(content, startX, startY, 0, 1)
				dc.Pop()
			}
		}

		dc.SavePNG(fmt.Sprintf("%s%sout%d.png", savePath, string(os.PathSeparator), counter))
		counter++
	}

	return "Success, saved in " + savePath
}
