package dbtype

type Model interface {
	ModelID() string
	TableName() string
} 