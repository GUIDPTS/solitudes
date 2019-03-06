package solitudes

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

// ArticleTOC 文章标题
type ArticleTOC struct {
	Title     string
	Slug      string
	SubTitles []*ArticleTOC
	Parent    *ArticleTOC `gorm:"-" json:"-"`
	Level     int         `gorm:"-" json:"-"`
	ShowLevel int         `gorm:"-" json:"-"`
}

// SibilingArticle 相邻文章
type SibilingArticle struct {
	Next Article
	Prev Article
}

// Article 文章表
type Article struct {
	gorm.Model
	Slug    string `form:"slug" binding:"required" gorm:"unique_index" json:"slug,omitempty"`
	Title   string `form:"title" binding:"required" json:"title,omitempty"`
	Content string `form:"content" binding:"required" gorm:"text" json:"content,omitempty"`

	TemplateID byte           `form:"template" binding:"required" json:"template_id,omitempty"`
	IsBook     bool           `form:"is_book" json:"is_book,omitempty"`
	RawTags    string         `form:"tags" gorm:"-" json:"-"`
	Tags       pq.StringArray `gorm:"index;type:varchar(255)[]" json:"tags,omitempty"`
	ReadNum    uint           `gorm:"default:0;" json:"read_num,omitempty"`
	CommentNum uint           `gorm:"default:0;"`
	Version    uint           `form:"version" gorm:"default:1;"`
	BookRefer  uint           `form:"book_refer" gorm:"index" json:"book_refer,omitempty"`

	Comments         []*Comment `json:"comments,omitempty"`
	ArticleHistories []*ArticleHistory

	Toc             []*ArticleTOC    `gorm:"-"`
	Chapters        []*Article       `gorm:"foreignkey:BookRefer" form:"-" binding:"-"`
	Book            *Article         `gorm:"-" binding:"-" form:"-" json:"-"`
	SibilingArticle *SibilingArticle `gorm:"-" binding:"-" form:"-" json:"-"`
}

// SID string id
func (t *Article) SID() string {
	return fmt.Sprintf("%d", t.ID)
}

// ArticleIndex index data
type ArticleIndex struct {
	Slug    string
	Version string
	Title   string
	Content string
}

// ToIndexData to index data
func (t *Article) ToIndexData() ArticleIndex {
	return ArticleIndex{
		Slug:    t.Slug,
		Version: fmt.Sprintf("%d", t.Version),
		Content: t.Content,
		Title:   t.Title,
	}
}

// GetIndexID get index data id
func (t *Article) GetIndexID() string {
	return fmt.Sprintf("%d.%d", t.ID, t.Version)
}

// BeforeSave hook
func (t *Article) BeforeSave() {
	t.Tags = strings.Split(t.RawTags, ",")
}

// AfterFind hook
func (t *Article) AfterFind() {
	t.RawTags = strings.Join(t.Tags, ",")
}

var titleRegex = regexp.MustCompile(`^\s{0,2}(#{1,6})\s(.*)$`)
var whitespaces = regexp.MustCompile(`[\s|\.]{1,}`)

// GenTOC 生成标题树
func (t *Article) GenTOC() {
	lines := strings.Split(t.Content, "\n")
	var matches []string
	var currentToc *ArticleTOC
	for j := 0; j < len(lines); j++ {
		matches = titleRegex.FindStringSubmatch(lines[j])
		if len(matches) != 3 {
			continue
		}
		var toc ArticleTOC
		toc.Level = len(matches[1])
		toc.ShowLevel = 2
		toc.Title = string(matches[2])
		toc.Slug = string(whitespaces.ReplaceAllString(matches[2], "-"))
		if currentToc == nil {
			t.Toc = append(t.Toc, &toc)
			currentToc = &toc
			continue
		}
		parent := currentToc
		if currentToc.Level > toc.Level {
			// 父节点
			for i := -1; i < currentToc.Level-toc.Level; i++ {
				parent = parent.Parent
				if parent == nil || parent.Level < toc.Level {
					break
				}
			}
			if parent == nil {
				t.Toc = append(t.Toc, &toc)
			} else {
				toc.Parent = parent
				toc.ShowLevel = parent.ShowLevel + 1
				parent.SubTitles = append(parent.SubTitles, &toc)
			}
		} else if currentToc.Level == toc.Level {
			// 兄弟节点
			if parent.Parent == nil {
				t.Toc = append(t.Toc, &toc)
			} else {
				toc.Parent = parent.Parent
				toc.ShowLevel = parent.ShowLevel + 1
				parent.Parent.SubTitles = append(parent.Parent.SubTitles, &toc)
			}
		} else {
			// 子节点
			toc.Parent = parent
			toc.ShowLevel = parent.ShowLevel + 1
			parent.SubTitles = append(parent.SubTitles, &toc)
		}
		currentToc = &toc
	}
}

// BuildArticleIndex 重建索引
func BuildArticleIndex() {
	var as []Article
	System.DB.Find(&as)
	for i := 0; i < len(as); i++ {
		err := System.Search.Index(as[i].GetIndexID(), as[i].ToIndexData())
		if err != nil {
			panic(err)
		}
	}
}
