package repository

import "errors"

var ErrBlockNotFound = errors.New("block not found")
var ErrLCANotFound = errors.New("LCA not found")
