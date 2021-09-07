// Copyright 2021 CloudJ Company Limited. All rights reserved.

package types

import "cloudiac/portal/libs/db"

type MigrateFunc func(tx *db.Session) error

type Migration struct {
	Version string
	Actions []MigrateFunc
}
