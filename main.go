package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"math/rand"
	"time"
)

type Student struct {
	name      string
	startYear int
}

type Course struct {
	title string
	hours int
}

type Exam struct {
	sID   int
	cNo   int
	score int
}

func main() {
	connStr := "user=agudkov password=testpg dbname=academy host=localhost port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("Failed to ping database: %v\n", err)
		return
	}

	_, err = db.Exec(`
		DELETE FROM Exams;
		DELETE FROM Students;
		DELETE FROM Courses;
	`)
	if err != nil {
		fmt.Printf("Failed to clear tables: %v\n", err)
		return
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))

	students := generateStudents(10)
	for _, student := range students {
		_, err := db.Exec(
			`INSERT INTO Students (name, start_year) VALUES ($1, $2)`,
			student.name, student.startYear,
		)
		if err != nil {
			fmt.Printf("Failed to insert student %s: %v\n", student.name, err)
			return
		}
	}
	fmt.Printf("Inserted %d students\n", len(students))

	courses := generateCourses(5)
	for _, course := range courses {
		_, err := db.Exec(
			`INSERT INTO Courses (title, hours) VALUES ($1, $2)`,
			course.title, course.hours,
		)
		if err != nil {
			fmt.Printf("Failed to insert course %s: %v\n", course.title, err)
			return
		}
	}
	fmt.Printf("Inserted %d courses\n", len(courses))

	exams := generateExams(db, 20)
	for _, exam := range exams {
		_, err := db.Exec(
			`INSERT INTO Exams (s_id, c_no, score) VALUES ($1, $2, $3)`,
			exam.sID, exam.cNo, exam.score,
		)
		if err != nil {
			fmt.Printf("Failed to insert exam (s_id=%d, c_no=%d): %v\n", exam.sID, exam.cNo, err)
			return
		}
	}
	fmt.Printf("Inserted %d exams\n", len(exams))
}

func generateStudents(count int) []Student {
	surnames := []string{"Иванов", "Петров", "Александров", "Кузнецов", "Смирнов"}
	names := []string{"Иван", "Александр", "Михаил", "Елена", "Алексей"}
	middleName := []string{"Иванович", "Сергеевна", "Павлович", "Александрович", "Дмитриевич"}

	students := make([]Student, count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf(
			"%s %s %s",
			surnames[rand.Intn(len(surnames))],
			names[rand.Intn(len(names))],
			middleName[rand.Intn(len(middleName))],
		)
		startYear := 2020 + rand.Intn(6) // 2020–2025
		students[i] = Student{name: name, startYear: startYear}
	}
	return students
}

func generateCourses(count int) []Course {
	first := []string{"Матан", "Программирование", "Физика", "Базы данных", "Линал"}
	second := []string{"1", "2", "3", "4"}

	courses := make([]Course, count)
	usedTitles := make(map[string]bool)
	for i := 0; i < count; i++ {
		var title string
		for {
			title = fmt.Sprintf(
				"%s %s",
				first[rand.Intn(len(first))],
				second[rand.Intn(len(second))],
			)
			if !usedTitles[title] {
				usedTitles[title] = true
				break
			}
		}
		hours := rand.Intn(100) + 1
		courses[i] = Course{title: title, hours: hours}
	}
	return courses
}

func generateExams(db *sql.DB, count int) []Exam {
	var studentIDs []int
	rows, err := db.Query(`SELECT s_id FROM Students`)
	if err != nil {
		fmt.Printf("Failed to query students: %v\n", err)
		return nil
	}
	for rows.Next() {
		var sID int
		if err := rows.Scan(&sID); err != nil {
			fmt.Printf("Failed to get student ID: %v\n", err)
			return nil
		}
		studentIDs = append(studentIDs, sID)
	}
	rows.Close()

	var courseIDs []int
	rows, err = db.Query(`SELECT c_no FROM Courses`)
	if err != nil {
		fmt.Printf("Failed to query courses: %v\n", err)
		return nil
	}
	for rows.Next() {
		var cNo int
		if err := rows.Scan(&cNo); err != nil {
			fmt.Printf("Failed to get course ID: %v\n", err)
			return nil
		}
		courseIDs = append(courseIDs, cNo)
	}
	rows.Close()

	exams := make([]Exam, 0, count)
	usedPairs := make(map[string]bool)
	for len(exams) < count && len(usedPairs) < len(studentIDs)*len(courseIDs) {
		sID := studentIDs[rand.Intn(len(studentIDs))]
		cNo := courseIDs[rand.Intn(len(courseIDs))]
		pairKey := fmt.Sprintf("%d-%d", sID, cNo)

		if !usedPairs[pairKey] {
			usedPairs[pairKey] = true
			score := 2 + rand.Intn(4)
			exams = append(exams, Exam{sID: sID, cNo: cNo, score: score})
		}
	}
	return exams
}
