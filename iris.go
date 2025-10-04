package iris

import (
	"database/sql"
	"fmt"

	_ "github.com/caretdev/go-irisnative"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	. "gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

const (
	DefaultDriverName = "iris"
)

var (
	currentTable = Table{Name: CurrentTable}
)

type Config struct {
	DriverName    string
	ServerVersion string
	DSN           string
	Conn          gorm.ConnPool
}

type Dialector struct {
	*Config
}

// QuoteTo implements gorm.Dialector.
func (dialector Dialector) QuoteTo(writer Writer, str string) {
	var (
		underQuoted, selfQuoted bool
		continuousBacktick      int8
		shiftDelimiter          int8
	)

	for _, v := range []byte(str) {
		switch v {
		case '"':
			continuousBacktick++
			if continuousBacktick == 2 {
				writer.WriteString(`""`)
				continuousBacktick = 0
			}
		case '.':
			if continuousBacktick > 0 || !selfQuoted {
				shiftDelimiter = 0
				underQuoted = false
				continuousBacktick = 0
				writer.WriteByte('"')
			}
			writer.WriteByte(v)
			continue
		default:
			if shiftDelimiter-continuousBacktick <= 0 && !underQuoted {
				writer.WriteByte('"')
				underQuoted = true
				if selfQuoted = continuousBacktick > 0; selfQuoted {
					continuousBacktick -= 1
				}
			}

			for ; continuousBacktick > 0; continuousBacktick -= 1 {
				writer.WriteString(`""`)
			}

			writer.WriteByte(v)
		}
		shiftDelimiter++
	}

	if continuousBacktick > 0 && !selfQuoted {
		writer.WriteString(`""`)
	}
	writer.WriteByte('"')
}

func Open(dsn string) gorm.Dialector {
	return &Dialector{&Config{DSN: dsn}}
}

func New(config Config) gorm.Dialector {
	return &Dialector{Config: &config}
}

func (dialector Dialector) Name() string {
	return "iris"
}

func (dialector Dialector) BindVarTo(writer Writer, stmt *gorm.Statement, v interface{}) {
	writer.WriteByte('?')
}

func (dialector Dialector) DataTypeOf(field *schema.Field) string {
	switch field.DataType {
	case schema.Bool:
		return "BIT"
	case schema.String:
		if field.Size > 0 {
			return fmt.Sprintf("varchar(%d)", field.Size)
		}
		return "varchar(65535)"
	case schema.Int, schema.Uint:
		size := field.Size
		if field.DataType == schema.Uint {
			size++
		}
		if field.AutoIncrement {
			// WITH %CLASSPARAMETER ALLOWIDENTITYINSERT = 1
			return "IDENTITY"
		} else {
			switch {
			case size <= 16:
				return "smallint"
			case size <= 32:
				return "integer"
			default:
				return "bigint"
			}
		}
	case schema.Float:
		if field.Precision > 0 {
			if field.Scale > 0 {
				return fmt.Sprintf("numeric(%d, %d)", field.Precision, field.Scale)
			}
			return fmt.Sprintf("numeric(%d)", field.Precision)
		}
		return "decimal"
	case schema.Time:
		return "timestamp"
	case schema.Bytes:
		return "BINARY(65535)"
	default:
		return dialector.getSchemaCustomType(field)
		// panic("unimplemented: DataTypeOf for " + field.DataType)
		// return "DEMO"
	}
}

func (dialector Dialector) getSchemaCustomType(field *schema.Field) string {
	sqlType := string(field.DataType)

	// if field.AutoIncrement && !strings.Contains(strings.ToLower(sqlType), " auto_increment") {
	// 	sqlType += " AUTO_INCREMENT"
	// }

	return sqlType
}

func (dialector Dialector) DefaultValueOf(field *schema.Field) Expression {
	return nil
	// return Expr{SQL: "DEFAULT"}
}

func (dialector Dialector) Apply(config *gorm.Config) error {
	return nil
}

func (dialector Dialector) Initialize(db *gorm.DB) (err error) {
	// log.Printf("initialize: %#v; %#v\n", db, dialector)
	if dialector.DriverName == "" {
		dialector.DriverName = DefaultDriverName
	}
	callbackConfig := &callbacks.Config{
		CreateClauses:        []string{"INSERT", "VALUES", "ON CONFLICT"},
		UpdateClauses:        []string{"UPDATE", "SET", "FROM", "WHERE"},
		DeleteClauses:        []string{"DELETE", "FROM", "WHERE"},
		LastInsertIDReversed: true,
	}
	callbacks.RegisterDefaultCallbacks(db, callbackConfig)

	for k, v := range dialector.ClauseBuilders() {
		if _, ok := db.ClauseBuilders[k]; !ok {
			db.ClauseBuilders[k] = v
		}
	}

	db.ConnPool, err = sql.Open(dialector.DriverName, dialector.Config.DSN)
	if err != nil {
		panic(err)
	}
	// db.Set("gorm:table_options", " WITH %CLASSPARAMETER ALLOWIDENTITYINSERT = 1")
	return
}

type IRISDB struct {
	*gorm.DB
}

func (irisdb *IRISDB) CreateInBatches(value interface{}, batchSize int) (tx *gorm.DB) {
	return irisdb.DB.CreateInBatches(value, 1)
}

func (dialector Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return Migrator{
		Migrator: migrator.Migrator{
			Config: migrator.Config{
				DB:                          db.Set("gorm:table_options", " WITH %CLASSPARAMETER ALLOWIDENTITYINSERT = 1"),
				Dialector:                   dialector,
				CreateIndexAfterCreateTable: true,
			},
		},
		Dialector: dialector,
	}
}

func (dialector Dialector) Explain(sql string, avars ...interface{}) string {
	var (
		convertParams func(interface{}, int)
		vars          = make([]interface{}, len(avars))
	)

	convertParams = func(v interface{}, idx int) {
		switch v := v.(type) {
		case bool:
			if v {
				vars[idx] = "1"
			} else {
				vars[idx] = "0"
			}
		default:
			vars[idx] = v
		}
	}
	for idx, v := range avars {
		convertParams(v, idx)
	}

	return logger.ExplainSQL(sql, nil, `'`, vars...)
}

type DopeWriter struct {
}

func (stmt DopeWriter) WriteString(string) (int, error) {
	return 0, nil
}

func (stmt DopeWriter) WriteByte(c byte) error {
	return nil
}

func (dialector Dialector) ClauseBuilders() map[string]ClauseBuilder {
	clauseBuilders := map[string]ClauseBuilder{
		"WHERE": func(c Clause, builder Builder) {
			if where, ok := c.Expression.(Where); ok && len(where.Exprs) > 0 {
				var exprs = where.Exprs
				for idx, expr := range where.Exprs {
					// Replace multicolumn IN with AND
					if in, ok := expr.(IN); ok {
						if columns, ok := in.Column.([]Column); ok {
							var values = in.Values[0].([]interface{})
							if len(columns) == len(values) {
								var newExprs = make([]Expression, len(columns))
								for i := 0; i < len(columns); i++ {
									newExprs[i] = Eq{
										Column: columns[i],
										Value: values[i],
									}
								}
								exprs[idx] = AndConditions{
									Exprs: newExprs,
								}
							}
						}
					}
				}
			}
			c.Build(builder)
		},
		"INSERT": func(c Clause, builder Builder) {
			if insert, ok := c.Expression.(Insert); ok {
				builder.WriteString("INSERT OR UPDATE ")
				if insert.Table.Name == "" {
					builder.WriteQuoted(currentTable)
				} else {
					builder.WriteQuoted(insert.Table)
				}
			}
		},
		"VALUES": func(c Clause, builder Builder) {
			var dopeWriter = DopeWriter{}
			if values, ok := c.Expression.(Values); ok {
				// if len(values.Values) > 1 {
				// 	panic(fmt.Sprintf("Create in batches not supported by IRIS: %d;", len(values.Values)))
				// }
				if len(values.Columns) == 0 {
					builder.WriteString("DEFAULT VALUES")
					return
				}
				builder.WriteByte('(')
				for idx, column := range values.Columns {
					if idx > 0 {
						builder.WriteByte(',')
					}
					builder.WriteQuoted(column)
				}
				builder.WriteByte(')')
				builder.WriteString(" VALUES ")
				for idx, value := range values.Values {
					if idx > 0 {
						builder.AddVar(dopeWriter, value...)
					} else {
						builder.WriteByte('(')
						builder.AddVar(builder, value...)
						builder.WriteByte(')')
					}
				}
				return
			}
			c.Build(builder)
		},
		"GROUP BY": func(c Clause, builder Builder) {
			if groupBy, ok := c.Expression.(GroupBy); ok {
				builder.WriteString("GROUP BY ")
				for idx, column := range groupBy.Columns {
					if idx > 0 {
						builder.WriteByte(',')
					}
					// Trick to keep original case for groupped column
					builder.WriteString(`''||`)
					builder.WriteQuoted(column)
				}

				if len(groupBy.Having) > 0 {
					builder.WriteString(" HAVING ")
					Where{Exprs: groupBy.Having}.Build(builder)
				}
				return
			}
			c.Build(builder)
		},
		"SELECT": func(c Clause, builder Builder) {
			if s, ok := c.Expression.(Select); ok {
				builder.WriteString("SELECT ")
				if len(s.Columns) > 0 {
					if s.Distinct {
						builder.WriteString("DISTINCT ")
					}

					for idx, column := range s.Columns {
						if idx > 0 {
							builder.WriteByte(',')
						}
						if s.Distinct {
							builder.WriteString("%EXACT(")
							builder.WriteQuoted(column)
							builder.WriteString(") AS ")
						}
						builder.WriteQuoted(column)
					}
				} else {
					builder.WriteByte('*')
				}

				return
			}
			c.Build(builder)
		},
		"ON CONFLICT": func(c Clause, builder Builder) {
			if onConflict, ok := c.Expression.(OnConflict); ok {
				// Some tricks to make it working with IRIS
				if onConflict.DoNothing {
					builder.WriteString(";\n-- ON CONFLICT DO NOTHING")
					return
				} else {
					builder.WriteString(";\n-- ON CONFLICT UPDATE")
					return
				}
			}
		},
		"RETURNING": func(c Clause, builder Builder) {
			if returning, ok := c.Expression.(Returning); ok {
				builder.WriteString(";\nSELECT ")
				var table string
				if len(returning.Columns) > 0 {
					for idx, column := range returning.Columns {
						table = column.Table
						if idx > 0 {
							builder.WriteByte(',')
						}

						builder.WriteQuoted(column)
					}
				} else {
					builder.WriteByte('*')
				}
				builder.WriteString(" FROM ")
				if table == "" {
					builder.WriteQuoted(currentTable)
				} else {
					builder.WriteQuoted(table)
				}
				builder.WriteString(" WHERE %ID = LAST_IDENTITY()")
				return
			}
			// c.Build(builder)
		},
	}

	return clauseBuilders
}

func (dialector Dialector) SavePoint(tx *gorm.DB, name string) error {
	tx.Exec("SAVEPOINT " + name)
	return nil
}

func (dialector Dialector) RollbackTo(tx *gorm.DB, name string) error {
	tx.Exec("ROLLBACK TO SAVEPOINT " + name)
	return nil
}
