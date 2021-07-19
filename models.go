package main

import "time"

// Course represents a course available for enrollment.
type Course struct {
	CourseCode string
	Name       string
}

// Enrollment represents a course that a student is enrolled in.
type Enrollment struct {
	StudentID    uint64
	CourseCode   string
	DateEnrolled time.Time
	FinalGrade   string
}

// Student represents a university
type Student struct {
	ID     uint64
	Name   string
	Mobile string
}
