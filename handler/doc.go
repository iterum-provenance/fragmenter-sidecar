// Package handler contains the handler functions for the connection with the user-defined fragmenter
// It has a producer and consumer function.
//
// The handler in producer.go sends over the list of files, followed by a kill_message
// The handler in consumer.go consumes messages of the Subfragment type and passes these on
// Fragmenter input is the input structure used by producer.go
package handler
