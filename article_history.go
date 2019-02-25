package solitudes

import (
	"fmt"
	"time"
)

// ArticleHistory 文章修订历史
type ArticleHistory struct {
	ArticleID uint
	Article   Article
	Version   uint   `gorm:"index"`
	Desc      string `gorm:"text"`
	Content   string `gorm:"text"`
	CreatedAt time.Time
}

// GetIndexID get index data id
func (t *ArticleHistory) GetIndexID() string {
	return fmt.Sprintf("%d.%d", t.ArticleID, t.Version)
}
