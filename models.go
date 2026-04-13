package main

type User struct {
	Id             int
	FullName       string
	hashedPassword string
	Grade          string
	Points         int
	Streak         int
	Questions_done []int
}

type Question struct {
	Id             int
	Question       string
	Subject        string
	For_grade      string
	A              string
	B              string
	C              string
	D              string
	Correct_answer string
	Points_worth   int
}

type QuizPageData struct {
	User      User
	Questions []Question
}

type QuestionResult struct {
	Question      string
	GivenAnswer   string
	CorrectAnswer string
	PointsEarned  int
	IsCorrect     bool
}

type LeaderBoardUsers struct {
	Rank      int
	Full_Name string
	Points    int
	Grade     string
}
