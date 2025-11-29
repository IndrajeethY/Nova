package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func checkImageMagick() bool {
	_, err := utils.RunCommand("convert -version")
	return err == nil
}

func imageGreyCommand(m *telegram.NewMessage) error {
	return processImage(m, "grey", func(inputPath, outputPath string) error {
		cmd := fmt.Sprintf("convert %q -colorspace Gray %q", inputPath, outputPath)
		_, err := utils.RunCommand(cmd)
		return err
	})
}

func imageBlurCommand(m *telegram.NewMessage) error {
	return processImage(m, "blur", func(inputPath, outputPath string) error {
		cmd := fmt.Sprintf("convert %q -blur 0x8 %q", inputPath, outputPath)
		_, err := utils.RunCommand(cmd)
		return err
	})
}

func imageNegativeCommand(m *telegram.NewMessage) error {
	return processImage(m, "negative", func(inputPath, outputPath string) error {
		cmd := fmt.Sprintf("convert %q -negate %q", inputPath, outputPath)
		_, err := utils.RunCommand(cmd)
		return err
	})
}

func imageMirrorCommand(m *telegram.NewMessage) error {
	return processImage(m, "mirror", func(inputPath, outputPath string) error {
		cmd := fmt.Sprintf("convert %q -flop %q", inputPath, outputPath)
		_, err := utils.RunCommand(cmd)
		return err
	})
}

func imageFlipCommand(m *telegram.NewMessage) error {
	return processImage(m, "flip", func(inputPath, outputPath string) error {
		cmd := fmt.Sprintf("convert %q -flip %q", inputPath, outputPath)
		_, err := utils.RunCommand(cmd)
		return err
	})
}

func imageRotateCommand(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	angle := 90
	if args != "" {
		if a, err := strconv.Atoi(args); err == nil {
			angle = a
		}
	}

	return processImage(m, "rotate", func(inputPath, outputPath string) error {
		cmd := fmt.Sprintf("convert %q -rotate %d %q", inputPath, angle, outputPath)
		_, err := utils.RunCommand(cmd)
		return err
	})
}

func imageSketchCommand(m *telegram.NewMessage) error {
	return processImage(m, "sketch", func(inputPath, outputPath string) error {
		cmd := fmt.Sprintf("convert %q -colorspace Gray -sketch 0x20+120 %q", inputPath, outputPath)
		_, err := utils.RunCommand(cmd)
		return err
	})
}

func imageBorderCommand(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	borderColor := "white"
	borderWidth := 20

	if args != "" {
		parts := strings.Split(args, ";")
		borderColor = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			if w, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil && w > 0 {
				borderWidth = w
			}
		}
	}

	return processImage(m, "border", func(inputPath, outputPath string) error {
		cmd := fmt.Sprintf("convert %q -bordercolor %q -border %d %q", inputPath, borderColor, borderWidth, outputPath)
		_, err := utils.RunCommand(cmd)
		return err
	})
}

func imagePixelateCommand(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	scale := 10
	if args != "" {
		if s, err := strconv.Atoi(args); err == nil && s > 0 && s <= 100 {
			scale = s
		}
	}

	return processImage(m, "pixelate", func(inputPath, outputPath string) error {
		cmd := fmt.Sprintf("convert %q -scale %d%% -scale 1000%% %q", inputPath, scale, outputPath)
		_, err := utils.RunCommand(cmd)
		return err
	})
}

func imageSepiaCommand(m *telegram.NewMessage) error {
	return processImage(m, "sepia", func(inputPath, outputPath string) error {
		cmd := fmt.Sprintf("convert %q -sepia-tone 80%% %q", inputPath, outputPath)
		_, err := utils.RunCommand(cmd)
		return err
	})
}

func imageEmbossCommand(m *telegram.NewMessage) error {
	return processImage(m, "emboss", func(inputPath, outputPath string) error {
		cmd := fmt.Sprintf("convert %q -emboss 0x1 %q", inputPath, outputPath)
		_, err := utils.RunCommand(cmd)
		return err
	})
}

func imageSharpenCommand(m *telegram.NewMessage) error {
	return processImage(m, "sharpen", func(inputPath, outputPath string) error {
		cmd := fmt.Sprintf("convert %q -sharpen 0x2 %q", inputPath, outputPath)
		_, err := utils.RunCommand(cmd)
		return err
	})
}

