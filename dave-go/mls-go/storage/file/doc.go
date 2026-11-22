// Package file stores MLS group state and leaf encryption keys in files.
//
// The store persists sensitive state on disk. Applications are responsible for
// protecting its storage location. Signature private keys are held only in
// memory and are lost when the process exits.
package file
