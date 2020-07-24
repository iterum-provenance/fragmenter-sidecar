// Package data contains two main concepts:
// ```
// 		1. The data.Mover goroutine, responsible for downloading files from the Daemon and uploading them to MinIO
// 		2. The data.Tracker goroutine, responsible for tracking file uploads and subfragments generating, combining these into desc.RemoteFragmentDesc
// ```
// The mover uploads files and posts Upload structures for the Tracker
// The generated Subfragment structures from the fragmenter are send to the tracker by the handler package
// The Tracker consumes both of these to perform its task
package data
