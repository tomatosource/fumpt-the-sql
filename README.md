# fumpt-the-sql

Opionated sql formatter.

Uses https://sqlformat.darold.net/ as formatter.

## Usage

1. `make`
2. `docker run --rm --name sqlfumpt_run --mount type=bind,source="$(pwd)",target=/repo sqlfumpt` in repository to be formatted root

Example Make target

```
.PHONY: formatsql
formatsql:
	docker run --rm --name sqlfumpt_run --mount type=bind,source=`pwd`,target=/repo sqlfumpt
```

## TODO

- Sort out the before/after backticks
- Add the command to repo `Makefile`s
