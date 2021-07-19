-- assume sqlite db for this example

CREATE TABLE IF NOT EXISTS courses (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS students (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    mobile TEXT
);

CREATE TABLE IF NOT EXISTS enrollment (
    student_id INTEGER NOT NULL,
    course_code TEXT NOT NULL,
    date_enrolled INTEGER NOT NULL, -- Epoch time (seconds)
    final_grade TEXT,
    PRIMARY KEY (student_id, course_code),
    FOREIGN KEY (student_id) REFERENCES students (id),
    FOREIGN KEY (course_code) REFERENCES sources (code)
) WITHOUT ROWID;

-- Index the date_enrolled table to help with filters on date/time
CREATE INDEX enrollment_date_enrolled ON enrollment(date_enrolled);


