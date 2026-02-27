package gen

//go:generate go tool github.com/go-swagger/go-swagger/cmd/swagger generate server -m models -s restapi --exclude-main --name Stork --regenerate-configureapi --spec ../../../api/swagger.yaml --template stratoscale
