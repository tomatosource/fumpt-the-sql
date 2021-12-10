# fumpt-the-sql

Opionated sql formatter for use with `.go` files containing backticked queries.

Uses https://sqlformat.darold.net/ for the actual sql formtatting, formatter documentation here: https://github.com/darold/pgFormatter.

Arguments set:

- `--comma-break`
- `--tabs`

## Installation

```
make
```

## Usage

From root of directory containing `.go` files with queries.

```
docker run --rm --name sqlfumpt_run --mount type=bind,source="$(pwd)",target=/repo sqlfumpt
```

Example Make target

```
.PHONY: formatsql
formatsql:
	docker run --rm --name sqlfumpt_run --mount type=bind,source=`pwd`,target=/repo sqlfumpt
```

## Notes

Queries must be in backticks.

Queries must be an argument to one of the following functions:

```
Get
Select
Exec
NamedExec
NamedQuery
Query
Prepare
GetContext
SelectContext
ExecContext
NamedExecContext
QueryContext
PrepareContext
PrepareNamedContext
```

Vendor directory will be ignored.

## TODO

- Sort out the before/after backticks
