# builder-backend
ILLA Builder Backend

# Development 
## Prerequisite
```
go get github.com/google/wire/cmd/wire
go install github.com/google/wire/cmd/wire@latest
```

if still not working, please check your GOBIN settings.

## After Dev
Add new wire set to `/cmd/http-server/wire.go` if necessary

## Before Compile
run ```wire``` command in dir `/cmd/http-server`
