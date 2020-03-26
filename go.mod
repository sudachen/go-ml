module github.com/sudachen/go-ml

go 1.13

replace github.com/sudachen/go-foo => ./go-foo

require (
	github.com/getsentry/sentry-go v0.5.1
	github.com/gomarkdown/markdown v0.0.0-20200316172748-fd1f3374857d
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/sudachen/go-dl v1.0.1
	github.com/sudachen/go-foo v0.0.0-00010101000000-000000000000
	github.com/ulikunitz/xz v0.5.7
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	gotest.tools v2.2.0+incompatible
)