func imageResizeCommand(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	if args == "" {
		_, err := eOR(m, locales.Tr("imagetools.resize_usage"))
		return err
	}

	size := args
	if !strings.Contains(size, "x") && !strings.HasSuffix(size, "%") {
		size = size + "x" + size
	}

	return processImage(m, "resize", func(inputPath, outputPath string) error {
		cmd := fmt.Sprintf("convert %q -resize %s %q", inputPath, size, outputPath)
		_, err := utils.RunCommand(cmd)
		return err
	})
}

func colorSampleCommand(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	if args == "" {
		_, err := eOR(m, locales.Tr("imagetools.csample_usage"))
		return err
	}

	if !checkImageMagick() {
		_, err := eOR(m, locales.Tr("imagetools.imagemagick_missing"))
		return err
	}

	color := args
	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("csample_%d.png", m.ID))
	defer os.Remove(outputPath)

	cmd := fmt.Sprintf("convert -size 200x100 xc:%q %q", color, outputPath)
	_, err := utils.RunCommand(cmd)
	if err != nil {
		_, err := eOR(m, locales.Tr("imagetools.invalid_color"))
		return err
	}

	_, err = m.Respond(fmt.Sprintf(locales.Tr("imagetools.csample_result"), color), &telegram.SendOptions{
		Media: outputPath,
	})
	if m.Sender.ID == ubId {
		m.Delete()
	}
	return err
}

func processImage(m *telegram.NewMessage, operation string, processor func(inputPath, outputPath string) error) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("imagetools.reply_required"))
		return err
	}

	if !checkImageMagick() {
		_, err := eOR(m, locales.Tr("imagetools.imagemagick_missing"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("imagetools.fetch_error"))
		return err
	}

	if reply.Photo() == nil && reply.Sticker() == nil && reply.Document() == nil {
		_, err := eOR(m, locales.Tr("imagetools.no_image"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("imagetools.downloading"))

	inputPath, err := reply.Download()
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("imagetools.download_error"))
		}
		return err
	}
	defer os.Remove(inputPath)

	if msg != nil {
		msg.Edit(locales.Tr("imagetools.processing"))
	}

	ext := filepath.Ext(inputPath)
	if ext == "" || ext == ".tgs" {
		ext = ".png"
	}
	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%d%s", operation, m.ID, ext))
	defer os.Remove(outputPath)

	if err := processor(inputPath, outputPath); err != nil {
		if msg != nil {
			msg.Edit(fmt.Sprintf(locales.Tr("imagetools.process_error"), err.Error()))
		}
		return err
	}

	if msg != nil {
		msg.Edit(locales.Tr("imagetools.uploading"))
	}

	_, err = m.Respond("", &telegram.SendOptions{Media: outputPath})
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("imagetools.upload_error"))
		}
		return err
	}

	if msg != nil {
		msg.Delete()
	}
	return nil
}

func LoadImageToolsModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "grey", Func: imageGreyCommand, Description: "Convert image to grayscale", ModuleName: "ImageTools"},
		{Command: "blur", Func: imageBlurCommand, Description: "Apply blur effect to image", ModuleName: "ImageTools"},
		{Command: "negative", Func: imageNegativeCommand, Description: "Create negative of image", ModuleName: "ImageTools"},
		{Command: "mirror", Func: imageMirrorCommand, Description: "Mirror image horizontally", ModuleName: "ImageTools"},
		{Command: "flip", Func: imageFlipCommand, Description: "Flip image vertically", ModuleName: "ImageTools"},
		{Command: "rotate", Func: imageRotateCommand, Description: "Rotate image by angle (default: 90)", ModuleName: "ImageTools"},
		{Command: "sketch", Func: imageSketchCommand, Description: "Convert image to sketch", ModuleName: "ImageTools"},
		{Command: "border", Func: imageBorderCommand, Description: "Add border (color ; width)", ModuleName: "ImageTools"},
		{Command: "pixelate", Func: imagePixelateCommand, Description: "Pixelate image (scale 1-100)", ModuleName: "ImageTools"},
		{Command: "sepia", Func: imageSepiaCommand, Description: "Apply sepia tone effect", ModuleName: "ImageTools"},
		{Command: "emboss", Func: imageEmbossCommand, Description: "Apply emboss effect", ModuleName: "ImageTools"},
		{Command: "sharpen", Func: imageSharpenCommand, Description: "Sharpen image", ModuleName: "ImageTools"},
		{Command: "resize", Func: imageResizeCommand, Description: "Resize image (WxH or %)", ModuleName: "ImageTools"},
		{Command: "csample", Func: colorSampleCommand, Description: "Create color sample (color name/hex)", ModuleName: "ImageTools"},
	}
	AddHandlers(handlers, c)
}
