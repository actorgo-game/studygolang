package apiv1

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"

	"github.com/studygolang/studygolang/context"
	"github.com/studygolang/studygolang/internal/logic"
	"github.com/studygolang/studygolang/internal/model"

	echo "github.com/labstack/echo/v4"
)

type ImageController struct{}

func (self ImageController) RegisterRoute(g *echo.Group) {
	g.POST("/image/upload", self.Upload)
}

func (ImageController) Upload(ctx echo.Context) error {
	meVal, ok := ctx.Get("user").(*model.Me)
	if !ok || meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	file, err := ctx.FormFile("img")
	if err != nil {
		return fail(ctx, "请选择图片")
	}
	ext := filepath.Ext(file.Filename)
	if !isImageExt(ext) {
		return fail(ctx, "不支持的图片格式")
	}

	src, err := file.Open()
	if err != nil {
		return fail(ctx, "读取图片失败")
	}
	defer src.Close()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, src)
	if err != nil {
		return fail(ctx, "读取图片失败")
	}

	imgUrl, err := logic.DefaultUploader.UploadImage(context.EchoContext(ctx), buf, "img", buf.Bytes(), ext)
	if err != nil {
		return fail(ctx, "上传图片失败")
	}
	return success(ctx, map[string]interface{}{"url": imgUrl})
}

func isImageExt(ext string) bool {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		return true
	}
	return false
}
