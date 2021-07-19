package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// PLEASE NOTE that I would absolutely not organize in this way an application
// destined for a production environment. Typically, the application
// structure would be much better organized, with separate packages
// based on functional area. However, for an exercise such as this I
// opted to keep it really simple.

// global variables (also, I don't generally use globals in production applications)
var pm PartitionManager

// to build app: make
// to start app and build sqlite db: ./enrollment build_db=true
// to start app and use existing db: ./enrollment
func main() {
	var (
		buildDB = flag.Bool("build_db", false, "Set to true to build the sqlite databases and populate them with test data")
	)
	flag.Parse()

	log.Println("Define db partitions...")
	dbs, err := createDBPartitions()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Building partition manager...")
	pm = NewPartitionManager(dbs)
	defer pm.CloseConnections()

	log.Printf("Build sqlite databases? %t", *buildDB)
	if *buildDB {
		buildDBAndPopulate()
	}

	showGetCoursesOutput()
	showGetStudentsInCourseOutput()
	showGetCoursesForStudentsOutput()
}

func buildDBAndPopulate() {
	err := createDatabases(pm.DBs)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = createSchema(pm.DBs)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = createTables(pm.DBs)
	if err != nil {
		log.Fatal(err.Error())
	}
	addSampleCourses()
	addSampleStudents()
	addSampleEnrollments()
}

func showGetCoursesOutput() {
	log.Println("*** Output from getCourses(): ***")
	c, err := getCourses("Ken Thompson")
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
	for _, v := range c {
		log.Printf("%+v", v)
	}
}

func showGetStudentsInCourseOutput() {
	log.Println("*** Output from getStudentsInCourse(): ***")
	s, err := getStudentsInCourse("DB101")
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
	for _, v := range s {
		log.Printf("%+v", v)
	}
}

func showGetCoursesForStudentsOutput() {
	log.Println("*** Output from getetCoursesForStudents(): ***")
	s1 := Student{1, "Ken Thompson", "8885551112"}
	s2 := Student{1, "Rob Pike", "Rob Pike"}
	s := []Student{s1, s2}
	res, err := getCoursesForStudents(s)
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
	for k, v := range res {
		log.Printf("key: %+v; val: %+v", k, v)
	}
}

func createDBPartitions() ([]*Database, error) {
	db1, err := NewDatabase("enrollment1.db", "./enrollment1.db", 65, 77)
	if err != nil {
		return nil, err
	}

	db2, err := NewDatabase("enrollment2.db", "./enrollment2.db", 78, 90)
	if err != nil {
		return nil, err
	}

	dbs := make([]*Database, 2)
	dbs[0] = db1
	dbs[1] = db2

	return dbs, nil
}

// AddCourse inserts a new course into the database.
func addCourse(course Course) error {
	sql := `INSERT INTO courses(code, name) VALUES (?, ?)`
	// add course info to each db
	for i := range pm.DBs {
		s, err := pm.DBs[i].db.Prepare(sql)
		if err != nil {
			return err
		}
		defer s.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := s.ExecContext(ctx, course.CourseCode, course.Name)
		if err != nil {
			return err
		}

		cnt, err := res.RowsAffected()
		if err != nil || cnt == 0 {
			return fmt.Errorf("Unable to add course: %s", course.Name)
		}
	}
	return nil
}

// getCourses fetches courses a student is taking.
func getCourses(studentName string) ([]Course, error) {
	partition := pm.GetDatabaseByPartitionString(studentName)
	sql := `SELECT c.code, c.name 
			FROM enrollment AS e
				JOIN courses AS c ON e.course_code = c.code 
				JOIN students AS s ON e.student_id = s.id
			WHERE s.name = ?`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return execGetCoursesSql(ctx, partition, sql, studentName)
}

// execGetCoursesSql helper function that accepts a context to limit query run time, a pointer to the correct
// database partition, a query, and arguments that will be safely merged into the query to avoid sql injection.
func execGetCoursesSql(ctx context.Context, partition *Database, query string, args ...interface{}) ([]Course, error) {
	rows, err := partition.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []Course
	for rows.Next() {
		c := Course{}
		err := rows.Scan(&c.CourseCode, &c.Name)
		if err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return courses, nil
}

// getStudentsInCourse queries the SQL server to list all the students taking a certain course.
func getStudentsInCourse(courseCode string) ([]Student, error) {
	sql := `SELECT s.id, s.name, s.mobile
			FROM enrollment AS e
				JOIN students AS s ON e.student_id = s.id
			WHERE e.course_code = ?`

	return execGetStudentsInCourseSql(sql, courseCode)
}

// execGetStudentsInCourseSql helper function that accepts a query, and arguments that will be safely
// merged into the query to avoid sql injection.
// Also, if we needed ultra-high performance, this could be rewritten to utilize go routines to run the
// queries to each database partition concurrently.
func execGetStudentsInCourseSql(query string, args ...interface{}) ([]Student, error) {
	var students []Student
	for i := range pm.DBs {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rows, err := pm.DBs[i].db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			s := Student{}
			err := rows.Scan(&s.ID, &s.Name, &s.Mobile)
			if err != nil {
				return nil, err
			}
			students = append(students, s)
		}
		err = rows.Err()
		if err != nil {
			return nil, err
		}
	}

	return students, nil
}

