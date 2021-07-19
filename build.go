package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

func createDatabases(dbs []*Database) error {
	for i := range dbs {
		// remove db if it exists already to ensure we
		// don't duplicate data.
		if err := os.Remove(dbs[i].Name); err != nil {
			log.Printf("Attempted to remove existing database: %s, Error: %s", dbs[i].Name, err.Error())
		}

		file, err := os.Create(dbs[i].Name)
		if err != nil {
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}
	return nil
}

func createSchema(dbs []*Database) error {
	for _, v := range dbs {
		db, err := sql.Open("sqlite3", fmt.Sprintf("./%s", v.Name))
		if err != nil {
			log.Println(err.Error())
			return err
		}
		defer db.Close()

		log.Printf("Creating schema for %s...", v.Name)
	}

	return nil
}

func createTables(dbs []*Database) error {
	courses := `CREATE TABLE IF NOT EXISTS courses (
		code TEXT PRIMARY KEY,
		name TEXT NOT NULL
	) WITHOUT ROWID;`

	students := `CREATE TABLE IF NOT EXISTS students (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		mobile TEXT
	);`

	enrollment := `CREATE TABLE IF NOT EXISTS enrollment (
		student_id INTEGER NOT NULL,
		course_code TEXT NOT NULL,
		date_enrolled INTEGER NOT NULL, -- Epoch time
		final_grade TEXT,
		PRIMARY KEY (student_id, course_code),
		FOREIGN KEY (student_id) REFERENCES students (id),
		FOREIGN KEY (course_code) REFERENCES sources (code)
	) WITHOUT ROWID;`

	enrollmentIdx := `CREATE INDEX enrollment_date_enrolled ON enrollment(date_enrolled);`

	queries := []string{courses, students, enrollment, enrollmentIdx}

	for _, partition := range dbs {
		for _, query := range queries {
			log.Printf("Creating db object: %s", query)
			stmnt, err := partition.db.Prepare(query)
			if err != nil {
				log.Println(err.Error())
				return err
			}
			_, err = stmnt.Exec()
			if err != nil {
				log.Println(err.Error())
				return err
			}
		}
	}

	return nil
}

func addSampleCourses() {
	log.Println("Adding sample courses...")
	c1 := Course{"DB101", "Databases 101"}
	c2 := Course{"ALGO201", "Algorithms 201"}
	c3 := Course{"ML301", "Machine Learning 301"}

	err := addCourse(c1)
	handleError(err)
	err = addCourse(c2)
	handleError(err)
	err = addCourse(c3)
	handleError(err)
}

func addSampleStudents() {
	log.Println("Adding sample students...")
	s1 := Student{
		Name:   "Rob Pike",
		Mobile: "8885551111",
	}
	s2 := Student{
		Name:   "Ken Thompson",
		Mobile: "8885551112",
	}
	s3 := Student{
		Name:   "Robert Griesemer",
		Mobile: "8885551113",
	}
	s4 := Student{
		Name:   "Russ Cox",
		Mobile: "8885551114",
	}
	s5 := Student{
		Name:   "Ian Taylor",
		Mobile: "8885551115",
	}
	s6 := Student{
		Name:   "Guido van Rossum",
		Mobile: "8885551116",
	}
	err := addStudent(s1)
	handleError(err)
	err = addStudent(s2)
	handleError(err)
	err = addStudent(s3)
	handleError(err)
	err = addStudent(s4)
	handleError(err)
	err = addStudent(s5)
	handleError(err)
	err = addStudent(s6)
	handleError(err)
}

func addSampleEnrollments() {
	log.Println("Adding sample enrollments...")
	s, err := getStudents()
	handleError(err)

	if len(s) == 0 {
		log.Println("Error: no students returned!")
	}
	c1 := Course{"DB101", "Databases 101"}
	c2 := Course{"ALGO201", "Algorithms 201"}
	c3 := Course{"ML301", "Machine Learning 301"}
	cs1 := []Course{c1, c2}
	cs2 := []Course{c2, c3}

	for i, v := range s {
		if i%2 == 0 {
			err := enrollStudent(v, cs1)
			if err != nil {
				log.Printf("%v\n", err)
			}
		} else {
			err := enrollStudent(v, cs2)
			if err != nil {
				log.Printf("%v\n", err)
			}
		}
	}
}

func handleError(err error) {
	if err != nil {
		log.Printf("Error: %s\n", err.Error())
	}
}
