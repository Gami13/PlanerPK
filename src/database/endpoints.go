package database

import (
	"fmt"
	"net/http"
)

func (ec *EndpointContext) AuthorizeEndpoint(w http.ResponseWriter, r *http.Request) {

	fmt.Println("UwU")
	w.Write([]byte("UwU"))
}
