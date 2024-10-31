package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"

	_ "github.com/lib/pq"
)

var DB *sql.DB

type PostgresRepo struct {
}

func Connect() error {
	var (
		host     = "localhost"
		port     = 5432
		user     = "postgres"
		password = "Monday@01"
		dbname   = "splitwise"
	)

	// Connect to the default PostgreSQL database (e.g., 'postgres')
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password)
	var err error
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	// Check if the desired database exists
	var exists bool
	err = DB.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", dbname).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking database existence: %w", err)
	}

	// Create the database if it does not exist
	if !exists {
		_, err = DB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
		if err != nil {
			return fmt.Errorf("error creating database: %w", err)
		}
		log.Printf("Database %s created successfully\n", dbname)
	} else {
		log.Printf("Database %s already exists\n", dbname)
	}

	// Now connect to the actual database
	psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	// You can start your CRUD operations from here
	log.Println("Connected to the database successfully!")
	return nil
}

func (m *PostgresRepo) Insert(e interface{}) error {
	tname := typeName(e)
	// Build the INSERT query dynamically based on the fields of the struct
	fields := []string{}
	placeholders := []string{}
	values := []interface{}{}

	v := reflect.ValueOf(e).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i).Name
		value := v.Field(i).Interface()

		fields = append(fields, strings.ToLower(field))
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		values = append(values, value)
	}
	log.Println("tnameeeeeeeeeeeeeeeeeeee", tname)
	// Wrap tname in double quotes to handle reserved keywords
	insertQuery := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		tname,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
	)
	log.Println("$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$", insertQuery)
	// Execute the query
	_, err := DB.Exec(insertQuery, values...)
	if err != nil {
		return fmt.Errorf("error inserting data into %s: %w", tname, err)
	}

	log.Println("Data inserted successfully.")
	return nil
}

func (m *PostgresRepo) Find(e interface{}, columnName string, value interface{}) error {
	tname := typeName(e)
	log.Println("tname", tname)

	// Check if columnName is empty
	if columnName == "" {
		return fmt.Errorf("column name cannot be empty")
	}

	// Ensure the table and column names are correctly quoted
	log.Println("cccccccccccccccccc", columnName)
	query := fmt.Sprintf(`SELECT * FROM %s WHERE %s = $1 LIMIT 1`, tname, columnName)
	log.Println("queryyyyyyyyyyyy", query)

	// Execute the query to get a single row
	log.Println("vvvvvvvvvvvvvvvvvvvvvvv", value)
	row := DB.QueryRow(query, value)

	// Ensure `e` is a pointer to a struct so we can set its fields
	v := reflect.ValueOf(e)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct, got %T", e)
	}

	v = v.Elem() // Dereference to get the actual struct

	// Prepare a slice to hold pointers to each field in the struct
	columns := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		columns[i] = v.Field(i).Addr().Interface() // Set up each field to be populated by row.Scan
	}

	// Scan the row into the struct fields
	if err := row.Scan(columns...); err != nil {
		return fmt.Errorf("error finding data in %s: %w", tname, err)
	}

	log.Println("Data found successfully and set in struct.")
	return nil
}

func (m *PostgresRepo) FindAll(e interface{}, tname string) ([]interface{}, error) {

	query := fmt.Sprintf("SELECT * FROM %s", tname)

	rows, err := DB.Query(query)
	log.Println("FindAllFindAllFindAllFindAllFindAllFindAllFindAllFindAllFindAllFindAllFindAllFindAllFindAllFindAll", rows)
	if err != nil {
		return nil, fmt.Errorf("error finding data in %s: %w", tname, err)
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		newElem := reflect.New(reflect.TypeOf(e).Elem()).Interface()
		v := reflect.ValueOf(newElem).Elem()

		columns := make([]interface{}, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			columns[i] = v.Field(i).Addr().Interface()
		}

		if err := rows.Scan(columns...); err != nil {
			return nil, fmt.Errorf("error scanning row in %s: %w", tname, err)
		}

		results = append(results, newElem)
	}

	log.Println("Data found successfully.")
	return results, nil
}

func (m *PostgresRepo) Delete(e interface{}, columnName string, value interface{}) error {
	tname := typeName(e)

	deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", tname, columnName)
	_, err := DB.Exec(deleteQuery, value)
	if err != nil {
		return fmt.Errorf("error deleting data from %s: %w", tname, err)
	}

	log.Println("Data deleted successfully.")
	return nil
}

func typeName(i interface{}) string {
	t := reflect.TypeOf(i)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	if isSlice(t) {
		t = t.Elem()
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}
	return t.Name()
}
func isSlice(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Slice
}
