// Package http provides HTTP client utilities for interacting with the Discord API.
// It includes features like automatic retries with exponential backoff and custom transports.
package http

// Version is the library version. It is included in the default User-Agent
// header sent with REST requests so Discord can identify the library.
const Version = "0.11.7-coffee"
