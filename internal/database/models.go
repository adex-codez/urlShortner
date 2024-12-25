package database

type Url struct {
	UniqueCode string `db:"unique_code"`

	LongUrl string `db:"long_url"`
}
