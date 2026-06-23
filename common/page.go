package common

import (
	"fmt"
	"strconv"

	"github.com/labstack/echo/v5"
)

type Page struct {
	current uint
	size    uint
}

// Current 返回当前页码
func (p *Page) Current() uint {
	return p.current
}

// Size 返回每页大小
func (p *Page) Size() uint {
	return p.size
}

// PageBuilder 分页参数构建器
type PageBuilder struct {
	c           *echo.Context
	defaultPage uint
	defaultSize uint
	maxSize     uint
	withError   bool
}

// NewPageBuilder 创建分页构建器，默认 (page=1, size=10, maxSize=100)
func NewPageBuilder(c *echo.Context) *PageBuilder {
	return &PageBuilder{
		c:           c,
		defaultPage: 1,
		defaultSize: 10,
		maxSize:     100,
		withError:   false,
	}
}

// WithDefault 设置默认页码和每页大小
func (b *PageBuilder) WithDefault(page, size uint) *PageBuilder {
	if page > 0 {
		b.defaultPage = page
	}
	if size > 0 {
		b.defaultSize = size
	}
	return b
}

// WithDefaultPage 设置默认页码
func (b *PageBuilder) WithDefaultPage(page uint) *PageBuilder {
	if page > 0 {
		b.defaultPage = page
	}
	return b
}

// WithDefaultSize 设置默认每页大小
func (b *PageBuilder) WithDefaultSize(size uint) *PageBuilder {
	if size > 0 {
		b.defaultSize = size
	}
	return b
}

// WithMaxSize 设置最大每页大小限制
func (b *PageBuilder) WithMaxSize(maxSize uint) *PageBuilder {
	if maxSize > 0 {
		b.maxSize = maxSize
	}
	return b
}

// WithError 启用错误处理，参数无效时返回错误
func (b *PageBuilder) WithError(enable bool) *PageBuilder {
	b.withError = enable
	return b
}

// Build 构建分页对象
// 如果启用了错误处理，返回 (*Page, error)；否则返回 (*Page, nil)
func (b *PageBuilder) Build() (*Page, error) {
	pageStr := b.c.QueryParam("page")
	pageSizeStr := b.c.QueryParam("page_size")

	page := b.defaultPage
	size := b.defaultSize

	// 解析页码
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			if b.withError {
				return nil, fmt.Errorf("invalid page parameter: %w", err)
			}
			// 无效则使用默认值
		} else if p > 0 {
			page = uint(p)
		} else if b.withError {
			return nil, fmt.Errorf("page must be greater than 0")
		}
	}

	// 解析每页大小
	if pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			if b.withError {
				return nil, fmt.Errorf("invalid page_size parameter: %w", err)
			}
			// 无效则使用默认值
		} else if ps > 0 {
			size = uint(ps)
			if size > b.maxSize {
				size = b.maxSize
			}
		} else if b.withError {
			return nil, fmt.Errorf("page_size must be greater than 0")
		}
	}

	return &Page{
		current: page,
		size:    size,
	}, nil
}

// MustBuild 构建分页对象，如果出错则 panic
func (b *PageBuilder) MustBuild() *Page {
	page, err := b.Build()
	if err != nil {
		panic(err)
	}
	return page
}
