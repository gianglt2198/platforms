package mydatabase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	myerrors "github.com/gianglt2198/platforms/internal/errors"
)

const (
	KEY_AUTH_USER    = "auth_user_key"
	KEY_CURRENT_TRAN = "currrent_transaction_key"
)

type WhereOption struct {
	Where  string
	Params []interface{}
}

type PaginationQuery struct {
	CountSQL    string
	QuerySQL    string
	Params      []interface{}
	Page        int
	Take        int
	ItemMappers func(*sql.Rows) interface{}
}

type PreloadWithCond struct {
	Query string
	Args  []interface{}
}

type FindOption struct {
	Where           string
	Params          []interface{}
	Page            int
	Take            int
	Order           string
	Preload         []string
	PreloadWithCond []PreloadWithCond
	Select          *[]string
	ExcludeDeleted  bool
	Joins           []string
}

type (
	RepositoryIf[T any] interface {
		CreateOne(ctx context.Context, entity *T) (*T, *myerrors.AppError)
		Create(ctx context.Context, entities ...*T) ([]*T, *myerrors.AppError)
		CreateWithOnConflicting(ctx context.Context, conflictColumns []string, needUpdateColumns []string, entities ...*T) ([]*T, *myerrors.AppError)
		UpdateById(ctx context.Context, id int, updatedFields *T) *myerrors.AppError
		UpdateBy(ctx context.Context, updateValues *T, cond WhereOption) *myerrors.AppError
		UpdateByMap(ctx context.Context, values map[string]interface{}, cond WhereOption) *myerrors.AppError
		DeleteById(ctx context.Context, id int) *myerrors.AppError
		DeleteBy(ctx context.Context, cond WhereOption) *myerrors.AppError
		FindById(ctx context.Context, id int) (*T, *myerrors.AppError)
		FindAll(ctx context.Context) (*[]T, *myerrors.AppError)
		Pagination(ctx context.Context, option *FindOption) (int, *[]T, *myerrors.AppError)
		PaginationQuery(ctx context.Context, option *PaginationQuery) (int, *[]interface{}, *myerrors.AppError)
		FindOneBy(ctx context.Context, option *FindOption) (*T, *myerrors.AppError)
		FindBy(ctx context.Context, option *FindOption) (*[]T, *myerrors.AppError)
		FirstOrInitBy(ctx context.Context, options FindOption, entity *T) (*T, *myerrors.AppError)
		FirstOrCreateBy(ctx context.Context, options FindOption, entity *T) (*T, *myerrors.AppError)
		CountAll(ctx context.Context) int
		CountBy(ctx context.Context, cond WhereOption) int
		QueryBuilder(ctx context.Context) *gorm.DB
		GetExistsIdsByIds(context.Context, []uint) (*[]uint, *myerrors.AppError)
		IsExistById(context.Context, int) (*bool, *myerrors.AppError)
	}

	Repository[T any] struct {
		db *gorm.DB
	}
)

func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{
		db: db,
	}
}

func (r *Repository[T]) QueryBuilder(ctx context.Context) *gorm.DB {
	var model T

	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	return db.WithContext(ctx).Model(&model)
}

func (r *Repository[T]) CreateOne(ctx context.Context, entity *T) (*T, *myerrors.AppError) {
	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	err := db.WithContext(ctx).Create(entity).Error

	if err != nil {
		return nil, myerrors.QueryInvalid(err.Error())
	}

	return entity, nil
}

func (r *Repository[T]) Create(ctx context.Context, entities ...*T) ([]*T, *myerrors.AppError) {
	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	currentUser := ctx.Value(KEY_AUTH_USER).(map[string]interface{})
	if currentUser != nil {
		for _, e := range entities {
			now := time.Now().UTC()

			SetAttribute(e, "CreatedAt", now)
			SetAttribute(e, "CreatedBy", currentUser["id"])

			if hasAttribute(e, "UpdatedAt") {
				SetAttribute(e, "UpdatedAt", &now)
			}

			if hasAttribute(e, "UpdatedBy") {
				SetAttribute(e, "UpdatedBy", currentUser["id"])
			}
		}
	}

	err := db.WithContext(ctx).Create(&entities).Error

	if err != nil {
		return nil, myerrors.QueryInvalid(err.Error())
	}

	return entities, nil
}