// enrollStudent enrolls a student into one or more courses.
func enrollStudent(student Student, courses []Course) error {
	sql := `INSERT INTO enrollment VALUES (?, ?, ?, ?)`
	partition := pm.GetDatabaseByPartitionString(student.Name)

	for i := range courses {

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		now := time.Now().UTC()

		err := execEnrollStudentSql(ctx, partition, sql, student.ID, courses[i].CourseCode, now.Unix(), nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// execEnrollStudentSql helper function that accepts a context to limit query run time, a pointer to the correct
// database partition, a query, and arguments that will be safely merged into the query to avoid sql injection.
func execEnrollStudentSql(ctx context.Context, partition *Database, query string, args ...interface{}) error {
	s, err := partition.db.Prepare(query)
	if err != nil {
		return err
	}
	defer s.Close()

	res, err := s.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	cnt, err := res.RowsAffected()
	if err != nil || cnt == 0 {
		return fmt.Errorf("Unable to enroll student")
	}

	return nil
}

// StudentCourses represents students that may or may not be enrolled.
type StudentCourses struct {
	StudentID     uint64
	StudentName   string
	StudentMobile string
	CourseCode    sql.NullString
	CourseName    sql.NullString
}

// getCoursesForStudents fetches the courses each student is enrolled in.
// In this example, all students that were provided are returned, regardless
// if they are enrolled. Depending on the callers needs, this could easily
// be modified to return only those enrolled in a course.
func getCoursesForStudents(students []Student) (map[Student][]Course, error) {
	var studentPartitionMap map[string][]Student = make(map[string][]Student)

	// iterate students and build a map of students ids for each db partition. That way we only
	// have to run one query for each relevant partition to fetch all courses per student.
	// Will be faster than running query for each student.
	for i := range students {
		db := pm.GetDatabaseByPartitionString(students[i].Name)
		studentPartitionMap[db.Name] = append(studentPartitionMap[db.Name], students[i])
	}

	// create map to contain final results
	var finalResults map[Student][]Course = make(map[Student][]Course)

	// now range over student-partition map to build & execute queries; this map
	// will contain only the db partitions which contain students that are in those partitions.
	for k, v := range studentPartitionMap {
		// get the student id's for all students stored in this particular db partition
		// we'll use this to build our IN clause further below
		ids := make([]string, len(v))
		for i, v := range v {
			ids[i] = strconv.FormatUint(v.ID, 10)
		}

		// make sure we have ID's before attempting to build sql query to avoid panic
		if len(ids) == 0 {
			continue
		}

		// fetch all data using an IN clause, so fewer queries to relevant partitioned dbs.
		// this query ensures we get all students back that we asked for, regardless if
		// they are enrolled. Makes returning final results easier.
		sql := `SELECT s.id, s.name, s.mobile, c.code, c.name 
				FROM students AS s
					LEFT JOIN enrollment AS e ON s.id = e.student_id
					LEFT JOIN courses AS c on e.course_code = c.code
				WHERE s.id IN (` + strings.Join(ids, ", ") + `)`

		// get db connection for partition that these students are in
		partition := pm.GetDatabaseByName(k)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rows, err := partition.db.QueryContext(ctx, sql)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		// iterate query results and map of results
		for rows.Next() {
			sc := StudentCourses{}
			err := rows.Scan(&sc.StudentID, &sc.StudentName, &sc.StudentMobile, &sc.CourseCode, &sc.CourseName)
			if err != nil {
				return nil, err
			}

			s := Student{sc.StudentID, sc.StudentName, sc.StudentMobile}
			c := Course{}
			if sc.CourseCode.Valid && sc.CourseName.Valid {
				c.CourseCode = sc.CourseCode.String
				c.Name = sc.CourseName.String
			}

			finalResults[s] = append(finalResults[s], c)
		}
		err = rows.Err()
		if err != nil {
			return nil, err
		}
	}

	return finalResults, nil
}

// addStudent writes a new student to the appropriate database and
// adds the student identifer to the provided struct.
func addStudent(student Student) error {
	sql := "INSERT INTO students(name, mobile) VALUES (?, ?)"
	partition := pm.GetDatabaseByPartitionString(student.Name)

	s, err := partition.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := s.ExecContext(ctx, student.Name, student.Mobile)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	student.ID = uint64(id)

	return nil
}

// getStudents fetches all students.
func getStudents() ([]Student, error) {
	sql := `SELECT id, name, mobile
			FROM students`

	var students []Student
	for i := range pm.DBs {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rows, err := pm.DBs[i].db.QueryContext(ctx, sql)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			s := Student{}
			err := rows.Scan(&s.ID, &s.Name, &s.Mobile)
			if err != nil {
				return nil, err
			}
			students = append(students, s)
		}
		err = rows.Err()
		if err != nil {
			return nil, err
		}
	}

	return students, nil
}
