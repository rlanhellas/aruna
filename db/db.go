package db

import (
	"context"
	"github.com/rlanhellas/aruna/domain"
	"github.com/rlanhellas/aruna/httpbridge"
	"github.com/rlanhellas/aruna/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"strings"
)

var client *gorm.DB

type Pageable struct {
	Content          any  `json:"content"`
	Page             int  `json:"page"`
	TotalPages       int  `json:"totalPages"`
	Last             bool `json:"last"`
	TotalElements    int  `json:"totalElements"`
	Size             int  `json:"size"`
	First            bool `json:"first"`
	NumberOfElements int  `json:"numberOfElements"`
	Empty            bool `json:"empty"`
}

// SetClient configure database client
func SetClient(c *gorm.DB) {
	client = c
}

// CreateWithBindHandlerHttp create entity and return the response to be used by HTTP handlers
func CreateWithBindHandlerHttp(ctx context.Context, domain domain.BaseDomain) *httpbridge.HandlerHttpResponse {
	result := Create(ctx, domain)
	return resolveHandlerResponse(result.Error, http.StatusCreated, domain)
}

// Create entity on db
func Create(ctx context.Context, domain domain.BaseDomain) *gorm.DB {
	logger.Debug(ctx, "creating entity[%s]: %+v", domain.TableName(), domain)
	return client.Create(domain)
}

// UpdateWithBindHandlerHttp update entity and return the response to be used by HTTP handlers
func UpdateWithBindHandlerHttp(ctx context.Context, domain domain.BaseDomain) *httpbridge.HandlerHttpResponse {
	result := Update(ctx, domain)
	if result.RowsAffected > 0 {
		return resolveHandlerResponse(nil, http.StatusOK, domain)
	} else {
		return resolveHandlerResponse(result.Error, http.StatusNotFound, nil)
	}
}

// UpdateSpecificAttributesWithBindHandlerHttp update entity and return the response to be used by HTTP handlers
func UpdateSpecificAttributesWithBindHandlerHttp(ctx context.Context, domain domain.BaseDomain, updateinformation map[string]interface{}) *httpbridge.HandlerHttpResponse {
	result := UpdateSpecificAttributes(ctx, domain, updateinformation)
	if result.RowsAffected > 0 {
		return resolveHandlerResponse(nil, http.StatusOK, domain)
	} else {
		return resolveHandlerResponse(result.Error, http.StatusNotFound, nil)
	}
}

// Update entity on db
func Update(ctx context.Context, domain domain.BaseDomain) *gorm.DB {
	logger.Debug(ctx, "updating entity[%s]: %+v", domain.TableName(), domain)
	r, exist := EntityExist(ctx, domain)
	if exist {
		return client.Save(domain)
	} else {
		return r
	}
}

// UpdateSpecificAttributes specific attributes on table db
func UpdateSpecificAttributes(ctx context.Context, domain domain.BaseDomain, updateinformation map[string]interface{}) *gorm.DB {
	logger.Debug(ctx, "updating attributes entity[%s]: %+v", domain.TableName(), domain)
	r, exist := EntityExist(ctx, domain)
	if exist {
		return client.Model(domain).Updates(updateinformation)
	} else {
		return r
	}
}

// GetByIdWithBindHandlerHttp return entity by id to be used by HTTP handlers
func GetByIdWithBindHandlerHttp(ctx context.Context, domain domain.BaseDomain, preload []string) *httpbridge.HandlerHttpResponse {
	r, e := GetById(ctx, domain, preload)
	if r.RowsAffected > 0 {
		return resolveHandlerResponse(nil, http.StatusOK, e)
	} else {
		return resolveHandlerResponse(r.Error, http.StatusNotFound, nil)
	}
}

// EntityExist check if entity exist in db searching by id
func EntityExist(ctx context.Context, domain domain.BaseDomain) (*gorm.DB, bool) {
	logger.Debug(ctx, "checking if entity[%s] %+v exists", domain.TableName(), domain)
	result := client.Find(domain.Clone())
	if result.RowsAffected > 0 {
		return result, true
	} else {
		return result, false
	}
}

// GetById get entity by id
func GetById(ctx context.Context, domain domain.BaseDomain, preload []string) (*gorm.DB, any) {
	logger.Debug(ctx, "getting entity[%s] %+v by id", domain.TableName(), domain)
	var result *gorm.DB
	if preload != nil && len(preload) > 0 {
		tx := client.Preload(preload[0])
		for i, p := range preload {
			if i == 0 {
				continue
			}
			tx = tx.Preload(p)
		}
		result = tx.Find(domain)
	} else {
		result = client.Preload(clause.Associations).Find(domain)
	}
	return result, domain
}

