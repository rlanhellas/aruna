package db

import (
	"context"
	"net/http"
	"strings"

	"github.com/rlanhellas/aruna/domain"
	"github.com/rlanhellas/aruna/httpbridge"
	"github.com/rlanhellas/aruna/logger"
	"gorm.io/gorm"
)

var client *gorm.DB

type Pageable struct {
	Results     any
	CurrentPage int
	CurrentRows int
	TotalPages  int
	TotalRows   int
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
		return resolveHandlerResponse(result.Error, http.StatusOK, domain)
	} else {
		return resolveHandlerResponse(nil, http.StatusNotFound, nil)
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

// GetByIdWithBindHandlerHttp return entity by id to be used by HTTP handlers
func GetByIdWithBindHandlerHttp(ctx context.Context, domain domain.BaseDomain, preload []string) *httpbridge.HandlerHttpResponse {
	r, e := GetById(ctx, domain, preload)
	if r.RowsAffected > 0 {
		return resolveHandlerResponse(r.Error, http.StatusOK, e)
	} else {
		return resolveHandlerResponse(nil, http.StatusNotFound, nil)
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
		result = client.Find(domain)
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
		return resolveHandlerResponse(result.Error, http.StatusOK, nil)
	} else {
		return resolveHandlerResponse(result.Error, http.StatusNotFound, nil)
	}
}

// Delete entity on db
func Delete(ctx context.Context, domain domain.BaseDomain) *gorm.DB {
	logger.Debug(ctx, "deleting entity[%s] %+v", domain.TableName(), domain)
	return client.Delete(domain)
}

// ListWithBindHandlerHttp entities based on where and return response to be used by HTTP handlers
func ListWithBindHandlerHttp(ctx context.Context, where string, whereArgs []string, pageSize int, page int, domain domain.BaseDomain, results any) *httpbridge.HandlerHttpResponse {
	r, db := List(ctx, where, whereArgs, pageSize, page, domain, results)
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
func List(ctx context.Context, where string, whereArgs []string, pageSize int, page int, domain domain.BaseDomain, results any) (*Pageable, *gorm.DB) {

	if page <= 0 {
		page = 1
	}

	logger.Debug(ctx, "listing based on where [%v], whereArgs[%v], pageSize[%d], page[%d], domain[%s]", where, whereArgs, pageSize, page, domain.TableName())
	totalRows := int64(0)
	pageSize64 := int64(pageSize)

	txDB := client.Model(domain)

	if where != "" {
		txDB.Where(where, whereArgs)
	}

	txDB.Count(&totalRows)

	pages := totalRows / pageSize64
	if (totalRows % pageSize64) > 0 {
		pages++
	}

	if int64(page) > pages {
		return &Pageable{
			CurrentRows: 0,
			CurrentPage: page,
			TotalPages:  int(pages),
			TotalRows:   int(totalRows),
		}, nil
	}

	offset := (page - 1) * pageSize

	txDB = client.Offset(offset).Limit(pageSize)
	if where != "" {
		txDB.Where(where, whereArgs)
	}

	dbResult := txDB.Find(&results)

	return &Pageable{
		CurrentRows: int(dbResult.RowsAffected),
		Results:     results,
		CurrentPage: page,
		TotalPages:  int(pages),
		TotalRows:   int(totalRows),
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
