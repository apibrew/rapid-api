package data

type Record map[string]interface{}

type Interface interface {
	Init() error
	WriteRecord(path string, record Record) (Record, error)
	GetRecords(path string) ([]Record, bool, error)
	DeleteRecord(path string) error // record, isCollection, error
}
