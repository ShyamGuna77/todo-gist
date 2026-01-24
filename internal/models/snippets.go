package models

import (
	"database/sql"
	"time"
)



type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Insert (title string, content string, expiers int) (int , error){

	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`


	result, err := m.DB.Exec(stmt, title, content, expiers)
	if err != nil {
	    return 0, err
	}
	// Use the LastInsertId() method on the result to get the ID of our
	// newly inserted record in the snippets table.
	id, err := result.LastInsertId()
	if err != nil {
	    return 0, err
	}
	// The ID returned has the type int64, so we convert it to an int type
	// before returning.
	return int(id), nil
}

func (m *SnippetModel) Get ( id int , ) (Snippet ,error) {
  return  Snippet{} ,nil
}

func (m *SnippetModel) Latest() ([]Snippet , error) {
	return nil,nil
}