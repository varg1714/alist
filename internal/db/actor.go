package db

import (
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/pkg/errors"
)

func CreateActor(dir string, name string, url string) error {

	actor := model.Actor{
		Dir:  dir,
		Name: name,
		Url:  url,
	}

	err := db.Create(&actor).Error
	return errors.WithStack(err)

}

func QueryActor(source string) []model.Actor {

	actors := make([]model.Actor, 0)
	actor := model.Actor{
		Dir: source,
	}

	db.Where(actor).Order("updated_at desc").Find(&actors)

	return actors

}

func DeleteActor(source string, actor string) error {

	return errors.WithStack(db.Where("dir = ?", source).Where("name = ?", actor).Delete(&model.Actor{}).Error)

}
