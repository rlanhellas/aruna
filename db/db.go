package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/rlanhellas/aruna/domain"
	"github.com/rlanhellas/aruna/httpbridge"
	"github.com/rlanhellas/aruna/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"reflect"
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
	if result.Error != nil {
		return resolveHandlerResponse(result.Error, http.StatusNotAcceptable, domain)
	}
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
	//logger.Debug(ctx, "getting entity[%s] %+v by id", domain.TableName(), domain)
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
func ListWithBindHandlerHttp(ctx context.Context, where []string, orderBy string, whereArgs []string, pageSize, page int, domain domain.BaseDomain, results any, join ...interface{}) *httpbridge.HandlerHttpResponse {

	var r *Pageable
	var db *gorm.DB
	var erroMenssage error
	var isEmptyJoin = false

	for _, v := range join {
		if v != nil {
			if reflect.ValueOf(v).IsZero() {
				fmt.Println("O valor na interface está vazio")
				isEmptyJoin = true
			}
		}
	}

	if isEmptyJoin {
		r, db, erroMenssage = List(ctx, where, orderBy, whereArgs, pageSize, page, domain, results, nil)
	} else {
		r, db, erroMenssage = List(ctx, where, orderBy, whereArgs, pageSize, page, domain, results, join)
	}

	if db != nil {
		if db.RowsAffected == 0 {
			return resolveHandlerResponse(db.Error, http.StatusNoContent, nil)
		}
		return resolveHandlerResponse(db.Error, http.StatusOK, r)
	} else {
		return resolveHandlerResponse(erroMenssage, http.StatusOK, r)
	}

}

// List entities based on where
func List(ctx context.Context, where []string, orderBy string, whereArgs []string, pageSize, page int, domain domain.BaseDomain, results any, join []interface{}) (*Pageable, *gorm.DB, error) {

	if page <= 0 {
		page = 1
	}
	logger.Debug(ctx, "listing based on where [%v], whereArgs[%v], pageSize[%d], page[%d], domain[%s]", where, whereArgs, pageSize, page, domain.TableName())

	totalElements := int64(0)
	pageSize64 := int64(pageSize)
	txDB := client.Model(domain)

	if len(where) > 0 {
		for i, whereName := range where {
			txDB.Where(whereName, whereArgs[i])
		}
	}

	txDB.Count(&totalElements)

	logger.Debug(ctx, "list return case nil totalElements[%d]  pageSize[%d]", totalElements, pageSize64)
	var emptyAny interface{} = make([]interface{}, 0)

	if totalElements == 0 { // Ver se não vai dar algum tipo de bug aqui¹
		return &Pageable{
			Content: emptyAny,
			Empty:   true,
		}, nil, errors.New("Nothing was found with the search term!")
	}

	pages := totalElements / pageSize64
	if (totalElements % pageSize64) > 0 {
		pages++
	}

	isFirstPage := false
	if page == 1 {
		isFirstPage = true
	}
	if int64(page) > pages {
		return &Pageable{
			Content:          results,
			NumberOfElements: 0,
			Last:             false,
			Size:             int(pageSize64),
			Page:             page,
			First:            isFirstPage,
			Empty:            true,
			TotalPages:       int(pages),
			TotalElements:    int(totalElements),
		}, nil, errors.New("Pagination out of correct range.")
	}
	offset := (page - 1) * pageSize
	txDB = client.Offset(offset).Limit(pageSize)
	if len(where) > 0 {
		for i, whereName := range where {
			txDB.Where(whereName, whereArgs[i])
		}
	}

	if orderBy != "" {
		txDB.Order(orderBy)
	}

	if join != nil {
		fmt.Println("Realizar o processamento do join")
	} else {
		fmt.Println("Não existem valores a ser usado o Join")
	}

	dbResult := txDB.Find(&results)

	numberOfElements := pageSize
	last := false
	if page == int(pages) {
		last = true
		numberOfElements = int(totalElements) - offset
	}

	return &Pageable{
		Content:          results,
		Page:             page,
		First:            isFirstPage,
		Empty:            false,
		NumberOfElements: numberOfElements,
		TotalPages:       int(pages),
		Last:             last,
		TotalElements:    int(totalElements),
		Size:             int(pageSize64),
	}, dbResult, nil
}

func resolveHandlerResponse(err error, successStatus int, data any) *httpbridge.HandlerHttpResponse {

	if err != nil {
		if successStatus >= 200 && successStatus < 300 {
			return &httpbridge.HandlerHttpResponse{
				Error:      err,
				Data:       data,
				StatusCode: successStatus,
			}
		}
		if successStatus == 406 {
			return &httpbridge.HandlerHttpResponse{
				Error:      err,
				Data:       data,
				StatusCode: successStatus,
			}
		}
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
