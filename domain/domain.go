package domain

type BaseDomain interface {
	TableName() string
	Clone() any
}
