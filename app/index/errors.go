package index

import "fmt"

type BadRequest struct {
	URL string
}

type MethodNotAllowed struct {
	URL string
}

type UnProcessableEntity struct {
	CAUSE string
}

type NotFound struct {
	ID string
}

func (e BadRequest) Error() string {
	return fmt.Sprintf("Bad Request: %v", e.URL)
}

func (e MethodNotAllowed) Error() string {
	return fmt.Sprintf("Method not allowed for request: %v", e.URL)
}

func (e UnProcessableEntity) Error() string {
	return fmt.Sprintf("Request could not be processed: %v", e.CAUSE)
}

func (e NotFound) Error() string {
	return fmt.Sprintf("Item not in index: %v", e.ID)
}