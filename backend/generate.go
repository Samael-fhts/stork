//go:build ignore

package main

//go:generate go tool swagger generate server -m server/gen/models -s server/gen/restapi --exclude-main --name Stork --regenerate-configureapi --spec ../api/swagger.yaml --template stratoscale
