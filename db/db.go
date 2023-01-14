package db

import (
	"context"
	"github.com/rlanhellas/aruna/domain"
	"github.com/rlanhellas/aruna/httpbridge"
	"github.com/rlanhellas/aruna/logger"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

var client *gorm.DB

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
func GetByIdWithBindHandlerHttp(ctx context.Context, domain domain.BaseDomain) *httpbridge.HandlerHttpResponse {
	r, e := GetById(ctx, domain)
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
func GetById(ctx context.Context, domain domain.BaseDomain) (*gorm.DB, any) {
	logger.Debug(ctx, "getting entity[%s] %+v by id", domain.TableName(), domain)
	result := client.Find(domain)
	return result, domain
}

// GetSequenceId Get a ID sequence on database
func GetSequenceId(sequenceName string) uint64 {
	var nextval uint64
	client.Raw("SELECT nextval('?') as nextval;", sequenceName).Scan(&nextval)
	return nextval
}

// ExecSQL execute a native sql and store the result in dest variable
func ExecSQL(sql string, dest any, args ...string) {
	client.Raw(sql, args).Scan(dest)
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
