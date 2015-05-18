package empire

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// Scope is an interface that scopes a gorm.DB. Scopes are used in
// ThingsFirst and ThingsAll methods on the store for filtering/querying.
type Scope interface {
	Scope(*gorm.DB) *gorm.DB
}

// ScopeFunc implements the Scope interface for functions.
type ScopeFunc func(*gorm.DB) *gorm.DB

// Scope implements the Scope interface.
func (f ScopeFunc) Scope(db *gorm.DB) *gorm.DB {
	return f(db)
}

// ComposedScope is an implementation of the Scope interface that chains the
// scopes together.
type ComposedScope []Scope

// Scope implements the Scope interface.
func (s ComposedScope) Scope(db *gorm.DB) *gorm.DB {
	for _, s := range s {
		db = s.Scope(db)
	}

	return db
}

// FieldEquals returns a Scope that filters on a field.
func FieldEquals(field string, v interface{}) Scope {
	return ScopeFunc(func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s = ?", field), v)
	})
}

// Preload returns a Scope that preloads the associations.
func Preload(associations ...string) Scope {
	var scope ComposedScope

	for _, a := range associations {
		aa := a
		scope = append(scope, ScopeFunc(func(db *gorm.DB) *gorm.DB {
			return db.Preload(aa)
		}))
	}

	return scope
}

// Order returns a Scope that orders the results.
func Order(order string) Scope {
	return ScopeFunc(func(db *gorm.DB) *gorm.DB {
		return db.Order(order)
	})
}

// store provides methods for CRUD'ing things.
type store struct {
	db *gorm.DB
}

// Scope applies the scope to the gorm.DB.
func (s *store) Scope(scope Scope) *gorm.DB {
	return scope.Scope(s.db)
}

// First applies the scope to the gorm.DB and finds the first record, populating
// v.
func (s *store) First(scope Scope, v interface{}) error {
	return s.Scope(scope).First(v).Error
}

// Find applies the scope to the gorm.DB and finds the matching records,
// populating v.
func (s *store) Find(scope Scope, v interface{}) error {
	return s.Scope(scope).Find(v).Error
}

func (s *store) Reset() error {
	var err error
	exec := func(sql string) {
		if err == nil {
			err = s.db.Exec(sql).Error
		}
	}

	exec(`TRUNCATE TABLE apps CASCADE`)
	exec(`TRUNCATE TABLE ports CASCADE`)
	exec(`INSERT INTO ports (port) (SELECT generate_series(9000,10000))`)

	return err
}

func (s *store) IsHealthy() bool {
	return s.db.DB().Ping() == nil
}
