package apiv1

import (
	"github.com/studygolang/studygolang/context"
	. "github.com/studygolang/studygolang/internal/http"
	"github.com/studygolang/studygolang/internal/logic"
	"github.com/studygolang/studygolang/internal/model"

	echo "github.com/labstack/echo/v4"
	"github.com/polaris1119/goutils"
)

type BookController struct{}

func (self BookController) RegisterRoute(g *echo.Group) {
	g.GET("/books", self.ReadList)
	g.GET("/book/:id", self.Detail)
	g.POST("/book/new", self.Create)
	g.POST("/book/delete", self.Delete)
}

func (BookController) ReadList(ctx echo.Context) error {
	curPage := goutils.MustInt(ctx.QueryParam("p"), 1)
	paginator := logic.NewPaginatorWithPerPage(curPage, perPage)
	books := logic.DefaultGoBook.FindAll(context.EchoContext(ctx), paginator, "")
	total := logic.DefaultGoBook.Count(context.EchoContext(ctx))
	return success(ctx, map[string]interface{}{
		"list":     books,
		"total":    total,
		"page":     curPage,
		"per_page": perPage,
	})
}

func (BookController) Detail(ctx echo.Context) error {
	id := goutils.MustInt(ctx.Param("id"))
	book, err := logic.DefaultGoBook.FindById(context.EchoContext(ctx), id)
	if err != nil || book.Id == 0 {
		return fail(ctx, "图书不存在")
	}
	logic.Views.Incr(Request(ctx), model.TypeBook, id)
	return success(ctx, map[string]interface{}{"book": book})
}

func (BookController) Create(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	formParams, _ := ctx.FormParams()
	err := logic.DefaultGoBook.Publish(context.EchoContext(ctx), meVal, formParams)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}

func (BookController) Delete(ctx echo.Context) error {
	meVal := me(ctx)
	if meVal.Uid == 0 {
		return fail(ctx, "请先登录")
	}
	if !meVal.IsRoot {
		return fail(ctx, "无权操作")
	}
	id := goutils.MustInt(ctx.FormValue("id"))
	err := logic.DefaultGoBook.Delete(context.EchoContext(ctx), id)
	if err != nil {
		return fail(ctx, err.Error())
	}
	return success(ctx, nil)
}