func (r *Repository[T]) CreateWithOnConflicting(ctx context.Context, conflictColumns []string, needUpdateColumns []string, entities ...*T) ([]*T, *myerrors.AppError) {
	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	currentUser := ctx.Value(KEY_AUTH_USER).(map[string]interface{})
	if currentUser != nil {
		for _, e := range entities {
			now := time.Now().UTC()

			SetAttribute(e, "CreatedAt", now)
			SetAttribute(e, "CreatedBy", currentUser["id"])

			if hasAttribute(e, "UpdatedAt") {
				SetAttribute(e, "UpdatedAt", &now)
			}

			if hasAttribute(e, "UpdatedBy") {
				SetAttribute(e, "UpdatedBy", currentUser["id"])
			}
		}
	}

	clauseColumns := make([]clause.Column, len(conflictColumns))

	for i, c := range conflictColumns {
		clauseColumns[i] = clause.Column{Name: c}
	}

	err := db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   clauseColumns,
		DoNothing: len(needUpdateColumns) == 0,
		DoUpdates: clause.AssignmentColumns(needUpdateColumns),
	}).Create(&entities).Error

	if err != nil {
		return nil, myerrors.QueryInvalid(err.Error())
	}

	return entities, nil
}

func (r *Repository[T]) UpdateById(ctx context.Context, id int, updatedFields *T) *myerrors.AppError {
	var model T

	db := r.db

	currentUser := ctx.Value(KEY_AUTH_USER).(map[string]interface{})

	if currentUser != nil {
		currentTime := time.Now().UTC()
		SetAttribute(updatedFields, "UpdatedAt", &currentTime)
		SetAttribute(updatedFields, "UpdatedBy", currentUser["id"])
	}

	var err error
	if hasAttribute(model, "DeletedAt") {
		err = db.WithContext(ctx).Model(&model).Where("id = ? AND deleted_at IS NULL", id).Updates(updatedFields).Error
	} else {
		err = db.WithContext(ctx).Model(&model).Where("id = ?", id).Updates(updatedFields).Error
	}

	if err != nil {
		return myerrors.QueryInvalid(err.Error())
	}

	return nil
}

func (r *Repository[T]) UpdateBy(
	ctx context.Context,
	updateValues *T,
	cond WhereOption,
) *myerrors.AppError {
	var model T

	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	currentUser := ctx.Value(KEY_AUTH_USER).(map[string]interface{})

	if currentUser != nil {
		currentTime := time.Now().UTC()
		SetAttribute(updateValues, "UpdatedAt", &currentTime)
		SetAttribute(updateValues, "UpdatedBy", currentUser["id"])
	}

	err := db.WithContext(ctx).
		Model(&model).
		Where(cond.Where, cond.Params...).
		Updates(updateValues).
		Error

	if err != nil {
		return myerrors.QueryInvalid(err.Error())
	}

	return nil
}

func (r *Repository[T]) UpdateByMap(ctx context.Context, values map[string]interface{}, cond WhereOption) *myerrors.AppError {
	var model T

	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	if hasAttribute(model, "DeletedAt") {
		cond.Where = strings.Join([]string{cond.Where, " AND deleted_at IS NULL"}, " and ")
	}

	if hasAttribute(model, "UpdatedAt") {
		currentTime := time.Now().UTC()
		values["updated_at"] = &currentTime
	}

	if hasAttribute(model, "UpdatedBy") {
		currentUser := ctx.Value(KEY_AUTH_USER).(map[string]interface{})

		if currentUser != nil {
			values["updated_by"] = currentUser["id"]
		}
	}

	err := db.WithContext(ctx).Model(&model).Where(cond.Where, cond.Params...).Updates(values).Error
	if err != nil {
		return myerrors.QueryInvalid(err.Error())
	}

	return nil
}

func (r *Repository[T]) DeleteById(ctx context.Context, id int) *myerrors.AppError {
	var entity T

	db := r.db
	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	// Retrieve the record by ID to ensure AfterDelete hook can access its fields
	if err := db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return myerrors.QueryInvalid("record not found")
		}
		return myerrors.QueryInvalid(err.Error())
	}

	var err error
	if hasAttribute(entity, "DeletedAt") {
		currentUser := ctx.Value(KEY_AUTH_USER).(map[string]interface{})

		if currentUser != nil {
			err = db.WithContext(ctx).
				Model(&entity).
				Where("id = ?", id).
				Updates(map[string]interface{}{
					"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
					"deleted_by": currentUser["id"],
				},
				).Error
		} else {
			err = db.WithContext(ctx).
				Model(&entity).
				Where("id = ?", id).
				Updates(map[string]interface{}{
					"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
				},
				).Error
		}
	} else {
		err = db.WithContext(ctx).Delete(&entity).Error
	}

	if err != nil {
		return myerrors.QueryInvalid(err.Error())
	}

	return nil
}

func (r *Repository[T]) DeleteBy(ctx context.Context, cond WhereOption) *myerrors.AppError {
	var entity T

	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}
	var err error
	if hasAttribute(entity, "DeletedAt") {
		currentUser := ctx.Value(KEY_AUTH_USER).(map[string]interface{})

		if currentUser != nil {
			err = db.WithContext(ctx).Where(cond.Where, cond.Params...).
				Model(&entity).
				Updates(map[string]interface{}{
					"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
					"deleted_by": currentUser["id"],
				},
				).Error
		} else {
			err = db.WithContext(ctx).Where(cond.Where, cond.Params...).
				Model(&entity).
				Updates(map[string]interface{}{
					"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
				},
				).Error
		}
	} else {
		err = db.WithContext(ctx).Where(cond.Where, cond.Params...).Delete(&entity).Error
	}

	if err != nil {
		return myerrors.QueryInvalid(err.Error())
	}

	return nil
}

