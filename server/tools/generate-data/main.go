package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"slices"
	"strings"

	"github.com/jaswdr/faker/v2"
)

func main() {
	candidates, err := GetData()
	if err != nil {
		log.Fatalf("could not generate data: %s", err)
	}

	data, err := json.MarshalIndent(candidates, "", "  ")
	if err != nil {
		log.Fatalf("could not marshal data: %s", err)
	}

	_ = os.Remove("data.json")

	err = os.WriteFile("data.json", data, 0644)
	if err != nil {
		log.Fatalf("could not write data file: %s", err)
	}
}

type Candidate struct {
	ID        int      `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Picture   string   `json:"picture"`
	Skills    []string `json:"skills"`
}

func GetData() ([]Candidate, error) {
	var result []Candidate
	faker := faker.New()
	for i := 0; i < 80; i++ {
		parts := strings.Split(faker.Person().Name(), " ")

		picture, err := os.ReadFile(faker.ProfileImage().Image().Name())
		if err != nil {
			return nil, fmt.Errorf("could not generate image: %w", err)
		}

		result = append(result, Candidate{
			ID:        i + 1,
			FirstName: parts[0],
			LastName:  parts[1],
			Email:     faker.Person().Contact().Email,
			Phone:     faker.Person().Contact().Phone,
			Picture:   fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(picture)),
			Skills:    randomSkills(),
		})
	}

	return result, nil
}

var skillNames = []string{
	"JavaScript",
	"Go",
	"Ruby",
	"Python",
	"PHP",
	"C++",
	"C",
	"TypeScript",
	"Scala",
	"Swift",
	"Android",
	"iOS",
	"Java",
	"Kotlin",
	"React",
	"Vue",
	"Angular",
	"Svelte",
	"MySQL",
	"PostgreSQL",
	"MongoDB",
	"AWS",
	"GCP",
	"Azure",
	"Pandas",
	"TensorFlow",
}

func randomSkills() []string {
	var skills []string
	target := rand.Intn(8)
	for len(skills) < target {
		skill := skillNames[rand.Intn(len(skillNames))]
		if !slices.Contains(skills, skill) {
			skills = append(skills, skill)
		}
	}
	return skills
}
