package main

import (
	"fmt"
	"os"
	"text/template"

	"github.com/google/uuid"
	"gitlab.com/rainbird-ai/sdk-go"
)

const (
	kmID = "bfaab567-6494-4dd3-bbdb-0915a32da0f7"
)

var (
	questionsTemplate = template.Must(template.New("questions").
				Parse(`{{.Prompt}}{{range $i, $c := .Concepts}}
{{"\t"}}{{$i}}{{"\t"}}{{$c.Value}}{{end}}{{if .CanAdd}}
{{"\t"}}{{len .Concepts}}{{"\t"}}Other{{end}}
`))
	answersTemplate = template.Must(template.New("answers").
			Parse(`{{range .}}{{.Subject}} {{.Relationship}} {{.Object}} ({{.Certainty}} certainty){{end}}`))
)

// readInteger reads an integer from standard input
// returns an error if the value is not parsed as an integer or if the integer is outside of the bounds
func readInteger(l, u int) (int, error) {
	var i int
	_, err := fmt.Scanf("%d", &i)
	if err != nil {
		return i, err
	}
	if i < l || i > u {
		err = fmt.Errorf("integer not between %v and %v", l, u)
	}
	return i, err
}

// askForObject asks the user a question to find out the object value
// restricts the user to a single answer
func askForObject(question *sdk.Question) (sdk.QAnswer, error) {
	answer := sdk.QAnswer{
		Subject:      question.Subject,
		Relationship: question.Relationship,
	}

	// display Rainbird's question
	err := questionsTemplate.Execute(os.Stdout, question)
	if err != nil {
		return answer, err
	}

	// get option from user
	n := len(question.Concepts) - 1
	if question.CanAdd {
		n++
	}
	var i int
	for {
		i, err = readInteger(0, n)
		if err == nil {
			break
		}
		fmt.Println(err)
	}
	if question.CanAdd && i == n {
		// ask for user supplied answer
		fmt.Print("Please enter: ")
		fmt.Scanln(&answer.Object)
	} else {
		var ok bool
		answer.Object, ok = question.Concepts[i].Value.(string)
		if !ok {
			return answer, fmt.Errorf("answer suggestion is not a string")
		}
	}
	answer.CF = "100"
	return answer, nil
}

// makeDecision completes a single decision tree
func makeDecision(client sdk.Client, subject, relationship, object string) error {
	// create session
	session, err := client.NewSession(kmID, uuid.New().String())
	if err != nil {
		panic(err)
	}

	// make initial query
	question, answers, err := session.Query(subject, relationship, object)
	if err != nil {
		return err
	}

	for answers == nil || len(*answers) == 0 {
		// ask user a question
		questionAnswer, err := askForObject(question)
		if err != nil {
			return err
		}

		// send user response to Rainbird
		question, answers, err = session.Response([]sdk.QAnswer{questionAnswer})
		if err != nil {
			return err
		}
	}

	// display Rainbird's answer
	return answersTemplate.Execute(os.Stdout, *answers)
}

func main() {
	key, ok := os.LookupEnv("RB_API_KEY")
	if !ok {
		panic("The environment variable RB_API_KEY needs to be set to a valid Rainbird API key")
	}
	client := sdk.Client{
		APIKey:         key,
		EnvironmentURL: sdk.EnvCommunity,
	}

	err := makeDecision(client, "Fred", "speaks", "")
	if err != nil {
		panic(err)
	}
}
