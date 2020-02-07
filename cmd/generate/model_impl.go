package generate

import (
	"context"
	"fmt"

	"github.com/LyricTian/gin-admin-cli/util"
)

func getModelImplFileName(dir, name string) string {
	fullname := fmt.Sprintf("%s/model/m_%s.go", dir, util.ToLowerUnderlinedNamer(name))
	return fullname
}

// 生成model实现文件
func genModelImpl(ctx context.Context, pkgName, dir, name, comment string) error {
	data := map[string]interface{}{
		"PkgName":    pkgName,
		"Name":       name,
		"PluralName": util.ToPlural(name),
		"Comment":    comment,
	}

	buf, err := execParseTpl(modelImplTpl, data)
	if err != nil {
		return err
	}

	fullname := getModelImplFileName(dir, name)
	err = createFile(ctx, fullname, buf)
	if err != nil {
		return err
	}

	fmt.Printf("文件[%s]写入成功\n", fullname)

	return execGoFmt(fullname)
}

const modelImplTpl = `
package model

import (
	"context"

	"{{.PkgName}}/errors"
	"{{.PkgName}}/model/entity"
	"{{.PkgName}}/schema"
	"github.com/jinzhu/gorm"
)

// {{.Name}} {{.Comment}}存储
type {{.Name}} struct {
	db *gorm.DB
}

func (a *{{.Name}}) getQueryOption(opts ...schema.{{.Name}}QueryOptions) schema.{{.Name}}QueryOptions {
	var opt schema.{{.Name}}QueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	return opt
}

// Query 查询数据
func (a *{{.Name}}) Query(ctx context.Context, params schema.{{.Name}}QueryParam, opts ...schema.{{.Name}}QueryOptions) (*schema.{{.Name}}QueryResult, error) {
	opt := a.getQueryOption(opts...)
	db := entity.Get{{.Name}}DB(ctx, DB)
	// TODO: 查询条件
	db = db.Order("id DESC")

	var list entity.{{.PluralName}}
	pr, err := WrapPageQuery(ctx, db, opt.PageParam, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	qr := &schema.{{.Name}}QueryResult{
		PageResult: pr,
		Data:       list.ToSchema{{.PluralName}}(),
	}

	return qr, nil
}

// Get 查询指定数据
func (a *{{.Name}}) Get(ctx context.Context, recordID string, opts ...schema.{{.Name}}QueryOptions) (*schema.{{.Name}}, error) {
	db := entity.Get{{.Name}}DB(ctx, DB).Where("record_id=?", recordID)
	var item entity.{{.Name}}
	ok, err := FindOne(ctx, db, &item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}

	return item.ToSchema{{.Name}}(), nil
}

// Create 创建数据
func (a *{{.Name}}) Create(ctx context.Context, item schema.{{.Name}}) error {
	eitem := entity.Schema{{.Name}}(item).To{{.Name}}()
	result := entity.Get{{.Name}}DB(ctx, DB).Create(eitem)
	if err := result.Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Update 更新数据
func (a *{{.Name}}) Update(ctx context.Context, recordID string, item schema.{{.Name}}) error {
	eitem := entity.Schema{{.Name}}(item).To{{.Name}}()
	result := entity.Get{{.Name}}DB(ctx, DB).Where("record_id=?", recordID).Omit("record_id").Updates(eitem)
	if err := result.Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Delete 删除数据
func (a *{{.Name}}) Delete(ctx context.Context, recordID string) error {
	result := entity.Get{{.Name}}DB(ctx, DB).Where("record_id=?", recordID).Delete(entity.{{.Name}}{})
	if err := result.Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

`
