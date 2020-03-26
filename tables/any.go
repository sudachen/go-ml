package tables

/*
AnyData represents both tables and lazy-streams as common data source type.
Any one of them can be used as Table or Lazy stream via this interface.
Although, if it's matter, user can decided to use real form of the data by IsLazy selector.
*/
type AnyData interface {
	// IsLazy specifies is it lazy datasource or a table
	IsLazy() bool
	// Use it as lazy datasource
	Lazy() Lazy
	// Use it as a table, if it's really lazy and collect returns an error panic will occur
	Table() *Table
	// Use it as a table, if it's a lazy data source it will be collected to a table
	Collect() (*Table, error)
}