func (r *Repository[T]) FindById(ctx context.Context, id int) (*T, *myerrors.AppError) {
	var entity T

	cond := "id = ?"
	if hasAttribute(entity, "DeletedAt") {
		cond = "id = ? AND deleted_at IS NULL"
	}

	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	err := db.WithContext(ctx).Model(&entity).Where(cond, id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.QueryNotFound(err.Error())
		}
		return nil, myerrors.QueryInvalid(err.Error())
	}

	return &entity, nil
}

func (r *Repository[T]) FindAll(ctx context.Context) (*[]T, *myerrors.AppError) {
	var model T
	var entities []T
	var err error

	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	if hasAttribute(model, "DeletedAt") {
		err = db.WithContext(ctx).Where("deleted_at IS NULL").Find(&entities).Error
	} else {
		err = db.WithContext(ctx).Find(&entities).Error
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &entities, nil
		}
		return nil, myerrors.QueryInvalid(err.Error())
	}

	return &entities, nil
}

func (r *Repository[T]) Pagination(
	ctx context.Context,
	option *FindOption,
) (int, *[]T, *myerrors.AppError) {
	zeroItems := make([]T, 0)

	totalItems := r.CountBy(ctx, WhereOption{
		Where:  option.Where,
		Params: option.Params,
	})

	if totalItems == 0 || totalItems < ((option.Page-1)*option.Take) {
		return totalItems, &zeroItems, nil
	}

	entities, err := r.FindBy(ctx, option)

	return totalItems, entities, err
}

func (r *Repository[T]) PaginationQuery(ctx context.Context, option *PaginationQuery) (int, *[]interface{}, *myerrors.AppError) {
	var countValue int64

	err := r.QueryBuilder(ctx).Raw(option.CountSQL, option.Params...).First(&countValue).Error

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil, myerrors.QueryInvalid(err.Error())
		}
	}

	totalItems := int(countValue)
	take := option.Take
	offset := (option.Page - 1) * option.Take
	items := make([]interface{}, 0)

	if totalItems == 0 || totalItems < offset {
		return 0, &items, nil
	}

	rows, err := r.QueryBuilder(ctx).Raw(option.QuerySQL+fmt.Sprintf(" LIMIT %v OFFSET %v", take, offset), option.Params...).Rows()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, &items, nil
		}
		return 0, &items, myerrors.QueryInvalid(err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		it := option.ItemMappers(rows)
		items = append(items, it)
	}

	return totalItems, &items, nil
}

func (r *Repository[T]) FindOneBy(ctx context.Context, option *FindOption) (*T, *myerrors.AppError) {
	var entity T

	cond := option.Where
	if hasAttribute(entity, "DeletedAt") && !option.ExcludeDeleted {
		cond = strings.Join([]string{option.Where, " AND deleted_at IS NULL"}, " and ")
	}

	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	query := db.Debug().WithContext(ctx).Where(cond, option.Params...)

	if option.Order != "" {
		query = query.Order(option.Order)
	}

	if len(option.Preload) > 0 {
		for _, r := range option.Preload {
			query = query.Preload(r)
		}
	}

	if len(option.PreloadWithCond) > 0 {
		for _, cond := range option.PreloadWithCond {
			query = query.Preload(cond.Query, cond.Args...)
		}
	}

	err := query.First(&entity).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerrors.QueryNotFound(err.Error())
		}
		return nil, myerrors.QueryInvalid(err.Error())
	}

	return &entity, nil
}

func (r *Repository[T]) FindBy(ctx context.Context, option *FindOption) (*[]T, *myerrors.AppError) {
	var entities []T
	var model T

	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	query := db.WithContext(ctx)

	if option.Where != "" {
		cond := option.Where
		if hasAttribute(model, "DeletedAt") && !option.ExcludeDeleted {
			cond = strings.Join([]string{option.Where, " AND deleted_at IS NULL"}, " and ")
		}
		query = query.Where(cond, option.Params...)
	}

	if len(option.Joins) > 0 {
		for _, joinQuery := range option.Joins {
			query = query.Joins(joinQuery)
		}
	}

	if option.Take != 0 {
		query = query.Limit(option.Take)

		if option.Page != 0 {
			query = query.Offset((option.Page - 1) * option.Take)
		}
	}

	if option.Order != "" {
		query = query.Order(option.Order)
	}

	if len(option.Preload) > 0 {
		for _, r := range option.Preload {
			query = query.Preload(r)
		}
	}

	if len(option.PreloadWithCond) > 0 {
		for _, cond := range option.PreloadWithCond {
			query = query.Preload(cond.Query, cond.Args...)
		}
	}

	if option.Select != nil {
		query = query.Select(*option.Select)
	}

	err := query.Find(&entities).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &entities, nil
		}
		return nil, myerrors.QueryInvalid(err.Error())
	}

	return &entities, nil
}

