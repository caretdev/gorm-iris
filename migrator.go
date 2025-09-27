package iris

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

type Migrator struct {
	migrator.Migrator
	Dialector
}

var defaultSchema = "SQLUser"

func (m Migrator) queryRaw(sql string, values ...interface{}) (tx *gorm.DB) {
	queryTx := m.DB
	if m.DB.DryRun {
		queryTx = m.DB.Session(&gorm.Session{})
		queryTx.DryRun = false
	}
	// log.Println("queryRaw:", m, sql)
	return queryTx.Raw(sql, values...)
}

// AddColumn implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).AddColumn of Migrator.Migrator.
func (m Migrator) AddColumn(dst interface{}, field string) error {
	return m.Migrator.AddColumn(dst, field)
}

// AlterColumn implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).AlterColumn of Migrator.Migrator.
func (m Migrator) AlterColumn(dst interface{}, field string) error {
	return m.Migrator.AlterColumn(dst, field)
}

// ColumnTypes implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).ColumnTypes of Migrator.Migrator.
func (m Migrator) ColumnTypes(dst interface{}) ([]gorm.ColumnType, error) {
	return m.Migrator.ColumnTypes(dst)
}

// CreateConstraint implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).CreateConstraint of Migrator.Migrator.
func (m Migrator) CreateConstraint(dst interface{}, name string) error {
	return m.Migrator.CreateConstraint(dst, name)
}

// CreateIndex implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).CreateIndex of Migrator.Migrator.
func (m Migrator) CreateIndex(dst interface{}, name string) error {
	return m.Migrator.CreateIndex(dst, name)
}

// CreateTable implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).CreateTable of Migrator.Migrator.
func (m Migrator) CreateTable(values ...interface{}) error {
	return m.Migrator.CreateTable(values...)
}

// CreateView implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).CreateView of Migrator.Migrator.
func (m Migrator) CreateView(name string, option gorm.ViewOption) error {
	return m.Migrator.CreateView(name, option)
}

// CurrentDatabase implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).CurrentDatabase of Migrator.Migrator.
func (m Migrator) CurrentDatabase() string {
	return ""
}

// DropColumn implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).DropColumn of Migrator.Migrator.
func (m Migrator) DropColumn(dst interface{}, field string) error {
	return m.Migrator.DropColumn(dst, field)
}

// DropConstraint implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).DropConstraint of Migrator.Migrator.
func (m Migrator) DropConstraint(dst interface{}, name string) error {
	return m.Migrator.DropConstraint(dst, name)
}

// DropIndex implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).DropIndex of Migrator.Migrator.
func (m Migrator) DropIndex(dst interface{}, name string) error {
	return m.Migrator.DropIndex(dst, name)
}

// DropTable implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).DropTable of Migrator.Migrator.
func (m Migrator) DropTable(values ...interface{}) error {
	values = m.ReorderModels(values, false)
	tx := m.DB.Session(&gorm.Session{})
	for i := len(values) - 1; i >= 0; i-- {
		if err := m.RunWithValue(values[i], func(stmt *gorm.Statement) error {
			return tx.Exec("DROP TABLE IF EXISTS ? CASCADE", m.CurrentTable(stmt)).Error
		}); err != nil {
			return err
		}
	}
	return nil

}

// DropView implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).DropView of Migrator.Migrator.
func (m Migrator) DropView(name string) error {
	return m.Migrator.DropView(name)
}

// GetIndexes implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).GetIndexes of Migrator.Migrator.
func (m Migrator) GetIndexes(dst interface{}) ([]gorm.Index, error) {
	return m.Migrator.GetIndexes(dst)
}

// GetTables implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).GetTables of Migrator.Migrator.
func (m Migrator) GetTables() (tableList []string, err error) {
	return m.Migrator.GetTables()
}

// GetTypeAliases implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).GetTypeAliases of Migrator.Migrator.
func (m Migrator) GetTypeAliases(databaseTypeName string) []string {
	return m.Migrator.GetTypeAliases(databaseTypeName)
}

// HasColumn implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).HasColumn of Migrator.Migrator.
func (m Migrator) HasColumn(dst interface{}, field string) bool {
	var count int64
	m.RunWithValue(dst, func(stmt *gorm.Statement) error {
		currentSchema, currentTable := m.CurrentSchema(stmt, stmt.Table)
		name := field
		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(field); field != nil {
				name = field.DBName
			}
		}

		return m.DB.Raw(
			"SELECT count(*) FROM INFORMATION_SCHEMA.columns WHERE table_schema = ? AND table_name = ? AND column_name = ?",
			currentSchema, currentTable, name,
		).Row().Scan(&count)
	})

	return count > 0}

// HasConstraint implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).HasConstraint of Migrator.Migrator.
func (m Migrator) HasConstraint(dst interface{}, name string) bool {
	return m.Migrator.HasConstraint(dst, name)
}

// HasIndex implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).HasIndex of Migrator.Migrator.
func (m Migrator) HasIndex(dst interface{}, name string) bool {
	return m.Migrator.HasIndex(dst, name)
}

func (m Migrator) CurrentSchema(stmt *gorm.Statement, table string) (interface{}, interface{}) {
	if strings.Contains(table, ".") {
		if tables := strings.Split(table, `.`); len(tables) == 2 {
			return tables[0], tables[1]
		}
	}

	if stmt.TableExpr != nil {
		if tables := strings.Split(stmt.TableExpr.SQL, `"."`); len(tables) == 2 {
			return strings.TrimPrefix(tables[0], `"`), table
		}
	}
	return defaultSchema, table
}

// HasTable implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).HasTable of Migrator.Migrator.
func (m Migrator) HasTable(value interface{}) bool {
	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		currentSchema, curTable := m.CurrentSchema(stmt, stmt.Table)
		return m.queryRaw("SELECT count(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ? AND table_type = ?", currentSchema, curTable, "BASE TABLE").Row().Scan(&count)
	})
	return count > 0

}

// MigrateColumn implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).MigrateColumn of Migrator.Migrator.
func (m Migrator) MigrateColumn(dst interface{}, field *schema.Field, columnType gorm.ColumnType) error {
	return m.Migrator.MigrateColumn(dst, field, columnType)
}

// MigrateColumnUnique implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).MigrateColumnUnique of Migrator.Migrator.
func (m Migrator) MigrateColumnUnique(dst interface{}, field *schema.Field, columnType gorm.ColumnType) error {
	return m.Migrator.MigrateColumnUnique(dst, field, columnType)
}

// RenameColumn implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).RenameColumn of Migrator.Migrator.
func (m Migrator) RenameColumn(dst interface{}, oldName string, field string) error {
	return m.Migrator.RenameColumn(dst, oldName, field)
}

// RenameIndex implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).RenameIndex of Migrator.Migrator.
func (m Migrator) RenameIndex(dst interface{}, oldName string, newName string) error {
	return m.Migrator.RenameIndex(dst, oldName, newName)
}

// RenameTable implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).RenameTable of Migrator.Migrator.
func (m Migrator) RenameTable(oldName interface{}, newName interface{}) error {
	return m.Migrator.RenameTable(oldName, newName)
}

// TableType implements gorm.Migrator.
// Subtle: this method shadows the method (Migrator).TableType of Migrator.Migrator.
func (m Migrator) TableType(dst interface{}) (gorm.TableType, error) {
	return m.Migrator.TableType(dst)
}
