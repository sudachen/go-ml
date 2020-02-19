module github.com/sudachen/go-ml

go 1.13

require (
	github.com/getsentry/sentry-go v0.5.1
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/sudachen/go-fp v0.0.0-20200216225720-c33a3e35e157
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	gopkg.in/yaml.v3 v3.0.0-20200121175148-a6ecf24a6d71
	gotest.tools v2.2.0+incompatible
)

replace github.com/sudachen/go-fp => ./go-fp