func (r *Repository[T]) CountAll(ctx context.Context) int {
	var entity T
	var count int64

	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	if hasAttribute(entity, "DeletedAt") {
		db.WithContext(ctx).Model(&entity).Where("deleted_at IS NULL").Count(&count)
	} else {
		db.WithContext(ctx).Model(&entity).Count(&count)
	}

	db.WithContext(ctx).Model(&entity).Count(&count)

	return int(count)
}

func (r *Repository[T]) CountBy(ctx context.Context, cond WhereOption) int {
	var entity T
	var count int64

	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	query := db.WithContext(ctx).Model(&entity)

	if cond.Where != "" {
		where := cond.Where
		if hasAttribute(entity, "DeletedAt") {
			where = strings.Join([]string{cond.Where, " AND deleted_at IS NULL"}, " and ")
		}

		query = query.Where(where, cond.Params...)
	}

	query.Count(&count)

	return int(count)
}

func (r *Repository[T]) FirstOrInitBy(ctx context.Context, option FindOption, entity *T) (*T, *myerrors.AppError) {
	var model T

	query := r.db.WithContext(ctx)

	if option.Where != "" {
		cond := option.Where
		if hasAttribute(model, "DeletedAt") {
			cond = strings.Join([]string{option.Where, " AND deleted_at IS NULL"}, " and ")
		}
		query = query.Where(cond, option.Params...)
	}

	if len(option.Preload) > 0 {
		for _, r := range option.Preload {
			query = query.Preload(r)
		}
	}

	err := query.FirstOrInit(entity).Error

	if err != nil {
		return nil, myerrors.QueryInvalid(err.Error())
	}

	return entity, nil
}

func (r *Repository[T]) FirstOrCreateBy(ctx context.Context, option FindOption, entity *T) (*T, *myerrors.AppError) {
	var model T

	db := r.db

	if currentTran, ok := ctx.Value(KEY_CURRENT_TRAN).(*gorm.DB); ok {
		if currentTran != nil {
			db = currentTran
		}
	}

	query := db.WithContext(ctx)

	if option.Where != "" {
		cond := option.Where
		if hasAttribute(model, "DeletedAt") {
			cond = strings.Join([]string{option.Where, " AND deleted_at IS NULL"}, " and ")
		}
		query = query.Where(cond, option.Params...)
	}

	if len(option.Preload) > 0 {
		for _, r := range option.Preload {
			query = query.Preload(r)
		}
	}

	currentUser := ctx.Value(KEY_AUTH_USER).(map[string]interface{})

	if currentUser != nil {
		SetAttribute(entity, "CreatedAt", time.Now().UTC())
		SetAttribute(entity, "CreatedBy", currentUser["id"])
	}

	err := query.FirstOrCreate(entity).Error

	if err != nil {
		return nil, myerrors.QueryInvalid(err.Error())
	}

	return entity, nil
}

func (r *Repository[T]) GetExistsIdsByIds(ctx context.Context, ids []uint) (*[]uint, *myerrors.AppError) {
	var model T

	type IdOnly struct {
		ID uint
	}

	var selectedIds []IdOnly
	err := r.db.WithContext(ctx).Model(&model).Where("id IN ?", ids).Find(&selectedIds).Error

	if err != nil {
		return nil, myerrors.QueryInvalid(err.Error())
	}

	var existsIds []uint
	for _, it := range selectedIds {
		existsIds = append(existsIds, it.ID)
	}

	return &existsIds, nil
}

func (r *Repository[T]) IsExistById(ctx context.Context, id int) (*bool, *myerrors.AppError) {
	var entity T
	var exists bool

	db := r.db.WithContext(ctx).Model(&entity).Select("TRUE AS isExist").Where("id = ?", id)

	if hasAttribute(entity, "DeletedAt") {
		db = db.Where("deleted_at IS NULL")
	}

	err := db.Find(&exists).Error

	if err != nil {
		return nil, myerrors.QueryInvalid(err.Error())
	}

	return &exists, nil
}

func hasAttribute[T any](obj T, attributeName string) bool {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val.FieldByName(attributeName).IsValid()
}

func SetAttribute[T any](obj T, fieldName string, value interface{}) {
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	field := v.FieldByName(fieldName)

	if !field.IsValid() || !field.CanSet() {
		return
	}

	field.Set(reflect.ValueOf(value))
}