// GetSequenceId Get a ID sequence on database
func GetSequenceId(sequenceName string) uint64 {
	var nextval uint64
	client.Raw("SELECT nextval(?) as nextval;", sequenceName).Scan(&nextval)
	return nextval
}

// ExecSQL execute a native sql and store the result in dest variable
func ExecSQL(sql string, dest any, args ...any) {
	if args != nil {
		client.Raw(sql, args).Scan(dest)
	} else {
		client.Raw(sql).Scan(dest)
	}
}

// DeleteWithBindHandlerHttp delete entity and return the response to be used by HTTP handlers
func DeleteWithBindHandlerHttp(ctx context.Context, domain domain.BaseDomain) *httpbridge.HandlerHttpResponse {
	result := Delete(ctx, domain)
	if result.RowsAffected > 0 {
		return resolveHandlerResponse(nil, http.StatusOK, nil)
	} else {
		return resolveHandlerResponse(result.Error, http.StatusNotFound, nil)
	}
}

// Delete entity on db
func Delete(ctx context.Context, domain domain.BaseDomain) *gorm.DB {
	logger.Debug(ctx, "deleting entity[%s] %+v", domain.TableName(), domain)
	return client.Delete(domain)
}

// Delete entity with association on db
func DeleteWithAssociation(ctx context.Context, domain, target domain.BaseDomain, association string) error {
	logger.Debug(ctx, "deleting association[%s] in entity %+v", association, domain)
	return client.Model(domain).Association(association).Delete(target)
}

// ListWithBindHandlerHttp entities based on where and return response to be used by HTTP handlers
func ListWithBindHandlerHttp(ctx context.Context, where []string, orderBy string, whereArgs []string, pageSize, page int, domain domain.BaseDomain, results any) *httpbridge.HandlerHttpResponse {
	r, db := List(ctx, where, orderBy, whereArgs, pageSize, page, domain, results)
	if db != nil {
		if db.RowsAffected == 0 {
			return resolveHandlerResponse(db.Error, http.StatusNoContent, nil)
		}

		return resolveHandlerResponse(db.Error, http.StatusOK, r)
	} else {
		return resolveHandlerResponse(nil, http.StatusNoContent, nil)
	}

}

// List entities based on where
func List(ctx context.Context, where []string, orderBy string, whereArgs []string, pageSize, page int, domain domain.BaseDomain, results any) (*Pageable, *gorm.DB) {

	if page <= 0 {
		page = 1
	}

	logger.Debug(ctx, "listing based on where [%v], whereArgs[%v], pageSize[%d], page[%d], domain[%s]", where, whereArgs, pageSize, page, domain.TableName())
	TotalElements := int64(0)
	pageSize64 := int64(pageSize)

	txDB := client.Model(domain)

	if len(where) > 0 {
		for i, whereName := range where {
			txDB.Where(whereName, whereArgs[i])
		}
	}

	txDB.Count(&TotalElements)

	pages := TotalElements / pageSize64
	if (TotalElements % pageSize64) > 0 {
		pages++
	}

	isFirstPage := false
	if page == 1 {
		isFirstPage = true
	}

	if int64(page) > pages {
		return &Pageable{
			Page:          page,
			First:         isFirstPage,
			TotalPages:    int(pages),
			TotalElements: int(TotalElements),
		}, nil
	}

	offset := (page - 1) * pageSize

	txDB = client.Offset(offset).Limit(pageSize)

	if len(where) > 0 {
		for i, whereName := range where {
			txDB.Where(whereName, whereArgs[i]).Order(orderBy)
		}
	}

	if orderBy != "" {
		txDB.Order(orderBy)
	}

	dbResult := txDB.Find(&results)
	numberOfElements := pageSize
	last := false
	if page == int(pages) {
		last = true
		numberOfElements = int(TotalElements) - offset
	}

	return &Pageable{
		Content:          results,
		Page:             page,
		First:            isFirstPage,
		NumberOfElements: numberOfElements,
		TotalPages:       int(pages),
		Last:             last,
		TotalElements:    int(TotalElements),
		Size:             int(TotalElements),
	}, dbResult
}

func resolveHandlerResponse(err error, successStatus int, data any) *httpbridge.HandlerHttpResponse {
	if err != nil {
		statusCodeError := http.StatusInternalServerError
		if strings.Contains(err.Error(), "duplicate key") {
			statusCodeError = http.StatusConflict
		}
		return &httpbridge.HandlerHttpResponse{
			Error:      err,
			StatusCode: statusCodeError,
		}
	}
	return &httpbridge.HandlerHttpResponse{
		Data:       data,
		StatusCode: successStatus,
	}
}
