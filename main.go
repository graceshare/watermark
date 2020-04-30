package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

func outputFilename(input string) string {
	lastDot := strings.LastIndex(input, ".")
	return input[:lastDot] + ".watermarked" + input[lastDot:]
}

func readResizeParam(value string) (Size, error) {
	resizeRe := regexp.MustCompile(`^(\d+)x(\d+)$`)
	params := resizeRe.FindStringSubmatch(value)
	if len(params) == 0 {
		return Size{}, fmt.Errorf(`invalid resize value, should be like widthxheight`)
	}

	width, _ := strconv.Atoi(params[1])
	height, _ := strconv.Atoi(params[2])

	return Size{width, height}, nil
}

func main() {
	app := &cli.App{
		Name:      "watermark",
		Usage:     "add watermark on images",
		UsageText: "watermark [OPTIONS] TEXT FILE ...",
		HideHelp:  true,
		Version:   "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "color",
				Aliases: []string{"c"},
				Value:   "#ffffff",
				Usage:   "Color for watermark text, name or #rrggbb",
			},
			&cli.Float64Flag{
				Name:    "vertical-spacing",
				Aliases: []string{"e"},
				Value:   200.0,
				Usage:   "Vertical spacing between watermarks",
			},
			&cli.StringFlag{
				Name:    "resize",
				Aliases: []string{"r"},
				Value:   "0x0",
				Usage:   "Resize the output image to specified width and height. To keep aspect ratio, set either width or height to 0",
			},
			&cli.StringFlag{
				Name:    "font",
				Aliases: []string{"f"},
				Value:   "Courier",
				Usage:   "Font for watermark text",
			},
			&cli.StringFlag{
				Name:    "font-path",
				Aliases: []string{"af"},
				Value:   "/Users/vanilla/gopath/src/github.com/graceshare/watermark/RuiZiYunZiKuKaiTiGBK-1.ttf",
				Usage:   "Add local font for watermark text",
			},
			&cli.Float64Flag{
				Name:    "font-size",
				Aliases: []string{"S"},
				Value:   30.0,
				Usage:   "Font size for watermark text",
			},
			&cli.Float64Flag{
				Name:    "horizontal-spacing",
				Aliases: []string{"o"},
				Value:   25.0,
				Usage:   "Horizontal spacing between watermarks 横向间距",
			},
			&cli.Float64Flag{
				Name:    "scale",
				Aliases: []string{"s"},
				Value:   1.0,
				Usage:   "Scale watermarks",
			},
			&cli.Float64Flag{
				Name:    "transparency",
				Aliases: []string{"t"},
				Value:   0.30,
				Usage:   "Transparency of watermark",
			},
			&cli.StringSliceFlag{
				Name:  "output",
				Usage: `Specify output names. This flag can be repeated many times, input and output names will be matched in order (1st to 1st, 2nd to 2nd, ...). Unspecified outputs will have the name "input name.watermarked.extension"`,
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		if c.NArg() < 2 {
			return fmt.Errorf("not enough arguments")
		}

		text := c.Args().Get(0)

		color, err := parseColor(c.String("color"))
		if err != nil {
			return err
		}

		color.A = uint8((1 - c.Float64("transparency")) * 255.0)

		scale := c.Float64("scale")

		newSize, err := readResizeParam(c.String("resize"))
		if err != nil {
			return err
		}

		waterMarker := &WaterMarker{
			Text:              "此图片仅供办理" + text + "业务使用，他用无效",
			HorizontalSpacing: c.Float64("horizontal-spacing") * scale,
			VerticalSpacing:   c.Float64("vertical-spacing") * scale,
			FontSize:          c.Float64("font-size") * scale,
			OutputDPI:         72,
			FontName:          c.String("font"),
			Color:             color,
			FontPath:          c.String("font-path"),
			Resize:            newSize,
		}

		outputNames := c.StringSlice("output")

		for i := 1; i < c.NArg(); i++ {
			file := c.Args().Get(i)

			output := outputFilename(file)
			if i <= len(outputNames) {
				output = outputNames[i-1]
			}

			err := waterMarker.Mark(file, output)
			if err != nil {
				return err
			}

			fmt.Println(file, "->", output)
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println("error:", err)
	}
}
